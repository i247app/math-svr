package command

import (
	"context"
	"fmt"

	"math-ai.com/math-ai/internal/application/transaction"
	"math-ai.com/math-ai/internal/domain/prompt"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
)

// UpdatePromptCommand overwrites one prompt row and moves its
// prompt_version on by one. There is no history: the previous text is gone
// once this commits.
type UpdatePromptCommand struct {
	PromptID     int64
	SystemPrompt string
	AdminUID     *int64 // nil = no uid (API-key admin)
}

type UpdatePromptCommandHandler struct {
	uow transaction.UnitOfWork
}

func NewUpdatePromptCommandHandler(uow transaction.UnitOfWork) *UpdatePromptCommandHandler {
	return &UpdatePromptCommandHandler{uow: uow}
}

// Handle writes first and reads back after, in the same transaction: the
// UPDATE itself reports whether the row exists, and the read returns the
// version that this write produced.
func (h *UpdatePromptCommandHandler) Handle(ctx context.Context, cmd UpdatePromptCommand) (*prompt.Prompt, error) {
	var saved *prompt.Prompt
	err := h.uow.Do(ctx, func(ctx context.Context, repos transaction.Repositories) error {
		updated, err := repos.Prompt.UpdatePrompt(ctx, cmd.PromptID, cmd.SystemPrompt, cmd.AdminUID)
		if err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		if !updated {
			return errs.NewError(ctx, status.PROMPT_NOT_FOUND, nil,
				fmt.Errorf("prompt: ma_prompts has no row %d", cmd.PromptID))
		}
		if saved, err = repos.Prompt.FindByPromptId(ctx, cmd.PromptID); err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return saved, nil
}
