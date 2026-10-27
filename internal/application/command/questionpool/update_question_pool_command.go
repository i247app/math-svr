package command

import (
	"context"
	"errors"
	"fmt"

	"math-ai.com/math-ai/internal/application/transaction"
	"math-ai.com/math-ai/internal/domain/questionpool"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
)

// UpdateQuestionCommand edits one question. Apply turns the stored row into
// the edited one — merging the request and checking the result — and
// returns a MathError when the edit is refused. It runs on the row read
// under lock, so it always merges onto the latest version.
type UpdateQuestionPoolCommand struct {
	QuestionPoolID int64
	ActorUID       *int64 // modify_id; nil = no uid (API-key admin)
	Apply          func(stored *questionpool.QuestionPool) (*questionpool.QuestionPool, error)
}

type UpdateQuestionPoolCommandHandler struct {
	uow transaction.UnitOfWork
}

func NewUpdateQuestionPoolCommandHandler(uow transaction.UnitOfWork) *UpdateQuestionPoolCommandHandler {
	return &UpdateQuestionPoolCommandHandler{uow: uow}
}

// Handle locks the row first, so two admins editing the same question are
// serialised and the second merges onto the first's result instead of
// overwriting it. A new content that another live question already holds
// is refused (uk_live_question_hash).
func (h *UpdateQuestionPoolCommandHandler) Handle(ctx context.Context, cmd UpdateQuestionPoolCommand) (*questionpool.QuestionPool, error) {
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
		edited, err := cmd.Apply(stored)
		if err != nil {
			return err
		}
		edited.SetQuestionPoolId(stored.QuestionPoolId())
		if err := repos.QuestionPool.UpdateContent(ctx, edited, cmd.ActorUID); err != nil {
			if errors.Is(err, questionpool.ErrDuplicate) {
				return errs.NewError(ctx, status.QUESTION_POOL_DUPLICATE, nil,
					fmt.Errorf("another question in the bank already has this content: %w", err))
			}
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		if saved, err = repos.QuestionPool.LockById(ctx, cmd.QuestionPoolID); err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		if saved == nil {
			return errs.NewError(ctx, status.FAIL, nil, fmt.Errorf("question_pool: updated %d but cannot read it back", cmd.QuestionPoolID))
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return saved, nil
}
