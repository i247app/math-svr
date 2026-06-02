package profile

import (
	"context"
	"strings"

	dto "math-ai.com/math-ai/internal/application/dto/profile"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/shared/enum"
)

// profileSearchMaxLen bounds the LIKE %?% needle so a runaway client
// can't drive an unbounded table scan via an oversized search string.
const profileSearchMaxLen = 128

// validateLanguage allows an empty value (service substitutes vn) and
// otherwise restricts to the project's supported set. Shared by the get and
// list endpoints since both newly accept a language input.
func validateLanguage(ctx context.Context, lang enum.LanguageType) error {
	if lang == "" {
		return nil
	}
	switch lang {
	case enum.LanguageTypeVietnamese, enum.LanguageTypeEnglish:
		return nil
	default:
		return errs.NewError(ctx, status.PROFILE_INVALID_LANGUAGE, nil,
			ErrLanguageMustBeVnOrEn)
	}
}

func ValidateCreateProfile(ctx context.Context, req *dto.CreateProfileReq) error {
	if req.UserID == 0 {
		return errs.NewError(ctx, status.PROFILE_MISSING_USER_ID, nil,
			ErrUserIDRequired)
	}
	if strings.TrimSpace(req.Name) == "" {
		return errs.NewError(ctx, status.PROFILE_MISSING_NAME, nil,
			ErrNameRequired)
	}
	if strings.TrimSpace(req.Avatar) != "" && req.AvatarFile != nil {
		return errs.NewError(ctx, status.PROFILE_AVATAR_CONFLICT, nil,
			ErrProvideEitherAvatarFileOrAvatarReference)
	}
	// if req.ProgramID == nil {
	// 	return errs.NewError(ctx, status.PROFILE_MISSING_PROGRAM_ID, nil,
	// 		ErrProgramIDRequired)
	// }
	// if req.GradeID == nil {
	// 	return errs.NewError(ctx, status.PROFILE_MISSING_GRADE_ID, nil,
	// 		ErrGradeIDRequired)
	// }
	// if req.SemesterID == nil {
	// 	return errs.NewError(ctx, status.PROFILE_MISSING_SEMESTER_ID, nil,
	// 		ErrSemesterIDRequired)
	// }
	return nil
}

func ValidateUpdateProfile(ctx context.Context, req *dto.UpdateProfileReq) error {
	if req.ProfileID == 0 {
		return errs.NewError(ctx, status.PROFILE_NOT_FOUND, nil,
			ErrProfileIDRequired)
	}
	if req.Name != nil && strings.TrimSpace(*req.Name) == "" {
		return errs.NewError(ctx, status.PROFILE_MISSING_NAME, nil,
			ErrNameCannotBeBlank)
	}
	if req.Avatar != nil && strings.TrimSpace(*req.Avatar) == "" {
		return errs.NewError(ctx, status.PROFILE_AVATAR_INVALID_REFERENCE, nil,
			ErrAvatarReferenceMustBeNonEmptyWhenProvided)
	}
	if req.Avatar != nil && strings.TrimSpace(*req.Avatar) != "" && req.AvatarFile != nil {
		return errs.NewError(ctx, status.PROFILE_AVATAR_CONFLICT, nil,
			ErrProvideEitherAvatarFileOrAvatarReference)
	}
	return nil
}

func ValidateGetProfile(ctx context.Context, req *dto.GetProfileByIdReq) error {
	if req.ProfileID == 0 {
		return errs.NewError(ctx, status.PROFILE_NOT_FOUND, nil,
			ErrProfileIDRequired)
	}
	return validateLanguage(ctx, req.Language)
}

// ValidateListProfiles validates and normalises the list filters in
// place. None of the filters are required individually — callers may
// list every active profile. Blank optional values collapse to nil so
// the repo skips their predicate entirely. Role and ProfileStatus are
// validated against the enum sets so a typo surfaces as a typed error
// rather than a silently empty result.
func ValidateListProfiles(ctx context.Context, req *dto.ListProfilesReq) error {
	if req.UserID != nil && *req.UserID == 0 {
		req.UserID = nil
	}
	if req.SchoolID != nil && *req.SchoolID == 0 {
		req.SchoolID = nil
	}
	if req.ProgramID != nil && *req.ProgramID == 0 {
		req.ProgramID = nil
	}
	if req.GradeID != nil && *req.GradeID == 0 {
		req.GradeID = nil
	}
	if req.SemesterID != nil && *req.SemesterID == 0 {
		req.SemesterID = nil
	}
	if req.Role != nil {
		role := strings.TrimSpace(*req.Role)
		if role == "" {
			req.Role = nil
		} else {
			if !enum.RoleProfileType(role).IsValid() {
				return errs.NewError(ctx, status.FAIL, nil,
					ErrRoleInvalid)
			}
			req.Role = &role
		}
	}
	if req.ProfileStatus != nil {
		ps := strings.TrimSpace(*req.ProfileStatus)
		if ps == "" {
			req.ProfileStatus = nil
		} else {
			if !enum.ProfileStatusType(ps).IsValid() {
				return errs.NewError(ctx, status.FAIL, nil,
					ErrProfileStatusInvalid)
			}
			req.ProfileStatus = &ps
		}
	}
	if req.Search != nil {
		s := strings.TrimSpace(*req.Search)
		if s == "" {
			req.Search = nil
		} else if len(s) > profileSearchMaxLen {
			return errs.NewError(ctx, status.FAIL, nil,
				ErrSearchTooLong)
		} else {
			req.Search = &s
		}
	}
	return validateLanguage(ctx, req.Language)
}

func ValidateDeleteProfile(ctx context.Context, req *dto.DeleteProfileReq) error {
	if req.ProfileID == 0 {
		return errs.NewError(ctx, status.PROFILE_NOT_FOUND, nil,
			ErrProfileIDRequired)
	}
	return nil
}

func ValidateAssignSchool(ctx context.Context, req *dto.AssignSchoolReq) error {
	if req.ProfileID == 0 {
		return errs.NewError(ctx, status.PROFILE_NOT_FOUND, nil,
			ErrProfileIDRequired)
	}
	if req.SchoolID == 0 {
		return errs.NewError(ctx, status.SCHOOL_MISSING_ID, nil,
			ErrSchoolIDRequired)
	}
	return nil
}

func ValidateRemoveSchool(ctx context.Context, req *dto.RemoveSchoolReq) error {
	if req.ProfileID == 0 {
		return errs.NewError(ctx, status.PROFILE_NOT_FOUND, nil,
			ErrProfileIDRequired)
	}
	return nil
}
