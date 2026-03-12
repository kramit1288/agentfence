package audit

import (
	"strconv"
	"time"

	"github.com/agentfence/agentfence/internal/mcp/protocol"
	"github.com/agentfence/agentfence/internal/mcp/transport"
	"github.com/agentfence/agentfence/internal/policy"
)

// Builder constructs redacted audit events for request handling paths.
type Builder struct {
	now func() time.Time
}

func NewBuilder() Builder {
	return Builder{now: func() time.Time { return time.Now().UTC() }}
}

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

func (b Builder) BuildUpstreamCall(request protocol.Request, server string, tool string, args map[string]any, result transport.ForwardResult) Event {
	if b.now == nil {
		b = NewBuilder()
	}
	errText := ""
	if result.Err != nil {
		errText = RedactText(result.Err.Error())
	}
	return Event{
		Timestamp: b.now(),
		Kind:      EventKindUpstreamCall,
		Request: RequestContext{
			ID:        requestID(request.ID),
			Server:    server,
			Tool:      tool,
			Method:    request.Method,
			Arguments: RedactMap(args),
		},
		Upstream: UpstreamContext{
			Target:         result.Target,
			Outcome:        string(result.Outcome),
			HTTPStatusCode: result.HTTPStatusCode,
			Latency:        result.Latency,
			Error:          errText,
			Forwarded:      result.Err == nil,
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