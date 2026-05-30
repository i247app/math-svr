package classroom

import (
	"context"
	"errors"
	"strings"
	"time"

	dto "math-ai.com/math-ai/internal/application/dto/classroom"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/shared/enum"
)

// Column-width mirrors for migration 014. Keeping the bounds here lets
// the validator reject over-long input with a friendly status code
// instead of letting MySQL truncate.
const (
	nameMaxLen        = 128
	descriptionMaxLen = 500
	noteMaxLen        = 500
	coverKeyMaxLen    = 256
	searchMaxLen      = 128
)

func ValidateCreateClassroom(ctx context.Context, req *dto.CreateClassroomReq) error {
	if strings.TrimSpace(req.ProfileID) == "" {
		return errs.NewError(ctx, status.CLASSROOM_MISSING_OWNER_PROFILE_ID, nil,
			errors.New("profile_id is required"))
	}
	if strings.TrimSpace(req.Name) == "" {
		return errs.NewError(ctx, status.CLASSROOM_MISSING_NAME, nil,
			errors.New("name is required"))
	}
	if len(req.Name) > nameMaxLen {
		return errs.NewError(ctx, status.CLASSROOM_NAME_TOO_LONG, nil,
			errors.New("name too long"))
	}
	if req.Description != nil && len(*req.Description) > descriptionMaxLen {
		return errs.NewError(ctx, status.CLASSROOM_DESCRIPTION_TOO_LONG, nil,
			errors.New("description too long"))
	}
	if req.CoverKey != nil && len(*req.CoverKey) > coverKeyMaxLen {
		return errs.NewError(ctx, status.FAIL, nil,
			errors.New("cover_key too long"))
	}
	if req.Note != nil && len(*req.Note) > noteMaxLen {
		return errs.NewError(ctx, status.FAIL, nil,
			errors.New("note too long"))
	}
	if req.MaxMembers != nil && *req.MaxMembers <= 0 {
		return errs.NewError(ctx, status.CLASSROOM_INVALID_MAX_MEMBERS, nil,
			errors.New("max_members must be > 0"))
	}
	if req.InviteCode != nil {
		code := strings.TrimSpace(*req.InviteCode)
		if code == "" {
			// Treat a blank pointer as "no code supplied" so callers
			// using multipart forms don't need to omit the field.
			req.InviteCode = nil
		} else if len(code) > inviteCodeMaxLen {
			return errs.NewError(ctx, status.CLASSROOM_INVITE_CODE_INVALID, nil,
				errors.New("invite_code too long"))
		} else {
			req.InviteCode = &code
		}
	}
	if req.InviteCodeExpiresDt.IsValid() && req.InviteCodeExpiresDt.Time.Before(time.Now()) {
		return errs.NewError(ctx, status.CLASSROOM_INVITE_CODE_EXPIRED, nil,
			errors.New("invite_code_expires_dt must be in the future"))
	}
	return nil
}

func ValidateUpdateClassroom(ctx context.Context, req *dto.UpdateClassroomReq) error {
	if strings.TrimSpace(req.ProfileID) == "" {
		return errs.NewError(ctx, status.CLASSROOM_MISSING_OWNER_PROFILE_ID, nil,
			errors.New("profile_id is required"))
	}
	if strings.TrimSpace(req.ClassroomID) == "" {
		return errs.NewError(ctx, status.CLASSROOM_MISSING_ID, nil,
			errors.New("classroom_id is required"))
	}
	if req.Name != nil {
		if strings.TrimSpace(*req.Name) == "" {
			return errs.NewError(ctx, status.CLASSROOM_MISSING_NAME, nil,
				errors.New("name cannot be blank"))
		}
		if len(*req.Name) > nameMaxLen {
			return errs.NewError(ctx, status.CLASSROOM_NAME_TOO_LONG, nil,
				errors.New("name too long"))
		}
	}
	if req.Description != nil && len(*req.Description) > descriptionMaxLen {
		return errs.NewError(ctx, status.CLASSROOM_DESCRIPTION_TOO_LONG, nil,
			errors.New("description too long"))
	}
	if req.AvatarKey != nil && len(*req.AvatarKey) > coverKeyMaxLen {
		return errs.NewError(ctx, status.FAIL, nil,
			errors.New("cover_key too long"))
	}
	if req.Note != nil && len(*req.Note) > noteMaxLen {
		return errs.NewError(ctx, status.FAIL, nil,
			errors.New("note too long"))
	}
	if req.MaxMembers != nil && *req.MaxMembers <= 0 {
		return errs.NewError(ctx, status.CLASSROOM_INVALID_MAX_MEMBERS, nil,
			errors.New("max_members must be > 0"))
	}
	return nil
}

func ValidateGetClassroom(ctx context.Context, req *dto.GetClassroomReq) error {
	if strings.TrimSpace(req.ProfileID) == "" {
		return errs.NewError(ctx, status.CLASSROOM_MISSING_OWNER_PROFILE_ID, nil,
			errors.New("profile_id is required"))
	}
	if strings.TrimSpace(req.ClassroomID) == "" {
		return errs.NewError(ctx, status.CLASSROOM_MISSING_ID, nil,
			errors.New("classroom_id is required"))
	}
	return nil
}

func ValidateListClassrooms(ctx context.Context, req *dto.ListClassroomsReq) error {
	if strings.TrimSpace(req.ProfileID) == "" {
		return errs.NewError(ctx, status.CLASSROOM_MISSING_OWNER_PROFILE_ID, nil,
			errors.New("profile_id is required"))
	}
	// Collapse blank optional filters to nil so the repo skips the predicate.
	if req.OwnerProfileID != nil && strings.TrimSpace(*req.OwnerProfileID) == "" {
		req.OwnerProfileID = nil
	}
	if req.SchoolID != nil && strings.TrimSpace(*req.SchoolID) == "" {
		req.SchoolID = nil
	}
	if req.ProgramID != nil && strings.TrimSpace(*req.ProgramID) == "" {
		req.ProgramID = nil
	}
	if req.GradeID != nil && strings.TrimSpace(*req.GradeID) == "" {
		req.GradeID = nil
	}
	if req.Search != nil {
		s := strings.TrimSpace(*req.Search)
		if s == "" {
			req.Search = nil
		} else if len(s) > searchMaxLen {
			return errs.NewError(ctx, status.FAIL, nil,
				errors.New("search too long"))
		}
	}
	return nil
}

func ValidateArchiveClassroom(ctx context.Context, req *dto.ArchiveClassroomReq) error {
	if strings.TrimSpace(req.ProfileID) == "" {
		return errs.NewError(ctx, status.CLASSROOM_MISSING_OWNER_PROFILE_ID, nil,
			errors.New("profile_id is required"))
	}
	if strings.TrimSpace(req.ClassroomID) == "" {
		return errs.NewError(ctx, status.CLASSROOM_MISSING_ID, nil,
			errors.New("classroom_id is required"))
	}
	return nil
}

func ValidateRestoreClassroom(ctx context.Context, req *dto.RestoreClassroomReq) error {
	if strings.TrimSpace(req.ProfileID) == "" {
		return errs.NewError(ctx, status.CLASSROOM_MISSING_OWNER_PROFILE_ID, nil,
			errors.New("profile_id is required"))
	}
	if strings.TrimSpace(req.ClassroomID) == "" {
		return errs.NewError(ctx, status.CLASSROOM_MISSING_ID, nil,
			errors.New("classroom_id is required"))
	}
	return nil
}

func ValidateDeleteClassroom(ctx context.Context, req *dto.DeleteClassroomReq) error {
	if strings.TrimSpace(req.ProfileID) == "" {
		return errs.NewError(ctx, status.CLASSROOM_MISSING_OWNER_PROFILE_ID, nil,
			errors.New("profile_id is required"))
	}
	if strings.TrimSpace(req.ClassroomID) == "" {
		return errs.NewError(ctx, status.CLASSROOM_MISSING_ID, nil,
			errors.New("classroom_id is required"))
	}
	return nil
}

// inviteCodeMaxLen mirrors ma_classrooms.invite_code VARCHAR(16). Used
// by the join-by-code path to reject obviously-malformed input before
// it reaches the repo.
const inviteCodeMaxLen = 16

func ValidateJoinByCode(ctx context.Context, req *dto.JoinByCodeReq) error {
	if strings.TrimSpace(req.ProfileID) == "" {
		return errs.NewError(ctx, status.CLASSROOM_MISSING_OWNER_PROFILE_ID, nil,
			errors.New("profile_id is required"))
	}
	code := strings.TrimSpace(req.InviteCode)
	if code == "" {
		return errs.NewError(ctx, status.CLASSROOM_INVITE_CODE_INVALID, nil,
			errors.New("invite_code is required"))
	}
	if len(code) > inviteCodeMaxLen {
		return errs.NewError(ctx, status.CLASSROOM_INVITE_CODE_INVALID, nil,
			errors.New("invite_code too long"))
	}
	req.InviteCode = code
	return nil
}

func ValidateLeaveClassroom(ctx context.Context, req *dto.LeaveClassroomReq) error {
	if strings.TrimSpace(req.ProfileID) == "" {
		return errs.NewError(ctx, status.CLASSROOM_MISSING_OWNER_PROFILE_ID, nil,
			errors.New("profile_id is required"))
	}
	if strings.TrimSpace(req.ClassroomID) == "" {
		return errs.NewError(ctx, status.CLASSROOM_MISSING_ID, nil,
			errors.New("classroom_id is required"))
	}
	return nil
}

func ValidateRemoveMember(ctx context.Context, req *dto.RemoveMemberReq) error {
	if strings.TrimSpace(req.ProfileID) == "" {
		return errs.NewError(ctx, status.CLASSROOM_MISSING_OWNER_PROFILE_ID, nil,
			errors.New("profile_id is required"))
	}
	if strings.TrimSpace(req.ClassroomID) == "" {
		return errs.NewError(ctx, status.CLASSROOM_MISSING_ID, nil,
			errors.New("classroom_id is required"))
	}
	if strings.TrimSpace(req.TargetProfileID) == "" {
		return errs.NewError(ctx, status.CLASSROOM_MEMBER_MISSING_PROFILE_ID, nil,
			errors.New("target_profile_id is required"))
	}
	if strings.TrimSpace(req.TargetProfileID) == strings.TrimSpace(req.ProfileID) {
		return errs.NewError(ctx, status.CLASSROOM_MEMBER_INVALID_ROLE, nil,
			errors.New("cannot remove yourself; use leave instead"))
	}
	return nil
}

func ValidateUpdateMemberRole(ctx context.Context, req *dto.UpdateMemberRoleReq) error {
	if strings.TrimSpace(req.ProfileID) == "" {
		return errs.NewError(ctx, status.CLASSROOM_MISSING_OWNER_PROFILE_ID, nil,
			errors.New("profile_id is required"))
	}
	if strings.TrimSpace(req.ClassroomID) == "" {
		return errs.NewError(ctx, status.CLASSROOM_MISSING_ID, nil,
			errors.New("classroom_id is required"))
	}
	if strings.TrimSpace(req.TargetProfileID) == "" {
		return errs.NewError(ctx, status.CLASSROOM_MEMBER_MISSING_PROFILE_ID, nil,
			errors.New("target_profile_id is required"))
	}
	role := strings.TrimSpace(req.NewRole)
	if role == "" {
		return errs.NewError(ctx, status.CLASSROOM_MEMBER_INVALID_ROLE, nil,
			errors.New("new_role is required"))
	}
	// OWNER is intentionally not allowed here — transfer-ownership is
	// the only way to mint a new owner.
	if role != string(enum.ClassroomMemberRoleTypeCoTeacher) &&
		role != string(enum.ClassroomMemberRoleTypeStudent) {
		return errs.NewError(ctx, status.CLASSROOM_MEMBER_INVALID_ROLE, nil,
			errors.New("new_role must be CO_TEACHER or STUDENT"))
	}
	req.NewRole = role
	return nil
}

func ValidateTransferOwnership(ctx context.Context, req *dto.TransferOwnershipReq) error {
	if strings.TrimSpace(req.ProfileID) == "" {
		return errs.NewError(ctx, status.CLASSROOM_MISSING_OWNER_PROFILE_ID, nil,
			errors.New("profile_id is required"))
	}
	if strings.TrimSpace(req.ClassroomID) == "" {
		return errs.NewError(ctx, status.CLASSROOM_MISSING_ID, nil,
			errors.New("classroom_id is required"))
	}
	if strings.TrimSpace(req.NewOwnerProfileID) == "" {
		return errs.NewError(ctx, status.CLASSROOM_MEMBER_MISSING_PROFILE_ID, nil,
			errors.New("new_owner_profile_id is required"))
	}
	if strings.TrimSpace(req.NewOwnerProfileID) == strings.TrimSpace(req.ProfileID) {
		return errs.NewError(ctx, status.CLASSROOM_OWNER_TRANSFER_TO_NON_MEMBER, nil,
			errors.New("cannot transfer ownership to yourself"))
	}
	return nil
}

func ValidateListMembers(ctx context.Context, req *dto.ListMembersReq) error {
	if strings.TrimSpace(req.ProfileID) == "" {
		return errs.NewError(ctx, status.CLASSROOM_MISSING_OWNER_PROFILE_ID, nil,
			errors.New("profile_id is required"))
	}
	if strings.TrimSpace(req.ClassroomID) == "" {
		return errs.NewError(ctx, status.CLASSROOM_MISSING_ID, nil,
			errors.New("classroom_id is required"))
	}
	if req.Role != nil {
		r := strings.TrimSpace(*req.Role)
		if r == "" {
			req.Role = nil
		} else {
			req.Role = &r
		}
	}
	if req.Status != nil {
		s := strings.TrimSpace(*req.Status)
		if s == "" {
			req.Status = nil
		} else {
			req.Status = &s
		}
	}
	return nil
}
