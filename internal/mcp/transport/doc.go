// Package transport defines the minimal abstractions for talking to upstream
// MCP servers over different connection types.
//
// Implementations may later wrap remote HTTP endpoints or local stdio
// subprocesses, but the gateway only depends on the session behavior declared
// here.
package transport
