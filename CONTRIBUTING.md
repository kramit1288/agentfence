# Contributing

Thanks for considering a contribution to AgentFence.

## Project Priorities

AgentFence is trying to be technically serious and reviewable. Favor changes that improve:

- policy correctness
- auditability
- security boundaries
- local developer usability
- test quality

## Development Setup

Prerequisites:

- Go 1.22+
- Node.js 20+
- npm 10+
- optional: Docker for local Postgres

Useful commands:

- `make test`
- `make gateway`
- `make cli`
- `make web-install`
- `make web-dev`
- `make web-build`
- `make web-lint`

Direct commands:

- `go test ./...`
- `go run ./cmd/agentfence`
- `go run ./cmd/agentfence-cli`
- `cd web && npm install && npm run dev`

## Working Style

- prefer small, reviewable pull requests
- avoid broad refactors unless they clearly pay for themselves
- preserve separation between transport, policy, approval, audit, storage, and API logic
- keep the gateway path deterministic and easy to reason about
- avoid introducing dependencies without a strong reason

## Testing Expectations

At minimum, contributors should run the smallest relevant test set first, then broader validation when practical.

Backend changes:

- `go test ./...`
- `go test ./... -race` for concurrency-sensitive changes when practical

Web changes:

- `cd web && npm run lint`
- `cd web && npm run build`

If you cannot run a validation step, say so clearly in the PR.

## Security Expectations

Do not weaken these without explicit discussion:

- deny-by-default behavior
- redaction rules for secrets and credentials
- approval workflow semantics
- audit coverage for critical request paths
- trust-boundary assumptions around upstream MCP servers

## Documentation Expectations

If behavior changes, update the relevant docs. For first-order features, that usually means some combination of:

- `README.md`
- `QUICKSTART.md`
- `docs/architecture.md`
- `docs/threat-model.md`
- demo docs or example policies

## Reporting Bugs

Use the issue templates when possible. Security-sensitive issues should avoid posting raw secrets, credentials, or exploit details in public issues.

## Pull Requests

A good PR description should include:

- what changed
- why it changed
- tests run
- known risks or follow-up work
