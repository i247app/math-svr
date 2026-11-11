package command

import (
	"context"
	"errors"

	"math-ai.com/math-ai/internal/application/transaction"
	"math-ai.com/math-ai/internal/domain/questionpool"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
)

// ForceDeleteQuestionPoolsCommand removes, for good, every question that
// matches Filter (set fields ANDed), live or soft-deleted. Images stay in S3.
type ForceDeleteQuestionPoolsCommand struct {
	Filter questionpool.ListQuestionPoolsFilter
}

type ForceDeleteQuestionPoolsCommandHandler struct {
	uow transaction.UnitOfWork
}

func NewForceDeleteQuestionPoolsCommandHandler(uow transaction.UnitOfWork) *ForceDeleteQuestionPoolsCommandHandler {
	return &ForceDeleteQuestionPoolsCommandHandler{uow: uow}
}

// Handle is one DELETE and returns how many rows it removed (0 is not an
// error). An empty filter is refused before the statement.
func (h *ForceDeleteQuestionPoolsCommandHandler) Handle(ctx context.Context, cmd ForceDeleteQuestionPoolsCommand) (int64, error) {
	if cmd.Filter.IsEmpty() {
		return 0, errs.NewError(ctx, status.QUESTION_POOL_DELETE_FILTER_REQUIRED, nil, questionpool.ErrEmptyFilter)
	}
	var deleted int64
	err := h.uow.Do(ctx, func(ctx context.Context, repos transaction.Repositories) error {
		n, err := repos.QuestionPool.ForceDeleteByFilter(ctx, cmd.Filter)
		if errors.Is(err, questionpool.ErrEmptyFilter) {
			return errs.NewError(ctx, status.QUESTION_POOL_DELETE_FILTER_REQUIRED, nil, err)
		}
		if err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		deleted = n
		return nil
	})
	return deleted, err
}
