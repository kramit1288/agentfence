package transport

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/agentfence/agentfence/internal/mcp/protocol"
)

const (
	NetworkHTTP  = "http"
	NetworkStdio = "stdio"
)

type Outcome string

const (
	OutcomeSuccess        Outcome = "success"
	OutcomeRPCError       Outcome = "rpc_error"
	OutcomeHTTPError      Outcome = "http_error"
	OutcomeTransportError Outcome = "transport_error"
)

// ForwardResult captures the outcome metadata for one upstream call.
type ForwardResult struct {
	Response       protocol.Response
	HTTPStatusCode int
	Latency        time.Duration
	Outcome        Outcome
	Target         string
	Err            error
}

// Target describes how to reach an upstream MCP server.
type Target struct {
	Network string
	Address string
	Command string
	Args    []string
	Env     []string
	Headers map[string]string
}

func (t Target) Validate() error {
	switch t.Network {
	case NetworkHTTP:
		if t.Address == "" {
			return errors.New("http target address is required")
		}
	case NetworkStdio:
		if t.Command == "" {
			return errors.New("stdio target command is required")
		}
	default:
		return fmt.Errorf("unsupported target network %q", t.Network)
	}
	return nil
}

// Session represents one logical upstream MCP connection.
type Session interface {
	Call(ctx context.Context, request protocol.Request) (protocol.Response, error)
	Notify(ctx context.Context, notification protocol.Request) error
	Close() error
}

// Dialer opens sessions to upstream MCP servers.
type Dialer interface {
	Dial(ctx context.Context, target Target) (Session, error)
}

// Client narrows session behavior to the request/response surface the gateway
// will use most often when proxying to an upstream MCP server.
type Client interface {
	Call(ctx context.Context, request protocol.Request) (protocol.Response, error)
	Notify(ctx context.Context, notification protocol.Request) error
}
