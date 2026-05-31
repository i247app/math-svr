package command

import (
	"context"
	"errors"

	"math-ai.com/math-ai/internal/application/transaction"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/shared/enum"
)

// LeaveClassroomCommand transitions the caller's member row to LEFT and
// decrements the classroom's member_count. The OWNER guard runs at the
// module level — by the time we get here, we trust the caller is a
// non-owner active member.
type LeaveClassroomCommand struct {
	ProfileID   int64
	ClassroomID int64
}

type LeaveClassroomCommandHandler struct {
	uow transaction.UnitOfWork
}

func NewLeaveClassroomCommandHandler(uow transaction.UnitOfWork) *LeaveClassroomCommandHandler {
	return &LeaveClassroomCommandHandler{uow: uow}
}

func (h *LeaveClassroomCommandHandler) Handle(ctx context.Context, cmd LeaveClassroomCommand) error {
	return h.uow.Do(ctx, func(ctx context.Context, repos transaction.Repositories) error {
		m, err := repos.ClassroomMember.FindByClassroomAndProfile(ctx, cmd.ClassroomID, cmd.ProfileID)
		if err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		if m == nil {
			return errs.NewError(ctx, status.CLASSROOM_MEMBER_NOT_MEMBER, nil,
				errors.New("not a member of this classroom"))
		}
		if m.MemberStatus() == nil || *m.MemberStatus() != string(enum.ClassroomMemberStatusTypeActive) {
			return errs.NewError(ctx, status.CLASSROOM_MEMBER_NOT_MEMBER, nil,
				errors.New("membership is not active"))
		}
		if m.MemberRole() == string(enum.ClassroomMemberRoleTypeOwner) {
			return errs.NewError(ctx, status.CLASSROOM_OWNER_CANNOT_LEAVE, nil,
				errors.New("owner must transfer ownership before leaving"))
		}
		if err := repos.ClassroomMember.MarkLeft(ctx, m.MemberId()); err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		// Decrement the role-specific bucket the departing member
		// occupied. OWNER cannot reach here (guarded above), so the
		// only roles in play are STUDENT and CO_TEACHER.
		studentDelta, teacherDelta := roleCountDeltas(m.MemberRole(), -1)
		if err := repos.Classroom.IncCounts(ctx, cmd.ClassroomID, -1, studentDelta, teacherDelta); err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		return nil
	})
}
