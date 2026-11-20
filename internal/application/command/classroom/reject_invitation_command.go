package command

import (
	"context"

	"math-ai.com/math-ai/internal/application/transaction"
	classroomdom "math-ai.com/math-ai/internal/domain/classroom"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/shared/enum"
	errs "math-ai.com/math-ai/internal/shared/errs"
)

// RejectInvitationCommand flips a PENDING ma_classroom_members row to
// REJECTED. No counter change. Caller-is-invitee is enforced at the
// module layer; here we trust the (classroom, profile) pair and guard
// the state-machine: only PENDING rejects, never ACTIVE / etc.
type RejectInvitationCommand struct {
	ClassroomID      int64
	InviteeProfileID int64
	InviterProfileID int64
	ActorID          *int64
}

type RejectInvitationCommandHandler struct {
	uow transaction.UnitOfWork
}

func NewRejectInvitationCommandHandler(uow transaction.UnitOfWork) *RejectInvitationCommandHandler {
	return &RejectInvitationCommandHandler{uow: uow}
}

func (h *RejectInvitationCommandHandler) Handle(ctx context.Context, cmd RejectInvitationCommand) error {
	return h.uow.Do(ctx, func(ctx context.Context, repos transaction.Repositories) error {
		existing, err := repos.ClassroomMember.FindByClassroomAndProfileAndInvitedBy(ctx, cmd.ClassroomID, cmd.InviteeProfileID, cmd.InviterProfileID)
		if err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		if existing == nil {
			return errs.NewError(ctx, classroomdom.ClassroomInvitationNotFound, nil,
				ErrInvitationNotFound)
		}
		currentStatus := ""
		if existing.MemberStatus() != nil {
			currentStatus = *existing.MemberStatus()
		}
		if currentStatus != enum.ClassroomMemberStatusTypePendingInvitation.String() {
			return errs.NewError(ctx, classroomdom.ClassroomInvitationNotPending, nil,
				ErrInvitationNotPending)
		}
		if err := repos.ClassroomMember.Reject(ctx, existing.MemberId()); err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		return nil
	})
}
