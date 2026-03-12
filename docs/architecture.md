
---

## `docs/architecture.md`

```md
# Architecture

## Overview

AgentFence is a gateway that sits between an agent runtime and one or more upstream MCP servers.

Its primary responsibility is to make tool access safer and more governable by inserting a policy, approval, and audit layer before upstream MCP tools are invoked.

In the current technical preview, AgentFence focuses on:

- MCP JSON-RPC request handling for `tools/list` and `tools/call`
- allow / deny / require_approval policy evaluation
- redacted audit event recording
- approval workflow support
- forwarding allowed requests to an upstream HTTP MCP server
- admin APIs and a minimal admin UI

## Design goals

AgentFence is designed around these goals:

- deny dangerous actions by default where possible
- make risky actions visible and reviewable
- create audit trails without leaking secrets
- keep the request path deterministic and easy to review
- treat upstream MCP systems as potentially untrusted
- support a simple local developer flow first
- grow toward a stronger control plane later

## High-level flow

For a typical tool call:

1. the agent sends an MCP JSON-RPC request to AgentFence
2. AgentFence parses the request and identifies the operation
3. metadata is extracted for policy evaluation
4. the policy engine returns one of:
   - `allow`
   - `deny`
   - `require_approval`
5. AgentFence writes redacted audit events
6. if allowed, the request is forwarded upstream
7. if denied, the upstream is not called
8. if approval is required, an approval record is created and the upstream is not called
9. operators can inspect and resolve approval requests through the CLI and admin APIs

## Main modules

### `cmd/agentfence`

The main gateway entrypoint.

Responsibilities:

- process startup
- config loading
- dependency wiring
- starting HTTP APIs and the MCP endpoint

### `cmd/agentfence-cli`

The operator CLI.

Responsibilities:

- inspect pending approvals
- approve or deny requests
- provide a practical operator loop for the current technical preview

### `cmd/mock-github-mcp`

A local demo upstream MCP server for evaluation and examples.

Responsibilities:

- simulate an upstream MCP server
- make local development and demos easier
- provide a concrete scenario for policy, approval, and audit flows

### `internal/gateway`

The request-path orchestration layer.

Responsibilities:

- receive MCP requests
- parse and validate JSON-RPC envelopes
- connect policy decisions to upstream forwarding behavior
- invoke audit recording
- coordinate approval creation

### `internal/mcp`

Protocol and transport-facing types.

Responsibilities:

- represent MCP-related request and response shapes
- keep protocol logic separated from policy, storage, and admin layers
- define abstractions for current and future transport handling

Current scope is intentionally limited to the MCP surface needed for `tools/list` and `tools/call`.

### `internal/policy`

The policy loading and evaluation layer.

Responsibilities:

- load YAML policies
- validate supported policy shape
- deterministically evaluate actions
- return clear decisions and reasons

Current policy actions:

- `allow`
- `deny`
- `require_approval`

### `internal/audit`

Audit recording and redaction.

Responsibilities:

- represent audit event types
- mask likely sensitive fields
- preserve useful structure while redacting secrets
- support storage and later operator inspection

### `internal/approval`

Approval workflow domain.

Responsibilities:

- create approval records
- list pending approvals
- approve or deny requests
- preserve operator-facing workflow semantics

### `internal/api`

Admin and operator-facing HTTP APIs.

Responsibilities:

- expose health endpoints
- expose recent audit events
- expose pending approvals
- expose policy status or related operator views

### `internal/config`

Typed config loading and validation.

Responsibilities:

- load config from file and environment
- validate required values
- keep startup behavior explicit and predictable

### `internal/storage/postgres`

Postgres-backed repositories.

Responsibilities:

- durable approval storage
- durable audit storage
- support queryability beyond local JSON-file mode

### `internal/telemetry`

Structured operational signals.

Responsibilities:

- logging setup
- future hooks for tracing and metrics

## Data flow boundaries

AgentFence keeps core concerns separated:

- protocol handling
- policy evaluation
- approval workflow
- audit recording
- storage
- admin/API serving

This separation exists to reduce hidden coupling and keep the gateway data path reviewable.

## Request path behavior

### Allowed request

- request enters `POST /mcp`
- policy returns `allow`
- audit records the decision
- request is forwarded upstream
- response metadata is recorded
- response returns to the caller

### Denied request

- request enters `POST /mcp`
- policy returns `deny`
- audit records the decision
- request is not forwarded upstream
- caller receives a blocked response

### Approval-required request

- request enters `POST /mcp`
- policy returns `require_approval`
- audit records the decision
- an approval record is created
- request is not forwarded upstream
- caller receives a pending or blocked response according to current implementation

## Storage modes

### Local JSON-file mode

Useful for:

- local development
- demos
- quick evaluation
- early-stage iteration

Current limitations:

- not designed for broad concurrent usage
- weaker transactional guarantees
- not a full production durability model

### Postgres-backed mode

Useful for:

- stronger durability
- queryable audit and approval history
- future improvement toward a more complete control plane

Longer term, approvals and audit history should move together in a transactional path where creation, resolution, and audit events can be made more atomic and queryable.

## Current technical limits

The current public preview intentionally keeps scope narrow.

Notable limits:

- admin APIs and admin UI do not yet have authentication
- upstream routing is not yet a richer multi-server routing layer
- stdio transport is not yet implemented
- automatic replay of approved requests is not implemented
- current MCP support is intentionally centered on `tools/list` and `tools/call`

## Future architectural directions

Likely future areas include:

- authenticated admin/operator surfaces
- stronger multi-upstream routing
- richer transport support including stdio
- better concurrency and transactional guarantees
- automatic replay or resumable approval flows
- stronger observability and policy simulation
- stronger enterprise deployment posture

## Summary

AgentFence is not yet a complete production control plane, but it already demonstrates the core architecture required for one:

- explicit policy decisions before upstream tool execution
- redacted audit recording
- operator-visible approval workflows
- a narrow, understandable request path