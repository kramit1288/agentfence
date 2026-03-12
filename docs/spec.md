# AgentFence Spec

## Product
AgentFence is a gateway that sits between agent runtimes and MCP servers.

## v0.1 objective
Secure any MCP server locally or in staging with:
- tool allow/deny policies
- approval-required policies
- request/response audit logs
- secret redaction
- CLI
- local dashboard
- example integrations

## Non-goals for v0.1
- enterprise multi-tenancy
- advanced anomaly detection
- full compliance exports
- billing
- deep analytics

## Main modules
- gateway
- mcp transport/protocol
- policy engine
- approval service
- audit/event store
- config loader
- admin API
- CLI
- minimal web UI

## Core flow
1. Receive MCP request
2. Parse request
3. Identify tool/server/session metadata
4. Evaluate policy
5. Deny / require approval / allow
6. Redact sensitive fields for logs
7. Forward to upstream if allowed
8. Record audit event
9. Return response

## Critical invariants
- dangerous actions are not silently allowed
- secret values are not logged
- approval-required actions cannot bypass approval
- audit events are reconstructable
- malformed requests fail safely

## Initial integrations
- GitHub MCP
- Slack MCP
- Postgres/database MCP