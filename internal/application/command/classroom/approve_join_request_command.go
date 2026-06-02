package command

import (
	"context"

	"math-ai.com/math-ai/internal/application/transaction"
	"math-ai.com/math-ai/internal/domain/classroom"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/shared/enum"
)

// ApproveJoinRequestCommand flips a PENDING_REQUEST ma_classroom_members
// row to ACTIVE inside one UoW and advances the classroom counters in
// the same transaction. Owner permission is enforced at the module
// layer; here we trust the (classroom, target_profile) pair and guard
// the state-machine: only PENDING_REQUEST approves. The max_members
// check runs here (not at request time) so requests can queue while
// seats open up.
type ApproveJoinRequestCommand struct {
	ClassroomID     int64
	TargetProfileID int64
	ActorID         *int64
}

type ApproveJoinRequestCommandHandler struct {
	uow transaction.UnitOfWork
}

func NewApproveJoinRequestCommandHandler(uow transaction.UnitOfWork) *ApproveJoinRequestCommandHandler {
	return &ApproveJoinRequestCommandHandler{uow: uow}
}

func (h *ApproveJoinRequestCommandHandler) Handle(ctx context.Context, cmd ApproveJoinRequestCommand) (*classroom.Member, error) {
	var approved *classroom.Member

	err := h.uow.Do(ctx, func(ctx context.Context, repos transaction.Repositories) error {
		c, err := repos.Classroom.FindByClassroomId(ctx, cmd.ClassroomID)
		if err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		if c == nil {
			return errs.NewError(ctx, status.CLASSROOM_NOT_FOUND, nil,
				ErrClassroomNotFound)
		}
		if c.ClassroomStatus() != nil &&
			*c.ClassroomStatus() == string(enum.ClassroomStatusTypeArchived) {
			return errs.NewError(ctx, status.CLASSROOM_ALREADY_ARCHIVED, nil,
				ErrClassroomArchived)
		}
		if c.MaxMembers() != nil && c.MemberCount() >= *c.MaxMembers() {
			return errs.NewError(ctx, status.CLASSROOM_MAX_MEMBERS_REACHED, nil,
				ErrClassroomFull)
		}

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

		if err := repos.ClassroomMember.Activate(ctx, existing.MemberId()); err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		studentDelta, teacherDelta := roleCountDeltas(existing.MemberRole(), 1)
		if err := repos.Classroom.IncCounts(ctx, cmd.ClassroomID, 1, studentDelta, teacherDelta); err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		refreshed, err := repos.ClassroomMember.FindByMemberId(ctx, existing.MemberId())
		if err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		approved = refreshed
		return nil
	})
	if err != nil {
		return nil, err
	}
	return approved, nil
}
