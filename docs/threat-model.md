# Threat Model

This document describes the current security assumptions and major risks for AgentFence v0.1.

## Security Goals

AgentFence exists to reduce the risk of unsafe MCP tool use by:

- enforcing deny-by-default policy decisions
- requiring human approval for selected operations
- recording reconstructable audit events
- preventing obvious secrets from leaking into logs and stored audit data
- isolating the decision point from upstream MCP server behavior

## Assets To Protect

- credentials embedded in tool arguments or upstream connection strings
- integrity of policy decisions
- integrity of approval records
- integrity and usefulness of audit records
- downstream systems accessed through MCP tools

## Trust Boundaries

### Between agent runtime and AgentFence

The caller can send arbitrary HTTP and JSON-RPC input. Requests must be treated as untrusted.

### Between AgentFence and upstream MCP servers

Upstream MCP servers are treated as potentially untrusted. They may:

- return malformed JSON-RPC
- return misleading errors
- return non-2xx HTTP responses
- echo or embed sensitive material in error messages

### Between operators and admin surfaces

The current repository exposes admin APIs and an admin UI, but authentication is not implemented yet. In practice, that means the deployment boundary must provide protection for these endpoints.

## In-Scope Threats

- malformed JSON-RPC request bodies
- unsafe default allow behavior
- approval-required requests bypassing approval
- upstream failure modes that are accidentally treated as success
- secret leakage through audit storage or logs
- race conditions that overwrite a previous approval decision

## Current Mitigations

### Deny by default

- missing policy evaluation falls back to deny
- unsupported MCP methods are rejected
- malformed JSON-RPC requests fail safely
- denied and approval-required requests are not forwarded upstream

### Approval flow protections

- approval-required requests create durable approval records before returning blocked responses
- CLI resolution now requires an explicit actor
- conflicting repository updates are rejected instead of silently overwriting a terminal decision

### Audit protections

- request arguments are redacted recursively for likely secret keys
- upstream error text is redacted for inline secrets and URL userinfo
- upstream target metadata is sanitized before recording

### Upstream protections

- non-2xx upstream HTTP responses are treated as failures even if they contain JSON-RPC bodies
- response bodies are size-limited before decode
- decode failures and timeouts are surfaced as gateway errors and audited as failures

## Known Residual Risks

### Admin surface is unauthenticated

This is the largest current gap. Anyone who can reach the admin endpoints can read audit and approval data.

### Actor identity is caller-provided

The gateway records `X-AgentFence-Actor`, but does not authenticate or verify it. The value is useful for demos and trusted environments only.

### Local file storage is not production-grade

The JSON-file audit and approval repositories are good for local development and demos, but not strong enough for multi-process or hostile environments.

### Upstream output is not deeply normalized

AgentFence validates JSON-RPC structure, but it does not deeply inspect or rewrite all upstream result payloads.

## Deferred Work Before Stronger Production Claims

- authentication and authorization for admin and approval surfaces
- signed or otherwise authenticated operator identity
- automatic replay or resume semantics after approval
- per-server upstream routing and registration
- stdio transport implementation and its own hardening review
- broader end-to-end and race-focused test coverage
