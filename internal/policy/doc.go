// Package policy contains policy loading, compilation, and deterministic
// evaluation for tool access decisions.
//
// Precedence is intentionally simple: rules are evaluated in file order and the
// first matching rule wins. If no rule matches, evaluation denies by default.
package policy
