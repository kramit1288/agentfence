# AgentFence

AgentFence is an open-source gateway for securing, governing, and auditing MCP tool access in production agent systems.

It sits between agent runtimes and upstream MCP servers to enforce policy, approvals, and safe execution boundaries before tools touch real systems.

## Why AgentFence

As agents move from demos into production, tool access becomes a security, governance, and operability problem.

AgentFence is built to help teams:

- enforce deny-by-default access to MCP tools
- require approvals for risky actions
- audit tool usage and decision flow
- treat upstream MCP servers as potentially untrusted
- create safer operational boundaries for production agents

## Project Status

AgentFence is in early development.

The repository currently contains the initial scaffold, package boundaries, docs, and development structure. Core gateway, policy enforcement, approval flow, and audit capabilities are planned next.

## Planned v0.1

- MCP gateway request path
- allow / deny / approval policy engine
- audit event model
- secret redaction helpers
- CLI for operator workflows
- minimal admin UI
- example policies and integrations

## Architecture Overview

The repo separates core concerns so the gateway data path can remain deterministic and reviewable:

- `internal/gateway`: request orchestration and gateway flow coordination
- `internal/mcp`: MCP transport and protocol-facing types
- `internal/policy`: allow, deny, and approval policy evaluation
- `internal/approval`: approval workflow primitives
- `internal/audit`: durable audit event modeling
- `internal/api`: admin and operator-facing API surface
- `internal/config`: configuration loading and validation
- `internal/storage`: storage interfaces and implementations
- `internal/telemetry`: structured observability hooks
- `cmd/agentfence`: gateway entrypoint
- `cmd/agentfence-cli`: operator CLI entrypoint
- `web/`: minimal admin UI

## Repository Structure

```text
.
|-- cmd/
|   |-- agentfence/
|   `-- agentfence-cli/
|-- deploy/
|-- docs/
|-- internal/
|   |-- api/
|   |-- approval/
|   |-- audit/
|   |-- config/
|   |-- gateway/
|   |-- mcp/
|   |-- policy/
|   |-- storage/
|   `-- telemetry/
|-- policies/
|   `-- examples/
|-- tests/
|   |-- e2e/
|   |-- integration/
|   `-- unit/
`-- web/
```

## Local Development

Prerequisites:

- Go 1.22+
- Node.js 20+
- npm 10+
- GNU Make or compatible `make`

Common commands:

- `make fmt`: format Go code
- `make test`: run Go tests
- `make gateway`: run the gateway entrypoint
- `make cli`: run the CLI entrypoint
- `make web-dev`: start the web UI in development mode
- `make web-build`: build the web UI

The current scaffold intentionally keeps runtime dependencies minimal. The Go side uses the standard library only. The web UI uses the smallest Next.js + TypeScript setup needed to establish the admin surface.


## Roadmap

Near-term focus:

build the first end-to-end MCP request path

enforce allow / deny / approval policies

add audit logging and redaction

add approval workflow support

ship a usable local development flow

add example integrations and policies

## Contributing

Contributions, feedback, design discussion, and issue reports are welcome.

As the project evolves, contribution guidelines and issue templates will be added.

