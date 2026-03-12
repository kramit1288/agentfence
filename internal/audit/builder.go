package audit

import (
	"strconv"
	"time"

	"github.com/agentfence/agentfence/internal/mcp/protocol"
	"github.com/agentfence/agentfence/internal/policy"
)

// Builder constructs redacted audit events for request handling paths.
type Builder struct {
	now func() time.Time
}

// NewBuilder returns an audit event builder using the current UTC time.
func NewBuilder() Builder {
	return Builder{
		now: func() time.Time { return time.Now().UTC() },
	}
}

// BuildPolicyDecision records the redacted request metadata and policy outcome.
func (b Builder) BuildPolicyDecision(request protocol.Request, server string, tool string, args map[string]any, result policy.Result) Event {
	if b.now == nil {
		b = NewBuilder()
	}

	return Event{
		Timestamp: b.now(),
		Kind:      EventKindPolicyDecision,
		Request: RequestContext{
			ID:        requestID(request.ID),
			Server:    server,
			Tool:      tool,
			Method:    request.Method,
			Arguments: RedactMap(args),
		},
		Decision: DecisionContext{
			Action:   result.Action,
			Reason:   result.Reason,
			RuleName: result.RuleName,
			Allowed:  result.Action == policy.DecisionAllow,
		},
	}
}

func requestID(id *protocol.ID) string {
	if id == nil {
		return ""
	}
	if value, ok := id.StringValue(); ok {
		return value
	}
	if value, ok := id.IntValue(); ok {
		return strconv.FormatInt(value, 10)
	}
	return ""
}
