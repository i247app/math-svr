package pow

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"math/big"
	"strconv"
	"time"

	"math-ai.com/math-ai/internal/infrastructure/session"
)

type Service struct {
}

func NewService() *Service {
	return &Service{}
}

func generateRandomString(length int) (string, error) {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	result := make([]byte, length)

	for i := 0; i < length; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			return "", err
		}
		result[i] = letters[num.Int64()]
	}

	return string(result), nil
}

func (s *Service) VerifyChallenge(ctx context.Context, sess *session.AppSession, nonce int, message string, difficulty int) bool {
	// Implement the logic to verify the nonce with the message and difficulty
	// This is a placeholder implementation. Replace it with your actual verification logic.
	input := message + strconv.Itoa(nonce)
	println("Verifying challenge with input:", input, "and difficulty:", difficulty)
	hash := sha256.Sum256([]byte(input))
	println("Hash:", hex.EncodeToString(hash[:]))
	target := make([]byte, difficulty)
	for i := range target {
		target[i] = '0'
	}
	targetPrefix := string(target)
	if hex.EncodeToString(hash[:])[:difficulty] == targetPrefix {
		challengeData := session.Pow{
			Seed:       message,
			Difficulty: difficulty,
			Nonce:      nonce,
			CreatedAt:  time.Now(),
		}

		sessionData := session.InitData{
			Source:    "pow",
			IsSecure:  true,
			Challenge: challengeData,
		}

		sess.Init(sessionData)

		return true
	}
	return false
}

func (s *Service) GenerateChallenge(ctx context.Context, sess *session.AppSession) (string, int) {
	challenge, err := generateRandomString(5)
	if err != nil {
		return "", 0
	}

	difficulty := 4

	challengeData := session.Pow{
		Seed:       challenge,
		Difficulty: difficulty,
		CreatedAt:  time.Now(),
	}

	sessionData := session.InitData{
		Source:    "pow",
		IsSecure:  false,
		Challenge: challengeData,
	}

	sess.Init(sessionData)
	return challenge, difficulty
}
