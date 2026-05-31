package command

import (
	"context"
	"errors"

	"math-ai.com/math-ai/internal/application/transaction"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/shared/enum"
)

// CancelInvitationCommand flips a PENDING ma_classroom_members row to
// REMOVED, capturing the manager's profile id in
// removed_by_profile_id. No counter change (the target never joined).
// Manager permission is enforced at the module layer; here we trust
// the (classroom, target_profile) pair and guard the state-machine:
// only PENDING cancels, never ACTIVE / REJECTED / etc.
type CancelInvitationCommand struct {
	ClassroomID     int64
	TargetProfileID int64
	CancelledBy     int64
}

type CancelInvitationCommandHandler struct {
	uow transaction.UnitOfWork
}

func NewCancelInvitationCommandHandler(uow transaction.UnitOfWork) *CancelInvitationCommandHandler {
	return &CancelInvitationCommandHandler{uow: uow}
}

func (h *CancelInvitationCommandHandler) Handle(ctx context.Context, cmd CancelInvitationCommand) error {
	return h.uow.Do(ctx, func(ctx context.Context, repos transaction.Repositories) error {
		existing, err := repos.ClassroomMember.FindByClassroomAndProfile(ctx, cmd.ClassroomID, cmd.TargetProfileID)
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
		if err := repos.ClassroomMember.Cancel(ctx, existing.MemberId(), cmd.CancelledBy); err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		return nil
	})
}
