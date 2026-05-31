package query

import (
	"context"

	"math-ai.com/math-ai/internal/domain/classroom"
	"math-ai.com/math-ai/internal/shared/pagination"
)

// ListInvitationsByClassroomQuery powers the manager-side roster view
// of invitations for a single classroom. Permission gating lives at
// the module service — this handler trusts the caller and just runs
// the read against the existing ListInvitations repo method.
type ListInvitationsByClassroomQuery struct {
	ClassroomID int64
	Status      *string
	Page        int64
	Limit       int64
}

type ListInvitationsByClassroomQueryHandler struct {
	invitationRepo classroom.IInvitationRepository
}

func NewListInvitationsByClassroomQueryHandler(invitationRepo classroom.IInvitationRepository) *ListInvitationsByClassroomQueryHandler {
	return &ListInvitationsByClassroomQueryHandler{invitationRepo: invitationRepo}
}

func (h *ListInvitationsByClassroomQueryHandler) Handle(ctx context.Context, q ListInvitationsByClassroomQuery) ([]*classroom.Invitation, *pagination.Pagination, error) {
	cid := q.ClassroomID
	return h.invitationRepo.ListInvitations(ctx, &classroom.ListInvitationsParams{
		ClassroomId: &cid,
		Status:      q.Status,
		Page:        q.Page,
		Limit:       q.Limit,
	})
}
