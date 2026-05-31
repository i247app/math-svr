package command

import (
	"context"
	"errors"

	"math-ai.com/math-ai/internal/application/transaction"
	"math-ai.com/math-ai/internal/domain/classroom"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/shared/enum"
)

// AcceptInvitationCommand flips a PENDING ma_classroom_members row to
// ACTIVE inside one UoW and advances the classroom counters in the
// same transaction. Caller permission (caller == invitee) is enforced
// at the module layer; here we trust the (classroom, profile) pair and
// focus on the state-machine invariant: only PENDING accepts, never
// ACTIVE / REJECTED / LEFT / REMOVED.
type AcceptInvitationCommand struct {
	ClassroomID     int64
	CallerProfileID int64
	ActorID         *int64
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
			return errs.NewError(ctx, status.CLASSROOM_NOT_FOUND, nil,
				errors.New("classroom not found"))
		}
		if c.ClassroomStatus() != nil &&
			*c.ClassroomStatus() == string(enum.ClassroomStatusTypeArchived) {
			return errs.NewError(ctx, status.CLASSROOM_ALREADY_ARCHIVED, nil,
				errors.New("classroom is archived"))
		}
		if c.MaxMembers() != nil && c.MemberCount() >= *c.MaxMembers() {
			return errs.NewError(ctx, status.CLASSROOM_MAX_MEMBERS_REACHED, nil,
				errors.New("classroom is full"))
		}

		existing, err := repos.ClassroomMember.FindByClassroomAndProfile(ctx, cmd.ClassroomID, cmd.CallerProfileID)
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
		if currentStatus != string(enum.ClassroomMemberStatusTypePending) {
			return errs.NewError(ctx, status.CLASSROOM_INVITATION_NOT_PENDING, nil,
				errors.New("invitation is not pending"))
		}

		if err := repos.ClassroomMember.Accept(ctx, existing.MemberId()); err != nil {
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
		accepted = refreshed
		return nil
	})
	if err != nil {
		return nil, err
	}
	return accepted, nil
}
