package pow

type ChallengeReq struct {

}

type ChallengeResponse struct {
	Message string `json:"message"`
	Difficulty int    `json:"difficulty"`
}

type VerifyChallengeResponse struct {
	IsValid bool `json:"is_valid"`
}