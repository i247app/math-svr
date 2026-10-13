package prompt

import (
	domain "math-ai.com/math-ai/internal/domain/prompt"
)

// ListPromptsReq narrows /prompts/list; every field is optional.
type ListPromptsReq struct {
	PromptCode *string `json:"prompt_code,omitempty"`
	Grade      *int    `json:"grade,omitempty"`
	Language   *string `json:"language,omitempty"`
}

// GetPromptReq reads one prompt row.
type GetPromptReq struct {
	PromptID int64 `json:"prompt_id"`
}

// UpdatePromptReq replaces the text of one prompt row. The text is stored
// and sent to the model verbatim; nothing in it is filled in.
type UpdatePromptReq struct {
	PromptID     int64  `json:"prompt_id"`
	SystemPrompt string `json:"system_prompt"`
}

// PromptResponse is the static half of one kind of LLM call. grade and
// language are null when the prompt_code does not vary by them.
// prompt_version grows by one per rewrite; for exam generation it is part
// of the cache tag, so sets generated before a rewrite are not served after.
type PromptResponse struct {
	PromptID      int64   `json:"prompt_id"`
	PromptCode    string  `json:"prompt_code"`
	Grade         *int    `json:"grade"`
	Language      *string `json:"language"`
	Description   *string `json:"description"`
	SystemPrompt  string  `json:"system_prompt"`
	PromptVersion int     `json:"prompt_version"`
	ModifyID      *int64  `json:"modify_id,omitempty"`
	ModifyDt      string  `json:"modify_dt"`
}

type PromptRes struct {
	Prompt *PromptResponse `json:"prompt"`
}

type ListPromptsRes struct {
	Prompts []*PromptResponse `json:"prompts"`
}

func PromptToResponse(p *domain.Prompt) *PromptResponse {
	if p == nil {
		return nil
	}
	return &PromptResponse{
		PromptID:      p.PromptId(),
		PromptCode:    string(p.PromptCode()),
		Grade:         p.Grade(),
		Language:      p.Language(),
		Description:   p.Description(),
		SystemPrompt:  p.SystemPrompt(),
		PromptVersion: p.PromptVersion(),
		ModifyID:      p.ModifyId(),
		ModifyDt:      p.ModifyDt().String(),
	}
}

func PromptsToResponse(ps []*domain.Prompt) []*PromptResponse {
	out := make([]*PromptResponse, 0, len(ps))
	for _, p := range ps {
		out = append(out, PromptToResponse(p))
	}
	return out
}
