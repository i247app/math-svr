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

// UpdateMemberRoleCommand promotes / demotes a non-owner member
// between STUDENT and CO_TEACHER. OWNER cannot be set or unset through
// this path — use TransferOwnership instead. NewRole is validated at
// the module level (must be CO_TEACHER or STUDENT) and a TEACHER
// profile check guards promotion to CO_TEACHER.
type UpdateMemberRoleCommand struct {
	ClassroomID     string
	TargetProfileID string
	NewRole         string
	ActorID         *string
}

type UpdateMemberRoleCommandHandler struct {
	uow transaction.UnitOfWork
}

func NewUpdateMemberRoleCommandHandler(uow transaction.UnitOfWork) *UpdateMemberRoleCommandHandler {
	return &UpdateMemberRoleCommandHandler{uow: uow}
}

func (h *UpdateMemberRoleCommandHandler) Handle(ctx context.Context, cmd UpdateMemberRoleCommand) (*classroom.Member, error) {
	var updated *classroom.Member

	err := h.uow.Do(ctx, func(ctx context.Context, repos transaction.Repositories) error {
		target, err := repos.ClassroomMember.FindByClassroomAndProfile(ctx, cmd.ClassroomID, cmd.TargetProfileID)
		if err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		if target == nil {
			return errs.NewError(ctx, status.CLASSROOM_MEMBER_NOT_FOUND, nil,
				errors.New("target member not found"))
		}
		if target.MemberStatus() == nil || *target.MemberStatus() != string(enum.ClassroomMemberStatusTypeActive) {
			return errs.NewError(ctx, status.CLASSROOM_MEMBER_NOT_MEMBER, nil,
				errors.New("target member is not active"))
		}
		if target.MemberRole() == string(enum.ClassroomMemberRoleTypeOwner) {
			return errs.NewError(ctx, status.CLASSROOM_MEMBER_CANNOT_DEMOTE_OWNER, nil,
				errors.New("owner role cannot be changed via update; use transfer"))
		}
		if target.MemberRole() == cmd.NewRole {
			updated = target
			return nil
		}
		if err := repos.ClassroomMember.SetRole(ctx, target.MemberId(), cmd.NewRole); err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		// Shift counters between buckets in lockstep with the role flip.
		// member_count stays put — the head count is unchanged.
		oldStudentDelta, oldTeacherDelta := roleCountDeltas(target.MemberRole(), -1)
		newStudentDelta, newTeacherDelta := roleCountDeltas(cmd.NewRole, 1)
		if err := repos.Classroom.IncCounts(ctx, cmd.ClassroomID, 0,
			oldStudentDelta+newStudentDelta,
			oldTeacherDelta+newTeacherDelta); err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		refreshed, err := repos.ClassroomMember.FindByMemberId(ctx, target.MemberId())
		if err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		updated = refreshed
		return nil
	})
	if err != nil {
		return nil, err
	}
	return updated, nil
}
