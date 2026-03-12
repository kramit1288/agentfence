package gateway

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/agentfence/agentfence/internal/audit"
	"github.com/agentfence/agentfence/internal/approval"
	"github.com/agentfence/agentfence/internal/config"
	"github.com/agentfence/agentfence/internal/mcp/protocol"
	"github.com/agentfence/agentfence/internal/mcp/transport"
	"github.com/agentfence/agentfence/internal/policy"
)

func TestGatewayAllowedRequestForwardsUpstream(t *testing.T) {
	engine := mustCompilePolicy(t, `
version: v1
rules:
  - name: allow-deploy
    action: allow
    match:
      server: deployer
      tool: deploy
`)
	var hits atomic.Int32
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		var request protocol.Request
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatalf("json.Decode() error = %v", err)
		}
		_ = json.NewEncoder(w).Encode(protocol.Response{JSONRPC: protocol.JSONRPCVersion, ID: *request.ID, Result: mustMarshal(map[string]any{"ok": true})})
	}))
	defer upstream.Close()
	forwarder := mustHTTPForwarder(t, upstream.URL, time.Second)
	auditSink := &recordingAuditSink{}
	gateway := New(config.Default(), testLogger(), WithPolicyEvaluator(engine), WithForwarder(forwarder), WithAuditSink(auditSink))

	response := performMCPRequest(t, gateway.Handler(), "deployer", "", `{"jsonrpc":"2.0","id":"req-1","method":"tools/call","params":{"name":"deploy","arguments":{"environment":"staging","api_key":"top-secret"}}}`)
	if response.StatusCode != http.StatusOK {
		t.Fatalf("StatusCode = %d, want %d", response.StatusCode, http.StatusOK)
	}
	if hits.Load() != 1 {
		t.Fatalf("upstream hits = %d, want 1", hits.Load())
	}
	if len(auditSink.events) != 2 || auditSink.events[1].Kind != audit.EventKindUpstreamCall {
		t.Fatalf("audit events = %+v, want decision and upstream events", auditSink.events)
	}
	if auditSink.events[1].Upstream.Outcome != string(transport.OutcomeSuccess) {
		t.Fatalf("upstream event = %+v, want success outcome", auditSink.events[1])
	}
}

func TestGatewayUpstreamFailureReturnsBadGateway(t *testing.T) {
	engine := mustCompilePolicy(t, `
version: v1
rules:
  - name: allow-deploy
    action: allow
    match:
      server: deployer
      tool: deploy
`)
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte(`not-jsonrpc`))
	}))
	defer upstream.Close()
	forwarder := mustHTTPForwarder(t, upstream.URL, time.Second)
	auditSink := &recordingAuditSink{}
	gateway := New(config.Default(), testLogger(), WithPolicyEvaluator(engine), WithForwarder(forwarder), WithAuditSink(auditSink))

	response := performMCPRequest(t, gateway.Handler(), "deployer", "", `{"jsonrpc":"2.0","id":"req-2","method":"tools/call","params":{"name":"deploy","arguments":{"environment":"staging"}}}`)
	if response.StatusCode != http.StatusBadGateway {
		t.Fatalf("StatusCode = %d, want %d", response.StatusCode, http.StatusBadGateway)
	}
	var payload protocol.Response
	decodeBody(t, response, &payload)
	if payload.Error == nil || payload.Error.Code != jsonRPCForwardingFailure {
		t.Fatalf("Error = %+v, want forwarding failure", payload.Error)
	}
	if len(auditSink.events) != 2 || auditSink.events[1].Upstream.Outcome != string(transport.OutcomeHTTPError) {
		t.Fatalf("audit events = %+v, want upstream http_error", auditSink.events)
	}
}

func TestGatewayUpstreamNon2xxJSONRPCBodyReturnsBadGateway(t *testing.T) {
	engine := mustCompilePolicy(t, `
version: v1
rules:
  - name: allow-deploy
    action: allow
    match:
      server: deployer
      tool: deploy
`)
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		_ = json.NewEncoder(w).Encode(protocol.Response{JSONRPC: protocol.JSONRPCVersion, ID: protocol.StringID("req-2b"), Error: &protocol.Error{Code: -32099, Message: "upstream failed"}})
	}))
	defer upstream.Close()
	forwarder := mustHTTPForwarder(t, upstream.URL, time.Second)
	auditSink := &recordingAuditSink{}
	gateway := New(config.Default(), testLogger(), WithPolicyEvaluator(engine), WithForwarder(forwarder), WithAuditSink(auditSink))

	response := performMCPRequest(t, gateway.Handler(), "deployer", "", `{"jsonrpc":"2.0","id":"req-2b","method":"tools/call","params":{"name":"deploy","arguments":{"environment":"staging"}}}`)
	if response.StatusCode != http.StatusBadGateway {
		t.Fatalf("StatusCode = %d, want %d", response.StatusCode, http.StatusBadGateway)
	}
	var payload protocol.Response
	decodeBody(t, response, &payload)
	if payload.Error == nil || payload.Error.Code != jsonRPCForwardingFailure {
		t.Fatalf("Error = %+v, want forwarding failure", payload.Error)
	}
	if len(auditSink.events) != 2 || auditSink.events[1].Upstream.Outcome != string(transport.OutcomeHTTPError) {
		t.Fatalf("audit events = %+v, want upstream http_error", auditSink.events)
	}
}

func TestGatewayUpstreamTimeoutReturnsBadGateway(t *testing.T) {
	engine := mustCompilePolicy(t, `
version: v1
rules:
  - name: allow-deploy
    action: allow
    match:
      server: deployer
      tool: deploy
`)
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		_ = json.NewEncoder(w).Encode(protocol.Response{JSONRPC: protocol.JSONRPCVersion, ID: protocol.StringID("req-3"), Result: mustMarshal(map[string]any{"ok": true})})
	}))
	defer upstream.Close()
	forwarder := mustHTTPForwarder(t, upstream.URL, 10*time.Millisecond)
	auditSink := &recordingAuditSink{}
	gateway := New(config.Default(), testLogger(), WithPolicyEvaluator(engine), WithForwarder(forwarder), WithAuditSink(auditSink))

	response := performMCPRequest(t, gateway.Handler(), "deployer", "", `{"jsonrpc":"2.0","id":"req-3","method":"tools/call","params":{"name":"deploy","arguments":{"environment":"staging"}}}`)
	if response.StatusCode != http.StatusBadGateway {
		t.Fatalf("StatusCode = %d, want %d", response.StatusCode, http.StatusBadGateway)
	}
	if len(auditSink.events) != 2 || auditSink.events[1].Upstream.Outcome != string(transport.OutcomeTransportError) {
		t.Fatalf("audit events = %+v, want upstream transport_error", auditSink.events)
	}
}

func TestGatewayDeniedRequestDoesNotForward(t *testing.T) {
	engine := mustCompilePolicy(t, `
version: v1
rules:
  - name: deny-deploy
    action: deny
    reason: deploys are blocked
    match:
      server: deployer
      tool: deploy
`)
	var hits atomic.Int32
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer upstream.Close()
	forwarder := mustHTTPForwarder(t, upstream.URL, time.Second)
	gateway := New(config.Default(), testLogger(), WithPolicyEvaluator(engine), WithForwarder(forwarder))

	response := performMCPRequest(t, gateway.Handler(), "deployer", "", `{"jsonrpc":"2.0","id":"req-4","method":"tools/call","params":{"name":"deploy","arguments":{"environment":"prod"}}}`)
	if response.StatusCode != http.StatusForbidden {
		t.Fatalf("StatusCode = %d, want %d", response.StatusCode, http.StatusForbidden)
	}
	if hits.Load() != 0 {
		t.Fatalf("upstream hits = %d, want 0", hits.Load())
	}
}

func TestGatewayApprovalRequiredRequestCreatesApproval(t *testing.T) {
	engine := mustCompilePolicy(t, `
version: v1
rules:
  - name: approval-prod
    action: require_approval
    reason: production deploy needs approval
    match:
      server: deployer
      tool: deploy
      args:
        environment: prod
`)
	repo := approval.NewMemoryRepository()
	service := approval.NewService(repo)
	gateway := New(config.Default(), testLogger(), WithPolicyEvaluator(engine), WithApprovalManager(service))

	response := performMCPRequest(t, gateway.Handler(), "deployer", "alice", `{"jsonrpc":"2.0","id":"req-1","method":"tools/call","params":{"name":"deploy","arguments":{"environment":"prod","api_key":"shh"}}}`)
	if response.StatusCode != http.StatusForbidden {
		t.Fatalf("StatusCode = %d, want %d", response.StatusCode, http.StatusForbidden)
	}
	var payload protocol.Response
	decodeBody(t, response, &payload)
	var data map[string]any
	if err := json.Unmarshal(payload.Error.Data, &data); err != nil {
		t.Fatalf("json.Unmarshal(data) error = %v", err)
	}
	approvalID, ok := data["approval_id"].(string)
	if !ok || approvalID == "" {
		t.Fatalf("data = %+v, want non-empty approval_id", data)
	}
	stored, err := repo.Get(context.Background(), approvalID)
	if err != nil {
		t.Fatalf("repo.Get() error = %v", err)
	}
	if stored.CreatedBy != "alice" || stored.Arguments["api_key"] != audit.RedactedValue {
		t.Fatalf("stored = %+v, want actor alice and redacted args", stored)
	}
}

func TestGatewayMalformedRequest(t *testing.T) {
	gateway := New(config.Default(), testLogger())
	request := httptest.NewRequest(http.MethodPost, "/mcp", bytes.NewBufferString(`{"jsonrpc":"2.0"`))
	request.Header.Set(headerServerID, "deployer")
	response := httptest.NewRecorder()
	gateway.Handler().ServeHTTP(response, request)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("StatusCode = %d, want %d", response.Code, http.StatusBadRequest)
	}
	var payload protocol.Response
	decodeRecorderBody(t, response, &payload)
	if payload.Error == nil || payload.Error.Code != jsonRPCParseError {
		t.Fatalf("Error = %+v, want parse error", payload.Error)
	}
}

func performMCPRequest(t *testing.T, handler http.Handler, server string, actor string, body string) *http.Response {
	t.Helper()
	request := httptest.NewRequest(http.MethodPost, "/mcp", bytes.NewBufferString(body))
	request.Header.Set(headerServerID, server)
	if actor != "" {
		request.Header.Set(headerActor, actor)
	}
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	return response.Result()
}

func decodeBody(t *testing.T, response *http.Response, target any) {
	t.Helper()
	defer response.Body.Close()
	if err := json.NewDecoder(response.Body).Decode(target); err != nil {
		t.Fatalf("json.Decode() error = %v", err)
	}
}

func decodeRecorderBody(t *testing.T, response *httptest.ResponseRecorder, target any) {
	t.Helper()
	if err := json.Unmarshal(response.Body.Bytes(), target); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
}

func mustCompilePolicy(t *testing.T, raw string) *policy.Engine {
	t.Helper()
	parsed, err := policy.ParseYAML([]byte(raw))
	if err != nil {
		t.Fatalf("ParseYAML() error = %v", err)
	}
	engine, err := policy.Compile(parsed)
	if err != nil {
		t.Fatalf("Compile() error = %v", err)
	}
	return engine
}

func mustHTTPForwarder(t *testing.T, url string, timeout time.Duration) *transport.HTTPForwarder {
	t.Helper()
	forwarder, err := transport.NewHTTPForwarder(transport.Target{Address: url}, &http.Client{Timeout: timeout})
	if err != nil {
		t.Fatalf("NewHTTPForwarder() error = %v", err)
	}
	return forwarder
}

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

type recordingAuditSink struct{ events []audit.Event }

func (s *recordingAuditSink) Record(_ context.Context, event audit.Event) error {
	s.events = append(s.events, event)
	return nil
}

