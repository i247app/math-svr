package command

import (
	"context"

	"math-ai.com/math-ai/internal/application/transaction"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/shared/enum"
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
			return errs.NewError(ctx, status.CLASSROOM_NOT_FOUND, nil,
				ErrClassroomNotFound)
		}
		if existing.ClassroomStatus() != nil && *existing.ClassroomStatus() == string(enum.ClassroomStatusTypeArchived) {
			return errs.NewError(ctx, status.CLASSROOM_ALREADY_ARCHIVED, nil,
				ErrClassroomAlreadyArchived)
		}
		if err := repos.Classroom.ArchiveByClassroomId(ctx, cmd.ClassroomID); err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		return nil
	})
}
