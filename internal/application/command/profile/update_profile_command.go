package command

import (
	"context"

	"math-ai.com/math-ai/internal/application/transaction"
	"math-ai.com/math-ai/internal/domain/profile"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	mtime "math-ai.com/math-ai/internal/domain/shared/time"
	"math-ai.com/math-ai/internal/shared/enum"
)

type UpdateProfileCommand struct {
	ProfileID  int64
	Name       *string
	Phone      *string
	Email      *string
	Role       *string
	IsDefault  *bool
	Dob        *mtime.MathTime
	SchoolID   *int64
	ProgramID  *int64
	GradeID    *int64
	SemesterID *int64
	IDType     *string // TEACHER only
	TeacherID  *string // TEACHER only
	StudentID  *string // STUDENT only
	Note       *string
	AvatarKey  *string
}

type UpdateProfileCommandHandler struct {
	uow transaction.UnitOfWork
}

func NewUpdateProfileCommandHandler(uow transaction.UnitOfWork) *UpdateProfileCommandHandler {
	return &UpdateProfileCommandHandler{uow: uow}
}

// Handle mutates the existing profile row. Per the project rule, advancing
// a child's (grade, semester) NEVER inserts a new row.
func (h *UpdateProfileCommandHandler) Handle(ctx context.Context, cmd UpdateProfileCommand) (*profile.Profile, error) {
	var updated *profile.Profile

	handler := func(ctx context.Context, repos transaction.Repositories) error {
		existing, err := repos.Profile.FindByProfileId(ctx, cmd.ProfileID)
		if err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}
		if existing == nil {
			return errs.NewError(ctx, status.PROFILE_NOT_FOUND, nil,
				ErrProfileNotFound)
		}

		patch := BuildUpdateProfile(cmd)

		// profile_status is derived from the post-patch state — fields not
		// present in the patch keep their existing value. We compute it
		// here (vs. in BuildUpdateProfile) because we need the existing
		// row to merge the patch over.
		derived := deriveUpdatedProfileStatus(existing, cmd).String()
		patch.SetProfileStatus(&derived)

		if err := repos.Profile.Update(ctx, patch); err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}

		updated, err = repos.Profile.FindByProfileId(ctx, cmd.ProfileID)
		if err != nil {
			return errs.NewError(ctx, status.FAIL, nil, err)
		}

		if cmd.IsDefault != nil && *cmd.IsDefault {
			if err := repos.Profile.MarkDefaultByProfileId(ctx, updated.UserId(), cmd.ProfileID); err != nil {
				return errs.NewError(ctx, status.FAIL, nil, err)
			}
		}

		return nil
	}

	if err := h.uow.Do(ctx, handler); err != nil {
		return nil, err
	}
	return updated, nil
}

func BuildUpdateProfile(cmd UpdateProfileCommand) *profile.Profile {
	patch := profile.NewProfile()
	patch.SetProfileId(cmd.ProfileID)
	if cmd.Name != nil {
		patch.SetName(*cmd.Name)
	}
	if cmd.Phone != nil {
		patch.SetPhone(cmd.Phone)
	}
	if cmd.Email != nil {
		patch.SetEmail(cmd.Email)
	}
	if cmd.Role != nil {
		patch.SetRole(*cmd.Role)
	}
	if cmd.IsDefault != nil {
		patch.SetIsDefault(*cmd.IsDefault)
	}
	if cmd.Dob != nil {
		patch.SetDob(*cmd.Dob)
	}
	if cmd.SchoolID != nil {
		patch.SetSchoolId(cmd.SchoolID)
	}
	if cmd.ProgramID != nil {
		patch.SetProgramId(cmd.ProgramID)
	}
	if cmd.GradeID != nil {
		patch.SetGradeId(cmd.GradeID)
	}
	if cmd.SemesterID != nil {
		patch.SetSemesterId(cmd.SemesterID)
	}
	if cmd.IDType != nil {
		patch.SetIdType(cmd.IDType)
	}
	if cmd.TeacherID != nil {
		patch.SetTeacherId(cmd.TeacherID)
	}
	if cmd.StudentID != nil {
		patch.SetStudentId(cmd.StudentID)
	}
	if cmd.Note != nil {
		patch.SetNote(cmd.Note)
	}
	if cmd.AvatarKey != nil && *cmd.AvatarKey != "" {
		patch.SetAvatarKey(cmd.AvatarKey)
	}
	return patch
}

// deriveUpdatedProfileStatus folds the patch over the existing row, then
// applies DeriveProfileStatus. A nil pointer in the patch means "keep
// the existing column"; a non-nil pointer (even pointing to "") becomes
// the new value, which can downgrade an OFFICIAL profile to INCOMPLETE
// (e.g. role flipped from STUDENT to TEACHER without a teacher_id yet).
func deriveUpdatedProfileStatus(existing *profile.Profile, cmd UpdateProfileCommand) enum.ProfileStatusType {
	role := existing.Role()
	if cmd.Role != nil && *cmd.Role != "" {
		role = *cmd.Role
	}

	idType := derefOrEmpty(existing.IdType())
	if cmd.IDType != nil {
		idType = *cmd.IDType
	}

	teacherId := derefOrEmpty(existing.TeacherId())
	if cmd.TeacherID != nil {
		teacherId = *cmd.TeacherID
	}

	studentId := derefOrEmpty(existing.StudentId())
	if cmd.StudentID != nil {
		studentId = *cmd.StudentID
	}

	return DeriveProfileStatus(role, idType, teacherId, studentId)
}
