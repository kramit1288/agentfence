package policy

// RuleCount returns the number of compiled rules in the engine.
func (e *Engine) RuleCount() int {
	if e == nil {
		return 0
	}
	return len(e.rules)
}
