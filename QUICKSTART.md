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
2. Install web dependencies
cd web
npm install
cd ..
3. Create local working directories
mkdir -p .local
mkdir -p .local/data

On Windows PowerShell:

New-Item -ItemType Directory -Force .local | Out-Null
New-Item -ItemType Directory -Force .local\data | Out-Null
4. Start the mock GitHub MCP server

This repo includes a demo upstream MCP server for local evaluation.

go run ./cmd/mock-github-mcp

By default, keep it running in a separate terminal.

Example upstream URL used below:

http://localhost:8081/mcp

If your mock server runs on a different port, adjust the config values accordingly.

5. Create a local config file

Create agentfence.dev.json in the repository root:

{
  "listen_addr": ":8080",
  "upstream_url": "http://localhost:8081/mcp",
  "upstream_timeout": "10s",
  "approval_store": ".local/data/approvals.json",
  "audit_store": ".local/data/audit.json",
  "policy_file": "policies/examples/github-readonly.yaml"
}
6. Set environment variables

Linux/macOS:

export AGENTFENCE_CONFIG_FILE=agentfence.dev.json
export AGENTFENCE_API_BASE=http://localhost:8080

Windows PowerShell:

$env:AGENTFENCE_CONFIG_FILE="agentfence.dev.json"
$env:AGENTFENCE_API_BASE="http://localhost:8080"
7. Start the gateway
go run ./cmd/agentfence

The gateway should now be listening on:

http://localhost:8080

Useful endpoints:

health: GET /healthz

MCP entrypoint: POST /mcp

approvals admin API

audit admin API

8. Start the admin UI

In a separate terminal:

cd web
npm run dev

The UI should now be available at:

http://localhost:3000
9. Run the operator CLI

You can inspect or resolve approval requests with the CLI.

Examples:

go run ./cmd/agentfence-cli --help
go run ./cmd/agentfence-cli approvals list
go run ./cmd/agentfence-cli approvals approve <approval-id>
go run ./cmd/agentfence-cli approvals deny <approval-id>

Adjust flags or environment variables as required by the current CLI implementation.

10. Try the GitHub MCP demo

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

policies/examples/github-readonly.yaml

You can later switch to other example policies or create your own YAML file.

Local validation

From the repo root:

go test ./...

From the web directory:

npm run build
Common problems
Build fails because of missing Go dependencies

Run:

go mod tidy
go test ./...
Web build fails because of missing npm dependencies

Run:

cd web
npm install
npm run build
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