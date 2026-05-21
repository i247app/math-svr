package command

import (
	"context"

	"github.com/google/uuid"
	"math-ai.com/math-ai/internal/application/transaction"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
)

type SoftDeleteQuizCommand struct {
	QuizID uuid.UUID
}

type SoftDeleteQuizCommandHandler struct {
	uow transaction.UnitOfWork
}

func NewSoftDeleteQuizCommandHandler(uow transaction.UnitOfWork) *SoftDeleteQuizCommandHandler {
	return &SoftDeleteQuizCommandHandler{uow: uow}
}

func (h *SoftDeleteQuizCommandHandler) Handle(ctx context.Context, cmd SoftDeleteQuizCommand) error {
	return h.uow.Do(ctx, func(ctx context.Context, repos transaction.Repositories) error {
		if err := repos.Quiz.SoftDelete(ctx, cmd.QuizID); err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		return nil
	})
}
