package audit

import (
	"strings"
	"unicode"
)

const RedactedValue = "[REDACTED]"

var sensitiveTokens = []string{
	"token",
	"secret",
	"password",
	"apikey",
	"authorization",
}

// RedactValue returns a copy of value with sensitive fields masked while
// preserving the original structure where practical.
func RedactValue(value any) any {
	switch v := value.(type) {
	case map[string]any:
		return RedactMap(v)
	case []any:
		items := make([]any, len(v))
		for i := range v {
			items[i] = RedactValue(v[i])
		}
		return items
	case []string:
		items := make([]any, len(v))
		for i := range v {
			items[i] = v[i]
		}
		return items
	default:
		return value
	}
}

// RedactMap returns a deep redacted copy of a generic map.
func RedactMap(input map[string]any) map[string]any {
	if input == nil {
		return nil
	}

	output := make(map[string]any, len(input))
	for key, value := range input {
		if isSensitiveKey(key) {
			output[key] = RedactedValue
			continue
		}
		output[key] = RedactValue(value)
	}
	return output
}

func isSensitiveKey(key string) bool {
	normalized := normalizeKey(key)
	for _, token := range sensitiveTokens {
		if strings.Contains(normalized, token) {
			return true
		}
	}
	return false
}

func normalizeKey(key string) string {
	var builder strings.Builder
	builder.Grow(len(key))
	for _, r := range key {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			builder.WriteRune(unicode.ToLower(r))
		}
	}
	return builder.String()
}
