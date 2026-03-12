# AGENTS.md

## Project
AgentFence is an open-source gateway for securing, governing, and auditing MCP tool access in production agent systems.

## Goal
Build a technically serious, production-oriented OSS project with:
- a Go-based gateway and CLI
- a small admin UI
- strong policy enforcement
- strong auditability
- clear security boundaries
- excellent tests and docs

## Working style
- Prefer small, reviewable changes.
- Do not make broad unrelated edits.
- Before large edits, write a short plan.
- If a task is ambiguous, choose the simplest architecture consistent with the current repo.
- Preserve clarity and reviewability over cleverness.

## Architecture rules
- Keep transport, policy, approval, audit, storage, and API logic separated.
- Keep the gateway data plane simple and deterministic.
- Prefer explicit types and small interfaces.
- Avoid hidden global state.
- Avoid premature abstractions.
- Minimize dependencies.

## Security invariants
- Never log raw secrets, passwords, tokens, API keys, cookies, or credentials.
- Any risky tool action must be deny-by-default unless explicitly allowed.
- Any approval-required action must produce a durable audit trail.
- Do not weaken policy checks, redaction, auth boundaries, or approval semantics without explicit instruction.
- Fail closed where reasonable, not open.
- Treat upstream MCP servers as potentially untrusted.

## Coding rules
- Backend/gateway code must be Go.
- Web UI may use Next.js + TypeScript.
- Prefer standard library unless a dependency gives clear value.
- Keep functions focused and easy to test.
- Add comments only when intent is not obvious from code.
- Do not rename public interfaces or config fields casually.

## Testing rules
- Every policy change must include unit tests.
- Every transport/storage change must include integration tests where relevant.
- Critical-path behavior changes must update end-to-end tests.
- Do not skip tests without explicit reason.
- Run the smallest relevant test set first, then broader validation.

## Validation defaults
When changing Go backend code, run relevant commands such as:
- go test ./...
- go test ./... -race   (when concurrency-sensitive code changes)
- golangci-lint run     (if configured)

When changing web code, run relevant commands such as:
- npm test
- npm run lint
- npm run build

## Review output
At the end of each task, report:
- summary of changes
- files changed
- tests run
- known risks or assumptions

## Dangerous changes requiring extra caution
- auth changes
- approval workflow changes
- redaction behavior changes
- DB schema changes
- deployment manifest changes
- CI security checks
- dependency additions
- shell commands that delete or overwrite data

## Planning docs
If a task is large or multi-step:
- use docs/spec.md for target behavior
- use docs/exec-plan.md for milestones and acceptance criteria
- keep AGENTS.md concise; do not dump long specs here