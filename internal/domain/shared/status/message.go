package status

import (
	"fmt"
	"strings"

	"math-ai.com/math-ai/internal/shared/enum"
)

type StatusMessage string

func (s StatusMessage) IsEmpty() bool {
	return string(s) == ""
}

func GetMessage(language enum.LanguageType, statusCode StatusCode, args map[string]any) StatusMessage {
	english := language == enum.LanguageTypeEnglish || language == enum.LanguageTypeEnglishV2

	// Anything but English is Vietnamese; an unregistered code has no
	// message ("") and the response falls back to the error's text.
	var message StatusMessage
	if m, ok := registered[statusCode]; ok {
		message = m[1]
		if english {
			message = m[0]
		}
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
