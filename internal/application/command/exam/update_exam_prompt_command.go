package command

import (
	"context"
	"fmt"

	"math-ai.com/math-ai/internal/application/transaction"
	"math-ai.com/math-ai/internal/domain/exam"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
)

// UpdateExamPromptCommand overwrites one grade's system prompt and moves
// its prompt_version on by one. There is no history: the previous text is
// gone once this commits.
type UpdateExamPromptCommand struct {
	Grade        int
	SystemPrompt string
	AdminUID     *int64 // nil = no uid (API-key admin)
}

type UpdateExamPromptCommandHandler struct {
	uow transaction.UnitOfWork
}

func NewUpdateExamPromptCommandHandler(uow transaction.UnitOfWork) *UpdateExamPromptCommandHandler {
	return &UpdateExamPromptCommandHandler{uow: uow}
}

// Handle writes first and reads back after, in the same transaction: the
// UPDATE itself reports whether the grade exists, and the read returns the
// version that this write produced.
func (h *UpdateExamPromptCommandHandler) Handle(ctx context.Context, cmd UpdateExamPromptCommand) (*exam.ExamPrompt, error) {
	var saved *exam.ExamPrompt
	err := h.uow.Do(ctx, func(ctx context.Context, repos transaction.Repositories) error {
		updated, err := repos.ExamPrompt.UpdatePrompt(ctx, cmd.Grade, cmd.SystemPrompt, cmd.AdminUID)
		if err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		if !updated {
			return errs.NewError(ctx, status.EXAM_PROMPT_NOT_FOUND, nil,
				fmt.Errorf("exam: ma_exam_prompts has no row for grade %d", cmd.Grade))
		}
		if saved, err = repos.ExamPrompt.FindByGrade(ctx, cmd.Grade); err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return saved, nil
}
