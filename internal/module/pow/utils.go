package pow

import (
	"crypto/sha256"
	"encoding/hex"
	"strconv"
	"strings"
)

// Solves reports whether SHA-256(seed + nonce), in hex, starts with
// difficulty zeros.
func Solves(seed string, nonce int64, difficulty int) bool {
	sum := sha256.Sum256([]byte(seed + strconv.FormatInt(nonce, 10)))
	return strings.HasPrefix(hex.EncodeToString(sum[:]), strings.Repeat("0", difficulty))
}
