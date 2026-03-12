package audit

import (
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/agentfence/agentfence/internal/mcp/protocol"
	"github.com/agentfence/agentfence/internal/mcp/transport"
	"github.com/agentfence/agentfence/internal/policy"
)

func TestRedactMapMasksSensitiveKeysCaseInsensitive(t *testing.T) {
	input := map[string]any{"token": "abc", "Authorization": "Bearer abc", "API_KEY": "key-123", "safe": "value"}
	redacted := RedactMap(input)
	if redacted["token"] != RedactedValue || redacted["Authorization"] != RedactedValue || redacted["API_KEY"] != RedactedValue {
		t.Fatalf("redacted = %#v, want sensitive fields masked", redacted)
	}
	if redacted["safe"] != "value" || input["token"] != "abc" {
		t.Fatalf("redacted = %#v input = %#v, want safe preserved and input untouched", redacted, input)
	}
}

func TestRedactValueHandlesNestedMapsAndSlices(t *testing.T) {
	input := map[string]any{"config": map[string]any{"password": "secret-pass", "nested": []any{map[string]any{"client_secret": "top-secret", "name": "svc"}, "ok"}}}
	redacted := RedactMap(input)
	config := redacted["config"].(map[string]any)
	items := config["nested"].([]any)
	first := items[0].(map[string]any)
	if config["password"] != RedactedValue || first["client_secret"] != RedactedValue || first["name"] != "svc" {
		t.Fatalf("redacted = %#v, want nested redaction", redacted)
	}
}

func TestRedactMapMatchesNormalizedSensitiveKeys(t *testing.T) {
	input := map[string]any{"x-authORIZATION-token": "abc", "dbSecretValue": "hidden"}
	redacted := RedactMap(input)
	if redacted["x-authORIZATION-token"] != RedactedValue || redacted["dbSecretValue"] != RedactedValue {
		t.Fatalf("redacted = %#v, want normalized redaction", redacted)
	}
}

func TestRedactTextMasksInlineSecrets(t *testing.T) {
	input := "token=abc authorization=Bearer foo api_key=bar password=hunter2"
	redacted := RedactText(input)
	if strings.Contains(redacted, "abc") || strings.Contains(redacted, "Bearer foo") || strings.Contains(redacted, "bar") || strings.Contains(redacted, "hunter2") {
		t.Fatalf("RedactText() = %q, want masked secrets", redacted)
	}
}

func TestRedactTextMasksURLUserInfo(t *testing.T) {
	input := "call upstream: Post \"http://user:secret@example.com/mcp?token=abc\""
	redacted := RedactText(input)
	if strings.Contains(redacted, "user:secret") || strings.Contains(redacted, "token=abc") {
		t.Fatalf("RedactText() = %q, want URL secrets masked", redacted)
	}
	if !strings.Contains(redacted, RedactedValue) {
		t.Fatalf("RedactText() = %q, want redaction marker", redacted)
	}
}

func TestBuilderBuildPolicyDecisionRedactsArguments(t *testing.T) {
	fixed := time.Date(2026, time.March, 12, 12, 0, 0, 0, time.UTC)
	builder := Builder{now: func() time.Time { return fixed }}
	request := protocol.Request{JSONRPC: protocol.JSONRPCVersion, ID: idPtr(protocol.IntID(7)), Method: protocol.MethodToolsCall}
	result := policy.Result{Action: policy.DecisionRequireApproval, Reason: "needs approval", RuleName: "approval-prod"}
	args := map[string]any{"password": "secret", "env": "prod"}
	event := builder.BuildPolicyDecision(request, "deployer", "deploy", args, result)
	wantArgs := map[string]any{"password": RedactedValue, "env": "prod"}
	if event.Timestamp != fixed || event.Request.ID != "7" || !reflect.DeepEqual(event.Request.Arguments, wantArgs) || event.Decision.Allowed {
		t.Fatalf("event = %#v, want redacted decision event", event)
	}
}

func TestBuilderBuildUpstreamCall(t *testing.T) {
	fixed := time.Date(2026, time.March, 12, 12, 1, 0, 0, time.UTC)
	builder := Builder{now: func() time.Time { return fixed }}
	request := protocol.Request{JSONRPC: protocol.JSONRPCVersion, ID: idPtr(protocol.StringID("req-1")), Method: protocol.MethodToolsCall}
	result := transport.ForwardResult{HTTPStatusCode: 200, Latency: 25 * time.Millisecond, Outcome: transport.OutcomeSuccess, Target: "http://upstream"}
	event := builder.BuildUpstreamCall(request, "deployer", "deploy", map[string]any{"api_key": "secret"}, result)
	if event.Kind != EventKindUpstreamCall || event.Upstream.Target != "http://upstream" || event.Upstream.Outcome != string(transport.OutcomeSuccess) || event.Upstream.Forwarded != true {
		t.Fatalf("event = %#v, want upstream call event", event)
	}
	if event.Request.Arguments["api_key"] != RedactedValue {
		t.Fatalf("arguments = %#v, want redacted api_key", event.Request.Arguments)
	}
}

func TestBuilderBuildUpstreamCallRedactsErrorText(t *testing.T) {
	builder := Builder{now: func() time.Time { return time.Date(2026, time.March, 12, 12, 1, 0, 0, time.UTC) }}
	request := protocol.Request{JSONRPC: protocol.JSONRPCVersion, ID: idPtr(protocol.StringID("req-2")), Method: protocol.MethodToolsCall}
	result := transport.ForwardResult{Outcome: transport.OutcomeTransportError, Target: "http://example.com", Err: errors.New("authorization=Bearer abc token=secret")}
	event := builder.BuildUpstreamCall(request, "deployer", "deploy", nil, result)
	if strings.Contains(event.Upstream.Error, "abc") || strings.Contains(event.Upstream.Error, "secret") {
		t.Fatalf("event.Upstream.Error = %q, want redacted error text", event.Upstream.Error)
	}
	if !strings.Contains(event.Upstream.Error, RedactedValue) {
		t.Fatalf("event.Upstream.Error = %q, want redaction marker", event.Upstream.Error)
	}
}

func idPtr(id protocol.ID) *protocol.ID { return &id }