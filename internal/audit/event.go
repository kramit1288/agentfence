package audit

import (
	"context"
	"time"

	"github.com/agentfence/agentfence/internal/policy"
)

// Event captures one policy decision in the gateway request path.
type Event struct {
	Timestamp time.Time
	Server    string
	Tool      string
	Decision  policy.Decision
	Reason    string
	RuleName  string
	Method    string
	Allowed   bool
}

// Sink records durable audit events.
type Sink interface {
	Record(ctx context.Context, event Event) error
}
