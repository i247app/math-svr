package chatbox_helper

import "regexp"

// JSONSanitizer handles sanitization of JSON responses from AI providers
type JSONSanitizer struct{}

// NewJSONSanitizer creates a new JSONSanitizer instance
func NewJSONSanitizer() *JSONSanitizer {
	return &JSONSanitizer{}
}

// SanitizeJSONResponse fixes common JSON escaping issues in AI responses,
// particularly for LaTeX expressions where backslashes need to be escaped.
func (s *JSONSanitizer) SanitizeJSONResponse(jsonStr string) string {
	// Pattern to find backslashes in JSON string values that are not already escaped
	// This regex looks for single backslashes followed by common LaTeX commands
	patterns := []struct {
		pattern string
		replace string
	}{
		// Fix unescaped backslashes before common LaTeX commands
		{`([^\\])\\frac`, `$1\\frac`},
		{`([^\\])\\sqrt`, `$1\\sqrt`},
		{`([^\\])\\int`, `$1\\int`},
		{`([^\\])\\{`, `$1\\{`},
		{`([^\\])\\}`, `$1\\}`},
		{`([^\\])\\left`, `$1\\left`},
		{`([^\\])\\right`, `$1\\right`},
		// Fix at start of string values (after quotes)
		{`"\\frac`, `"\\\\frac`},
		{`"\\sqrt`, `"\\\\sqrt`},
		{`"\\int`, `"\\\\int`},
		{`"\\{`, `"\\\\{`},
		{`"\\}`, `"\\\\}`},
		{`"\\left`, `"\\\\left`},
		{`"\\right`, `"\\\\right`},
	}

	result := jsonStr
	for _, p := range patterns {
		re := regexp.MustCompile(p.pattern)
		result = re.ReplaceAllString(result, p.replace)
	}

	return result
}
