// Package protocol defines the minimal MCP JSON-RPC types AgentFence needs in
// v0.1 to parse envelopes and reason about tool-related traffic.
//
// Assumptions for v0.1:
// - AgentFence validates generic JSON-RPC 2.0 envelopes for all methods.
// - The only method-specific payloads modeled here are `tools/list` and
//   `tools/call`, because policy and audit decisions in v0.1 are tool-centric.
// - Other MCP methods may still transit the gateway later, but they are not yet
//   represented as typed params/results in this package.
package protocol
