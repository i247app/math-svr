package command

import (
	"context"
	"errors"

	"math-ai.com/math-ai/internal/application/transaction"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/shared/enum"
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
			return errs.NewError(ctx, status.CLASSROOM_INVITATION_NOT_FOUND, nil,
				errors.New("invitation not found"))
		}
		currentStatus := ""
		if existing.MemberStatus() != nil {
			currentStatus = *existing.MemberStatus()
		}
		if currentStatus != string(enum.ClassroomMemberStatusTypePendingInvitation) {
			return errs.NewError(ctx, status.CLASSROOM_INVITATION_NOT_PENDING, nil,
				errors.New("invitation is not pending"))
		}
		if err := repos.ClassroomMember.Reject(ctx, existing.MemberId()); err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		return nil
	})
}
