package pagination

import (
	"context"
	"errors"
	"strings"

	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/shared/enum"
)

var (
	ErrPaginationTypeInvalid      = errors.New("pagination_type is invalid")
	ErrPageWithCursorPagination   = errors.New("page is not accepted with CURSOR pagination")
	ErrCursorWithOffsetPagination = errors.New("next / previous is not accepted with OFFSET pagination")
	ErrNextAndPrevious            = errors.New("send either next or previous, not both")
)

// Request is the paging part of a list request that supports both styles
// (enum.PaginationType). Embed it in the list's request DTO: its fields sit
// at the top level of the JSON body.
//
// CURSOR (default) reads size plus at most one of next (the previous
// response's end_cursor → next page) and previous (its start_cursor →
// previous page), neither = the first page; OFFSET reads page + size.
type Request struct {
	PaginationType enum.PaginationType `json:"pagination_type"`
	Page           int64               `json:"page"`
	Size           int64               `json:"size"`
	Next           string              `json:"next"`
	Previous       string              `json:"previous"`
}

// DefaultType is the style of a request that names none.
const DefaultType = enum.PaginationTypeCursor

// IsCursor reports whether the request pages by cursor; an empty type is
// DefaultType, validated or not.
func (r *Request) IsCursor() bool {
	t := r.PaginationType
	if t == "" {
		t = DefaultType
	}
	return t == enum.PaginationTypeCursor
}

// Validate resolves the paging style (empty = CURSOR) and refuses next AND
// previous together — they are two directions, both at once has no
// meaning. The other style's parameter (page with CURSOR, next/previous
// with OFFSET) is currently ignored, not refused; the checks are kept
// commented below. The cursors' content is checked where they are decoded
// (KeysetPage).
func (r *Request) Validate(ctx context.Context) error {
	r.PaginationType = enum.PaginationType(strings.TrimSpace(string(r.PaginationType)))
	if r.PaginationType == "" {
		r.PaginationType = DefaultType
	}
	if !r.PaginationType.IsValid() {
		args := map[string]any{
			"pagination_types": enum.ListPaginationTypes(),
		}
		return errs.NewError(ctx, status.PAGINATION_INVALID_TYPE, args, ErrPaginationTypeInvalid)
	}
	r.Next = strings.TrimSpace(r.Next)
	r.Previous = strings.TrimSpace(r.Previous)

	switch r.PaginationType {
	case enum.PaginationTypeCursor:
		// TODO: Turn on if we need, now don't care
		// if r.Page != 0 {
		// 	return errs.NewError(ctx, status.PAGINATION_PARAMS_CONFLICT, nil, ErrPageWithCursorPagination)
		// }
		if r.Next != "" && r.Previous != "" {
			return errs.NewError(ctx, status.PAGINATION_PARAMS_CONFLICT, nil, ErrNextAndPrevious)
		}
	case enum.PaginationTypeOffset:
		// TODO: Turn on if we need, now don't care
		// if r.Next != "" || r.Previous != "" {
		// 	return errs.NewError(ctx, status.PAGINATION_PARAMS_CONFLICT, nil, ErrCursorWithOffsetPagination)
		// }
	}
	return nil
}
