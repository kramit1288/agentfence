# Threat Model

## Status

This threat model describes the current AgentFence technical preview and the risks it is explicitly trying to reduce.

It is not a claim that all security problems are solved.

AgentFence should currently be treated as an early-stage gateway for evaluation, local development, and staging-style demos.

## Security posture goals

AgentFence exists to improve the safety and governability of MCP tool access by introducing:

- policy checks before tool calls are forwarded
- approval workflows for risky actions
- audit visibility
- redaction of likely sensitive fields
- clearer trust boundaries

## Primary assets

The most important things to protect are:

- tool execution authority
- credentials and secrets in requests or responses
- audit integrity
- approval integrity
- operator trust in recorded history
- upstream system safety
- internal data reachable through MCP tools

## Trust boundaries

AgentFence treats these boundaries seriously:

### 1. Agent runtime -> AgentFence

The agent runtime may be buggy, misconfigured, or influenced by adversarial prompts.

AgentFence should not assume the caller is always safe or correct.

### 2. AgentFence -> upstream MCP server

The upstream MCP server should be treated as potentially untrusted or at least not fully controlled.

Why:

- server behavior can change
- tool metadata can drift
- responses can contain dangerous or sensitive content
- upstream failures may be malformed or inconsistent

### 3. Operator/admin surfaces

Admin APIs and the admin UI expose operationally sensitive state such as:

- pending approvals
- recent audit history
- policy information

In the current technical preview, these surfaces are not yet authenticated. This is a known limitation and a major reason the project should not yet be treated as a complete production control plane.

## Threat actors

Potential threat actors include:

- a malicious or compromised upstream MCP server
- a prompt-injected or misdirected agent
- an operator making unsafe decisions
- a developer misconfiguring policy
- a user sending malformed MCP requests
- an attacker with access to local storage or admin endpoints
- a future multi-tenant misuse scenario

## Main threat categories

## 1. Dangerous tool execution

### Risk

An agent attempts to invoke a dangerous tool or dangerous tool arguments reach an upstream server.

### Examples

- destructive repo operation
- write or merge action
- sensitive message posting
- dangerous database query

### Current mitigations

- policy evaluation before forwarding
- deny behavior
- require_approval behavior
- denied and approval-required requests do not forward upstream

### Remaining gaps

- policy quality depends on configuration correctness
- there is not yet a richer policy language for deeper semantic controls
- approved requests do not automatically replay today

## 2. Secret leakage in audit trails

### Risk

Sensitive values appear in logs or stored audit data.

### Examples

- API keys
- tokens
- passwords
- authorization headers
- secret-like argument values

### Current mitigations

- redaction of likely sensitive keys
- structured audit event handling instead of naive raw-body logging

### Remaining gaps

- pattern coverage is not perfect
- future richer content types may require stronger redaction logic
- operators should still treat stored data carefully

## 3. Malformed or adversarial MCP requests

### Risk

Malformed JSON-RPC or weird input causes unsafe behavior, panics, or accidental allow paths.

### Current mitigations

- explicit protocol parsing
- narrow current scope
- deterministic policy actions
- fail-closed behavior where practical

### Remaining gaps

- wider MCP surface is not yet implemented
- broader fuzzing and adversarial testing can improve confidence

## 4. Malicious or compromised upstream MCP server

### Risk

The upstream returns dangerous, misleading, or malformed responses, or changes behavior unexpectedly.

### Examples

- tool schema drift
- unexpected response structure
- dangerous content in responses
- unstable or inconsistent behavior

### Current mitigations

- clear trust-boundary framing
- upstream sits behind AgentFence, not directly behind the agent
- audit visibility around decisions and outcomes

### Remaining gaps

- upstream trust is still only partially controlled
- response validation can be strengthened
- richer policy by tool/server identity is still early

## 5. Approval abuse or integrity loss

### Risk

Approval records are created, modified, or resolved incorrectly.

### Examples

- unsafe approval resolution
- missing audit linkage
- inconsistent state between approval history and audit history
- concurrent resolution issues

### Current mitigations

- durable approval records
- operator CLI resolution path
- Postgres-backed repositories available for stronger durability

### Remaining gaps

- atomicity between approval state and audit history can be improved
- no authentication yet on admin/operator surfaces
- replay after approval is not automatic

## 6. Admin surface exposure

### Risk

Unauthenticated admin endpoints or UI pages reveal sensitive operational data.

### Current mitigations

- currently mostly controlled by deployment context, not by built-in auth

### Remaining gaps

- this is a real current risk
- the admin UI and APIs should not be exposed broadly on the internet in the current preview
- authentication and authorization are future priorities

## 7. Local storage compromise

### Risk

JSON-file approval and audit stores are tampered with or read by unauthorized local users.

### Current mitigations

- local file mode is positioned for development and demos, not broad production use
- Postgres-backed repositories are available for stronger durability

### Remaining gaps

- file integrity and access control depend on host environment
- transactional safety is limited in local file mode

## Security assumptions

Current assumptions include:

- local development users control their own machine
- demo deployments are in limited-trust environments
- operators understand the project is a technical preview
- admins do not expose unauthenticated endpoints publicly
- policy files are reviewed before use

## Non-goals in the current preview

The current technical preview does not yet claim to provide:

- a hardened internet-exposed admin surface
- full enterprise authn/authz
- complete replay/resume semantics for approvals
- complete MCP transport coverage
- complete multi-tenant isolation
- complete protection from prompt injection or malicious upstream behavior

## Operational guidance

For current users:

- do not expose admin APIs/UI publicly without additional protection
- treat upstream MCP servers as potentially untrusted
- review policy files carefully
- prefer Postgres-backed storage over local file mode for stronger durability
- treat audit data as sensitive operational information
- use the project for evaluation and controlled demos, not as a finished control plane

## Priority future security work

The most important next security steps are:

1. add authentication and authorization for admin APIs and UI
2. strengthen the relationship between approval state and audit history
3. improve upstream trust controls and routing controls
4. expand adversarial testing and malformed-input testing
5. improve redaction coverage and response handling
6. support stronger deployment guidance and hardening

## Summary

AgentFence’s current security value comes from introducing a reviewable control point between agents and MCP tools.

Its strongest current protections are:

- deny / approval gating before forwarding
- non-forwarding of blocked requests
- redacted audit records
- a clear trust-boundary model

Its biggest current gaps are:

- unauthenticated admin surfaces
- incomplete transport and routing support
- limited replay/resume behavior
- early-stage operational hardening