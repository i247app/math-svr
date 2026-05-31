package query

import (
	"context"

	"math-ai.com/math-ai/internal/domain/classroom"
	"math-ai.com/math-ai/internal/shared/enum"
	"math-ai.com/math-ai/internal/shared/pagination"
)

// ListPendingInvitationsByClassroomQuery returns every PENDING
// ma_classroom_members row attached to the given classroom — the
// manager-side view ("who have I invited?"). The module service runs
// the requireManager gate before invoking this query.
type ListPendingInvitationsByClassroomQuery struct {
	ClassroomID int64
	Page        int64
	Limit       int64
}

type ListPendingInvitationsByClassroomQueryHandler struct {
	memberRepo classroom.IMemberRepository
}

func NewListPendingInvitationsByClassroomQueryHandler(memberRepo classroom.IMemberRepository) *ListPendingInvitationsByClassroomQueryHandler {
	return &ListPendingInvitationsByClassroomQueryHandler{memberRepo: memberRepo}
}

func (h *ListPendingInvitationsByClassroomQueryHandler) Handle(ctx context.Context, q ListPendingInvitationsByClassroomQuery) ([]*classroom.Member, *pagination.Pagination, error) {
	cid := q.ClassroomID
	pendingStatus := string(enum.ClassroomMemberStatusTypePendingInvitation)
	return h.memberRepo.ListMembers(ctx, &classroom.ListMembersParams{
		ClassroomId: &cid,
		Status:      &pendingStatus,
		Page:        q.Page,
		Limit:       q.Limit,
	})
}
