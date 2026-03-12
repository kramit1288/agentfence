package policy

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

const Version = "v1"

type Decision string

const (
	DecisionAllow           Decision = "allow"
	DecisionDeny            Decision = "deny"
	DecisionRequireApproval Decision = "require_approval"
)

// Policy is the human-authored policy document format.
type Policy struct {
	Version string `yaml:"version"`
	Rules   []Rule `yaml:"rules"`
}

// Rule describes one ordered policy decision.
type Rule struct {
	Name   string    `yaml:"name"`
	Action Decision  `yaml:"action"`
	Reason string    `yaml:"reason,omitempty"`
	Match  RuleMatch `yaml:"match"`
}

// RuleMatch selects requests by server, tool, and argument values.
type RuleMatch struct {
	Server string            `yaml:"server,omitempty"`
	Tool   string            `yaml:"tool,omitempty"`
	Args   map[string]string `yaml:"args,omitempty"`
}

// Input is the policy evaluation input.
type Input struct {
	Server string
	Tool   string
	Args   map[string]any
}

// Result is the deterministic outcome of policy evaluation.
type Result struct {
	Action   Decision
	Matched  bool
	RuleName string
	Reason   string
}

// Engine is a compiled policy matcher.
type Engine struct {
	rules []compiledRule
}

type compiledRule struct {
	name   string
	action Decision
	reason string
	match  compiledMatch
}

type compiledMatch struct {
	server pattern
	tool   pattern
	args   []argPattern
}

type pattern struct {
	set   bool
	value string
}

type argPattern struct {
	key     string
	pattern string
}

// ValidationError aggregates configuration problems in a policy file.
type ValidationError struct {
	Problems []string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("invalid policy: %s", strings.Join(e.Problems, "; "))
}

// ParseYAML parses a policy document from YAML and validates its shape.
func ParseYAML(data []byte) (Policy, error) {
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)

	var policy Policy
	if err := decoder.Decode(&policy); err != nil {
		return Policy{}, fmt.Errorf("decode policy yaml: %w", err)
	}
	if err := policy.Validate(); err != nil {
		return Policy{}, err
	}
	return policy, nil
}

// LoadFile loads a policy document from a YAML file.
func LoadFile(path string) (Policy, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Policy{}, fmt.Errorf("read policy file %q: %w", path, err)
	}
	return ParseYAML(data)
}

// Compile validates and compiles a policy into an evaluation engine.
func Compile(policy Policy) (*Engine, error) {
	if err := policy.Validate(); err != nil {
		return nil, err
	}

	compiled := make([]compiledRule, 0, len(policy.Rules))
	for _, rule := range policy.Rules {
		compiled = append(compiled, compiledRule{
			name:   rule.Name,
			action: rule.Action,
			reason: rule.effectiveReason(),
			match:  compileMatch(rule.Match),
		})
	}

	return &Engine{rules: compiled}, nil
}

func (p Policy) Validate() error {
	var problems []string

	if p.Version != Version {
		problems = append(problems, fmt.Sprintf("version must be %q", Version))
	}
	if len(p.Rules) == 0 {
		problems = append(problems, "at least one rule is required")
	}

	seenNames := make(map[string]struct{}, len(p.Rules))
	for i, rule := range p.Rules {
		prefix := fmt.Sprintf("rules[%d]", i)
		if rule.Name == "" {
			problems = append(problems, prefix+".name is required")
		} else {
			if _, exists := seenNames[rule.Name]; exists {
				problems = append(problems, prefix+".name must be unique")
			}
			seenNames[rule.Name] = struct{}{}
		}
		switch rule.Action {
		case DecisionAllow, DecisionDeny, DecisionRequireApproval:
		default:
			problems = append(problems, prefix+".action must be allow, deny, or require_approval")
		}
		if rule.Match.isEmpty() {
			problems = append(problems, prefix+".match must include server, tool, or args")
		}
		for key, value := range rule.Match.Args {
			if key == "" {
				problems = append(problems, prefix+".match.args keys must not be empty")
			}
			if value == "" {
				problems = append(problems, prefix+".match.args["+key+"] pattern is required")
			}
		}
	}

	if len(problems) > 0 {
		return &ValidationError{Problems: problems}
	}
	return nil
}

func (m RuleMatch) isEmpty() bool {
	return m.Server == "" && m.Tool == "" && len(m.Args) == 0
}

func (r Rule) effectiveReason() string {
	if r.Reason != "" {
		return r.Reason
	}
	return fmt.Sprintf("matched policy rule %q", r.Name)
}

func compileMatch(match RuleMatch) compiledMatch {
	compiled := compiledMatch{
		server: pattern{set: match.Server != "", value: match.Server},
		tool:   pattern{set: match.Tool != "", value: match.Tool},
	}

	keys := make([]string, 0, len(match.Args))
	for key := range match.Args {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	compiled.args = make([]argPattern, 0, len(keys))
	for _, key := range keys {
		compiled.args = append(compiled.args, argPattern{key: key, pattern: match.Args[key]})
	}

	return compiled
}

// Evaluate returns the first matching rule in policy order or denies by default.
func (e *Engine) Evaluate(input Input) Result {
	for _, rule := range e.rules {
		if rule.match.matches(input) {
			return Result{
				Action:   rule.action,
				Matched:  true,
				RuleName: rule.name,
				Reason:   rule.reason,
			}
		}
	}

	return Result{
		Action:  DecisionDeny,
		Matched: false,
		Reason:  "no matching policy rule; deny by default",
	}
}

func (m compiledMatch) matches(input Input) bool {
	if m.server.set && !globMatch(m.server.value, input.Server) {
		return false
	}
	if m.tool.set && !globMatch(m.tool.value, input.Tool) {
		return false
	}
	for _, arg := range m.args {
		actual, ok := input.Args[arg.key]
		if !ok {
			return false
		}
		if !globMatch(arg.pattern, stringifyArg(actual)) {
			return false
		}
	}
	return true
}

func stringifyArg(value any) string {
	switch v := value.(type) {
	case string:
		return v
	case fmt.Stringer:
		return v.String()
	case nil:
		return "null"
	default:
		encoded, err := json.Marshal(v)
		if err == nil {
			if len(encoded) >= 2 && encoded[0] == '"' && encoded[len(encoded)-1] == '"' {
				return string(encoded[1 : len(encoded)-1])
			}
			return string(encoded)
		}
		return fmt.Sprint(v)
	}
}

func globMatch(pattern string, value string) bool {
	patternRunes := []rune(pattern)
	valueRunes := []rune(value)

	dp := make([][]bool, len(patternRunes)+1)
	for i := range dp {
		dp[i] = make([]bool, len(valueRunes)+1)
	}
	dp[0][0] = true

	for i := 1; i <= len(patternRunes); i++ {
		if patternRunes[i-1] == '*' {
			dp[i][0] = dp[i-1][0]
		}
	}

	for i := 1; i <= len(patternRunes); i++ {
		for j := 1; j <= len(valueRunes); j++ {
			switch patternRunes[i-1] {
			case '*':
				dp[i][j] = dp[i-1][j] || dp[i][j-1]
			case '?':
				dp[i][j] = dp[i-1][j-1]
			default:
				dp[i][j] = dp[i-1][j-1] && patternRunes[i-1] == valueRunes[j-1]
			}
		}
	}

	return dp[len(patternRunes)][len(valueRunes)]
}

func IsValidationError(err error) bool {
	var validationErr *ValidationError
	return errors.As(err, &validationErr)
}
