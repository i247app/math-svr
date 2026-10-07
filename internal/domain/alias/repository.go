package alias

import (
	"context"

	"math-ai.com/math-ai/internal/shared/enum"
)

type IRepository interface {
	Create(ctx context.Context, alias *Alias) (*Alias, error)
	FindByAid(ctx context.Context, aid int64) (*Alias, error)
	FindByAka(ctx context.Context, alias string) (*Alias, error)
	FindByUid(ctx context.Context, uid int64) ([]*Alias, error)
	UpdateByAid(ctx context.Context, alias *Alias) error
	// SoftDeleteByAid retires ONE login key while leaving the user's
	// other keys alone — used when a guest registers and their stand-in
	// device alias must stop resolving. It stamps deleted_dt and flips
	// status, which is what the active filter actually reads; setting
	// alias_status alone would leave the row findable.
	SoftDeleteByAid(ctx context.Context, aid int64) error
	DeleteByUid(ctx context.Context, uid int64) error
	MarkStatusByUid(ctx context.Context, uid int64, status enum.UserAliasStatusType) error
	SoftDeleteByUid(ctx context.Context, uid int64) error
}
