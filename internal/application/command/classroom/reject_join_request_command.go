package command

import (
	"context"

	"math-ai.com/math-ai/internal/application/transaction"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/shared/enum"
)

// RejectJoinRequestCommand flips a PENDING_REQUEST row to REJECTED.
// Owner permission is enforced at the module layer; here we trust
// the (classroom, target_profile) pair and guard the state-machine.
// No counter change (the requester never joined).
type RejectJoinRequestCommand struct {
	ClassroomID     int64
	TargetProfileID int64
	ActorID         *int64
}

type RejectJoinRequestCommandHandler struct {
	uow transaction.UnitOfWork
}

func NewRejectJoinRequestCommandHandler(uow transaction.UnitOfWork) *RejectJoinRequestCommandHandler {
	return &RejectJoinRequestCommandHandler{uow: uow}
}

func (h *RejectJoinRequestCommandHandler) Handle(ctx context.Context, cmd RejectJoinRequestCommand) error {
	return h.uow.Do(ctx, func(ctx context.Context, repos transaction.Repositories) error {
		existing, err := repos.ClassroomMember.FindByClassroomAndProfile(ctx, cmd.ClassroomID, cmd.TargetProfileID)
		if err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		if existing == nil {
			return errs.NewError(ctx, status.CLASSROOM_JOIN_REQUEST_NOT_FOUND, nil,
				ErrJoinRequestNotFound)
		}
		currentStatus := ""
		if existing.MemberStatus() != nil {
			currentStatus = *existing.MemberStatus()
		}
		if currentStatus != string(enum.ClassroomMemberStatusTypePendingRequest) {
			return errs.NewError(ctx, status.CLASSROOM_JOIN_REQUEST_NOT_PENDING, nil,
				ErrJoinRequestNotPending)
		}
		if err := repos.ClassroomMember.Reject(ctx, existing.MemberId()); err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		return nil
	})
}
