# AgentFence Spec

## Product definition

AgentFence is an open-source gateway for securing, governing, and auditing MCP tool access.

It sits between an agent runtime and upstream MCP servers and adds:

- deterministic policy evaluation
- approval workflows for risky actions
- redacted audit recording
- operator-facing visibility into decisions and pending actions

## Product intention

Most agent tooling focuses on connectivity or orchestration.

AgentFence focuses on the production control layer around tool access:

- what should be allowed
- what should be denied
- what should require human approval
- what should be recorded for later review

## Current release posture

This repository is an early OSS technical preview.

It is suitable for:

- local development
- staging-style demos
- early evaluation
- architectural exploration

It is not yet a complete production control plane.

## Problem statement

MCP gives agents a standardized way to access tools and external systems. That power creates operational and security problems:

- dangerous tools should not be callable by default
- sensitive writes may need human approval
- tool usage should be auditable
- secrets should not spill into logs
- upstream MCP servers should not be trusted blindly
- malformed requests should fail predictably

AgentFence exists to address those needs.

## Goals

The main goals are:

- intercept MCP tool access before forwarding
- enforce `allow`, `deny`, and `require_approval` decisions
- record audit events with redaction
- create a usable operator loop through API, UI, and CLI
- provide a simple local and demo-friendly path first
- establish a clean architecture that can grow into a stronger control plane

## Non-goals for the current preview

Current non-goals include:

- full enterprise authn/authz
- complete transport coverage
- full multi-upstream routing
- a full policy language for every possible tool behavior
- automatic replay after approval
- a finished production deployment story

## Supported behavior today

### MCP scope

Current typed support focuses on:

- `tools/list`
- `tools/call`

Other MCP methods may still exist as generic JSON-RPC envelopes, but the preview is intentionally centered on the core tool execution path.

### Policy actions

Current policy actions:

- `allow`
- `deny`
- `require_approval`

### Storage modes

Current storage modes:

- local JSON-file audit and approval storage
- Postgres-backed audit and approval repositories

### Upstream forwarding

Current upstream forwarding:

- one upstream HTTP MCP server
- allowed requests only

### Operator interfaces

Current operator interfaces:

- operator CLI for approval resolution
- admin APIs for approvals, audit events, and policy status
- minimal Next.js admin UI

## Functional requirements

## FR-1 MCP entrypoint

The system must expose an MCP-facing HTTP entrypoint that accepts JSON-RPC requests.

## FR-2 request parsing

The system must parse requests for the supported current MCP surface and reject malformed input safely.

## FR-3 policy evaluation

Before forwarding a tool call, the system must evaluate policy and return one of the supported actions.

## FR-4 no-forwarding for blocked requests

Denied and approval-required requests must not be forwarded upstream.

## FR-5 approval record creation

When policy returns `require_approval`, the system must create a durable approval record.

## FR-6 audit recording

The system must record audit events for significant request-path decisions and outcomes.

## FR-7 redaction

The system must redact likely secret-bearing fields before storing or displaying audit data.

## FR-8 upstream forwarding

Allowed requests must be able to proxy to the configured upstream HTTP MCP endpoint.

## FR-9 operator queryability

Operators must be able to inspect recent audit events and pending approvals.

## FR-10 local usability

The system must support a simple local mode without requiring Postgres.

## Quality requirements

## QR-1 deterministic behavior

The request path should be explicit, reviewable, and deterministic.

## QR-2 fail-safe behavior

Malformed or unsupported requests should fail safely and should not accidentally bypass policy.

## QR-3 separation of concerns

Transport, policy, audit, approval, config, and storage concerns should remain separated.

## QR-4 developer usability

Local development should remain simple enough to support fast iteration and demos.

## QR-5 honest posture

Docs and release messaging should remain honest about current limitations.

## Current limitations

The current technical preview still has important limits:

- admin APIs and UI do not have built-in authentication
- upstream routing is currently narrow
- stdio transport is not implemented
- automatic replay after approval is not implemented
- operational hardening is still early

## Future directions

Important likely next steps:

- authenticated admin and operator surfaces
- stronger upstream routing and multi-server support
- stdio transport support
- tighter approval/audit transactional behavior
- richer policy controls
- stronger observability and deployment hardening

## Summary

AgentFence already proves the core design:

- requests are intercepted
- policy is evaluated before forwarding
- risky actions can be moved into approval flow
- audit events are recorded with redaction
- operators can inspect and resolve workflow state

The current preview is intentionally narrow so the control layer remains understandable and testable.