package query

import (
	"context"

	"math-ai.com/math-ai/internal/domain/classroom"
	"math-ai.com/math-ai/internal/shared/pagination"
)

// ListMembersQuery powers both "members of classroom" (ClassroomID set)
// and "my classrooms via membership" (ProfileID set) reads. The module
// service guarantees at least one is non-nil before calling.
type ListMembersQuery struct {
	ClassroomID *string
	ProfileID   *string
	Role        *string
	Status      *string
	Page        int64
	Limit       int64
}

type ListMembersQueryHandler struct {
	memberRepo classroom.IMemberRepository
}

func NewListMembersQueryHandler(memberRepo classroom.IMemberRepository) *ListMembersQueryHandler {
	return &ListMembersQueryHandler{memberRepo: memberRepo}
}

func (h *ListMembersQueryHandler) Handle(ctx context.Context, q ListMembersQuery) ([]*classroom.Member, *pagination.Pagination, error) {
	return h.memberRepo.ListMembers(ctx, &classroom.ListMembersParams{
		ClassroomId: q.ClassroomID,
		ProfileId:   q.ProfileID,
		Role:        q.Role,
		Status:      q.Status,
		Page:        q.Page,
		Limit:       q.Limit,
	})
}
