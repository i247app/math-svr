package classroom

import (
	"context"
	"strings"
	"time"

	dto "math-ai.com/math-ai/internal/application/dto/classroom"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	mtime "math-ai.com/math-ai/internal/domain/shared/time"
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
	if req.ProfileID == 0 {
		return errs.NewError(ctx, status.CLASSROOM_MISSING_OWNER_PROFILE_ID, nil,
			ErrProfileIDRequired)
	}
	if strings.TrimSpace(req.Name) == "" {
		return errs.NewError(ctx, status.CLASSROOM_MISSING_NAME, nil,
			ErrNameRequired)
	}
	if len(req.Name) > nameMaxLen {
		return errs.NewError(ctx, status.CLASSROOM_NAME_TOO_LONG, nil,
			ErrNameTooLong)
	}
	if req.Description != nil && len(*req.Description) > descriptionMaxLen {
		return errs.NewError(ctx, status.CLASSROOM_DESCRIPTION_TOO_LONG, nil,
			ErrDescriptionTooLong)
	}
	if req.CoverKey != nil && len(*req.CoverKey) > coverKeyMaxLen {
		return errs.NewError(ctx, status.FAIL, nil,
			ErrCoverKeyTooLong)
	}
	if req.Note != nil && len(*req.Note) > noteMaxLen {
		return errs.NewError(ctx, status.FAIL, nil,
			ErrNoteTooLong)
	}
	if req.MaxMembers != nil && *req.MaxMembers <= 0 {
		return errs.NewError(ctx, status.CLASSROOM_INVALID_MAX_MEMBERS, nil,
			ErrMaxMembersInvalid)
	}
	if normalized, err := normalizeProgramIDList(ctx, req.ProgramIDs); err != nil {
		return err
	} else {
		req.ProgramIDs = normalized
	}
	if req.ClassroomCode != nil {
		code := strings.TrimSpace(*req.ClassroomCode)
		if code == "" {
			// Treat a blank pointer as "no code supplied" so callers
			// using multipart forms don't need to omit the field.
			req.ClassroomCode = nil
		} else if len(code) > classroomCodeMaxLen {
			return errs.NewError(ctx, status.CLASSROOM_CODE_INVALID, nil,
				ErrClassroomCodeTooLong)
		} else {
			req.ClassroomCode = &code
		}
	}
	if req.ClassroomCodeExpiresDt != "" {
		parseTime, err := mtime.ParseFromString(req.ClassroomCodeExpiresDt)
		if err != nil {
			return err
		}
		if parseTime.IsValid() && parseTime.Time.Before(time.Now()) {
			return errs.NewError(ctx, status.CLASSROOM_CODE_EXPIRED, nil, ErrClassroomCodeExpiresMustBeFuture)
		}
	}
	return nil
}

func ValidateUpdateClassroom(ctx context.Context, req *dto.UpdateClassroomReq) error {
	if req.ProfileID == 0 {
		return errs.NewError(ctx, status.CLASSROOM_MISSING_OWNER_PROFILE_ID, nil,
			ErrProfileIDRequired)
	}
	if req.ClassroomID == 0 {
		return errs.NewError(ctx, status.CLASSROOM_MISSING_ID, nil,
			ErrClassroomIDRequired)
	}
	if req.Name != nil {
		if strings.TrimSpace(*req.Name) == "" {
			return errs.NewError(ctx, status.CLASSROOM_MISSING_NAME, nil,
				ErrNameBlank)
		}
		if len(*req.Name) > nameMaxLen {
			return errs.NewError(ctx, status.CLASSROOM_NAME_TOO_LONG, nil,
				ErrNameTooLong)
		}
	}
	if req.Description != nil && len(*req.Description) > descriptionMaxLen {
		return errs.NewError(ctx, status.CLASSROOM_DESCRIPTION_TOO_LONG, nil,
			ErrDescriptionTooLong)
	}
	if req.AvatarKey != nil && len(*req.AvatarKey) > coverKeyMaxLen {
		return errs.NewError(ctx, status.FAIL, nil,
			ErrCoverKeyTooLong)
	}
	if req.Note != nil && len(*req.Note) > noteMaxLen {
		return errs.NewError(ctx, status.FAIL, nil,
			ErrNoteTooLong)
	}
	if req.MaxMembers != nil && *req.MaxMembers <= 0 {
		return errs.NewError(ctx, status.CLASSROOM_INVALID_MAX_MEMBERS, nil,
			ErrMaxMembersInvalid)
	}
	if req.ProgramIDs != nil {
		normalized, err := normalizeProgramIDList(ctx, *req.ProgramIDs)
		if err != nil {
			return err
		}
		// Hold onto the same pointer-to-non-nil-slice so the command
		// can still distinguish "clear all" (non-nil empty) from "no-op"
		// (nil). A user-supplied [a, b] with duplicates collapses to
		// the deduped slice in place.
		req.ProgramIDs = &normalized
	}
	return nil
}

// ValidateGetClassroom keeps classroom_id required but treats profile_id
// as optional. When supplied, the service still resolves and
// session-validates it so the per-caller relationship/my_role fields
// hydrate; when absent, the response carries Relationship=NONE and the
// classroom itself is still returned. classroom_id is the only hard
// requirement because the endpoint is fundamentally an id-lookup.
func ValidateGetClassroom(ctx context.Context, req *dto.GetClassroomReq) error {
	if req.ClassroomID == 0 {
		return errs.NewError(ctx, status.CLASSROOM_MISSING_ID, nil,
			ErrClassroomIDRequired)
	}
	return nil
}

func ValidateListClassrooms(ctx context.Context, req *dto.ListClassroomsReq) error {
	// if req.ProfileID == 0 {
	// 	return errs.NewError(ctx, status.CLASSROOM_MISSING_OWNER_PROFILE_ID, nil,
	// 		ErrProfileIDRequired)
	// }
	// Collapse blank optional filters to nil so the repo skips the predicate.
	if req.OwnerProfileID != nil && *req.OwnerProfileID == 0 {
		req.OwnerProfileID = nil
	}
	if req.SchoolID != nil && *req.SchoolID == 0 {
		req.SchoolID = nil
	}
	if req.ProgramID != nil && *req.ProgramID == 0 {
		req.ProgramID = nil
	}
	if len(req.ProgramIDs) > 0 {
		// On the list path we don't surface CLASSROOM_PROGRAM_DUPLICATE
		// — duplicates in the filter are harmless. Just normalize the
		// slice so the repo only has to deal with non-blank, deduped
		// ids.
		seen := make(map[int64]struct{}, len(req.ProgramIDs))
		out := make([]int64, 0, len(req.ProgramIDs))
		for _, id := range req.ProgramIDs {
			if _, ok := seen[id]; ok {
				continue
			}
			seen[id] = struct{}{}
			out = append(out, id)
		}
		req.ProgramIDs = out
	}
	if req.GradeID != nil && *req.GradeID == 0 {
		req.GradeID = nil
	}
	if req.Search != nil {
		s := strings.TrimSpace(*req.Search)
		if s == "" {
			req.Search = nil
		} else if len(s) > searchMaxLen {
			return errs.NewError(ctx, status.FAIL, nil,
				ErrSearchTooLong)
		}
	}
	return nil
}

// ValidateListMyJoinedClassrooms enforces profile_id (the caller must
// disambiguate which of their profiles is asking) and normalises the
// optional search filter the same way ValidateListClassrooms does.
func ValidateListMyJoinedClassrooms(ctx context.Context, req *dto.ListMyJoinedClassroomsReq) error {
	if req.ProfileID == 0 {
		return errs.NewError(ctx, status.CLASSROOM_MISSING_OWNER_PROFILE_ID, nil,
			ErrProfileIDRequired)
	}
	if req.Search != nil {
		s := strings.TrimSpace(*req.Search)
		if s == "" {
			req.Search = nil
		} else if len(s) > searchMaxLen {
			return errs.NewError(ctx, status.FAIL, nil,
				ErrSearchTooLong)
		}
	}
	return nil
}

func ValidateArchiveClassroom(ctx context.Context, req *dto.ArchiveClassroomReq) error {
	if req.ProfileID == 0 {
		return errs.NewError(ctx, status.CLASSROOM_MISSING_OWNER_PROFILE_ID, nil,
			ErrProfileIDRequired)
	}
	if req.ClassroomID == 0 {
		return errs.NewError(ctx, status.CLASSROOM_MISSING_ID, nil,
			ErrClassroomIDRequired)
	}
	return nil
}

func ValidateRestoreClassroom(ctx context.Context, req *dto.RestoreClassroomReq) error {
	if req.ProfileID == 0 {
		return errs.NewError(ctx, status.CLASSROOM_MISSING_OWNER_PROFILE_ID, nil,
			ErrProfileIDRequired)
	}
	if req.ClassroomID == 0 {
		return errs.NewError(ctx, status.CLASSROOM_MISSING_ID, nil,
			ErrClassroomIDRequired)
	}
	return nil
}

func ValidateDeleteClassroom(ctx context.Context, req *dto.DeleteClassroomReq) error {
	if req.ProfileID == 0 {
		return errs.NewError(ctx, status.CLASSROOM_MISSING_OWNER_PROFILE_ID, nil,
			ErrProfileIDRequired)
	}
	if req.ClassroomID == 0 {
		return errs.NewError(ctx, status.CLASSROOM_MISSING_ID, nil,
			ErrClassroomIDRequired)
	}
	return nil
}

// classroomCodeMaxLen mirrors ma_classrooms.classroom_code VARCHAR(16). Used
// by the join-by-code path to reject obviously-malformed input before
// it reaches the repo.
const classroomCodeMaxLen = 16

// normalizeProgramIDList trims, drops blanks, and rejects duplicates
// with CLASSROOM_PROGRAM_DUPLICATE. Used by Create + Update so the
// service-level validatePrograms loop never wastes a roundtrip on a
// duplicate id, and the command's INSERT path never has to handle a
// UNIQUE-violation error from the DB on the happy path.
func normalizeProgramIDList(ctx context.Context, in []int64) ([]int64, error) {
	if len(in) == 0 {
		return []int64{}, nil
	}
	seen := make(map[int64]struct{}, len(in))
	out := make([]int64, 0, len(in))
	for _, raw := range in {
		if _, dup := seen[raw]; dup {
			return nil, errs.NewError(ctx, status.CLASSROOM_PROGRAM_DUPLICATE, nil,
				ErrDuplicateProgramIDInList)
		}
		seen[raw] = struct{}{}
		out = append(out, raw)
	}
	return out, nil
}

// ValidateFindClassroomByCode trims the supplied code and rejects blank
// or over-long input. Mirrors ValidateJoinByCode so the read-only
// preview endpoint surfaces the same CLASSROOM_CODE_INVALID code as
// the join path for malformed input.
func ValidateFindClassroomByCode(ctx context.Context, req *dto.FindClassroomByCodeReq) error {
	code := strings.TrimSpace(req.ClassCode)
	if code == "" {
		return errs.NewError(ctx, status.CLASSROOM_CODE_INVALID, nil,
			ErrClassroomCodeRequired)
	}
	if len(code) > classroomCodeMaxLen {
		return errs.NewError(ctx, status.CLASSROOM_CODE_INVALID, nil,
			ErrClassroomCodeTooLong)
	}
	req.ClassCode = code
	return nil
}

func ValidateJoinByCode(ctx context.Context, req *dto.JoinByCodeReq) error {
	if req.ProfileID == 0 {
		return errs.NewError(ctx, status.CLASSROOM_MISSING_OWNER_PROFILE_ID, nil,
			ErrProfileIDRequired)
	}
	code := strings.TrimSpace(req.ClassroomCode)
	if code == "" {
		return errs.NewError(ctx, status.CLASSROOM_CODE_INVALID, nil,
			ErrClassroomCodeRequired)
	}
	if len(code) > classroomCodeMaxLen {
		return errs.NewError(ctx, status.CLASSROOM_CODE_INVALID, nil,
			ErrClassroomCodeTooLong)
	}
	req.ClassroomCode = code
	return nil
}

func ValidateLeaveClassroom(ctx context.Context, req *dto.LeaveClassroomReq) error {
	if req.ProfileID == 0 {
		return errs.NewError(ctx, status.CLASSROOM_MISSING_OWNER_PROFILE_ID, nil,
			ErrProfileIDRequired)
	}
	if req.ClassroomID == 0 {
		return errs.NewError(ctx, status.CLASSROOM_MISSING_ID, nil,
			ErrClassroomIDRequired)
	}
	return nil
}

func ValidateRemoveMember(ctx context.Context, req *dto.RemoveMemberReq) error {
	if req.ProfileID == 0 {
		return errs.NewError(ctx, status.CLASSROOM_MISSING_OWNER_PROFILE_ID, nil,
			ErrProfileIDRequired)
	}
	if req.ClassroomID == 0 {
		return errs.NewError(ctx, status.CLASSROOM_MISSING_ID, nil,
			ErrClassroomIDRequired)
	}
	if req.TargetProfileID == 0 {
		return errs.NewError(ctx, status.CLASSROOM_MEMBER_MISSING_PROFILE_ID, nil,
			ErrTargetProfileIDRequired)
	}
	if req.TargetProfileID == req.ProfileID {
		return errs.NewError(ctx, status.CLASSROOM_MEMBER_INVALID_ROLE, nil,
			ErrCannotRemoveSelf)
	}
	return nil
}

func ValidateUpdateMemberRole(ctx context.Context, req *dto.UpdateMemberRoleReq) error {
	if req.ProfileID == 0 {
		return errs.NewError(ctx, status.CLASSROOM_MISSING_OWNER_PROFILE_ID, nil,
			ErrProfileIDRequired)
	}
	if req.ClassroomID == 0 {
		return errs.NewError(ctx, status.CLASSROOM_MISSING_ID, nil,
			ErrClassroomIDRequired)
	}
	if req.TargetProfileID == 0 {
		return errs.NewError(ctx, status.CLASSROOM_MEMBER_MISSING_PROFILE_ID, nil,
			ErrTargetProfileIDRequired)
	}
	role := strings.TrimSpace(req.NewRole)
	if role == "" {
		return errs.NewError(ctx, status.CLASSROOM_MEMBER_INVALID_ROLE, nil,
			ErrNewRoleRequired)
	}
	// OWNER is intentionally not allowed here — transfer-ownership is
	// the only way to mint a new owner.
	if role != string(enum.ClassroomMemberRoleTypeCoTeacher) &&
		role != string(enum.ClassroomMemberRoleTypeStudent) {
		return errs.NewError(ctx, status.CLASSROOM_MEMBER_INVALID_ROLE, nil,
			ErrNewRoleInvalid)
	}
	req.NewRole = role
	return nil
}

func ValidateTransferOwnership(ctx context.Context, req *dto.TransferOwnershipReq) error {
	if req.ProfileID == 0 {
		return errs.NewError(ctx, status.CLASSROOM_MISSING_OWNER_PROFILE_ID, nil,
			ErrProfileIDRequired)
	}
	if req.ClassroomID == 0 {
		return errs.NewError(ctx, status.CLASSROOM_MISSING_ID, nil,
			ErrClassroomIDRequired)
	}
	if req.NewOwnerProfileID == 0 {
		return errs.NewError(ctx, status.CLASSROOM_MEMBER_MISSING_PROFILE_ID, nil,
			ErrNewOwnerProfileIDRequired)
	}
	if req.NewOwnerProfileID == req.ProfileID {
		return errs.NewError(ctx, status.CLASSROOM_OWNER_TRANSFER_TO_NON_MEMBER, nil,
			ErrCannotTransferToSelf)
	}
	return nil
}

func ValidateListMembers(ctx context.Context, req *dto.ListMembersReq) error {
	// profile_id is optional — when omitted the service falls back to a
	// profile owned by the authenticated session user that is an ACTIVE
	// member of the classroom. classroom_id stays required.
	if req.ClassroomID == 0 {
		return errs.NewError(ctx, status.CLASSROOM_MISSING_ID, nil,
			ErrClassroomIDRequired)
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
