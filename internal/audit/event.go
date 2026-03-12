package audit

import (
	"context"
	"time"

	"github.com/agentfence/agentfence/internal/policy"
)

type EventKind string

const EventKindPolicyDecision EventKind = "policy.decision"

// Event captures one redacted gateway activity record.
type Event struct {
	Timestamp time.Time        `json:"timestamp"`
	Kind      EventKind        `json:"kind"`
	Request   RequestContext   `json:"request"`
	Decision  DecisionContext  `json:"decision"`
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
	Action  policy.Decision `json:"action"`
	Reason  string          `json:"reason"`
	RuleName string         `json:"rule_name,omitempty"`
	Allowed bool            `json:"allowed"`
}

// Sink records durable audit events.
type Sink interface {
	Record(ctx context.Context, event Event) error
}
