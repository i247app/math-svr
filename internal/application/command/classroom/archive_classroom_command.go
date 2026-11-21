package command

import (
	"context"

	"math-ai.com/math-ai/internal/application/transaction"
	classroomdom "math-ai.com/math-ai/internal/domain/classroom"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/shared/enum"
	"math-ai.com/math-ai/internal/shared/errs"
)

// ArchiveClassroomCommand flips classroom_status to ARCHIVED. Archived
// classrooms remain readable (history view) but mutations are blocked at
// the module level. Idempotency: re-archiving an ARCHIVED classroom
// returns ALREADY_ARCHIVED rather than silently succeeding.
type ArchiveClassroomCommand struct {
	ClassroomID int64
}

type ArchiveClassroomCommandHandler struct {
	uow transaction.UnitOfWork
}

func NewArchiveClassroomCommandHandler(uow transaction.UnitOfWork) *ArchiveClassroomCommandHandler {
	return &ArchiveClassroomCommandHandler{uow: uow}
}

func (h *ArchiveClassroomCommandHandler) Handle(ctx context.Context, cmd ArchiveClassroomCommand) error {
	return h.uow.Do(ctx, func(ctx context.Context, repos transaction.Repositories) error {
		existing, err := repos.Classroom.FindByClassroomId(ctx, cmd.ClassroomID)
		if err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		if existing == nil {
			return errs.NewError(ctx, classroomdom.ClassroomNotFound, nil,
				ErrClassroomNotFound)
		}
		if existing.ClassroomStatus() != nil && *existing.ClassroomStatus() == enum.ClassroomStatusTypeArchived.String() {
			return errs.NewError(ctx, classroomdom.ClassroomAlreadyArchived, nil,
				ErrClassroomAlreadyArchived)
		}
		if err := repos.Classroom.ArchiveByClassroomId(ctx, cmd.ClassroomID); err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		return nil
	})
}
