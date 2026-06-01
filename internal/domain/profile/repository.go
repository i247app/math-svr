package profile

import (
	"context"

	"math-ai.com/math-ai/internal/shared/pagination"
)

// ListProfilesParams narrows the listing query. Every filter is optional:
// when all are zero/nil the list returns every active profile, paginated.
// Search matches case-insensitively against BOTH p.name and p.profile_code
// (LIKE %?% on each, OR'd together) so a single client-supplied needle
// covers full-code, partial-code, and name-fragment lookups uniformly.
// IsDefault is a pointer so callers can distinguish "filter to
// is_default=false" from "no filter".
type ListProfilesParams struct {
	UserId        *int64
	Role          *string
	ProfileStatus *string
	SchoolId      *int64
	ProgramId     *int64
	GradeId       *int64
	SemesterId    *int64
	IsDefault     *bool
	Search        *string
	Page          int64
	Limit         int64
	TakeAll       bool
}

// IRepository owns all profile persistence. UpdateAvatarKey is split from
// Update so the upload-avatar flow can set just the key without forcing
// COALESCE on every other column.
type IRepository interface {
	FindByProfileId(ctx context.Context, profileId int64) (*Profile, error)
	// FindByProfileCode looks up by the human-readable code (e.g.
	// "AA-1234"). Used by the create command to probe for uniqueness
	// before insert so a colliding code surfaces PROFILE_CODE_TAKEN
	// rather than a raw constraint error. Returns (nil, nil) when no
	// active row matches.
	FindByProfileCode(ctx context.Context, profileCode string) (*Profile, error)
	ListByProfileIds(ctx context.Context, profileIds []int64) ([]*Profile, error)
	ListByUserId(ctx context.Context, userId int64) ([]*Profile, error)
	ListProfiles(ctx context.Context, params *ListProfilesParams) ([]*Profile, *pagination.Pagination, error)
	ListAvatarKeysByUserId(ctx context.Context, userId int64) ([]string, error)
	Create(ctx context.Context, profile *Profile) (*Profile, error)
	Update(ctx context.Context, profile *Profile) error
	UpdateAvatarKey(ctx context.Context, profileId int64, avatarKey string) error
	// SetSchoolId writes school_id directly (no COALESCE) so an explicit
	// nil clears the link. Used by the assign/remove flows; Update is
	// still the right tool for partial PATCH-style payloads.
	SetSchoolId(ctx context.Context, profileId int64, schoolId *int64) error
	MarkStatusByProfileId(ctx context.Context, profileId int64, profileStatus string) error
	MarkDefaultByProfileId(ctx context.Context, userId int64, profileId int64) error
	SoftDelete(ctx context.Context, profileId int64) error
	ForceDelete(ctx context.Context, profileId int64) error
	SoftDeleteByUserId(ctx context.Context, userId int64) error
	ForceDeleteByUserId(ctx context.Context, userId int64) error
}
