package gateway

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/agentfence/agentfence/internal/api"
	"github.com/agentfence/agentfence/internal/audit"
	"github.com/agentfence/agentfence/internal/approval"
	"github.com/agentfence/agentfence/internal/config"
	"github.com/agentfence/agentfence/internal/mcp/protocol"
	"github.com/agentfence/agentfence/internal/policy"
)

const (
	headerServerID = "X-AgentFence-Server"
	headerActor    = "X-AgentFence-Actor"
)

const (
	jsonRPCParseError         int64 = -32700
	jsonRPCInvalidRequest     int64 = -32600
	jsonRPCPolicyDenied       int64 = -32001
	jsonRPCApprovalRequired   int64 = -32002
	jsonRPCForwardingStub     int64 = -32003
	jsonRPCAuditFailure       int64 = -32004
	jsonRPCApprovalUnavailable int64 = -32005
)

// PolicyEvaluator captures the portion of the policy engine used by the gateway.
type PolicyEvaluator interface {
	Evaluate(input policy.Input) policy.Result
}

// Forwarder sends approved requests upstream.
type Forwarder interface {
	Forward(ctx context.Context, server string, request protocol.Request) (protocol.Response, error)
}

// ApprovalManager captures the approval workflow behavior used by the gateway.
type ApprovalManager interface {
	Create(ctx context.Context, input approval.CreateInput) (approval.Request, error)
}

// Gateway is the top-level runtime for the HTTP gateway process.
type Gateway struct {
	cfg       config.Config
	logger    *slog.Logger
	policy    PolicyEvaluator
	approvals ApprovalManager
	forwarder Forwarder
	auditSink audit.Sink
	builder   audit.Builder
	server    *http.Server
}

// Option customizes the gateway runtime.
type Option func(*Gateway)

func WithPolicyEvaluator(evaluator PolicyEvaluator) Option {
	return func(g *Gateway) { g.policy = evaluator }
}

func WithApprovalManager(manager ApprovalManager) Option {
	return func(g *Gateway) { g.approvals = manager }
}

func WithForwarder(forwarder Forwarder) Option {
	return func(g *Gateway) { g.forwarder = forwarder }
}

func WithAuditSink(sink audit.Sink) Option {
	return func(g *Gateway) { g.auditSink = sink }
}

func New(cfg config.Config, logger *slog.Logger, opts ...Option) *Gateway {
	gateway := &Gateway{
		cfg:     cfg,
		logger:  logger,
		builder: audit.NewBuilder(),
	}
	for _, opt := range opts {
		opt(gateway)
	}

	handler := api.NewHandler(logger, gateway.Handler())
	gateway.server = &http.Server{
		Addr:              cfg.HTTP.Address,
		Handler:           handler,
		ReadHeaderTimeout: cfg.HTTP.ReadHeaderTimeout,
		ReadTimeout:       cfg.HTTP.ReadTimeout,
		WriteTimeout:      cfg.HTTP.WriteTimeout,
		IdleTimeout:       cfg.HTTP.IdleTimeout,
	}

	return gateway
}

func (g *Gateway) Handler() http.Handler {
	return http.HandlerFunc(g.handleMCP)
}

func (g *Gateway) ListenAddr() string {
	return g.cfg.HTTP.Address
}

func (g *Gateway) Run(ctx context.Context) error {
	errCh := make(chan error, 1)

	g.logger.Info("starting gateway",
		"environment", g.cfg.Environment,
		"address", g.server.Addr,
	)

	go func() {
		err := g.server.ListenAndServe()
		if errors.Is(err, http.ErrServerClosed) {
			errCh <- nil
			return
		}
		errCh <- err
	}()

	select {
	case err := <-errCh:
		if err != nil {
			return fmt.Errorf("serve gateway: %w", err)
		}
		return nil
	case <-ctx.Done():
		g.logger.Info("gateway shutdown requested")
		if err := g.Shutdown(context.Background()); err != nil {
			return err
		}
		return <-errCh
	}
}

func (g *Gateway) Shutdown(ctx context.Context) error {
	shutdownCtx, cancel := context.WithTimeout(ctx, g.cfg.HTTP.ShutdownTimeout)
	defer cancel()

	if err := g.server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown gateway: %w", err)
	}

	g.logger.Info("gateway shutdown complete")
	return nil
}

func (g *Gateway) handleMCP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		g.writeJSONRPCError(w, http.StatusBadRequest, nil, jsonRPCParseError, "failed to read request body", nil)
		return
	}

	request, err := protocol.DecodeRequest(body)
	if err != nil {
		code := jsonRPCInvalidRequest
		if !json.Valid(body) {
			code = jsonRPCParseError
		}
		g.writeJSONRPCError(w, http.StatusBadRequest, nil, code, err.Error(), nil)
		return
	}
	if request.IsNotification() {
		g.writeJSONRPCError(w, http.StatusBadRequest, nil, jsonRPCInvalidRequest, "notifications are not supported", nil)
		return
	}

	metadata, err := extractMetadata(r, request)
	if err != nil {
		g.writeJSONRPCError(w, http.StatusBadRequest, request.ID, jsonRPCInvalidRequest, err.Error(), nil)
		return
	}

	decision := g.evaluatePolicy(metadata)
	event := g.builder.BuildPolicyDecision(request, metadata.Server, metadata.Tool, metadata.Args, decision)
	if err := g.recordAudit(r.Context(), event); err != nil {
		g.writeJSONRPCError(w, http.StatusInternalServerError, request.ID, jsonRPCAuditFailure, "failed to record audit event", map[string]any{"reason": err.Error()})
		return
	}

	switch decision.Action {
	case policy.DecisionAllow:
		g.handleAllowed(w, r, request, metadata, decision)
	case policy.DecisionRequireApproval:
		g.handleApprovalRequired(w, r, request, metadata, decision)
	default:
		g.writeJSONRPCError(w, http.StatusForbidden, request.ID, jsonRPCPolicyDenied, "request denied by policy", map[string]any{
			"reason": decision.Reason,
			"rule":   decision.RuleName,
			"server": metadata.Server,
			"tool":   metadata.Tool,
		})
	}
}

type requestMetadata struct {
	Server string
	Tool   string
	Args   map[string]any
	Actor  string
}

func extractMetadata(r *http.Request, request protocol.Request) (requestMetadata, error) {
	server := r.Header.Get(headerServerID)
	if server == "" {
		server = r.URL.Query().Get("server")
	}
	if server == "" {
		return requestMetadata{}, errors.New("server identifier is required via X-AgentFence-Server header or server query parameter")
	}

	actor := r.Header.Get(headerActor)

	switch request.Method {
	case protocol.MethodToolsCall:
		params, err := protocol.DecodeToolsCallParams(request.Params)
		if err != nil {
			return requestMetadata{}, err
		}
		args, err := decodeArguments(params.Arguments)
		if err != nil {
			return requestMetadata{}, err
		}
		return requestMetadata{Server: server, Tool: params.Name, Args: args, Actor: actor}, nil
	case protocol.MethodToolsList:
		return requestMetadata{Server: server, Tool: protocol.MethodToolsList, Args: map[string]any{}, Actor: actor}, nil
	default:
		return requestMetadata{}, fmt.Errorf("unsupported MCP method %q", request.Method)
	}
}

func decodeArguments(raw json.RawMessage) (map[string]any, error) {
	if len(raw) == 0 {
		return map[string]any{}, nil
	}
	var args map[string]any
	if err := json.Unmarshal(raw, &args); err != nil {
		return nil, fmt.Errorf("decode tools/call arguments: %w", err)
	}
	if args == nil {
		return map[string]any{}, nil
	}
	return args, nil
}

func (g *Gateway) evaluatePolicy(metadata requestMetadata) policy.Result {
	if g.policy == nil {
		return policy.Result{Action: policy.DecisionDeny, Reason: "policy engine not configured; deny by default"}
	}
	return g.policy.Evaluate(policy.Input{Server: metadata.Server, Tool: metadata.Tool, Args: metadata.Args})
}

func (g *Gateway) recordAudit(ctx context.Context, event audit.Event) error {
	if g.auditSink == nil {
		return nil
	}
	return g.auditSink.Record(ctx, event)
}

func (g *Gateway) handleApprovalRequired(w http.ResponseWriter, r *http.Request, request protocol.Request, metadata requestMetadata, decision policy.Result) {
	if g.approvals == nil {
		g.writeJSONRPCError(w, http.StatusInternalServerError, request.ID, jsonRPCApprovalUnavailable, "approval workflow unavailable", map[string]any{
			"reason": "approval service not configured",
		})
		return
	}

	approvalRequest, err := g.approvals.Create(r.Context(), approval.CreateInput{
		Server:    metadata.Server,
		Tool:      metadata.Tool,
		Method:    request.Method,
		Reason:    decision.Reason,
		RuleName:  decision.RuleName,
		RequestID: eventRequestID(request.ID),
		Arguments: audit.RedactMap(metadata.Args),
		Actor:     metadata.Actor,
	})
	if err != nil {
		g.writeJSONRPCError(w, http.StatusInternalServerError, request.ID, jsonRPCApprovalUnavailable, "failed to create approval request", map[string]any{"reason": err.Error()})
		return
	}

	g.writeJSONRPCError(w, http.StatusForbidden, request.ID, jsonRPCApprovalRequired, "request requires approval", map[string]any{
		"status":      string(approvalRequest.Status),
		"approval_id": approvalRequest.ID,
		"reason":      decision.Reason,
		"rule":        decision.RuleName,
		"server":      metadata.Server,
		"tool":        metadata.Tool,
		"forwarded":   false,
	})
}

func (g *Gateway) handleAllowed(w http.ResponseWriter, r *http.Request, request protocol.Request, metadata requestMetadata, decision policy.Result) {
	if g.forwarder == nil {
		g.writeJSONRPCResult(w, http.StatusOK, protocol.Response{
			JSONRPC: protocol.JSONRPCVersion,
			ID:      *request.ID,
			Result: mustMarshal(map[string]any{
				"status":    "allowed",
				"forwarded": false,
				"reason":    decision.Reason,
				"server":    metadata.Server,
				"tool":      metadata.Tool,
			}),
		})
		return
	}

	response, err := g.forwarder.Forward(r.Context(), metadata.Server, request)
	if err != nil {
		g.writeJSONRPCError(w, http.StatusBadGateway, request.ID, jsonRPCForwardingStub, "forwarding failed", map[string]any{"reason": err.Error()})
		return
	}
	g.writeJSONRPCResult(w, http.StatusOK, response)
}

func (g *Gateway) writeJSONRPCResult(w http.ResponseWriter, status int, response protocol.Response) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(response)
}

func (g *Gateway) writeJSONRPCError(w http.ResponseWriter, status int, id *protocol.ID, code int64, message string, data map[string]any) {
	responseID := protocol.ID{}
	if id != nil {
		responseID = *id
	}
	response := protocol.Response{JSONRPC: protocol.JSONRPCVersion, ID: responseID, Error: &protocol.Error{Code: code, Message: message}}
	if data != nil {
		response.Error.Data = mustMarshal(data)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(response)
}

func mustMarshal(value any) json.RawMessage {
	encoded, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	return encoded
}

func eventRequestID(id *protocol.ID) string {
	if id == nil {
		return ""
	}
	if value, ok := id.StringValue(); ok {
		return value
	}
	if value, ok := id.IntValue(); ok {
		return fmt.Sprintf("%d", value)
	}
	return ""
}
