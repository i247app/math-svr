package pow

import (
	"net/http"
	"fmt"

	"math-ai.com/math-ai/internal/application/resource"
	dto "math-ai.com/math-ai/internal/application/dto/pow"
	"math-ai.com/math-ai/internal/shared/response"
	// "math-ai.com/math-ai/internal/infrastructure/session"

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
	message, difficulty := p.service.GenerateChallenge(r.Context())
	res.Message = message
	res.Difficulty = difficulty
	// sess, err := p.appResource.GetRequestSession(r)
	// if err != nil {
	// 	response.WriteJson(w, nil, err)
	// 	return
	// }

	// sessionData := session.InitData{
	// 	Source:     "pow",
	// 	IsSecure:   true,
	// 	Message:    message,
	// 	Difficulty: difficulty,

	// }

	// sess.Init(sessionData)

	response.WriteJson(w, res, nil)
}


func (p *PowHandler) VerifyChallenge(w http.ResponseWriter, r *http.Request) {
	var res dto.VerifyChallengeResponse

	// Get session
	ctx := r.Context()
	sess, err := p.appResource.GetRequestSession(r)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}

	nonce := r.FormValue("nonce")
	messageAny, ok := sess.Get("message")
	if !ok {
		response.WriteJson(w, nil, fmt.Errorf("Message not found in session"))
		return
	}
	message, ok := messageAny.(string)
	if !ok {
		response.WriteJson(w, nil, fmt.Errorf("Message is not a string"))
		return
	}

	difficultyAny, ok := sess.Get("difficulty")
	if !ok {
		response.WriteJson(w, nil, fmt.Errorf("Difficulty not found in session"))
		return
	}
	difficulty, ok := difficultyAny.(int)
	if !ok {
		response.WriteJson(w, nil, fmt.Errorf("Difficulty is not an integer"))
		return
	}

	//solve challenge
	isValid := p.service.VerifyChallenge(ctx, nonce, message, difficulty)
	if !isValid {
		response.WriteJson(w, nil, fmt.Errorf("Invalid Proof"))
		return
	}
	res.IsValid = isValid
	response.WriteJson(w, res, nil)
}