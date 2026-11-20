package profile

import (
	"context"
	"strings"

	"math-ai.com/math-ai/internal/application/dto"
	domain "math-ai.com/math-ai/internal/domain/profile"
	"math-ai.com/math-ai/internal/domain/school"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/shared/enum"
	errs "math-ai.com/math-ai/internal/shared/errs"
)

// profileSearchMaxLen bounds the LIKE %?% needle so a runaway client
// can't drive an unbounded table scan via an oversized search string.
const profileSearchMaxLen = 128

func ValidateCreateProfile(ctx context.Context, req *dto.CreateProfileReq) error {
	if req.UID == 0 {
		return errs.NewError(ctx, domain.ProfileMissingUserID, nil,
			ErrUIDRequired)
	}
	if strings.TrimSpace(req.Name) == "" {
		return errs.NewError(ctx, domain.ProfileMissingName, nil,
			ErrNameRequired)
	}
	if strings.TrimSpace(req.Avatar) != "" && req.AvatarFile != nil {
		return errs.NewError(ctx, domain.ProfileAvatarConflict, nil,
			ErrProvideEitherAvatarFileOrAvatarReference)
	}
	// if req.ProgramID == nil {
	// 	return errs.NewError(ctx, domain.ProfileMissingProgramID, nil,
	// 		ErrProgramIDRequired)
	// }
	// if req.GradeID == nil {
	// 	return errs.NewError(ctx, domain.ProfileMissingGradeID, nil,
	// 		ErrGradeIDRequired)
	// }
	// if req.SemesterID == nil {
	// 	return errs.NewError(ctx, domain.ProfileMissingSemesterID, nil,
	// 		ErrSemesterIDRequired)
	// }
	return nil
}

func ValidateUpdateProfile(ctx context.Context, req *dto.UpdateProfileReq) error {
	if req.ProfileID == 0 {
		return errs.NewError(ctx, domain.ProfileNotFound, nil,
			ErrProfileIDRequired)
	}
	if req.Name != nil && strings.TrimSpace(*req.Name) == "" {
		return errs.NewError(ctx, domain.ProfileMissingName, nil,
			ErrNameCannotBeBlank)
	}
	if req.Avatar != nil && strings.TrimSpace(*req.Avatar) == "" {
		return errs.NewError(ctx, domain.ProfileAvatarInvalidReference, nil,
			ErrAvatarReferenceMustBeNonEmptyWhenProvided)
	}
	if req.Avatar != nil && strings.TrimSpace(*req.Avatar) != "" && req.AvatarFile != nil {
		return errs.NewError(ctx, domain.ProfileAvatarConflict, nil,
			ErrProvideEitherAvatarFileOrAvatarReference)
	}
	return nil
}

func ValidateGetProfile(ctx context.Context, req *dto.GetProfileByIdReq) error {
	if req.ProfileID == 0 {
		return errs.NewError(ctx, domain.ProfileNotFound, nil,
			ErrProfileIDRequired)
	}

	return nil
}

// ValidateListProfiles validates and normalises the list filters in
// place. None of the filters are required individually — callers may
// list every active profile. Blank optional values collapse to nil so
// the repo skips their predicate entirely. Role and ProfileStatus are
// validated against the enum sets so a typo surfaces as a typed error
// rather than a silently empty result.
func ValidateListProfiles(ctx context.Context, req *dto.ListProfilesReq) error {
	if req.UID != nil && *req.UID == 0 {
		req.UID = nil
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
			if !enum.RoleType(role).IsSelfAssignable() {
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

	return nil
}

func ValidateDeleteProfile(ctx context.Context, req *dto.DeleteProfileReq) error {
	if req.ProfileID == 0 {
		return errs.NewError(ctx, domain.ProfileNotFound, nil,
			ErrProfileIDRequired)
	}
	return nil
}

func ValidateAssignSchool(ctx context.Context, req *dto.AssignSchoolReq) error {
	if req.ProfileID == 0 {
		return errs.NewError(ctx, domain.ProfileNotFound, nil,
			ErrProfileIDRequired)
	}
	if req.SchoolID == 0 {
		return errs.NewError(ctx, school.SchoolMissingID, nil,
			ErrSchoolIDRequired)
	}
	return nil
}

func ValidateRemoveSchool(ctx context.Context, req *dto.RemoveSchoolReq) error {
	if req.ProfileID == 0 {
		return errs.NewError(ctx, domain.ProfileNotFound, nil,
			ErrProfileIDRequired)
	}
	return nil
}
