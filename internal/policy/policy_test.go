package policy

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAllowDecision(t *testing.T) {
	engine := mustCompilePolicy(t, `
version: v1
rules:
  - name: allow-readonly
    action: allow
    match:
      server: github
      tool: repos/get
`)

	result := engine.Evaluate(Input{Server: "github", Tool: "repos/get"})
	if result.Action != DecisionAllow {
		t.Fatalf("Action = %q, want %q", result.Action, DecisionAllow)
	}
	if !result.Matched || result.RuleName != "allow-readonly" {
		t.Fatalf("Result = %+v, want matched allow-readonly", result)
	}
}

func TestDenyDecision(t *testing.T) {
	engine := mustCompilePolicy(t, `
version: v1
rules:
  - name: deny-delete
    action: deny
    reason: destructive delete denied
    match:
      tool: repos/delete
`)

	result := engine.Evaluate(Input{Server: "github", Tool: "repos/delete"})
	if result.Action != DecisionDeny {
		t.Fatalf("Action = %q, want %q", result.Action, DecisionDeny)
	}
	if result.Reason != "destructive delete denied" {
		t.Fatalf("Reason = %q, want custom reason", result.Reason)
	}
}

func TestRequireApprovalDecision(t *testing.T) {
	engine := mustCompilePolicy(t, `
version: v1
rules:
  - name: approval-prod
    action: require_approval
    match:
      tool: deploy
      args:
        environment: prod
`)

	result := engine.Evaluate(Input{Server: "deployer", Tool: "deploy", Args: map[string]any{"environment": "prod"}})
	if result.Action != DecisionRequireApproval {
		t.Fatalf("Action = %q, want %q", result.Action, DecisionRequireApproval)
	}
}

func TestFirstMatchWinsPrecedence(t *testing.T) {
	engine := mustCompilePolicy(t, `
version: v1
rules:
  - name: broad-deny
    action: deny
    match:
      server: github
      tool: repos/*
  - name: later-allow
    action: allow
    match:
      server: github
      tool: repos/get
`)

	result := engine.Evaluate(Input{Server: "github", Tool: "repos/get"})
	if result.Action != DecisionDeny {
		t.Fatalf("Action = %q, want %q", result.Action, DecisionDeny)
	}
	if result.RuleName != "broad-deny" {
		t.Fatalf("RuleName = %q, want broad-deny", result.RuleName)
	}
}

func TestWildcardMatching(t *testing.T) {
	engine := mustCompilePolicy(t, `
version: v1
rules:
  - name: approval-prod-databases
    action: require_approval
    match:
      server: db-*
      tool: schema/*
      args:
        database: prod-*
`)

	result := engine.Evaluate(Input{
		Server: "db-primary",
		Tool:   "schema/migrate",
		Args:   map[string]any{"database": "prod-main"},
	})
	if result.Action != DecisionRequireApproval {
		t.Fatalf("Action = %q, want %q", result.Action, DecisionRequireApproval)
	}
}

func TestMissingArgsDoNotMatch(t *testing.T) {
	engine := mustCompilePolicy(t, `
version: v1
rules:
  - name: approval-prod
    action: require_approval
    match:
      tool: deploy
      args:
        environment: prod
`)

	result := engine.Evaluate(Input{Server: "deployer", Tool: "deploy", Args: map[string]any{"service": "web"}})
	if result.Action != DecisionDeny {
		t.Fatalf("Action = %q, want default deny", result.Action)
	}
	if result.Matched {
		t.Fatalf("Matched = %t, want false", result.Matched)
	}
}

func TestDefaultDenyWhenNoRuleMatches(t *testing.T) {
	engine := mustCompilePolicy(t, `
version: v1
rules:
  - name: allow-read
    action: allow
    match:
      tool: repos/get
`)

	result := engine.Evaluate(Input{Server: "github", Tool: "repos/delete"})
	if result.Action != DecisionDeny {
		t.Fatalf("Action = %q, want default deny", result.Action)
	}
	if result.Reason != "no matching policy rule; deny by default" {
		t.Fatalf("Reason = %q, want default deny reason", result.Reason)
	}
}

func TestLoadFileParsesYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "policy.yaml")
	content := `
version: v1
rules:
  - name: allow-read
    action: allow
    match:
      tool: repos/get
`
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("os.WriteFile() error = %v", err)
	}

	policy, err := LoadFile(path)
	if err != nil {
		t.Fatalf("LoadFile() error = %v", err)
	}
	if len(policy.Rules) != 1 || policy.Rules[0].Name != "allow-read" {
		t.Fatalf("Policy = %+v, want single allow-read rule", policy)
	}
}

func TestInvalidPolicyConfig(t *testing.T) {
	_, err := ParseYAML([]byte(`
version: v2
rules:
  - name: ""
    action: maybe
    match: {}
  - name: dup
    action: allow
    match:
      tool: a
  - name: dup
    action: deny
    match:
      args:
        "": x
        env: ""
`))
	if err == nil {
		t.Fatal("ParseYAML() error = nil, want error")
	}
	if !IsValidationError(err) {
		t.Fatalf("ParseYAML() error type = %T, want ValidationError", err)
	}

	var validationErr *ValidationError
	if !errors.As(err, &validationErr) {
		t.Fatalf("errors.As() failed for %T", err)
	}
	if len(validationErr.Problems) < 5 {
		t.Fatalf("Problems = %v, want multiple validation failures", validationErr.Problems)
	}
}

func TestUnknownYAMLFieldRejected(t *testing.T) {
	_, err := ParseYAML([]byte(`
version: v1
extra: true
rules:
  - name: allow-read
    action: allow
    match:
      tool: repos/get
`))
	if err == nil {
		t.Fatal("ParseYAML() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "field extra not found") {
		t.Fatalf("ParseYAML() error = %v, want unknown field failure", err)
	}
}

func mustCompilePolicy(t *testing.T, raw string) *Engine {
	t.Helper()

	policy, err := ParseYAML([]byte(raw))
	if err != nil {
		t.Fatalf("ParseYAML() error = %v", err)
	}

	engine, err := Compile(policy)
	if err != nil {
		t.Fatalf("Compile() error = %v", err)
	}

	return engine
}
