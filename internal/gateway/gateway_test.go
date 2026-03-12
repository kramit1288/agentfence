package gateway

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/agentfence/agentfence/internal/audit"
	"github.com/agentfence/agentfence/internal/config"
	"github.com/agentfence/agentfence/internal/mcp/protocol"
	"github.com/agentfence/agentfence/internal/policy"
)

func TestGatewayAllowedRequest(t *testing.T) {
	engine := mustCompilePolicy(t, `
version: v1
rules:
  - name: allow-deploy
    action: allow
    match:
      server: deployer
      tool: deploy
`)
	forwarder := &stubForwarder{response: protocol.Response{
		JSONRPC: protocol.JSONRPCVersion,
		ID:      protocol.StringID("req-1"),
		Result:  mustMarshal(map[string]any{"ok": true}),
	}}
	auditSink := &recordingAuditSink{}
	gateway := New(config.Default(), testLogger(), WithPolicyEvaluator(engine), WithForwarder(forwarder), WithAuditSink(auditSink))

	response := performMCPRequest(t, gateway.Handler(), "deployer", `{"jsonrpc":"2.0","id":"req-1","method":"tools/call","params":{"name":"deploy","arguments":{"environment":"staging"}}}`)
	if response.StatusCode != http.StatusOK {
		t.Fatalf("StatusCode = %d, want %d", response.StatusCode, http.StatusOK)
	}
	if !forwarder.called {
		t.Fatal("forwarder.called = false, want true")
	}
	if len(auditSink.events) != 1 || auditSink.events[0].Decision != policy.DecisionAllow {
		t.Fatalf("audit events = %+v, want one allow event", auditSink.events)
	}

	var payload protocol.Response
	decodeBody(t, response, &payload)
	if payload.Error != nil {
		t.Fatalf("Error = %+v, want nil", payload.Error)
	}
	var result map[string]any
	if err := json.Unmarshal(payload.Result, &result); err != nil {
		t.Fatalf("json.Unmarshal(result) error = %v", err)
	}
	if result["ok"] != true {
		t.Fatalf("result = %+v, want ok=true", result)
	}
}

func TestGatewayDeniedRequest(t *testing.T) {
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
	forwarder := &stubForwarder{}
	gateway := New(config.Default(), testLogger(), WithPolicyEvaluator(engine), WithForwarder(forwarder))

	response := performMCPRequest(t, gateway.Handler(), "deployer", `{"jsonrpc":"2.0","id":"req-1","method":"tools/call","params":{"name":"deploy","arguments":{"environment":"prod"}}}`)
	if response.StatusCode != http.StatusForbidden {
		t.Fatalf("StatusCode = %d, want %d", response.StatusCode, http.StatusForbidden)
	}
	if forwarder.called {
		t.Fatal("forwarder.called = true, want false")
	}

	var payload protocol.Response
	decodeBody(t, response, &payload)
	if payload.Error == nil || payload.Error.Code != jsonRPCPolicyDenied {
		t.Fatalf("Error = %+v, want policy denied error", payload.Error)
	}
}

func TestGatewayApprovalRequiredRequest(t *testing.T) {
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
	gateway := New(config.Default(), testLogger(), WithPolicyEvaluator(engine))

	response := performMCPRequest(t, gateway.Handler(), "deployer", `{"jsonrpc":"2.0","id":"req-1","method":"tools/call","params":{"name":"deploy","arguments":{"environment":"prod"}}}`)
	if response.StatusCode != http.StatusForbidden {
		t.Fatalf("StatusCode = %d, want %d", response.StatusCode, http.StatusForbidden)
	}

	var payload protocol.Response
	decodeBody(t, response, &payload)
	if payload.Error == nil || payload.Error.Code != jsonRPCApprovalRequired {
		t.Fatalf("Error = %+v, want approval required error", payload.Error)
	}
	var data map[string]any
	if err := json.Unmarshal(payload.Error.Data, &data); err != nil {
		t.Fatalf("json.Unmarshal(data) error = %v", err)
	}
	if data["status"] != "pending_approval" {
		t.Fatalf("data = %+v, want pending_approval status", data)
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

func performMCPRequest(t *testing.T, handler http.Handler, server string, body string) *http.Response {
	t.Helper()

	request := httptest.NewRequest(http.MethodPost, "/mcp", bytes.NewBufferString(body))
	request.Header.Set(headerServerID, server)
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

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

type stubForwarder struct {
	called   bool
	response protocol.Response
	err      error
}

func (s *stubForwarder) Forward(_ context.Context, _ string, _ protocol.Request) (protocol.Response, error) {
	s.called = true
	return s.response, s.err
}

type recordingAuditSink struct {
	events []audit.Event
}

func (s *recordingAuditSink) Record(_ context.Context, event audit.Event) error {
	s.events = append(s.events, event)
	return nil
}
