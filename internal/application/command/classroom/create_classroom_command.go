package command

import (
	"context"
	"errors"
	"strings"

	"math-ai.com/math-ai/internal/application/transaction"
	"math-ai.com/math-ai/internal/domain/classroom"
	"math-ai.com/math-ai/internal/domain/seq"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	mtime "math-ai.com/math-ai/internal/domain/shared/time"
	"math-ai.com/math-ai/internal/shared/enum"
)

// CreateClassroomCommand creates the classroom row AND the OWNER member
// row in a single UoW. Role/identity checks live at the module level —
// here we trust the caller and focus on atomicity: a classroom always
// ships with exactly one OWNER on disk.
type CreateClassroomCommand struct {
	ActorID             *string
	OwnerProfileID      string
	Name                string
	Description         *string
	SchoolID            *string
	ProgramID           *string
	GradeID             *string
	MaxMembers          *int64
	CoverKey            *string
	Note                *string
	InviteCode          *string
	InviteCodeExpiresDt mtime.MathTime
}

type CreateClassroomCommandHandler struct {
	uow transaction.UnitOfWork
}

func NewCreateClassroomCommandHandler(uow transaction.UnitOfWork) *CreateClassroomCommandHandler {
	return &CreateClassroomCommandHandler{uow: uow}
}

func (h *CreateClassroomCommandHandler) Handle(ctx context.Context, cmd CreateClassroomCommand) (*classroom.Classroom, error) {
	var created *classroom.Classroom

	err := h.uow.Do(ctx, func(ctx context.Context, repos transaction.Repositories) error {
		// Reject duplicate invite_code up front so the friendlier
		// CLASSROOM_INVITE_CODE_TAKEN propagates instead of a generic
		// repo-level constraint error. The DB UNIQUE key is still the
		// hard backstop for the race window.
		var inviteCodeArg *string
		if cmd.InviteCode != nil {
			trimmed := strings.TrimSpace(*cmd.InviteCode)
			if trimmed != "" {
				existing, err := repos.Classroom.FindByInviteCode(ctx, trimmed)
				if err != nil {
					return errs.NewError(ctx, status.FAIL, nil, err)
				}
				if existing != nil {
					return errs.NewError(ctx, status.CLASSROOM_INVITE_CODE_TAKEN, nil,
						errors.New("invite code already taken"))
				}
				inviteCodeArg = &trimmed
			}
		}

		classroomID, err := nextSeqID(ctx, repos, seq.NameClassroom)
		if err != nil {
			return err
		}

		c := classroom.NewClassroom()
		c.SetClassroomId(classroomID)
		c.SetOwnerProfileId(cmd.OwnerProfileID)
		c.SetName(cmd.Name)
		c.SetDescription(cmd.Description)
		c.SetSchoolId(cmd.SchoolID)
		c.SetProgramId(cmd.ProgramID)
		c.SetGradeId(cmd.GradeID)
		c.SetInviteCode(inviteCodeArg)
		if cmd.InviteCodeExpiresDt.IsValid() {
			c.SetInviteCodeExpiresDt(cmd.InviteCodeExpiresDt)
		}
		c.SetMaxMembers(cmd.MaxMembers)
		// Seed counters in sync with the OWNER member row inserted
		// below. OWNER counts as a teacher for role tallies.
		c.SetMemberCount(1)
		c.SetStudentCount(0)
		c.SetTeacherCount(1)
		c.SetCoverKey(cmd.CoverKey)
		c.SetNote(cmd.Note)
		c.SetCreateId(cmd.ActorID)

		saved, err := repos.Classroom.Create(ctx, c)
		if err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		if saved == nil {
			return errs.NewError(ctx, status.CLASSROOM_NOT_FOUND, nil,
				errors.New("classroom not found after insert"))
		}

		memberID, err := nextSeqID(ctx, repos, seq.NameClassroomMember)
		if err != nil {
			return err
		}

		now := mtime.Now()
		ownerStatus := string(enum.ClassroomMemberStatusTypeActive)

		m := classroom.NewMember()
		m.SetMemberId(memberID)
		m.SetClassroomId(saved.ClassroomId())
		m.SetProfileId(cmd.OwnerProfileID)
		m.SetMemberRole(string(enum.ClassroomMemberRoleTypeOwner))
		m.SetJoinedDt(now)
		m.SetMemberStatus(&ownerStatus)
		m.SetCreateId(cmd.ActorID)

		if _, err := repos.ClassroomMember.Create(ctx, m); err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}

		created = saved
		return nil
	})
	if err != nil {
		return nil, err
	}
	return created, nil
}
