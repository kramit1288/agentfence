# Manual Test Checklist

Run these manually for risky changes:

1. Allowed tool call succeeds
2. Denied tool call is blocked with clear reason
3. Approval-required call pauses and resumes after approval
4. Sensitive fields are redacted in logs and UI
5. Upstream MCP failure is handled and audited
6. Malformed request fails safely