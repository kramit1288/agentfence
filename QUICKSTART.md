# Quickstart

This quickstart gives you a working local AgentFence setup in a few minutes.

## Prerequisites

- Go 1.22+
- Node.js 20+
- npm 10+
- optional: Docker if you want Postgres instead of local JSON storage

## 1. Start a mock upstream MCP server

```powershell
go run ./cmd/mock-github-mcp -addr :8090
```

## 2. Start the AgentFence gateway

Use the included GitHub demo policy and local JSON-backed audit and approval storage:

```powershell
$env:AGENTFENCE_POLICY_FILE = "examples/github-mcp/policy.yaml"
$env:AGENTFENCE_UPSTREAM_URL = "http://127.0.0.1:8090"
$env:AGENTFENCE_APPROVAL_STORE = "data/demo-approvals.json"
$env:AGENTFENCE_AUDIT_STORE = "data/demo-audit.json"
go run ./cmd/agentfence -config examples/github-mcp/config.json
```

The gateway listens on `http://127.0.0.1:8080` by default.

## 3. Exercise the request path

Allowed read-only request:

```powershell
curl -sS -X POST "http://127.0.0.1:8080/mcp?server=github-demo" ^
  -H "Content-Type: application/json" ^
  --data-binary "@examples/github-mcp/requests/read-repo.json"
```

Approval-required write request:

```powershell
curl -sS -X POST "http://127.0.0.1:8080/mcp?server=github-demo" ^
  -H "Content-Type: application/json" ^
  -H "X-AgentFence-Actor: demo-user" ^
  --data-binary "@examples/github-mcp/requests/merge-pr.json"
```

Denied dangerous request:

```powershell
curl -sS -X POST "http://127.0.0.1:8080/mcp?server=github-demo" ^
  -H "Content-Type: application/json" ^
  --data-binary "@examples/github-mcp/requests/delete-repo.json"
```

## 4. Inspect approvals and audit data

List pending approvals:

```powershell
go run ./cmd/agentfence-cli list-approvals --store data/demo-approvals.json
```

Approve a request:

```powershell
go run ./cmd/agentfence-cli approve <approval-id> --store data/demo-approvals.json --actor reviewer --reason "approved for testing"
```

Admin API endpoints:

- `GET /healthz`
- `GET /api/admin/audit`
- `GET /api/admin/approvals`
- `GET /api/admin/policy`

## 5. Start the admin UI

```powershell
cd web
npm install
$env:AGENTFENCE_API_BASE = "http://127.0.0.1:8080"
npm run dev
```

Open `http://127.0.0.1:3000`.

## Optional: use Postgres instead of local JSON files

Start Postgres:

```powershell
docker compose -f deploy/docker-compose.yml up postgres
```

Then set:

```powershell
$env:AGENTFENCE_POSTGRES_DSN = "postgres://agentfence:agentfence@127.0.0.1:5432/agentfence?sslmode=disable"
```

The gateway runs migrations automatically on startup.

## What this quickstart does not cover yet

- authenticated admin access
- automatic replay after approval
- stdio upstream transport
- multi-upstream routing
