package command

import (
	"context"
	"strings"

	"math-ai.com/math-ai/internal/application/transaction"
	"math-ai.com/math-ai/internal/domain/classroom"
	"math-ai.com/math-ai/internal/domain/seq"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	mtime "math-ai.com/math-ai/internal/domain/shared/time"
	"math-ai.com/math-ai/internal/shared/enum"
)

// CreateClassroomCommand creates the classroom row, the OWNER member
// row, AND any ma_classroom_programs junction rows in a single UoW.
// Role/identity checks live at the module level — here we trust the
// caller and focus on atomicity: a classroom always ships with exactly
// one OWNER on disk, and its program associations land or fail as a
// unit. ProgramIDs may be empty; the validator dedupes before this
// runs, and the DB UNIQUE on (classroom_id, program_id) remains the
// hard backstop.
type CreateClassroomCommand struct {
	ActorID                *int64
	OwnerProfileID         int64
	Name                   string
	Description            *string
	SchoolID               *int64
	ProgramIDs             []int64
	GradeID                *int64
	MaxMembers             *int64
	CoverKey               *string
	Note                   *string
	ClassroomCode          *string
	ClassroomCodeExpiresDt string
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
		// Resolve classroom_code. If the client supplied one, precheck for
		// uniqueness so the friendlier CLASSROOM_CODE_TAKEN
		// propagates instead of a generic repo-level constraint error.
		// If they did not, mint an "AA-1111" code and retry on the rare
		// collision (≈ 1 in 6.76M per attempt). The DB UNIQUE key
		// remains the hard backstop for the race window in both paths.
		var classroomCodeArg *string
		if cmd.ClassroomCode != nil {
			trimmed := strings.TrimSpace(*cmd.ClassroomCode)
			if trimmed != "" {
				existing, err := repos.Classroom.FindByClassroomCode(ctx, trimmed)
				if err != nil {
					return errs.NewError(ctx, status.FAIL, nil, err)
				}
				if existing != nil {
					return errs.NewError(ctx, status.CLASSROOM_CODE_TAKEN, nil,
						ErrClassroomCodeTaken)
				}
				classroomCodeArg = &trimmed
			}
		}
		if classroomCodeArg == nil {
			minted, err := mintUniqueClassroomCode(ctx, repos)
			if err != nil {
				return err
			}
			classroomCodeArg = &minted
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
		c.SetGradeId(cmd.GradeID)
		c.SetClassroomCode(classroomCodeArg)
		if cmd.ClassroomCodeExpiresDt != "" {
			parseTime, err := mtime.ParseFromString(cmd.ClassroomCodeExpiresDt)
			if err != nil {
				return errs.NewError(ctx, status.FAIL, nil, err)
			}
			c.SetClassroomCodeExpiresDt(parseTime)
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
				ErrClassroomNotFoundAfterInsert)
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

		if len(cmd.ProgramIDs) > 0 {
			insertedProgramIDs, err := insertClassroomPrograms(ctx, repos,
				saved.ClassroomId(), cmd.ProgramIDs, cmd.ActorID)
			if err != nil {
				return err
			}
			saved.SetProgramIds(insertedProgramIDs)
		} else {
			// Hydrate as the empty slice (not nil) so DomainToResponse
			// emits []string{} rather than omitting the field.
			saved.SetProgramIds([]int64{})
		}

		created = saved
		return nil
	})
	if err != nil {
		return nil, err
	}
	return created, nil
}
