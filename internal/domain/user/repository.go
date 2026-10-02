package user

import (
	"context"

	"math-ai.com/math-ai/internal/shared/pagination"

	"math-ai.com/math-ai/internal/shared/enum"
)

type IRepository interface {
	// FindById(ctx context.Context, id int64) (*User, error)
	FindByUid(ctx context.Context, uid int64) (*User, error)
	FindByEmail(ctx context.Context, email string) (*User, error)
	FindByPhone(ctx context.Context, phone string) (*User, error)
	FindByUserName(ctx context.Context, userName string) (*User, error)
	ListUsers(ctx context.Context, params *ListUsersParams) ([]*User, *pagination.Pagination, error)
	// ListUsersByKeyset is the keyset (cursor) read behind /users/list with
	// pagination_type CURSOR. Up to Limit rows, ALWAYS returned uid DESC
	// (display order): with AfterUid, the rows right below it (next page);
	// with BeforeUid, the rows right above it (previous page); with neither,
	// the newest. No COUNT — the caller asks for one row more than it shows
	// to learn whether the list goes on in the direction it is reading.
	ListUsersByKeyset(ctx context.Context, params *ListUsersKeysetParams) ([]*User, error)
	// ExistsUserUidAtLeast / ExistsUserUidAtMost report whether any active
	// user has uid >= / <= uid — the keyset list's "is there a page on the
	// other side" probe. One PRIMARY KEY lookup each.
	ExistsUserUidAtLeast(ctx context.Context, uid int64) (bool, error)
	ExistsUserUidAtMost(ctx context.Context, uid int64) (bool, error)
	Create(ctx context.Context, user *User) (*User, error)
	Update(ctx context.Context, user *User) error
	UpdateAvatarKey(ctx context.Context, uid int64, avatarKey string) error
	DeleteByUid(ctx context.Context, uid int64) error
	MarkStatusByUid(ctx context.Context, uid int64, status enum.UserStatusType) error
	SoftDeleteByUid(ctx context.Context, uid int64) error
}

type IAliasRepository interface {
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

// ListUsersKeysetParams: at most one of AfterUid / BeforeUid is set.
type ListUsersKeysetParams struct {
	AfterUid  *int64 // next page: uid < AfterUid
	BeforeUid *int64 // previous page: uid > BeforeUid
	Limit     int64
}

type ListUsersParams struct {
	Search    string
	Page      int64
	Limit     int64
	OrderBy   string
	OrderDesc bool
	TakeAll   bool
}
