package pow

import "errors"

var (
	ErrChallengeNotFound = errors.New("no pending proof-of-work challenge")
	ErrChallengeExpired  = errors.New("proof-of-work challenge expired")
	ErrInvalidProof      = errors.New("nonce does not solve the challenge")
	ErrPowRequired       = errors.New("no unused proof of work in session")
	ErrNonceRequired     = errors.New("nonce is required")
)
