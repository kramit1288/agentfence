# GitHub MCP Demo

This demo shows a minimal AgentFence flow for a GitHub-style MCP server:

- read-only requests are allowed and proxied upstream
- write requests are blocked pending approval
- dangerous requests are denied immediately
- approvals and audit events are stored locally in JSON files by default

## Files

- Policy: `examples/github-mcp/policy.yaml`
- Gateway config: `examples/github-mcp/config.json`
- Sample MCP requests: `examples/github-mcp/requests/*.json`
- Mock upstream server: `cmd/mock-github-mcp`

## Start the mock GitHub MCP server

```powershell
go run ./cmd/mock-github-mcp -addr :8090
```

The mock server exposes these MCP tools:

- `repos/get`
- `pulls/list`
- `issues/comment`
- `pulls/merge`
- `repos/delete`

## Start AgentFence with the demo policy

```powershell
$env:AGENTFENCE_POLICY_FILE = "examples/github-mcp/policy.yaml"
$env:AGENTFENCE_UPSTREAM_URL = "http://127.0.0.1:8090"
$env:AGENTFENCE_APPROVAL_STORE = "data/demo-approvals.json"
$env:AGENTFENCE_AUDIT_STORE = "data/demo-audit.json"
go run ./cmd/agentfence -config examples/github-mcp/config.json
```

Optional: if you want Postgres-backed audit and approval storage instead of local JSON files, start the database in `deploy/docker-compose.yml` and set `AGENTFENCE_POSTGRES_DSN`.

## 1. Allowed read-only call

```powershell
curl -sS -X POST "http://127.0.0.1:8080/mcp?server=github-demo" ^
  -H "Content-Type: application/json" ^
  --data-binary "@examples/github-mcp/requests/read-repo.json"
```

Expected result:

- HTTP `200 OK`
- JSON-RPC success response from the mock upstream
- a `policy.decision` audit event and an `upstream.call` audit event are recorded

## 2. Approval-required write call

```powershell
curl -sS -X POST "http://127.0.0.1:8080/mcp?server=github-demo" ^
  -H "Content-Type: application/json" ^
  -H "X-AgentFence-Actor: demo-user" ^
  --data-binary "@examples/github-mcp/requests/merge-pr.json"
```

Expected result:

- HTTP `403 Forbidden`
- JSON-RPC error with code `-32002`
- response data includes `status: pending` and an `approval_id`
- no upstream forwarding occurs

List the pending approvals:

```powershell
go run ./cmd/agentfence-cli list-approvals --store data/demo-approvals.json
```

Approve or deny by ID:

```powershell
go run ./cmd/agentfence-cli approve <approval-id> --store data/demo-approvals.json --actor reviewer --reason "demo approved"
go run ./cmd/agentfence-cli deny <approval-id> --store data/demo-approvals.json --actor reviewer --reason "demo denied"
```

Current v0.1 constraint: approval resolution is durable and auditable, but the original MCP request is not replayed automatically after approval. Re-submit the request manually if you want to test the full operator loop.

## 3. Dangerous delete call

```powershell
curl -sS -X POST "http://127.0.0.1:8080/mcp?server=github-demo" ^
  -H "Content-Type: application/json" ^
  --data-binary "@examples/github-mcp/requests/delete-repo.json"
```

Expected result:

- HTTP `403 Forbidden`
- JSON-RPC error with code `-32001`
- reason shows that repository deletion is denied
- no upstream forwarding occurs

## Inspect audit and approval state

- Admin UI dashboard: `http://127.0.0.1:3000`
- Audit API: `http://127.0.0.1:8080/api/admin/audit`
- Pending approvals API: `http://127.0.0.1:8080/api/admin/approvals`
- Policy status API: `http://127.0.0.1:8080/api/admin/policy`

If you run the Next.js admin UI, set `AGENTFENCE_API_BASE=http://127.0.0.1:8080` before starting the web app.