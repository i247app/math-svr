package classroom

import (
	"context"
	"strings"

	"math-ai.com/math-ai/internal/application/dto"
	domain "math-ai.com/math-ai/internal/domain/classroom"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/shared/errs"
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
		return errs.NewError(ctx, domain.ClassroomMissingOwnerProfileID, nil,
			ErrProfileIDRequired)
	}
	if req.ClassroomID == 0 {
		return errs.NewError(ctx, domain.ClassroomMissingID, nil,
			ErrClassroomIDRequired)
	}
	if len(req.Targets) == 0 {
		return errs.NewError(ctx, domain.ClassroomInvitationMissingTarget, nil,
			ErrAtLeastOneTargetRequired)
	}
	if req.Note != nil {
		n := strings.TrimSpace(*req.Note)
		if n == "" {
			req.Note = nil
		} else if len(n) > invitationNoteMaxLen {
			return errs.NewError(ctx, status.FAIL, nil,
				ErrNoteTooLong)
		} else {
			req.Note = &n
		}
	}

	return nil
}

func ValidateListMyPendingInvitations(ctx context.Context, req *dto.ListMyPendingInvitationsReq) error {
	if req.ProfileID == 0 {
		return errs.NewError(ctx, domain.ClassroomMissingOwnerProfileID, nil, ErrProfileIDRequired)
	}
	return nil
}

func ValidateListClassroomInvitations(ctx context.Context, req *dto.ListClassroomInvitationsReq) error {
	if req.ProfileID == 0 {
		return errs.NewError(ctx, domain.ClassroomMissingOwnerProfileID, nil, ErrProfileIDRequired)
	}
	if req.ClassroomID == 0 {
		return errs.NewError(ctx, domain.ClassroomMissingID, nil, ErrClassroomIDRequired)
	}
	return nil
}

func ValidateAcceptInvitation(ctx context.Context, req *dto.AcceptInvitationReq) error {
	if req.InviterProfileID == 0 {
		return errs.NewError(ctx, domain.ClassroomMissingOwnerProfileID, nil, ErrProfileIDRequired)
	}
	if req.ClassroomID == 0 {
		return errs.NewError(ctx, domain.ClassroomMissingID, nil, ErrClassroomIDRequired)
	}
	return nil
}

func ValidateRejectInvitation(ctx context.Context, req *dto.RejectInvitationReq) error {
	if req.InviterProfileID == 0 {
		return errs.NewError(ctx, domain.ClassroomMissingOwnerProfileID, nil, ErrProfileIDRequired)
	}
	if req.ClassroomID == 0 {
		return errs.NewError(ctx, domain.ClassroomMissingID, nil, ErrClassroomIDRequired)
	}
	return nil
}

func ValidateCancelInvitation(ctx context.Context, req *dto.CancelInvitationReq) error {
	if req.ProfileID == 0 {
		return errs.NewError(ctx, domain.ClassroomMissingOwnerProfileID, nil, ErrProfileIDRequired)
	}
	if req.ClassroomID == 0 {
		return errs.NewError(ctx, domain.ClassroomMissingID, nil, ErrClassroomIDRequired)
	}
	if req.TargetProfileID == 0 {
		return errs.NewError(ctx, domain.ClassroomMemberMissingProfileID, nil, ErrTargetProfileIDRequired)
	}
	return nil
}
