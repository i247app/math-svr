package models

type PromptModel struct {
	Audit

	PromptId      int64
	PromptCode    string
	Grade         *int
	Language      *string
	SystemPrompt  string
	PromptVersion int
	Description   *string
	PromptStatus  *string
}
