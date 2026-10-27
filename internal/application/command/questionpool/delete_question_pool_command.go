package command

import (
	"context"
	"fmt"

	"math-ai.com/math-ai/internal/application/transaction"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
)

// DeleteQuestionCommand removes one question. Soft (Force false) hides the
// live row and frees its hash; Force deletes the row, live or soft-deleted.
// Uploaded images stay in S3 either way.
type DeleteQuestionPoolCommand struct {
	QuestionPoolID int64
	ActorUID       *int64 // modify_id of a soft delete; nil = no uid (API-key admin)
	Force          bool
}

type DeleteQuestionPoolCommandHandler struct {
	uow transaction.UnitOfWork
}

func NewDeleteQuestionPoolCommandHandler(uow transaction.UnitOfWork) *DeleteQuestionPoolCommandHandler {
	return &DeleteQuestionPoolCommandHandler{uow: uow}
}

// Handle is one statement; no row matched → QUESTION_POOL_NOT_FOUND.
func (h *DeleteQuestionPoolCommandHandler) Handle(ctx context.Context, cmd DeleteQuestionPoolCommand) error {
	return h.uow.Do(ctx, func(ctx context.Context, repos transaction.Repositories) error {
		var found bool
		var err error
		if cmd.Force {
			found, err = repos.QuestionPool.ForceDeleteById(ctx, cmd.QuestionPoolID)
		} else {
			found, err = repos.QuestionPool.SoftDeleteById(ctx, cmd.QuestionPoolID, cmd.ActorUID)
		}
		if err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		if !found {
			return errs.NewError(ctx, status.QUESTION_POOL_NOT_FOUND, nil, fmt.Errorf("question_pool: no question %d to delete", cmd.QuestionPoolID))
		}
		return nil
	})
}
