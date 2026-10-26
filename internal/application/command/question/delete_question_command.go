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
type DeleteQuestionCommand struct {
	QuestionID int64
	ActorUID   *int64 // modify_id of a soft delete; nil = no uid (API-key admin)
	Force      bool
}

type DeleteQuestionCommandHandler struct {
	uow transaction.UnitOfWork
}

func NewDeleteQuestionCommandHandler(uow transaction.UnitOfWork) *DeleteQuestionCommandHandler {
	return &DeleteQuestionCommandHandler{uow: uow}
}

// Handle is one statement; no row matched → QUESTION_NOT_FOUND.
func (h *DeleteQuestionCommandHandler) Handle(ctx context.Context, cmd DeleteQuestionCommand) error {
	return h.uow.Do(ctx, func(ctx context.Context, repos transaction.Repositories) error {
		var found bool
		var err error
		if cmd.Force {
			found, err = repos.Question.ForceDeleteById(ctx, cmd.QuestionID)
		} else {
			found, err = repos.Question.SoftDeleteById(ctx, cmd.QuestionID, cmd.ActorUID)
		}
		if err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		if !found {
			return errs.NewError(ctx, status.QUESTION_NOT_FOUND, nil, fmt.Errorf("question: no question %d to delete", cmd.QuestionID))
		}
		return nil
	})
}
