package command

import (
	"context"
	"errors"

	"math-ai.com/math-ai/internal/application/transaction"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/shared/enum"
)

// CancelJoinRequestCommand flips a PENDING_REQUEST row owned by the
// caller to REMOVED, capturing the caller's profile id in
// removed_by_profile_id. The user is cancelling their own request —
// caller-is-requester is enforced at the module layer; here we trust
// the (classroom, caller_profile) pair and guard the state-machine:
// only PENDING_REQUEST cancels.
type CancelJoinRequestCommand struct {
	ClassroomID     int64
	CallerProfileID int64
}

type CancelJoinRequestCommandHandler struct {
	uow transaction.UnitOfWork
}

func NewCancelJoinRequestCommandHandler(uow transaction.UnitOfWork) *CancelJoinRequestCommandHandler {
	return &CancelJoinRequestCommandHandler{uow: uow}
}

func (h *CancelJoinRequestCommandHandler) Handle(ctx context.Context, cmd CancelJoinRequestCommand) error {
	return h.uow.Do(ctx, func(ctx context.Context, repos transaction.Repositories) error {
		existing, err := repos.ClassroomMember.FindByClassroomAndProfile(ctx, cmd.ClassroomID, cmd.CallerProfileID)
		if err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		if existing == nil {
			return errs.NewError(ctx, status.CLASSROOM_JOIN_REQUEST_NOT_FOUND, nil,
				errors.New("join request not found"))
		}
		currentStatus := ""
		if existing.MemberStatus() != nil {
			currentStatus = *existing.MemberStatus()
		}
		if currentStatus != string(enum.ClassroomMemberStatusTypePendingRequest) {
			return errs.NewError(ctx, status.CLASSROOM_JOIN_REQUEST_NOT_PENDING, nil,
				errors.New("join request is not pending"))
		}
		if err := repos.ClassroomMember.Cancel(ctx, existing.MemberId(), cmd.CallerProfileID); err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		return nil
	})
}
