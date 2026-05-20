package otp

import (
	"encoding/json"
	"net/http"

	dto "math-ai.com/math-ai/internal/application/dto/otp"
	"math-ai.com/math-ai/internal/application/resource"
	"math-ai.com/math-ai/internal/shared/response"
)

type OtpHandler struct {
	appResource *resource.Resource
	otpSvc      *Service
}

func NewOtpHandler(appResource *resource.Resource, otpSvc *Service) *OtpHandler {
	return &OtpHandler{
		appResource: appResource,
		otpSvc:      otpSvc,
	}
}

// POST /otps/send and POST /otps/resend — same payload, same handler. Resend
// is just a semantic alias; the underlying SendOtpCommand always revokes any
// prior PENDING row before issuing a fresh one, so the behavior is correct
// for both intents.
func (h *OtpHandler) HandleSend(w http.ResponseWriter, r *http.Request) {
	var req dto.SendOtpReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}

	res, err := h.otpSvc.Send(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	response.WriteJson(w, res, nil)
}

// POST /otps/verify
func (h *OtpHandler) HandleVerify(w http.ResponseWriter, r *http.Request) {
	var req dto.VerifyOtpReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}

	// Get session
	session, err := h.appResource.GetRequestSession(r)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}

	res, err := h.otpSvc.Verify(r.Context(), session, &req)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	response.WriteJson(w, res, nil)
}
