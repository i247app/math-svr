package query

import (
	"context"

	"math-ai.com/math-ai/internal/domain/classroom"
	"math-ai.com/math-ai/internal/shared/enum"
	"math-ai.com/math-ai/internal/shared/pagination"
)

// ListMyPendingInvitationsQuery returns every PENDING ma_classroom_members
// row that lists ProfileID as the invitee. The caller is the invitee —
// no manager gate. ClassroomID is left nil so a profile can see
// invitations to every classroom in one shot.
type ListMyPendingInvitationsQuery struct {
	ProfileID int64
	Page      int64
	Limit     int64
}

type ListMyPendingInvitationsQueryHandler struct {
	memberRepo classroom.IMemberRepository
}

func NewListMyPendingInvitationsQueryHandler(memberRepo classroom.IMemberRepository) *ListMyPendingInvitationsQueryHandler {
	return &ListMyPendingInvitationsQueryHandler{memberRepo: memberRepo}
}

func (h *ListMyPendingInvitationsQueryHandler) Handle(ctx context.Context, q ListMyPendingInvitationsQuery) ([]*classroom.Member, *pagination.Pagination, error) {
	pid := q.ProfileID
	pendingStatus := string(enum.ClassroomMemberStatusTypePending)
	return h.memberRepo.ListMembers(ctx, &classroom.ListMembersParams{
		ProfileId: &pid,
		Status:    &pendingStatus,
		Page:      q.Page,
		Limit:     q.Limit,
	})
}
