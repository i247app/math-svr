package query

import (
	"context"

	"math-ai.com/math-ai/internal/domain/classroom"
	"math-ai.com/math-ai/internal/shared/enum"
	"math-ai.com/math-ai/internal/shared/pagination"
)

// ListMyPendingInvitationsQuery returns invitations directly addressed
// to ProfileID. Status defaults to PENDING when nil so the common case
// (the "do I have anything to act on?" call) stays a single
// per-aggregate query. Claimable-by-alias invitations (NULL
// invited_profile_id matched by the caller's email/phone) are out of
// scope for Phase 5.2 — that wires in with the Phase 5.6 notify work.
type ListMyPendingInvitationsQuery struct {
	ProfileID int64
	Status    *string
	Page      int64
	Limit     int64
}

type ListMyPendingInvitationsQueryHandler struct {
	invitationRepo classroom.IInvitationRepository
}

func NewListMyPendingInvitationsQueryHandler(invitationRepo classroom.IInvitationRepository) *ListMyPendingInvitationsQueryHandler {
	return &ListMyPendingInvitationsQueryHandler{invitationRepo: invitationRepo}
}

func (h *ListMyPendingInvitationsQueryHandler) Handle(ctx context.Context, q ListMyPendingInvitationsQuery) ([]*classroom.Invitation, *pagination.Pagination, error) {
	pid := q.ProfileID
	effectiveStatus := q.Status
	if effectiveStatus == nil {
		s := string(enum.ClassroomInvitationStatusTypePending)
		effectiveStatus = &s
	}
	return h.invitationRepo.ListInvitations(ctx, &classroom.ListInvitationsParams{
		InvitedProfileId: &pid,
		Status:           effectiveStatus,
		Page:             q.Page,
		Limit:            q.Limit,
	})
}
