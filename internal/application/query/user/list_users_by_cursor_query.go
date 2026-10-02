package query

import (
	"context"

	"math-ai.com/math-ai/internal/domain/user"
	"math-ai.com/math-ai/internal/shared/pagination"
)

// ListUsersByCursorQuery is /users/list with pagination_type CURSOR. Next
// is the previous response's end_cursor (next page), Previous its
// start_cursor (previous page); neither = the first page. The validator
// guarantees at most one is set.
type ListUsersByCursorQuery struct {
	Next     string
	Previous string
	Size     int64
}

// userCursor is what a /users/list cursor holds: the uid of one row. Rows
// are displayed uid DESC, so "after" a row means a lower uid and "before"
// it a higher one.
type userCursor struct {
	Uid int64 `json:"uid"`
}

type ListUsersByCursorQueryHandler struct {
	userRepo user.IRepository
}

func NewListUsersByCursorQueryHandler(userRepo user.IRepository) *ListUsersByCursorQueryHandler {
	return &ListUsersByCursorQueryHandler{userRepo: userRepo}
}

func (h *ListUsersByCursorQueryHandler) Handle(ctx context.Context, query *ListUsersByCursorQuery) ([]*user.User, *pagination.CursorPagination, error) {
	return pagination.KeysetPage(ctx, pagination.Keyset[*user.User, userCursor]{
		Read: func(ctx context.Context, after, before *userCursor, limit int64) ([]*user.User, error) {
			params := &user.ListUsersKeysetParams{Limit: limit}
			if after != nil {
				params.AfterUid = &after.Uid
			}
			if before != nil {
				params.BeforeUid = &before.Uid
			}
			return h.userRepo.ListUsersByKeyset(ctx, params)
		},
		// Display order is uid DESC: "before" a row means a higher uid.
		ExistsAtOrBefore: func(ctx context.Context, k userCursor) (bool, error) {
			return h.userRepo.ExistsUserUidAtLeast(ctx, k.Uid)
		},
		ExistsAtOrAfter: func(ctx context.Context, k userCursor) (bool, error) {
			return h.userRepo.ExistsUserUidAtMost(ctx, k.Uid)
		},
		Key:   func(u *user.User) userCursor { return userCursor{Uid: u.Uid()} },
		Valid: func(k userCursor) bool { return k.Uid > 0 },
	}, query.Next, query.Previous, query.Size)
}
