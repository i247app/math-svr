package pow

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
)

type Service struct {

}

func NewService() *Service {
	return &Service{}
}


func (s *Service) VerifyChallenge(ctx context.Context, nonce string, message string, difficulty int) bool {
	// Implement the logic to verify the nonce with the message and difficulty
	// This is a placeholder implementation. Replace it with your actual verification logic.
	input := message + nonce
	hash := sha256.Sum256([]byte(input))

	target := make([]byte, difficulty)
	for i := range target {
		target[i] = '0'
	}
	targetPrefix := string(target)
	if hex.EncodeToString(hash[:])[:difficulty] == targetPrefix {
		return true
	}
	return false
}

func (s *Service) GenerateChallenge(ctx context.Context) (string, int) {
	randomBytes := make([]byte, 16)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", 0
	}
	challenge := hex.EncodeToString(randomBytes)
	difficulty := 4
	return challenge, difficulty
}
