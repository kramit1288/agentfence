# GitHub MCP Demo

This demo shows how AgentFence behaves when placed in front of a GitHub-style MCP server.

The goal is to make three behaviors obvious:

- safe calls can be allowed
- risky calls can require approval
- dangerous calls can be denied

This demo uses the repository’s mock GitHub MCP server for local evaluation.

## What this demo proves

By the end of this demo you should be able to show:

- MCP requests are sent to AgentFence, not directly to the upstream
- policy decisions are applied before forwarding
- denied and approval-required requests do not reach the upstream
- approval requests appear in the operator workflow
- audit events are recorded with redaction

## Prerequisites

Before starting, complete the setup in [QUICKSTART.md](../../QUICKSTART.md).

You should have:

- the mock GitHub MCP server running
- the AgentFence gateway running
- the optional admin UI running
- a GitHub demo policy file available

## Example policy behavior

A typical GitHub demo policy should express ideas like:

- allow read-only repo inspection
- require approval for merge or write-like actions
- deny obviously dangerous administrative actions

For example:

- `list_pull_requests` -> allow
- `merge_pull_request` -> require_approval
- `delete_repository` -> deny

Use the existing policy examples in `policies/examples/` and adjust names to match the current mock server tool set if needed.

## Step 1: start the mock GitHub MCP server

From the repo root:

```bash
go run ./cmd/mock-github-mcp

Assume it is reachable at:

http://localhost:8081/mcp
Step 2: configure AgentFence to point to the mock server

Use a local config file such as:

{
  "listen_addr": ":8080",
  "upstream_url": "http://localhost:8081/mcp",
  "upstream_timeout": "10s",
  "approval_store": ".local/data/approvals.json",
  "audit_store": ".local/data/audit.json",
  "policy_file": "policies/examples/github-readonly.yaml"
}

Then start the gateway:

go run ./cmd/agentfence
Step 3: send a safe request

Send a read-style tool call through AgentFence.

Example shape:

{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "tools/call",
  "params": {
    "name": "list_pull_requests",
    "arguments": {
      "owner": "octo-org",
      "repo": "agentfence-demo"
    }
  }
}

POST this to:

http://localhost:8080/mcp

Expected result:

policy returns allow

request is forwarded upstream

response returns successfully

an audit event is recorded

Step 4: send a request that should require approval

Send a write-like tool call through AgentFence.

Example shape:

{
  "jsonrpc": "2.0",
  "id": 2,
  "method": "tools/call",
  "params": {
    "name": "merge_pull_request",
    "arguments": {
      "owner": "octo-org",
      "repo": "agentfence-demo",
      "pull_number": 42
    }
  }
}

Expected result:

policy returns require_approval

request is not forwarded upstream

a durable approval record is created

an audit event is recorded

the request appears in the approval workflow

Step 5: inspect pending approvals

Use the CLI:

go run ./cmd/agentfence-cli approvals list

Or use the admin UI or admin API to view pending approvals.

Expected result:

the pending approval is visible

the original request details are visible in a controlled form

the audit trail reflects the decision path

Step 6: approve or deny the request

Approve from the CLI:

go run ./cmd/agentfence-cli approvals approve <approval-id>

Or deny it:

go run ./cmd/agentfence-cli approvals deny <approval-id>

Current limitation:

approval resolution does not automatically replay the original request yet

So the demo should focus on:

durable workflow state

operator visibility

correct non-forwarding before approval

Step 7: send a denied request

Send a tool call that should be blocked entirely.

Example shape:

{
  "jsonrpc": "2.0",
  "id": 3,
  "method": "tools/call",
  "params": {
    "name": "delete_repository",
    "arguments": {
      "owner": "octo-org",
      "repo": "agentfence-demo"
    }
  }
}

Expected result:

policy returns deny

request is not forwarded upstream

an audit event is recorded

the caller receives a blocked response

Step 8: inspect audit history

Use the admin API or UI to inspect recent audit events.

You should be able to show:

allowed request path

approval-required request path

denied request path

redacted fields where applicable

What to emphasize in a demo

The strongest points to highlight are:

AgentFence is a control point, not just a proxy

not every tool call should reach the upstream

risky actions can be forced into a human review loop

audit history should be useful without exposing raw secrets

the current preview already proves the core workflow

Known limitations in this demo

mock server behavior is intentionally simplified

current upstream forwarding is HTTP-only

approval resolution does not auto-replay requests

current admin surfaces do not have built-in authentication

policy scope is intentionally narrow in this first preview

Suggested demo narrative

A strong short demo narrative is:

show a harmless read action passing through

show a sensitive write action getting stopped for approval

show a dangerous action getting denied

show the recorded audit and approval trail

explain that AgentFence creates a safer boundary between agents and tool execution