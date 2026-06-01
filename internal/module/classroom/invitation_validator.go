package classroom

import (
	"context"
	"errors"
	"strings"

	dto "math-ai.com/math-ai/internal/application/dto/classroom"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
)

const (
	// identifierMaxLen mirrors ma_classroom_members.note for safety and
	// keeps free-form identifier strings (emails, phones) bounded.
	identifierMaxLen = 128
	// invitationNoteMaxLen mirrors ma_classroom_members.note VARCHAR(500).
	invitationNoteMaxLen = 500
)

// ValidateSendInvitation normalizes the request in place: per-target
// trims, default ProposedRole = STUDENT, drops blank duplicates by
// (identifier_type, identifier) so the bulk batch can't double-count a
// recipient sent twice in the same call.
func ValidateSendInvitation(ctx context.Context, req *dto.SendInvitationReq) error {
	if req.ProfileID == 0 {
		return errs.NewError(ctx, status.CLASSROOM_MISSING_OWNER_PROFILE_ID, nil,
			errors.New("profile_id is required"))
	}
	if req.ClassroomID == 0 {
		return errs.NewError(ctx, status.CLASSROOM_MISSING_ID, nil,
			errors.New("classroom_id is required"))
	}
	if len(req.Targets) == 0 {
		return errs.NewError(ctx, status.CLASSROOM_INVITATION_MISSING_TARGET, nil,
			errors.New("at least one target is required"))
	}
	if req.Note != nil {
		n := strings.TrimSpace(*req.Note)
		if n == "" {
			req.Note = nil
		} else if len(n) > invitationNoteMaxLen {
			return errs.NewError(ctx, status.FAIL, nil,
				errors.New("note too long"))
		} else {
			req.Note = &n
		}
	}

	// seen := make(map[string]struct{}, len(req.Targets))
	// normalized := make([]dto.InvitationTarget, 0, len(req.Targets))
	// for _, t := range req.Targets {
	// 	// t := req.Targets[i]
	// 	t.IdentifierType = strings.TrimSpace(strings.ToUpper(t.IdentifierType))
	// 	t.Identifier = strings.TrimSpace(t.Identifier)
	// 	t.ProposedRole = strings.TrimSpace(strings.ToUpper(t.ProposedRole))

	// 	if t.Identifier == "" {
	// 		return errs.NewError(ctx, status.CLASSROOM_INVITATION_INVALID_IDENTIFIER, nil,
	// 			errors.New("identifier is required"))
	// 	}
	// 	if len(t.Identifier) > identifierMaxLen {
	// 		return errs.NewError(ctx, status.CLASSROOM_INVITATION_INVALID_IDENTIFIER, nil,
	// 			errors.New("identifier too long"))
	// 	}
	// 	if !enum.ClassroomInviteeIdentifierType(t.IdentifierType).IsValid() {
	// 		return errs.NewError(ctx, status.CLASSROOM_INVITATION_INVALID_IDENTIFIER_TYPE, nil,
	// 			errors.New("identifier_type must be EMAIL, PHONE, or PROFILE_ID"))
	// 	}
	// 	if t.ProposedRole == "" {
	// 		t.ProposedRole = string(enum.ClassroomMemberRoleTypeStudent)
	// 	}
	// 	// OWNER cannot be proposed — transfer-ownership is the only path.
	// 	if t.ProposedRole != string(enum.ClassroomMemberRoleTypeStudent) &&
	// 		t.ProposedRole != string(enum.ClassroomMemberRoleTypeCoTeacher) {
	// 		return errs.NewError(ctx, status.CLASSROOM_INVITATION_INVALID_ROLE, nil,
	// 			errors.New("proposed_role must be STUDENT or CO_TEACHER"))
	// 	}

	// 	dedupKey := t.IdentifierType + "|" + strings.ToLower(t.Identifier)
	// 	if _, dup := seen[dedupKey]; dup {
	// 		continue
	// 	}
	// 	seen[dedupKey] = struct{}{}
	// 	normalized = append(normalized, t)
	// }
	// req.Targets = normalized
	return nil
}

func ValidateListMyPendingInvitations(ctx context.Context, req *dto.ListMyPendingInvitationsReq) error {
	if req.ProfileID == 0 {
		return errs.NewError(ctx, status.CLASSROOM_MISSING_OWNER_PROFILE_ID, nil,
			errors.New("profile_id is required"))
	}
	return nil
}

func ValidateListClassroomInvitations(ctx context.Context, req *dto.ListClassroomInvitationsReq) error {
	if req.ProfileID == 0 {
		return errs.NewError(ctx, status.CLASSROOM_MISSING_OWNER_PROFILE_ID, nil,
			errors.New("profile_id is required"))
	}
	if req.ClassroomID == 0 {
		return errs.NewError(ctx, status.CLASSROOM_MISSING_ID, nil,
			errors.New("classroom_id is required"))
	}
	return nil
}

func ValidateAcceptInvitation(ctx context.Context, req *dto.AcceptInvitationReq) error {
	if req.InviterProfileID == 0 {
		return errs.NewError(ctx, status.CLASSROOM_MISSING_OWNER_PROFILE_ID, nil,
			errors.New("profile_id is required"))
	}
	if req.ClassroomID == 0 {
		return errs.NewError(ctx, status.CLASSROOM_MISSING_ID, nil,
			errors.New("classroom_id is required"))
	}
	return nil
}

func ValidateRejectInvitation(ctx context.Context, req *dto.RejectInvitationReq) error {
	if req.InviterProfileID == 0 {
		return errs.NewError(ctx, status.CLASSROOM_MISSING_OWNER_PROFILE_ID, nil,
			errors.New("profile_id is required"))
	}
	if req.ClassroomID == 0 {
		return errs.NewError(ctx, status.CLASSROOM_MISSING_ID, nil,
			errors.New("classroom_id is required"))
	}
	return nil
}

func ValidateCancelInvitation(ctx context.Context, req *dto.CancelInvitationReq) error {
	if req.ProfileID == 0 {
		return errs.NewError(ctx, status.CLASSROOM_MISSING_OWNER_PROFILE_ID, nil,
			errors.New("profile_id is required"))
	}
	if req.ClassroomID == 0 {
		return errs.NewError(ctx, status.CLASSROOM_MISSING_ID, nil,
			errors.New("classroom_id is required"))
	}
	if req.TargetProfileID == 0 {
		return errs.NewError(ctx, status.CLASSROOM_MEMBER_MISSING_PROFILE_ID, nil,
			errors.New("target_profile_id is required"))
	}
	return nil
}
