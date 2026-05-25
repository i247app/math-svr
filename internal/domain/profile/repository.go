package profile

import (
	"context"
)

// IRepository owns all profile persistence. UpdateAvatarKey is split from
// Update so the upload-avatar flow can set just the key without forcing
// COALESCE on every other column.
type IRepository interface {
	FindByProfileId(ctx context.Context, profileId string) (*Profile, error)
	ListByUserId(ctx context.Context, userId string) ([]*Profile, error)
	ListAvatarKeysByUserId(ctx context.Context, userId string) ([]string, error)
	Create(ctx context.Context, profile *Profile) (*Profile, error)
	Update(ctx context.Context, profile *Profile) error
	UpdateAvatarKey(ctx context.Context, profileId string, avatarKey string) error
	MarkStatusByProfileId(ctx context.Context, profileId string, profileStatus string) error
	SoftDelete(ctx context.Context, profileId string) error
	ForceDelete(ctx context.Context, profileId string) error
	SoftDeleteByUserId(ctx context.Context, userId string) error
	ForceDeleteByUserId(ctx context.Context, userId string) error
}
