
---

## `CONTRIBUTING.md`

```md
# Contributing

Thanks for your interest in contributing to AgentFence.

AgentFence is an early OSS technical preview focused on building a production-oriented control layer around MCP tool access.

At this stage, the most helpful contributions are the ones that improve:

- correctness
- clarity
- safety
- operator usability
- test coverage
- docs quality

## Before you contribute

Please read:

- [README.md](README.md)
- [QUICKSTART.md](QUICKSTART.md)
- [docs/architecture.md](docs/architecture.md)
- [docs/threat-model.md](docs/threat-model.md)
- [docs/spec.md](docs/spec.md)

The project intentionally keeps scope narrow in the current preview. Contributions that preserve that clarity are more likely to be accepted than broad feature dumps.

## What kinds of contributions are helpful

Examples of helpful contributions:

- bug fixes
- tests for policy, audit, approval, and gateway behavior
- docs improvements
- safer defaults
- redaction improvements
- local developer experience improvements
- demo improvements
- narrowly scoped transport or storage improvements
- issue reports with clear reproduction steps

Examples of contributions that may need discussion first:

- major architectural changes
- new transport layers
- auth/authz model changes
- policy language expansion
- broad UI redesigns
- major dependency additions

## Development setup

### Prerequisites

- Go 1.22+
- Node.js 20+
- npm 10+

### Install dependencies

Backend:

```bash
go mod tidy

Frontend:

cd web
npm install
cd ..
Run tests

From repo root:

go test ./...

Frontend build validation:

cd web
npm run build
Local workflow

A good local workflow is:

pull the latest changes

create a short-lived branch

make one focused change

add or update tests

run local validation

update docs if behavior changed

open a pull request

Coding principles

Please preserve these principles:

keep the request path explicit and reviewable

prefer small, focused changes

separate protocol, policy, approval, audit, and storage concerns

avoid unnecessary dependencies

keep docs honest about current limitations

prefer clarity over cleverness

do not weaken deny-by-default, redaction, or approval semantics casually

Security-sensitive areas

Be especially careful when changing:

policy evaluation

redaction behavior

approval workflow semantics

admin APIs

storage integrity

upstream forwarding behavior

config parsing and defaults

Changes in these areas should include tests and a clear explanation of the behavior change.

Pull request guidance

Good pull requests are:

scoped to one logical change

easy to review

tested

documented where needed

explicit about limitations or tradeoffs

A good PR description should include:

what changed

why it changed

how it was tested

whether docs were updated

any known limitations

Issues

When opening an issue, include as much of the following as possible:

expected behavior

actual behavior

steps to reproduce

relevant config

platform details

logs or error output

whether the issue affects local JSON-file mode, Postgres mode, or both

Docs contributions

Docs matter a lot in this project.

Helpful doc contributions include:

clearer setup steps

better examples

policy examples

threat-model clarifications

operator workflow docs

deployment caveats

Current project status

Please keep in mind:

AgentFence is still an early technical preview

the admin UI and admin APIs are not yet authenticated

stdio transport is not yet implemented

approval resolution does not automatically replay requests

the current focus is correctness and architecture, not breadth

Communication

When proposing a substantial change, open an issue or discussion first so the direction can be aligned before implementation work grows.