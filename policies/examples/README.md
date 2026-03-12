# Policy Examples

These files show the current AgentFence YAML policy format.

## Format summary

- `version` must be `v1`
- rules are evaluated in file order
- the first matching rule wins
- if no rule matches, the request is denied

Supported actions:

- `allow`
- `deny`
- `require_approval`

Supported match keys:

- `server`
- `tool`
- `args`

Wildcards use simple glob semantics with `*` and `?`.

## Included examples

- `allow-readonly.yaml` small read-only example
- `require-approval.yaml` simple approval example
- `deny-shell.yaml` shell denial example
- `github-production.yaml` GitHub-oriented production guardrails
- `database-readonly.yaml` database-oriented read-only guardrails

These are example starting points, not turnkey production policies.
