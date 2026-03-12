package policy

import (
	"path/filepath"
	"testing"
)

func TestGitHubDemoPolicyExamples(t *testing.T) {
	policyPath := filepath.Join("..", "..", "examples", "github-mcp", "policy.yaml")
	document, err := LoadFile(policyPath)
	if err != nil {
		t.Fatalf("LoadFile() error = %v", err)
	}
	engine, err := Compile(document)
	if err != nil {
		t.Fatalf("Compile() error = %v", err)
	}

	tests := []struct {
		name   string
		input  Input
		action Decision
		rule   string
	}{
		{
			name:   "read only repo metadata allowed",
			input:  Input{Server: "github-demo", Tool: "repos/get", Args: map[string]any{"owner": "agentfence", "repo": "agentfence"}},
			action: DecisionAllow,
			rule:   "allow-readonly-github-tools",
		},
		{
			name:   "pull request merge requires approval",
			input:  Input{Server: "github-demo", Tool: "pulls/merge", Args: map[string]any{"pull_number": 42}},
			action: DecisionRequireApproval,
			rule:   "require-approval-for-merges",
		},
		{
			name:   "repository delete denied",
			input:  Input{Server: "github-demo", Tool: "repos/delete", Args: map[string]any{"owner": "agentfence", "repo": "agentfence"}},
			action: DecisionDeny,
			rule:   "deny-repository-deletes",
		},
		{
			name:   "unknown github tool denied by catch all",
			input:  Input{Server: "github-demo", Tool: "actions/run", Args: map[string]any{}},
			action: DecisionDeny,
			rule:   "deny-everything-else",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := engine.Evaluate(tc.input)
			if result.Action != tc.action {
				t.Fatalf("Action = %q, want %q", result.Action, tc.action)
			}
			if result.RuleName != tc.rule {
				t.Fatalf("RuleName = %q, want %q", result.RuleName, tc.rule)
			}
		})
	}
}