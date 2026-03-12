# Review Checklist

## Security
- Did this weaken deny-by-default behavior?
- Could secrets leak into logs, traces, or errors?
- Could approval-required calls slip through?
- Are trust boundaries still clear?

## Correctness
- Are edge cases handled?
- Does malformed input fail safely?
- Are upstream failures surfaced correctly?

## Architecture
- Is logic separated cleanly?
- Is the change minimal and understandable?
- Are abstractions justified?

## Testing
- Did this change add/update the right tests?
- Do tests verify the real invariant instead of only the happy path?

## Operability
- Are logs structured?
- Are failures debuggable?
- Is audit data sufficient to explain what happened?