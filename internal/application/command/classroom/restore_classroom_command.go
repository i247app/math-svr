package command

import (
	"context"
	"errors"

	"math-ai.com/math-ai/internal/application/transaction"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/shared/enum"
)

// RestoreClassroomCommand reverses ArchiveClassroomCommand. Restoring
// from DELETED is intentionally NOT supported here — once a row has
// classroom_status=DELETED the classroomActiveWhere filter hides it,
// and recovery is a manual-admin task.
type RestoreClassroomCommand struct {
	ClassroomID string
}

type RestoreClassroomCommandHandler struct {
	uow transaction.UnitOfWork
}

func NewRestoreClassroomCommandHandler(uow transaction.UnitOfWork) *RestoreClassroomCommandHandler {
	return &RestoreClassroomCommandHandler{uow: uow}
}

func (h *RestoreClassroomCommandHandler) Handle(ctx context.Context, cmd RestoreClassroomCommand) error {
	return h.uow.Do(ctx, func(ctx context.Context, repos transaction.Repositories) error {
		existing, err := repos.Classroom.FindByClassroomId(ctx, cmd.ClassroomID)
		if err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		if existing == nil {
			return errs.NewError(ctx, status.CLASSROOM_NOT_FOUND, nil,
				errors.New("classroom not found"))
		}
		if existing.ClassroomStatus() == nil || *existing.ClassroomStatus() != string(enum.ClassroomStatusTypeArchived) {
			return errs.NewError(ctx, status.CLASSROOM_NOT_ARCHIVED, nil,
				errors.New("classroom is not archived"))
		}
		if err := repos.Classroom.RestoreByClassroomId(ctx, cmd.ClassroomID); err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		return nil
	})
}
