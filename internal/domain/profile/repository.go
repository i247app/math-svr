package profile

import (
	"context"

	"github.com/google/uuid"
)

// IRepository owns all profile persistence. UpdateAvatarKey is split from
// Update so the upload-avatar flow can set just the key without forcing
// COALESCE on every other column.
type IRepository interface {
	FindByProfileId(ctx context.Context, profileId uuid.UUID) (*Profile, error)
	ListByUserId(ctx context.Context, userId uuid.UUID) ([]*Profile, error)
	Create(ctx context.Context, profile *Profile) (*Profile, error)
	Update(ctx context.Context, profile *Profile) error
	UpdateAvatarKey(ctx context.Context, profileId uuid.UUID, avatarKey string) error
	MarkStatusByProfileId(ctx context.Context, profileId uuid.UUID, profileStatus string) error
	SoftDelete(ctx context.Context, profileId uuid.UUID) error
	SoftDeleteByUserId(ctx context.Context, userId uuid.UUID) error
	ForceDeleteByUserId(ctx context.Context, userId uuid.UUID) error
}
