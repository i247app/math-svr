package profile

import (
	"context"
)

// IRepository owns all profile persistence. UpdateAvatarKey is split from
// Update so the upload-avatar flow can set just the key without forcing
// COALESCE on every other column.
type IRepository interface {
	FindByProfileId(ctx context.Context, profileId int64) (*Profile, error)
	ListByUserId(ctx context.Context, userId int64) ([]*Profile, error)
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
