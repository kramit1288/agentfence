package audit

import (
	"regexp"
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

var inlineSecretPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)\b(token)\s*=\s*([^\s,;]+)`),
	regexp.MustCompile(`(?i)\b(secret)\s*=\s*([^\s,;]+)`),
	regexp.MustCompile(`(?i)\b(password)\s*=\s*([^\s,;]+)`),
	regexp.MustCompile(`(?i)\b(api_?key)\s*=\s*([^\s,;]+)`),
	regexp.MustCompile(`(?i)\b(authorization)\s*=\s*([^\s,;]+(?:\s+[^\s,;]+)?)`),
}

var urlUserInfoPattern = regexp.MustCompile(`(?i)(https?://)([^/@\s:]+:[^/@\s]+@)`)

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

// RedactText masks secret-looking key=value fragments in free-form text.
func RedactText(input string) string {
	output := urlUserInfoPattern.ReplaceAllString(input, `$1`+RedactedValue+`@`)
	for _, pattern := range inlineSecretPatterns {
		output = pattern.ReplaceAllString(output, `$1=`+RedactedValue)
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