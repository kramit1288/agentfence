package gateway

import "github.com/agentfence/agentfence/internal/config"

// Gateway is the top-level coordinator for MCP request handling.
type Gateway struct {
	cfg config.Config
}

// New constructs a Gateway with explicit configuration.
func New(cfg config.Config) *Gateway {
	return &Gateway{cfg: cfg}
}

// ListenAddr returns the configured bind address for the gateway process.
func (g *Gateway) ListenAddr() string {
	return g.cfg.ListenAddress
}
