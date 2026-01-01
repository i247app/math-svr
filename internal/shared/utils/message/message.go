package message

import (
	"fmt"
	"strings"

	"math-ai.com/math-ai/internal/shared/constant/status"
)

// GetLocalizedMessage returns a localized message for the given status code and language
// If args are provided and a template exists, it will interpolate the args into the template
// Otherwise, it falls back to the static message
func GetMessage(language LanguageType, statusCode status.Code, args map[string]interface{}) string {
	var template string
	var staticMessage string

	// Get template and static message based on language
	switch language {
	case EN:
		template = GetMessageTemplateEN(statusCode)
		staticMessage = GetMessageENFromStatus(statusCode)
	case VN:
		template = GetMessageTemplateVN(statusCode)
		staticMessage = GetMessageVNFromStatus(statusCode)
	default:
		template = GetMessageTemplateEN(statusCode)
		staticMessage = GetMessageENFromStatus(statusCode)
	}

	// If we have args and a template exists, use the template
	if len(args) > 0 && template != "" {
		return InterpolateMessage(template, args)
	}

	// Otherwise, use static message
	return staticMessage
}

// InterpolateMessage replaces placeholders in a message template with actual values
// Supports placeholders in the format: {key}
// Example: "User with email {email} already exists" + {"email": "john@example.com"}
//
//	-> "User with email john@example.com already exists"
func InterpolateMessage(template string, args map[string]interface{}) string {
	if len(args) == 0 {
		return template
	}

	result := template
	for key, value := range args {
		placeholder := fmt.Sprintf("{%s}", key)
		replacement := fmt.Sprintf("%v", value)
		result = strings.ReplaceAll(result, placeholder, replacement)
	}

	return result
}
