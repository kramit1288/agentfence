package audit

import (
	"reflect"
	"testing"
	"time"

	"github.com/agentfence/agentfence/internal/mcp/protocol"
	"github.com/agentfence/agentfence/internal/policy"
)

func TestRedactMapMasksSensitiveKeysCaseInsensitive(t *testing.T) {
	input := map[string]any{
		"token":         "abc",
		"Authorization": "Bearer abc",
		"API_KEY":       "key-123",
		"safe":          "value",
	}

	redacted := RedactMap(input)
	if redacted["token"] != RedactedValue {
		t.Fatalf("token = %v, want %q", redacted["token"], RedactedValue)
	}
	if redacted["Authorization"] != RedactedValue {
		t.Fatalf("Authorization = %v, want %q", redacted["Authorization"], RedactedValue)
	}
	if redacted["API_KEY"] != RedactedValue {
		t.Fatalf("API_KEY = %v, want %q", redacted["API_KEY"], RedactedValue)
	}
	if redacted["safe"] != "value" {
		t.Fatalf("safe = %v, want value", redacted["safe"])
	}
	if input["token"] != "abc" {
		t.Fatal("RedactMap mutated input")
	}
}

func TestRedactValueHandlesNestedMapsAndSlices(t *testing.T) {
	input := map[string]any{
		"config": map[string]any{
			"password": "secret-pass",
			"nested": []any{
				map[string]any{"client_secret": "top-secret", "name": "svc"},
				"ok",
			},
		},
	}

	redacted := RedactMap(input)
	config, ok := redacted["config"].(map[string]any)
	if !ok {
		t.Fatalf("config type = %T, want map[string]any", redacted["config"])
	}
	if config["password"] != RedactedValue {
		t.Fatalf("password = %v, want %q", config["password"], RedactedValue)
	}
	items, ok := config["nested"].([]any)
	if !ok {
		t.Fatalf("nested type = %T, want []any", config["nested"])
	}
	first, ok := items[0].(map[string]any)
	if !ok {
		t.Fatalf("nested[0] type = %T, want map[string]any", items[0])
	}
	if first["client_secret"] != RedactedValue {
		t.Fatalf("client_secret = %v, want %q", first["client_secret"], RedactedValue)
	}
	if first["name"] != "svc" {
		t.Fatalf("name = %v, want svc", first["name"])
	}
}

func TestRedactMapMatchesNormalizedSensitiveKeys(t *testing.T) {
	input := map[string]any{
		"x-authORIZATION-token": "abc",
		"dbSecretValue":        "hidden",
	}

	redacted := RedactMap(input)
	if redacted["x-authORIZATION-token"] != RedactedValue {
		t.Fatalf("x-authORIZATION-token = %v, want %q", redacted["x-authORIZATION-token"], RedactedValue)
	}
	if redacted["dbSecretValue"] != RedactedValue {
		t.Fatalf("dbSecretValue = %v, want %q", redacted["dbSecretValue"], RedactedValue)
	}
}

func TestBuilderBuildPolicyDecisionRedactsArguments(t *testing.T) {
	fixed := time.Date(2026, time.March, 12, 12, 0, 0, 0, time.UTC)
	builder := Builder{now: func() time.Time { return fixed }}
	request := protocol.Request{
		JSONRPC: protocol.JSONRPCVersion,
		ID:      idPtr(protocol.IntID(7)),
		Method:  protocol.MethodToolsCall,
	}
	result := policy.Result{Action: policy.DecisionRequireApproval, Reason: "needs approval", RuleName: "approval-prod"}
	args := map[string]any{"password": "secret", "env": "prod"}

	event := builder.BuildPolicyDecision(request, "deployer", "deploy", args, result)
	if event.Timestamp != fixed {
		t.Fatalf("Timestamp = %v, want %v", event.Timestamp, fixed)
	}
	if event.Kind != EventKindPolicyDecision {
		t.Fatalf("Kind = %q, want %q", event.Kind, EventKindPolicyDecision)
	}
	if event.Request.ID != "7" {
		t.Fatalf("Request.ID = %q, want 7", event.Request.ID)
	}
	wantArgs := map[string]any{"password": RedactedValue, "env": "prod"}
	if !reflect.DeepEqual(event.Request.Arguments, wantArgs) {
		t.Fatalf("Arguments = %#v, want %#v", event.Request.Arguments, wantArgs)
	}
	if event.Decision.Allowed {
		t.Fatal("Decision.Allowed = true, want false")
	}
}

func idPtr(id protocol.ID) *protocol.ID {
	return &id
}
