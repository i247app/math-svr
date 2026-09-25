package pow

import (
	"encoding/json"
	"net/http"

	dto "math-ai.com/math-ai/internal/application/dto/pow"
	"math-ai.com/math-ai/internal/application/resource"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/shared/response"
)

type PowHandler struct {
	appResource *resource.Resource
	service     *Service
}

func NewPowHandler(appResource *resource.Resource, service *Service) *PowHandler {
	return &PowHandler{
		appResource: appResource,
		service:     service,
	}
}

// POST /pow/challenge
func (p *PowHandler) GetChallenge(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req dto.ChallengeReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, errs.NewError(ctx, status.POW_INVALID_REQUEST, nil, err))
		return
	}

	sess, err := p.appResource.GetRequestSession(r)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}

	res, err := p.service.GenerateChallenge(ctx, sess)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}

	response.WriteJson(w, res, nil)
}

// POST /pow/verify
func (p *PowHandler) VerifyChallenge(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req dto.VerifyChallengeReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, errs.NewError(ctx, status.POW_INVALID_REQUEST, nil, err))
		return
	}

	sess, err := p.appResource.GetRequestSession(r)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}

	res, err := p.service.VerifyChallenge(ctx, sess, req.Nonce)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}

	response.WriteJson(w, res, nil)
}
