package prompt

import "errors"

var (
	ErrPromptIDRequired = errors.New("prompt_id is required")
	ErrPromptEmpty      = errors.New("system_prompt is required")
	ErrPromptTooLong    = errors.New("system_prompt is longer than 50000 characters")

	ErrAdminGateMissing = errors.New("prompt: route did not pass the admin gate")
)
