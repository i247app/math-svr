package command

import (
	"context"

	"math-ai.com/math-ai/internal/application/transaction"
	"math-ai.com/math-ai/internal/domain/profile"
	"math-ai.com/math-ai/internal/domain/seq"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	mtime "math-ai.com/math-ai/internal/domain/shared/time"
)

type CreateProfileCommand struct {
	UserID     int64
	Name       string
	Role       string
	IsDefault  bool
	Dob        *mtime.MathTime
	SchoolID   *int64
	ProgramID  *int64
	GradeID    *int64
	SemesterID *int64
	IDType     *string // TEACHER only
	TeacherID  *string // TEACHER only
	StudentID  *string // STUDENT only
	AvatarKey  *string
	Note       *string
}

type CreateProfileCommandHandler struct {
	uow transaction.UnitOfWork
}

func NewCreateProfileCommandHandler(uow transaction.UnitOfWork) *CreateProfileCommandHandler {
	return &CreateProfileCommandHandler{uow: uow}
}

func (h *CreateProfileCommandHandler) Handle(ctx context.Context, cmd CreateProfileCommand) (*profile.Profile, error) {
	var created *profile.Profile

	handler := func(ctx context.Context, repos transaction.Repositories) error {
		profileID, err := nextSeqID(ctx, repos, seq.NameUser)
		if err != nil {
			return err
		}
		// Mint a fresh AA-1234 profile_code inside the same tx so the
		// FindByProfileCode probe and the subsequent INSERT see one
		// snapshot. The DB UNIQUE constraint backstops the race
		// window; the generator retries on collision.
		profileCode, err := mintUniqueProfileCode(ctx, repos)
		if err != nil {
			return err
		}

		p := BuildProfile(cmd)
		p.SetProfileId(profileID)
		p.SetProfileCode(profileCode)

		saved, err := repos.Profile.Create(ctx, p)
		if err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}

		if cmd.IsDefault {
			if err := repos.Profile.MarkDefaultByProfileId(ctx, cmd.UserID, profileID); err != nil {
				return errs.NewError(ctx, status.FAIL, nil, err)
			}
		}

		created = saved
		created.SetIsDefault(cmd.IsDefault)
		return nil
	}

	if err := h.uow.Do(ctx, handler); err != nil {
		return nil, err
	}
	return created, nil
}

func BuildProfile(cmd CreateProfileCommand) *profile.Profile {
	p := profile.NewProfile()
	p.SetUserId(cmd.UserID)
	p.SetName(cmd.Name)
	p.SetRole(cmd.Role)
	p.SetSchoolId(cmd.SchoolID)
	p.SetProgramId(cmd.ProgramID)
	p.SetGradeId(cmd.GradeID)
	p.SetSemesterId(cmd.SemesterID)
	p.SetIdType(cmd.IDType)
	p.SetTeacherId(cmd.TeacherID)
	p.SetStudentId(cmd.StudentID)
	p.SetAvatarKey(cmd.AvatarKey)
	p.SetNote(cmd.Note)
	if cmd.Dob != nil {
		p.SetDob(*cmd.Dob)
	}

	// Derive profile_status from the role + identity fields the caller
	// provided. The DB column defaults to INCOMPLETE, but we set it
	// explicitly so the entity returned to the caller carries it too.
	derived := DeriveProfileStatus(cmd.Role,
		derefOrEmpty(cmd.IDType),
		derefOrEmpty(cmd.TeacherID),
		derefOrEmpty(cmd.StudentID),
	).String()
	p.SetProfileStatus(&derived)

	return p
}
