package command

import (
	"context"

	"math-ai.com/math-ai/internal/application/transaction"
	"math-ai.com/math-ai/internal/domain/exam"
	"math-ai.com/math-ai/internal/shared/enum"
)

// CreateExamPoolCommand stores a question set generated ahead of any
// child (POST /exams/pools/generate). No sitting and no journey is
// written: the set only joins the pool, where /exams/generate can later
// serve it from cache under its req_extras.
type CreateExamPoolCommand struct {
	ExamType  enum.ExamType
	Grade     int
	CreatedBy int64
	Content   *NewExamPoolContent
}

type CreateExamPoolCommandHandler struct {
	uow transaction.UnitOfWork
}

func NewCreateExamPoolCommandHandler(uow transaction.UnitOfWork) *CreateExamPoolCommandHandler {
	return &CreateExamPoolCommandHandler{uow: uow}
}

func (h *CreateExamPoolCommandHandler) Handle(ctx context.Context, cmd CreateExamPoolCommand) (*exam.ExamPool, error) {
	var saved *exam.ExamPool
	err := h.uow.Do(ctx, func(ctx context.Context, repos transaction.Repositories) error {
		var err error
		saved, err = insertExamPool(ctx, repos, cmd.ExamType, cmd.Grade, cmd.CreatedBy, cmd.Content)
		return err
	})
	if err != nil {
		return nil, err
	}
	return saved, nil
}
