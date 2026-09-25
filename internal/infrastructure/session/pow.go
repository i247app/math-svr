package session

import (
	"encoding/gob"
	"time"
)

// Pow is a pending proof-of-work challenge, stored in the session under the
// pow module's key between /pow/challenge and /pow/verify. The type name is
// part of the persisted session file (gob), so do not rename it.
type Pow struct {
	Seed       string    `json:"seed"`
	Difficulty int       `json:"difficulty"`
	CreatedAt  time.Time `json:"created_at"`
}

// Session values are stored as `any` and persisted with gob, which must know
// every concrete type it meets behind an interface. Register each struct type
// that is Put into a session, or the session file cannot be written.
func init() {
	gob.Register(Pow{})
}
