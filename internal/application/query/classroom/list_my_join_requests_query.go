package query

import (
	"context"

	"math-ai.com/math-ai/internal/domain/classroom"
	"math-ai.com/math-ai/internal/shared/enum"
	"math-ai.com/math-ai/internal/shared/pagination"
)

// ListMyJoinRequestsQuery returns every PENDING_REQUEST
// ma_classroom_members row created by ProfileID — the user-side view
// of "classrooms I have asked to join". No manager gate; the caller
// is the requester.
type ListMyJoinRequestsQuery struct {
	ProfileID int64
	Page      int64
	Limit     int64
}

type ListMyJoinRequestsQueryHandler struct {
	memberRepo classroom.IMemberRepository
}

func NewListMyJoinRequestsQueryHandler(memberRepo classroom.IMemberRepository) *ListMyJoinRequestsQueryHandler {
	return &ListMyJoinRequestsQueryHandler{memberRepo: memberRepo}
}

func (h *ListMyJoinRequestsQueryHandler) Handle(ctx context.Context, q ListMyJoinRequestsQuery) ([]*classroom.Member, *pagination.Pagination, error) {
	pid := q.ProfileID
	pendingStatus := string(enum.ClassroomMemberStatusTypePendingRequest)
	return h.memberRepo.ListMembers(ctx, &classroom.ListMembersParams{
		ProfileId: &pid,
		Status:    &pendingStatus,
		Page:      q.Page,
		Limit:     q.Limit,
	})
}
