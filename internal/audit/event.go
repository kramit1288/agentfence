package audit

import (
	"context"
	"time"

	"github.com/agentfence/agentfence/internal/policy"
)

type EventKind string

const (
	EventKindPolicyDecision EventKind = "policy.decision"
	EventKindUpstreamCall   EventKind = "upstream.call"
)

// Event captures one redacted gateway activity record.
type Event struct {
	Timestamp time.Time       `json:"timestamp"`
	Kind      EventKind       `json:"kind"`
	Request   RequestContext  `json:"request"`
	Decision  DecisionContext `json:"decision,omitempty"`
	Upstream  UpstreamContext `json:"upstream,omitempty"`
}

// RequestContext contains redacted request metadata that is safe to log or store.
type RequestContext struct {
	ID        string         `json:"id,omitempty"`
	Server    string         `json:"server"`
	Tool      string         `json:"tool"`
	Method    string         `json:"method"`
	Arguments map[string]any `json:"arguments,omitempty"`
}

// DecisionContext captures the policy outcome for a request.
type DecisionContext struct {
	Action   policy.Decision `json:"action"`
	Reason   string          `json:"reason"`
	RuleName string          `json:"rule_name,omitempty"`
	Allowed  bool            `json:"allowed"`
}

// UpstreamContext captures one proxy attempt to an upstream MCP server.
type UpstreamContext struct {
	Target         string        `json:"target"`
	Outcome        string        `json:"outcome"`
	HTTPStatusCode int           `json:"http_status_code,omitempty"`
	Latency        time.Duration `json:"latency,omitempty"`
	Error          string        `json:"error,omitempty"`
	Forwarded      bool          `json:"forwarded"`
}

// Sink records durable audit events.
type Sink interface {
	Record(ctx context.Context, event Event) error
}
