package prompt

import "context"

// ListFilter narrows /prompts/list; nil fields are not filtered on.
type ListFilter struct {
	Code     *Code
	Grade    *int
	Language *string
}

// IRepository owns ma_prompts. There is no Create or Delete: rows are
// seeded, and an admin can only rewrite one in place.
type IRepository interface {
	// FindByScope is the read every LLM call makes: an exact match on code,
	// grade and language (NULL matching NULL).
	FindByScope(ctx context.Context, scope Scope) (*Prompt, error)
	FindByPromptId(ctx context.Context, promptId int64) (*Prompt, error)
	// List orders by code, then grade, then language.
	List(ctx context.Context, filter ListFilter) ([]*Prompt, error)
	// UpdatePrompt overwrites system_prompt and adds 1 to prompt_version in
	// one statement, so two concurrent rewrites never share a version
	// number. It reports false when no live row has that id.
	UpdatePrompt(ctx context.Context, promptId int64, systemPrompt string, modifyId *int64) (bool, error)
}
