package command

import (
	"context"

	"math-ai.com/math-ai/internal/application/transaction"
	"math-ai.com/math-ai/internal/domain/classroom"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/shared/enum"
	"math-ai.com/math-ai/internal/shared/errs"
)

// AcceptInvitationCommand flips a PENDING ma_classroom_members row to
// ACTIVE inside one UoW and advances the classroom counters in the
// same transaction. Caller permission (caller == invitee) is enforced
// at the module layer; here we trust the (classroom, profile) pair and
// focus on the state-machine invariant: only PENDING accepts, never
// ACTIVE / REJECTED / LEFT / REMOVED.
type AcceptInvitationCommand struct {
	ClassroomID      int64
	InviteeProfileID int64
	InviterProfileID int64
	ActorID          *int64
}

type AcceptInvitationCommandHandler struct {
	uow transaction.UnitOfWork
}

func NewAcceptInvitationCommandHandler(uow transaction.UnitOfWork) *AcceptInvitationCommandHandler {
	return &AcceptInvitationCommandHandler{uow: uow}
}

func (h *AcceptInvitationCommandHandler) Handle(ctx context.Context, cmd AcceptInvitationCommand) (*classroom.Member, error) {
	var accepted *classroom.Member

	err := h.uow.Do(ctx, func(ctx context.Context, repos transaction.Repositories) error {
		c, err := repos.Classroom.FindByClassroomId(ctx, cmd.ClassroomID)
		if err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		if c == nil {
			return errs.NewError(ctx, classroom.ClassroomNotFound, nil,
				ErrClassroomNotFound)
		}
		if c.ClassroomStatus() != nil &&
			*c.ClassroomStatus() == enum.ClassroomStatusTypeArchived.String() {
			return errs.NewError(ctx, classroom.ClassroomAlreadyArchived, nil,
				ErrClassroomArchived)
		}
		if c.MaxMembers() != nil && c.MemberCount() >= *c.MaxMembers() {
			return errs.NewError(ctx, classroom.ClassroomMaxMembersReached, nil,
				ErrClassroomFull)
		}

		existing, err := repos.ClassroomMember.FindByClassroomAndProfileAndInvitedBy(ctx, cmd.ClassroomID, cmd.InviteeProfileID, cmd.InviterProfileID)
		if err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		if existing == nil {
			return errs.NewError(ctx, classroom.ClassroomInvitationNotFound, nil,
				ErrInvitationNotFound)
		}
		currentStatus := ""
		if existing.MemberStatus() != nil {
			currentStatus = *existing.MemberStatus()
		}
		if currentStatus != enum.ClassroomMemberStatusTypePendingInvitation.String() {
			return errs.NewError(ctx, classroom.ClassroomInvitationNotPending, nil,
				ErrInvitationNotPending)
		}

		if err := repos.ClassroomMember.Activate(ctx, existing.MemberId()); err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		studentDelta, teacherDelta := classroom.RoleCountDeltas(existing.MemberRole(), 1)
		if err := repos.Classroom.IncCounts(ctx, cmd.ClassroomID, 1, studentDelta, teacherDelta); err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		refreshed, err := repos.ClassroomMember.FindByMemberId(ctx, existing.MemberId())
		if err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		accepted = refreshed
		return nil
	})
	if err != nil {
		return nil, err
	}
	return accepted, nil
}
