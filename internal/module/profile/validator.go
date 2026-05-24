package profile

import (
	"context"
	"errors"
	"strings"

	dto "math-ai.com/math-ai/internal/application/dto/profile"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/shared/enum"
)

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
			errors.New("language must be 'vn' or 'en'"))
	}
}

func ValidateCreateProfile(ctx context.Context, req *dto.CreateProfileReq) error {
	if req.UserID == "" {
		return errs.NewError(ctx, status.PROFILE_MISSING_USER_ID, nil,
			errors.New("user_id is required"))
	}
	if strings.TrimSpace(req.Name) == "" {
		return errs.NewError(ctx, status.PROFILE_MISSING_NAME, nil,
			errors.New("name is required"))
	}
	if req.ProgramID == nil {
		return errs.NewError(ctx, status.PROFILE_MISSING_PROGRAM_ID, nil,
			errors.New("program_id is required"))
	}
	if req.GradeID == nil {
		return errs.NewError(ctx, status.PROFILE_MISSING_GRADE_ID, nil,
			errors.New("grade_id is required"))
	}
	if req.SemesterID == nil {
		return errs.NewError(ctx, status.PROFILE_MISSING_SEMESTER_ID, nil,
			errors.New("semester_id is required"))
	}
	return nil
}

func ValidateUpdateProfile(ctx context.Context, req *dto.UpdateProfileReq) error {
	if req.ProfileID == "" {
		return errs.NewError(ctx, status.PROFILE_NOT_FOUND, nil,
			errors.New("profile_id is required"))
	}
	if req.Name != nil && strings.TrimSpace(*req.Name) == "" {
		return errs.NewError(ctx, status.PROFILE_MISSING_NAME, nil,
			errors.New("name cannot be blank"))
	}
	return nil
}

func ValidateGetProfile(ctx context.Context, req *dto.GetProfileByIdReq) error {
	if req.ProfileID == "" {
		return errs.NewError(ctx, status.PROFILE_NOT_FOUND, nil,
			errors.New("profile_id is required"))
	}
	return validateLanguage(ctx, req.Language)
}

func ValidateListProfiles(ctx context.Context, req *dto.ListProfilesReq) error {
	if req.UserID == "" {
		return errs.NewError(ctx, status.PROFILE_MISSING_USER_ID, nil,
			errors.New("user_id is required"))
	}
	return validateLanguage(ctx, req.Language)
}

func ValidateDeleteProfile(ctx context.Context, req *dto.DeleteProfileReq) error {
	if req.ProfileID == "" {
		return errs.NewError(ctx, status.PROFILE_NOT_FOUND, nil,
			errors.New("profile_id is required"))
	}
	return nil
}
