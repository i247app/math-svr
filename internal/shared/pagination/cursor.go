package pagination

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
)

// ErrInvalidCursor: the cursor string is not one this server issued (bad
// base64, bad JSON, or a key the list does not recognise).
var ErrInvalidCursor = errors.New("pagination: invalid cursor")

// CursorPagination is the page info of a cursor-paged (keyset) list, in
// both directions. StartCursor / EndCursor point at the first / last row of
// the page as displayed: send EndCursor as `after` for the next page,
// StartCursor as `before` for the previous one. Both are nil when the page
// is empty. HasNext / HasPrev say whether a row exists past either end.
type CursorPagination struct {
	StartCursor *string `json:"start_cursor"`
	EndCursor   *string `json:"end_cursor"`
	HasNext     bool    `json:"has_next"`
	HasPrev     bool    `json:"has_prev"`
}

// EncodeCursor turns a list's position into the opaque string handed to the
// client: base64url (no padding) of the position as JSON. The client must
// treat it as opaque — the position may grow more keys (e.g. a sort column
// before the id) without the client changing.
func EncodeCursor(position any) (string, error) {
	raw, err := json.Marshal(position)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(raw), nil
}

// DecodeCursor reverses EncodeCursor into position (a pointer). Unknown keys
// are rejected, so a cursor minted by another list cannot be replayed here.
// It is not a security boundary — every list still applies its own filters
// in SQL — only a guard against garbage input.
func DecodeCursor(cursor string, position any) error {
	raw, err := base64.RawURLEncoding.DecodeString(cursor)
	if err != nil {
		return ErrInvalidCursor
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	if err := dec.Decode(position); err != nil {
		return ErrInvalidCursor
	}
	return nil
}
