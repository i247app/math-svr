package command

import (
	"context"
	"fmt"

	"math-ai.com/math-ai/internal/application/transaction"
	"math-ai.com/math-ai/internal/domain/questionpool"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
)

// ReviewQuestionCommand changes one question's review fields, never its
// content or hash: Status (when set) becomes question_status, Verify adds 1
// to verified_count.
type ReviewQuestionPoolCommand struct {
	QuestionPoolID int64
	ActorUID       *int64 // modify_id; nil = no uid (API-key admin)
	Status         *string
	Verify         bool
}

type ReviewQuestionPoolCommandHandler struct {
	uow transaction.UnitOfWork
}

func NewReviewQuestionPoolCommandHandler(uow transaction.UnitOfWork) *ReviewQuestionPoolCommandHandler {
	return &ReviewQuestionPoolCommandHandler{uow: uow}
}

// Handle locks the row first (unknown id → QUESTION_POOL_NOT_FOUND), writes, and
// returns the row read back in the same transaction.
func (h *ReviewQuestionPoolCommandHandler) Handle(ctx context.Context, cmd ReviewQuestionPoolCommand) (*questionpool.QuestionPool, error) {
	var saved *questionpool.QuestionPool
	err := h.uow.Do(ctx, func(ctx context.Context, repos transaction.Repositories) error {
		stored, err := repos.QuestionPool.LockById(ctx, cmd.QuestionPoolID)
		if err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		if stored == nil {
			return errs.NewError(ctx, status.QUESTION_POOL_NOT_FOUND, nil,
				fmt.Errorf("question_pool: no live question %d", cmd.QuestionPoolID))
		}
		if cmd.Status != nil {
			if err := repos.QuestionPool.MarkStatus(ctx, cmd.QuestionPoolID, *cmd.Status, cmd.ActorUID); err != nil {
				return errs.NewError(ctx, status.FAIL, nil, err)
			}
		}
		if cmd.Verify {
			if err := repos.QuestionPool.IncVerified(ctx, cmd.QuestionPoolID, cmd.ActorUID); err != nil {
				return errs.NewError(ctx, status.FAIL, nil, err)
			}
		}
		if saved, err = repos.QuestionPool.LockById(ctx, cmd.QuestionPoolID); err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		if saved == nil {
			return errs.NewError(ctx, status.FAIL, nil, fmt.Errorf("question_pool: reviewed %d but cannot read it back", cmd.QuestionPoolID))
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return saved, nil
}
