package query

import (
	"context"

	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
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

// Handle reads one page plus one row in the direction of travel — that extra
// row is the flag for that direction — and settles the flag for the opposite
// direction with one existence probe from the cursor's uid. The probe is
// exact: the page holds every row between the cursor and its far end, so a
// row beyond the near end exists iff one exists at or past the cursor.
func (h *ListUsersByCursorQueryHandler) Handle(ctx context.Context, query *ListUsersByCursorQuery) ([]*user.User, *pagination.CursorPagination, error) {
	size := pagination.ClampSize(query.Size)
	params := &user.ListUsersKeysetParams{Limit: size + 1}
	var err error
	if query.Next != "" {
		if params.AfterUid, err = decodeUserCursor(ctx, query.Next); err != nil {
			return nil, nil, err
		}
	} else if query.Previous != "" {
		if params.BeforeUid, err = decodeUserCursor(ctx, query.Previous); err != nil {
			return nil, nil, err
		}
	}

	users, err := h.userRepo.ListUsersByKeyset(ctx, params)
	if err != nil {
		return nil, nil, err
	}

	page := &pagination.CursorPagination{}
	more := int64(len(users)) > size
	if params.BeforeUid != nil {
		// Read upward: the extra row is the one farthest from the cursor,
		// i.e. the first in display order.
		if more {
			users = users[1:]
		}
		page.HasPrev = more
		if page.HasNext, err = h.userRepo.ExistsUserUidAtMost(ctx, *params.BeforeUid); err != nil {
			return nil, nil, err
		}
	} else {
		if more {
			users = users[:size]
		}
		page.HasNext = more
		if params.AfterUid != nil {
			if page.HasPrev, err = h.userRepo.ExistsUserUidAtLeast(ctx, *params.AfterUid); err != nil {
				return nil, nil, err
			}
		}
	}

	if len(users) > 0 {
		if page.StartCursor, err = encodeUserCursor(users[0]); err != nil {
			return nil, nil, err
		}
		if page.EndCursor, err = encodeUserCursor(users[len(users)-1]); err != nil {
			return nil, nil, err
		}
	}
	return users, page, nil
}

func decodeUserCursor(ctx context.Context, cursor string) (*int64, error) {
	var pos userCursor
	if err := pagination.DecodeCursor(cursor, &pos); err != nil || pos.Uid <= 0 {
		return nil, errs.NewError(ctx, status.PAGINATION_INVALID_CURSOR, nil, pagination.ErrInvalidCursor)
	}
	return &pos.Uid, nil
}

func encodeUserCursor(u *user.User) (*string, error) {
	c, err := pagination.EncodeCursor(userCursor{Uid: u.Uid()})
	if err != nil {
		return nil, err
	}
	return &c, nil
}
