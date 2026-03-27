# Quickstart

This guide gets AgentFence running locally in the shortest practical path.

AgentFence is currently an early OSS technical preview. The local flow is suitable for evaluation and staging-style demos.

## What you will run

You will start:

- the AgentFence gateway
- the optional admin UI
- the mock GitHub MCP demo server

The quickest local setup uses:

- local JSON-file audit storage
- local JSON-file approval storage
- one upstream HTTP MCP server
- a YAML policy file

## Prerequisites

- Go 1.22+
- Node.js 20+
- npm 10+
- optional: Docker if you want Postgres later

## 1. Clone and enter the repo

```bash
git clone https://github.com/<your-username>/agentfence.git
cd agentfence
```

## 2. Install web dependencies

```bash
cd web
npm install
cd ..
```

## 3. Create local working directories

Create `.local/data` if it does not already exist.

```bash
mkdir -p .local
mkdir -p .local/data
```

On Windows PowerShell:

```powershell
New-Item -ItemType Directory -Force .local | Out-Null
New-Item -ItemType Directory -Force .local\data | Out-Null
```

## 4. Start the mock GitHub MCP server

This repo includes a demo upstream MCP server for local evaluation.

```bash
go run ./cmd/mock-github-mcp
```

By default, keep it running in a separate terminal.

Example upstream URL used below:

http://localhost:8081/mcp

If your mock server runs on a different port, adjust the config values accordingly.

## 5. Create a local config file

Copy the checked-in example file in the repository root:

```bash
cp agentfence.dev.example.json agentfence.dev.json
```

On Windows PowerShell:

```powershell
Copy-Item agentfence.dev.example.json agentfence.dev.json
```

The example file uses the current local development shape:

```json
{
  "listen_addr": ":8080",
  "upstream_url": "http://localhost:8081/mcp",
  "upstream_timeout": "10s",
  "approval_store": ".local/data/approvals.json",
  "audit_store": ".local/data/audit.json",
  "policy_file": "policies/examples/github-production.yaml"
}
```

Adjust the copied file if your local upstream URL, ports, or file paths differ.

## 6. Set environment variables

The gateway currently reads the upstream, policy, audit, and approval settings from environment variables. Keep `agentfence.dev.json` as your local reference file and export matching values before you start the gateway.

Linux/macOS:

```bash
export AGENTFENCE_HTTP_ADDRESS=:8080
export AGENTFENCE_UPSTREAM_URL=http://localhost:8081/mcp
export AGENTFENCE_UPSTREAM_TIMEOUT=10s
export AGENTFENCE_APPROVAL_STORE=.local/data/approvals.json
export AGENTFENCE_AUDIT_STORE=.local/data/audit.json
export AGENTFENCE_POLICY_FILE=policies/examples/github-production.yaml
export AGENTFENCE_API_BASE=http://localhost:8080
```

Windows PowerShell:

```powershell
$env:AGENTFENCE_HTTP_ADDRESS=":8080"
$env:AGENTFENCE_UPSTREAM_URL="http://localhost:8081/mcp"
$env:AGENTFENCE_UPSTREAM_TIMEOUT="10s"
$env:AGENTFENCE_APPROVAL_STORE=".local/data/approvals.json"
$env:AGENTFENCE_AUDIT_STORE=".local/data/audit.json"
$env:AGENTFENCE_POLICY_FILE="policies/examples/github-production.yaml"
$env:AGENTFENCE_API_BASE="http://localhost:8080"
```

## 7. Start the gateway

```bash
go run ./cmd/agentfence
```

The gateway should now be listening on:

http://localhost:8080

Useful endpoints:

health: GET /healthz

MCP entrypoint: POST /mcp

approvals admin API

audit admin API

## 8. Start the admin UI

In a separate terminal:

```bash
cd web
npm run dev
```

The UI should now be available at:

http://localhost:3000

## 9. Run the operator CLI

You can inspect or resolve approval requests with the CLI.

Examples:

```bash
go run ./cmd/agentfence-cli --help
go run ./cmd/agentfence-cli list-approvals --store .local/data/approvals.json
go run ./cmd/agentfence-cli approve --store .local/data/approvals.json --actor alice <approval-id>
go run ./cmd/agentfence-cli deny --store .local/data/approvals.json --actor alice <approval-id>
```

Adjust flags or environment variables as required by the current CLI implementation.

## 10. Try the GitHub MCP demo

Use the demo described in docs/demos/github-mcp.md
.

Typical flow:

start the mock upstream MCP server

start AgentFence

send a safe MCP tool call

confirm it is allowed and audited

send a risky MCP tool call

confirm it is denied or moved into approval flow

inspect audit events and pending approvals

Example policy

For first local testing, use the example GitHub policy in:

policies/examples/github-production.yaml

You can later switch to other example policies or create your own YAML file.

Local validation

From the repo root:

```bash
go test ./...
```

From the web directory:

```bash
npm run build
```

Common problems

Build fails because of missing Go dependencies

Run:

```bash
go mod tidy
go test ./...
```

Web build fails because of missing npm dependencies

Run:

```bash
cd web
npm install
npm run build
```

No approvals are appearing

Check:

the policy actually uses require_approval

the gateway is using the expected policy file

the approval store path is writable

the tool call is reaching the gateway, not the upstream directly

Audit log is empty

Check:

the request is sent to POST /mcp

the audit store path is writable

the request was not rejected before request processing began

What this quickstart does not cover yet

This quickstart does not cover:

authentication for admin APIs or UI

multi-upstream routing

stdio transport

production deployment hardening

automatic replay of approved requests

Next reading

README.md

docs/architecture.md

docs/threat-model.md

docs/demos/github-mcp.md
