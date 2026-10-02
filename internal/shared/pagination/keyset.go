package pagination

import (
	"context"

	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
)

// Keyset describes one keyset-paged list to KeysetPage. T is a row, K the
// row's position in the list's display order — what a cursor holds, as
// JSON (so K's fields carry json tags, and every field must be set for a
// position to be Valid).
//
// "Before" and "after" below are in DISPLAY order: the first page starts
// the list, next pages move after it, previous pages before it.
type Keyset[T, K any] struct {
	// Read returns up to limit rows, always in display order: the rows
	// right after `after` (next page), the rows right before `before`
	// (previous page — read toward the start, then put back in display
	// order), or the first rows when both are nil. At most one is set.
	Read func(ctx context.Context, after, before *K, limit int64) ([]T, error)
	// ExistsAtOrBefore / ExistsAtOrAfter report whether any row of the
	// list sits at k or before / after it.
	ExistsAtOrBefore func(ctx context.Context, k K) (bool, error)
	ExistsAtOrAfter  func(ctx context.Context, k K) (bool, error)
	// Key is a row's position; Valid rejects a decoded position that is
	// well-formed JSON but could not have come from Key (e.g. id 0).
	Key   func(row T) K
	Valid func(k K) bool
}

// KeysetPage reads one page of a keyset list in either direction. It reads
// one row past the page in the direction of travel — that row is the flag
// for that direction — and settles the opposite flag with one existence
// probe from the cursor. The probe is exact: the page holds every row
// between the cursor and its far end, so a row beyond the near end exists
// iff one exists at or past the cursor. No COUNT.
//
// next / previous are the request's cursors (at most one set — see
// Request.Validate). A cursor that does not decode to a Valid position is
// PAGINATION_INVALID_CURSOR.
func KeysetPage[T, K any](ctx context.Context, ks Keyset[T, K], next, previous string, size int64) ([]T, *CursorPagination, error) {
	size = ClampSize(size)
	var after, before *K
	var err error
	if next != "" {
		if after, err = decodeKey(ctx, ks, next); err != nil {
			return nil, nil, err
		}
	} else if previous != "" {
		if before, err = decodeKey(ctx, ks, previous); err != nil {
			return nil, nil, err
		}
	}

	rows, err := ks.Read(ctx, after, before, size+1)
	if err != nil {
		return nil, nil, err
	}

	page := &CursorPagination{}
	more := int64(len(rows)) > size
	if before != nil {
		// Read toward the start: the extra row is the one farthest from the
		// cursor, i.e. the first in display order.
		if more {
			rows = rows[1:]
		}
		page.HasPrev = more
		if page.HasNext, err = ks.ExistsAtOrAfter(ctx, *before); err != nil {
			return nil, nil, err
		}
	} else {
		if more {
			rows = rows[:size]
		}
		page.HasNext = more
		if after != nil {
			if page.HasPrev, err = ks.ExistsAtOrBefore(ctx, *after); err != nil {
				return nil, nil, err
			}
		}
	}

	if len(rows) > 0 {
		if page.StartCursor, err = encodeKey(ks.Key(rows[0])); err != nil {
			return nil, nil, err
		}
		if page.EndCursor, err = encodeKey(ks.Key(rows[len(rows)-1])); err != nil {
			return nil, nil, err
		}
	}
	return rows, page, nil
}

func decodeKey[T, K any](ctx context.Context, ks Keyset[T, K], cursor string) (*K, error) {
	var k K
	if err := DecodeCursor(cursor, &k); err != nil || !ks.Valid(k) {
		return nil, errs.NewError(ctx, status.PAGINATION_INVALID_CURSOR, nil, ErrInvalidCursor)
	}
	return &k, nil
}

func encodeKey[K any](k K) (*string, error) {
	c, err := EncodeCursor(k)
	if err != nil {
		return nil, err
	}
	return &c, nil
}
