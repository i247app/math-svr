package command

import (
	"context"

	"math-ai.com/math-ai/internal/application/transaction"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/shared/enum"
)

// RemoveMemberCommand kicks a target profile out of the classroom.
// CallerProfileID is recorded as removed_by_profile_id so the audit
// trail captures who initiated the removal. Caller-vs-target role
// gating is enforced at the module level; here we only block the
// "cannot remove the owner" case as a defensive invariant.
type RemoveMemberCommand struct {
	ClassroomID     int64
	CallerProfileID int64
	TargetProfileID int64
}

type RemoveMemberCommandHandler struct {
	uow transaction.UnitOfWork
}

func NewRemoveMemberCommandHandler(uow transaction.UnitOfWork) *RemoveMemberCommandHandler {
	return &RemoveMemberCommandHandler{uow: uow}
}

func (h *RemoveMemberCommandHandler) Handle(ctx context.Context, cmd RemoveMemberCommand) error {
	return h.uow.Do(ctx, func(ctx context.Context, repos transaction.Repositories) error {
		target, err := repos.ClassroomMember.FindByClassroomAndProfile(ctx, cmd.ClassroomID, cmd.TargetProfileID)
		if err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		if target == nil {
			return errs.NewError(ctx, status.CLASSROOM_MEMBER_NOT_FOUND, nil,
				ErrTargetMemberNotFound)
		}
		if target.MemberStatus() == nil || *target.MemberStatus() != string(enum.ClassroomMemberStatusTypeActive) {
			return errs.NewError(ctx, status.CLASSROOM_MEMBER_NOT_MEMBER, nil,
				ErrTargetMemberNotActive)
		}
		if target.MemberRole() == string(enum.ClassroomMemberRoleTypeOwner) {
			return errs.NewError(ctx, status.CLASSROOM_MEMBER_CANNOT_REMOVE_OWNER, nil,
				ErrOwnerCannotBeRemoved)
		}
		if err := repos.ClassroomMember.MarkRemoved(ctx, target.MemberId(), cmd.CallerProfileID); err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		// OWNER is rejected above, so target.MemberRole() is either
		// STUDENT or CO_TEACHER — both buckets are covered by the helper.
		studentDelta, teacherDelta := roleCountDeltas(target.MemberRole(), -1)
		if err := repos.Classroom.IncCounts(ctx, cmd.ClassroomID, -1, studentDelta, teacherDelta); err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		return nil
	})
}
