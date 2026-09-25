package pow

type ChallengeReq struct {
}

type ChallengeRes struct {
	// Message is the challenge seed: find a nonce such that
	// SHA-256(message + decimal nonce), in hex, starts with Difficulty zeros.
	Message    string `json:"message"`
	Difficulty int    `json:"difficulty"`
}

type VerifyChallengeReq struct {
	Nonce int64 `json:"nonce"`
}
type VerifyChallengeRes struct {
	IsValid bool `json:"is_valid"`
}
