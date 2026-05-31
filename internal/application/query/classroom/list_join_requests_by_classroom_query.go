package query

import (
	"context"

	"math-ai.com/math-ai/internal/domain/classroom"
	"math-ai.com/math-ai/internal/shared/enum"
	"math-ai.com/math-ai/internal/shared/pagination"
)

// ListJoinRequestsByClassroomQuery returns every PENDING_REQUEST
// ma_classroom_members row attached to the given classroom — the
// owner-side view of who has asked to join. The module service runs
// the requireOwner gate before invoking this query.
type ListJoinRequestsByClassroomQuery struct {
	ClassroomID int64
	Page        int64
	Limit       int64
}

type ListJoinRequestsByClassroomQueryHandler struct {
	memberRepo classroom.IMemberRepository
}

func NewListJoinRequestsByClassroomQueryHandler(memberRepo classroom.IMemberRepository) *ListJoinRequestsByClassroomQueryHandler {
	return &ListJoinRequestsByClassroomQueryHandler{memberRepo: memberRepo}
}

func (h *ListJoinRequestsByClassroomQueryHandler) Handle(ctx context.Context, q ListJoinRequestsByClassroomQuery) ([]*classroom.Member, *pagination.Pagination, error) {
	cid := q.ClassroomID
	pendingStatus := string(enum.ClassroomMemberStatusTypePendingRequest)
	return h.memberRepo.ListMembers(ctx, &classroom.ListMembersParams{
		ClassroomId: &cid,
		Status:      &pendingStatus,
		Page:        q.Page,
		Limit:       q.Limit,
	})
}
