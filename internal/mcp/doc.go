// Package mcp contains transport and protocol-facing MCP primitives.
//
// Subpackages split JSON-RPC protocol modeling from transport concerns so the
// gateway can add HTTP and stdio client implementations without mixing wire
// parsing, session behavior, and policy logic.
package mcp
