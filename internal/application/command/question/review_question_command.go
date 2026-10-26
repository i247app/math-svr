package command

import (
	"context"
	"fmt"

	"math-ai.com/math-ai/internal/application/transaction"
	"math-ai.com/math-ai/internal/domain/question"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
)

// ReviewQuestionCommand changes one question's review fields, never its
// content or hash: Status (when set) becomes question_status, Verify adds 1
// to verified_count.
type ReviewQuestionCommand struct {
	QuestionID int64
	ActorUID   *int64 // modify_id; nil = no uid (API-key admin)
	Status     *string
	Verify     bool
}

type ReviewQuestionCommandHandler struct {
	uow transaction.UnitOfWork
}

func NewReviewQuestionCommandHandler(uow transaction.UnitOfWork) *ReviewQuestionCommandHandler {
	return &ReviewQuestionCommandHandler{uow: uow}
}

// Handle locks the row first (unknown id → QUESTION_NOT_FOUND), writes, and
// returns the row read back in the same transaction.
func (h *ReviewQuestionCommandHandler) Handle(ctx context.Context, cmd ReviewQuestionCommand) (*question.Question, error) {
	var saved *question.Question
	err := h.uow.Do(ctx, func(ctx context.Context, repos transaction.Repositories) error {
		stored, err := repos.Question.LockById(ctx, cmd.QuestionID)
		if err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		if stored == nil {
			return errs.NewError(ctx, status.QUESTION_NOT_FOUND, nil,
				fmt.Errorf("question: no live question %d", cmd.QuestionID))
		}
		if cmd.Status != nil {
			if err := repos.Question.MarkStatus(ctx, cmd.QuestionID, *cmd.Status, cmd.ActorUID); err != nil {
				return errs.NewError(ctx, status.FAIL, nil, err)
			}
		}
		if cmd.Verify {
			if err := repos.Question.IncVerified(ctx, cmd.QuestionID, cmd.ActorUID); err != nil {
				return errs.NewError(ctx, status.FAIL, nil, err)
			}
		}
		if saved, err = repos.Question.LockById(ctx, cmd.QuestionID); err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		if saved == nil {
			return errs.NewError(ctx, status.FAIL, nil, fmt.Errorf("question: reviewed %d but cannot read it back", cmd.QuestionID))
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return saved, nil
}
