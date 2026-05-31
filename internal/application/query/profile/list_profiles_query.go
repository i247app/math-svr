package query

import (
	"context"

	"math-ai.com/math-ai/internal/domain/profile"
	"math-ai.com/math-ai/internal/shared/pagination"
)

// ListProfilesQuery powers /profiles/list. Every filter is optional; the
// repo only adds a predicate when the field is non-zero / non-nil.
type ListProfilesQuery struct {
	UserID        *int64
	Role          *string
	ProfileStatus *string
	SchoolID      *int64
	ProgramID     *int64
	GradeID       *int64
	SemesterID    *int64
	IsDefault     *bool
	Search        *string
	Page          int64
	Limit         int64
}

type ListProfilesQueryHandler struct {
	profileRepo profile.IRepository
}

func NewListProfilesQueryHandler(profileRepo profile.IRepository) *ListProfilesQueryHandler {
	return &ListProfilesQueryHandler{profileRepo: profileRepo}
}

func (h *ListProfilesQueryHandler) Handle(ctx context.Context, q ListProfilesQuery) ([]*profile.Profile, *pagination.Pagination, error) {
	return h.profileRepo.ListProfiles(ctx, &profile.ListProfilesParams{
		UserId:        q.UserID,
		Role:          q.Role,
		ProfileStatus: q.ProfileStatus,
		SchoolId:      q.SchoolID,
		ProgramId:     q.ProgramID,
		GradeId:       q.GradeID,
		SemesterId:    q.SemesterID,
		IsDefault:     q.IsDefault,
		Search:        q.Search,
		Page:          q.Page,
		Limit:         q.Limit,
	})
}
