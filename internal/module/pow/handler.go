package pow

import (
	"fmt"
	"net/http"
	"time"

	dto "math-ai.com/math-ai/internal/application/dto/pow"
	"math-ai.com/math-ai/internal/application/resource"
	"math-ai.com/math-ai/internal/infrastructure/session"
	"math-ai.com/math-ai/internal/shared/response"
	"encoding/json"
)

type PowHandler struct {
	appResource *resource.Resource
	service     *Service
}

func NewPowHandler(appResource *resource.Resource, service *Service) *PowHandler {
	return &PowHandler{
		appResource: appResource,
		service:    service,
	}
}

// POST /getChallenge
func (p *PowHandler) GetChallenge(w http.ResponseWriter, r *http.Request) {
	var res dto.ChallengeResponse
	
	sess, err := p.appResource.GetRequestSession(r)
	message, difficulty := p.service.GenerateChallenge(r.Context(), sess)
	
	res.Message = message
	res.Difficulty = difficulty
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	
	response.WriteJson(w, res, nil)
}


func (p *PowHandler) VerifyChallenge(w http.ResponseWriter, r *http.Request) {
	var res dto.VerifyChallengeResponse
	var req dto.VerifyChallengeRequest
	// Get session
	ctx := r.Context()
	sess, err := p.appResource.GetRequestSession(r)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}

	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		response.WriteJson(w, nil, fmt.Errorf("Invalid request body"))
		return
	}
	nonce := req.Nonce

	challengeSession, ok := sess.Get("challenge")
	if !ok {
		response.WriteJson(w, nil, fmt.Errorf("Challenge not found in session"))
		return
	}

	challenge, ok := challengeSession.(session.Pow)
	if !ok {
		response.WriteJson(w, nil, fmt.Errorf("Invalid challenge data in session"))
		return
	}

	//check created at, if challenge is older than 24 hours, return error
	if time.Since(challenge.CreatedAt) > 24*time.Hour {
		response.WriteJson(w, nil, fmt.Errorf("Challenge has expired"))
		return
	}

	seed := challenge.Seed
	difficulty := challenge.Difficulty

	//solve challenge
	isValid := p.service.VerifyChallenge(ctx,sess, nonce, seed, difficulty)
	if !isValid {
		response.WriteJson(w, nil, fmt.Errorf("Invalid Proof"))
		return
	}
	res.IsValid = isValid
	response.WriteJson(w, res, nil)
}