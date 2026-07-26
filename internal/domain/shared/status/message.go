package status

import (
	"fmt"
	"strings"

	"math-ai.com/math-ai/internal/shared/enum"
)

type StatusMessage string

func GetMessage(language enum.LanguageType, statusCode StatusCode, args map[string]any) StatusMessage {
	var message StatusMessage

	// Get template and static message based on language
	switch language {
	case enum.LanguageTypeVietnamese, enum.LanguageTypeVietnameseV2:
		message = GetVNMessage(statusCode)
	case enum.LanguageTypeEnglish, enum.LanguageTypeEnglishV2:
		message = GetENMessage(statusCode)
	default:
		message = GetVNMessage(statusCode)
	}

	// If we have args and a template exists, use the template
	if len(args) > 0 && message != "" {
		return InterpolateMessage(message, args)
	}

	// Otherwise, use static message
	return message
}

func InterpolateMessage(template StatusMessage, args map[string]any) StatusMessage {
	if len(args) == 0 {
		return template
	}

	result := template
	for key, value := range args {
		placeholder := fmt.Sprintf("{%s}", key)
		replacement := fmt.Sprintf("%v", value)
		result = StatusMessage(strings.ReplaceAll(string(result), placeholder, replacement))
	}

	return result
}
