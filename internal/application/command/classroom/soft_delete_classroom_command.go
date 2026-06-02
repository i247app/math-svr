package command

import (
	"context"

	"math-ai.com/math-ai/internal/application/transaction"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
)

// SoftDeleteClassroomCommand flips the parent row to DELETED but leaves
// child member rows untouched. Reads filter both sides via the active
// where clause, and the cheaper restore path lets us recover without
// rebuilding membership history.
type SoftDeleteClassroomCommand struct {
	ClassroomID int64
}

type SoftDeleteClassroomCommandHandler struct {
	uow transaction.UnitOfWork
}

func NewSoftDeleteClassroomCommandHandler(uow transaction.UnitOfWork) *SoftDeleteClassroomCommandHandler {
	return &SoftDeleteClassroomCommandHandler{uow: uow}
}

func (h *SoftDeleteClassroomCommandHandler) Handle(ctx context.Context, cmd SoftDeleteClassroomCommand) error {
	return h.uow.Do(ctx, func(ctx context.Context, repos transaction.Repositories) error {
		existing, err := repos.Classroom.FindByClassroomId(ctx, cmd.ClassroomID)
		if err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		if existing == nil {
			return errs.NewError(ctx, status.CLASSROOM_NOT_FOUND, nil,
				ErrClassroomNotFound)
		}
		if err := repos.Classroom.SoftDeleteByClassroomId(ctx, cmd.ClassroomID); err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		return nil
	})
}
