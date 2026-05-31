package command

import (
	"context"
	"errors"

	"math-ai.com/math-ai/internal/application/transaction"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/shared/enum"
)

// TransferOwnershipCommand atomically moves the OWNER role from the
// current owner to a target ACTIVE member. The outgoing owner is
// demoted to CO_TEACHER so they keep manager rights but no longer
// block the leave path. The classroom row's owner_profile_id is kept
// in sync with the member rows so O(1) "who owns this" lookups stay
// accurate.
type TransferOwnershipCommand struct {
	ClassroomID       int64
	CurrentOwnerID    int64
	NewOwnerProfileID int64
}

type TransferOwnershipCommandHandler struct {
	uow transaction.UnitOfWork
}

func NewTransferOwnershipCommandHandler(uow transaction.UnitOfWork) *TransferOwnershipCommandHandler {
	return &TransferOwnershipCommandHandler{uow: uow}
}

func (h *TransferOwnershipCommandHandler) Handle(ctx context.Context, cmd TransferOwnershipCommand) error {
	return h.uow.Do(ctx, func(ctx context.Context, repos transaction.Repositories) error {
		if cmd.CurrentOwnerID == cmd.NewOwnerProfileID {
			return errs.NewError(ctx, status.CLASSROOM_OWNER_TRANSFER_TO_NON_MEMBER, nil,
				errors.New("new owner must differ from current owner"))
		}
		currentOwner, err := repos.ClassroomMember.FindByClassroomAndProfile(ctx, cmd.ClassroomID, cmd.CurrentOwnerID)
		if err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		if currentOwner == nil || currentOwner.MemberRole() != string(enum.ClassroomMemberRoleTypeOwner) {
			return errs.NewError(ctx, status.CLASSROOM_PERMISSION_DENIED, nil,
				errors.New("caller is not the current owner"))
		}

		newOwner, err := repos.ClassroomMember.FindByClassroomAndProfile(ctx, cmd.ClassroomID, cmd.NewOwnerProfileID)
		if err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		if newOwner == nil {
			return errs.NewError(ctx, status.CLASSROOM_OWNER_TRANSFER_TO_NON_MEMBER, nil,
				errors.New("new owner must be an existing member"))
		}
		if newOwner.MemberStatus() == nil || *newOwner.MemberStatus() != string(enum.ClassroomMemberStatusTypeActive) {
			return errs.NewError(ctx, status.CLASSROOM_OWNER_TRANSFER_TO_NON_MEMBER, nil,
				errors.New("new owner must be an active member"))
		}

		if err := repos.ClassroomMember.SetRole(ctx, currentOwner.MemberId(),
			string(enum.ClassroomMemberRoleTypeCoTeacher)); err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		if err := repos.ClassroomMember.SetRole(ctx, newOwner.MemberId(),
			string(enum.ClassroomMemberRoleTypeOwner)); err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		if err := repos.Classroom.SetOwnerProfileId(ctx, cmd.ClassroomID, cmd.NewOwnerProfileID); err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		return nil
	})
}
