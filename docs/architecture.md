# Architecture

This document describes the current v0.1 architecture as implemented in this repository.

## Design Principles

- deny by default where policy or infrastructure is missing
- keep transport, policy, approval, audit, and storage boundaries explicit
- prefer simple interfaces over framework-heavy indirection
- keep request handling deterministic and reviewable
- treat upstream MCP servers as potentially untrusted

## Main Components

### Gateway

`internal/gateway` owns the core request path.

Responsibilities:

- receive MCP HTTP requests
- decode JSON-RPC messages
- extract server, tool, argument, and actor metadata
- evaluate policy
- create approval records for approval-required requests
- forward only allowed requests
- emit audit events for both decisions and upstream outcomes

The gateway is intentionally thin. Policy, approval, audit, and storage logic live in separate packages and are injected through small interfaces.

### MCP protocol and transport

`internal/mcp/protocol` defines MCP-facing JSON-RPC envelopes and typed payloads for the v0.1 methods AgentFence cares about most:

- `tools/list`
- `tools/call`

`internal/mcp/transport` defines the transport contracts and the current HTTP forwarder implementation.

Current state:

- HTTP forwarding is implemented
- stdio transport is modeled but not implemented
- forwarding is a single configured upstream target, not dynamic per server

### Policy engine

`internal/policy` loads YAML policy files, validates them, compiles them, and evaluates requests deterministically.

Current semantics:

- actions: `allow`, `deny`, `require_approval`
- match dimensions: `server`, `tool`, and top-level argument keys
- wildcard matching via simple glob patterns
- first matching rule wins
- no match means deny

### Approval workflow

`internal/approval` manages approval records and operator resolution.

Current behavior:

- approval-required requests create durable approval records
- CLI operators can list pending approvals and resolve them
- resolution requires an explicit actor
- repository updates now reject stale conflicting overwrites
- approval resolution does not automatically replay the original MCP request

### Audit

`internal/audit` defines redacted event models and helpers.

Current event types:

- `policy.decision`
- `upstream.call`

Redaction goals:

- never persist obvious secrets in request arguments or audit error text
- preserve enough metadata to explain what happened
- record both decision context and upstream outcome metadata

### Storage

Two storage modes exist today.

Local development mode:

- audit events in a local JSON file
- approval requests in a local JSON file

Postgres mode:

- `internal/storage/postgres` implements audit and approval repositories
- migrations are embedded and also mirrored in `deploy/migrations`

### Admin surface

`internal/api` exposes:

- `GET /healthz`
- `GET /api/admin/audit`
- `GET /api/admin/approvals`
- `GET /api/admin/policy`

`web/` is a minimal Next.js UI for those endpoints.

## Request Flow

1. An agent runtime sends an MCP JSON-RPC request to `POST /mcp`.
2. The gateway decodes and validates the JSON-RPC envelope.
3. The gateway extracts the logical server ID plus tool and argument metadata.
4. Policy evaluation returns `allow`, `deny`, or `require_approval`.
5. A redacted `policy.decision` audit event is recorded.
6. If the action is `deny`, the gateway returns a policy error and stops.
7. If the action is `require_approval`, the gateway creates an approval record, returns a blocked response, and stops.
8. If the action is `allow`, the gateway forwards the request to the upstream MCP server.
9. The gateway records a redacted `upstream.call` audit event with latency and outcome metadata.
10. The upstream response is returned to the caller, or a gateway forwarding error is returned if the upstream call failed.

## Configuration Model

`internal/config` merges:

- built-in defaults
- optional JSON config file
- environment overrides

Important runtime environment variables outside the typed config file:

- `AGENTFENCE_POLICY_FILE`
- `AGENTFENCE_UPSTREAM_URL`
- `AGENTFENCE_UPSTREAM_TIMEOUT`
- `AGENTFENCE_POSTGRES_DSN`
- `AGENTFENCE_APPROVAL_STORE`
- `AGENTFENCE_AUDIT_STORE`

## Known Architectural Limits

- no authentication or RBAC yet for admin endpoints or approval actions
- no request replay after approval
- no stdio upstream transport yet
- no multi-upstream routing or per-server registry
- no durable session abstraction beyond the current HTTP forwarding path
- local file repositories are useful for demos, not multi-process production use
