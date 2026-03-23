# AgentFence

AgentFence is an open-source gateway for securing, governing, and auditing MCP tool access in production agent systems.

It sits between an agent runtime and one or more upstream MCP servers, evaluates policy before a tool call is forwarded, records redacted audit events, and supports approval workflows for risky actions.

## Current Scope

The repository already includes a usable v0.1 foundation:

- MCP JSON-RPC request parsing for `tools/list` and `tools/call`
- deterministic allow / deny / `require_approval` policy evaluation
- redacted audit event recording
- approval request creation and operator CLI resolution
- HTTP forwarding to an upstream MCP server
- admin APIs for recent audit events, pending approvals, and policy status
- a minimal Next.js admin UI
- a runnable GitHub MCP demo flow

This is an early OSS technical preview. AgentFence is suitable for local evaluation and staging-style demos, but it is not yet a complete production control plane.

## Why AgentFence

MCP servers expose powerful tool surfaces. In practice, production operators need more than raw connectivity:

- deny-by-default behavior for dangerous tools
- human approval for sensitive writes
- auditability that does not leak secrets
- a clear trust boundary between agents and upstream MCP servers
- predictable failure handling when inputs are malformed or upstreams misbehave

AgentFence is built around those constraints.

## What Works Today

- `POST /mcp` accepts MCP JSON-RPC requests
- policy decisions support `allow`, `deny`, and `require_approval`
- allowed requests can proxy to one upstream HTTP MCP server
- approval-required requests create durable approval records
- denied and approval-required requests never forward upstream
- recent audit events and pending approvals are queryable through admin APIs
- local development can use JSON-file audit and approval storage without Postgres
- Postgres-backed audit and approval repositories are available for durable storage

## Quickstart

Start with [QUICKSTART.md](QUICKSTART.md) for the shortest runnable path.

If you want a concrete scenario, use the GitHub MCP demo in [docs/demos/github-mcp.md](docs/demos/github-mcp.md).

## Repository Layout

```text
cmd/                   binary entrypoints
  agentfence/          gateway process
  agentfence-cli/      operator CLI
  mock-github-mcp/     demo upstream MCP server
internal/              Go application code
  api/                 health and admin HTTP APIs
  approval/            approval workflow domain + repositories
  audit/               redaction and audit event models
  config/              typed config loading + validation
  gateway/             MCP request path orchestration
  mcp/                 protocol models and transport abstractions
  policy/              YAML policy loading and evaluation
  storage/postgres/    Postgres repositories + migrations
  telemetry/           structured logging setup
web/                   minimal Next.js admin UI
examples/              runnable demo assets
policies/examples/     reusable example policy documents
docs/                  architecture, threat model, demos, checklists
```

## Main Commands

Prerequisites:

- Go 1.22+
- Node.js 20+
- npm 10+
- optional: Docker for local Postgres

From the repository root:

- `make test` runs Go tests
- `make gateway` runs the gateway
- `make cli` runs the operator CLI
- `make web-install` installs web dependencies
- `make web-dev` starts the admin UI
- `make web-build` builds the admin UI
- `make web-lint` lints the admin UI

Without `make`:

- `go run ./cmd/agentfence`
- `go run ./cmd/agentfence-cli`
- `cd web && npm install && npm run dev`

## Configuration Summary

The gateway reads config from JSON file plus environment variables.

Important environment variables:

- `AGENTFENCE_POLICY_FILE` path to a YAML policy file
- `AGENTFENCE_UPSTREAM_URL` upstream HTTP MCP endpoint
- `AGENTFENCE_POSTGRES_DSN` enable Postgres-backed audit and approvals
- `AGENTFENCE_APPROVAL_STORE` local JSON approval store path when Postgres is unset
- `AGENTFENCE_AUDIT_STORE` local JSON audit store path when Postgres is unset
- `AGENTFENCE_UPSTREAM_TIMEOUT` upstream HTTP client timeout
- `AGENTFENCE_API_BASE` backend base URL for the Next.js admin UI

## Docs

- [QUICKSTART.md](QUICKSTART.md)
- [docs/architecture.md](docs/architecture.md)
- [docs/threat-model.md](docs/threat-model.md)
- [docs/spec.md](docs/spec.md)
- [docs/demos/github-mcp.md](docs/demos/github-mcp.md)
- [CONTRIBUTING.md](CONTRIBUTING.md)

## Rough Edges

The first OSS release still has real limitations:

- admin APIs and UI do not have authentication yet
- upstream routing is a single HTTP target, not per-server routing
- stdio transport is modeled but not implemented
- approval resolution does not replay the original request automatically
- there is no generated `go.sum` in this repository snapshot yet
- web dependency lockfiles are not committed yet
- validation in this environment could not be completed because the Go toolchain is unavailable on `PATH`

Those gaps are intentional to call out before broader adoption.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md).

## License

By contributing to AgentFence, you agree that your contributions will be licensed under the repository’s Apache-2.0 license.

