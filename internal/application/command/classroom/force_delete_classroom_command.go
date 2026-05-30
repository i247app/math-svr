package command

import (
	"context"

	"math-ai.com/math-ai/internal/application/transaction"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
)

// ForceDeleteClassroomCommand removes the classroom and its child rows
// physically. Order matters: invitations → members → classroom so any
// future foreign keys (or just consistency-conscious operators) won't
// trip. Idempotent: missing rows are not an error.
type ForceDeleteClassroomCommand struct {
	ClassroomID string
}

type ForceDeleteClassroomCommandHandler struct {
	uow transaction.UnitOfWork
}

func NewForceDeleteClassroomCommandHandler(uow transaction.UnitOfWork) *ForceDeleteClassroomCommandHandler {
	return &ForceDeleteClassroomCommandHandler{uow: uow}
}

func (h *ForceDeleteClassroomCommandHandler) Handle(ctx context.Context, cmd ForceDeleteClassroomCommand) error {
	return h.uow.Do(ctx, func(ctx context.Context, repos transaction.Repositories) error {
		if err := repos.ClassroomInvitation.ForceDeleteByClassroomId(ctx, cmd.ClassroomID); err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		if err := repos.ClassroomMember.ForceDeleteByClassroomId(ctx, cmd.ClassroomID); err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		if err := repos.Classroom.ForceDeleteByClassroomId(ctx, cmd.ClassroomID); err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		return nil
	})
}
