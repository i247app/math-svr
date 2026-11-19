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
	// other side" probe. roles narrows them exactly like the list itself
	// (empty = every role), or has_prev / has_next would count rows the
	// list never shows.
	ExistsUserUidAtLeast(ctx context.Context, uid int64, roles []string) (bool, error)
	ExistsUserUidAtMost(ctx context.Context, uid int64, roles []string) (bool, error)
	Create(ctx context.Context, user *User) (*User, error)
	Update(ctx context.Context, user *User) error
	UpdateAvatarKey(ctx context.Context, uid int64, avatarKey string) error
	DeleteByUid(ctx context.Context, uid int64) error
	MarkStatusByUid(ctx context.Context, uid int64, status enum.UserStatusType) error
	SoftDeleteByUid(ctx context.Context, uid int64) error
}

// IOwnedDataRepository erases a user's rows in every OTHER table keyed by
// uid (logins, devices, login logs, otps, notifications, presence, chat
// participants, exam links, exam sessions). ma_users, ma_aliases and
// ma_profiles keep their own SoftDeleteByUid / DeleteByUid, which also set
// the business status. Both calls are idempotent.
type IOwnedDataRepository interface {
	// SoftDeleteByUid stamps status INACTIVE + deleted_dt on live rows.
	SoftDeleteByUid(ctx context.Context, uid int64) error
	// ForceDeleteByUid physically deletes the rows, live or soft-deleted.
	ForceDeleteByUid(ctx context.Context, uid int64) error
}

// ListUsersKeysetParams: at most one of AfterUid / BeforeUid is set.
type ListUsersKeysetParams struct {
	AfterUID  *int64 // next page: uid < AfterUid
	BeforeUID *int64 // previous page: uid > BeforeUid
	Limit     int64
	Roles     []string // ma_users.role IN Roles; empty = every role
}

type ListUsersParams struct {
	Roles     []string // ma_users.role IN Roles; empty = every role
	Search    string
	Page      int64
	Limit     int64
	OrderBy   string
	OrderDesc bool
	TakeAll   bool
}
