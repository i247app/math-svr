package auth

import (
	"encoding/json"
	"net/http"

	dto "math-ai.com/math-ai/internal/application/dto/auth"
	"math-ai.com/math-ai/internal/application/resource"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/infrastructure/metadata"
	"math-ai.com/math-ai/internal/infrastructure/session"
	"math-ai.com/math-ai/internal/shared/response"
)

type AuthHandler struct {
	appResource *resource.Resource
	service     *Service
}

func NewAuthHandler(appResource *resource.Resource, service *Service) *AuthHandler {
	return &AuthHandler{
		appResource: appResource,
		service:     service,
	}
}

// uid pulls the authenticated user id out of the request's session.
func (h *AuthHandler) uid(w http.ResponseWriter, r *http.Request) (*int64, bool) {
	ss, err := h.appResource.GetRequestSession(r)
	if err != nil {
		response.WriteJson(w, nil, err)
		return nil, false
	}
	id, ok := ss.UID()
	if !ok {
		response.WriteJson(w, nil, errs.NewError(r.Context(), status.UNAUTHORIZED, nil, session.ErrUidNotFoundFromSession))
		return nil, false
	}
	return &id, true
}

// POST /auth/login
func (h *AuthHandler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	var req dto.LoginReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	req.OTPEnabled = h.appResource.Env.EnableOTP

	// Get session
	session, err := h.appResource.GetRequestSession(r)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}

	res, err := h.service.Login(r.Context(), session, &req)
	if err != nil {
		if res != nil {
			response.WriteJson(w, res, err)
			return
		}
		response.WriteJson(w, nil, err)
		return
	}

	response.WriteJson(w, res, nil)
}

// POST /auth/otp
func (h *AuthHandler) HandleLoginOTP(w http.ResponseWriter, r *http.Request) {
	var req dto.LoginReq
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
	req.OTPEnabled = h.appResource.Env.EnableOTP

	res, err := h.service.Login(r.Context(), session, &req)
	if err != nil {
		if res != nil {
			response.WriteJson(w, res, err)
			return
		}

		response.WriteJson(w, nil, err)
		return
	}

	response.WriteJson(w, res, nil)
}

// POST /auth/resume-session
func (h *AuthHandler) HandleResumeSession(w http.ResponseWriter, r *http.Request) {
	// Get session
	session, err := h.appResource.GetRequestSession(r)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}

	res, err := h.service.ResumeSession(r.Context(), session)
	if err != nil {
		if res != nil {
			response.WriteJson(w, res, err)
			return
		}
		response.WriteJson(w, nil, err)
		return
	}

	response.WriteJson(w, res, nil)
}

// POST /auth/logout
func (h *AuthHandler) HandleLogout(w http.ResponseWriter, r *http.Request) {
	var req dto.LogoutReq
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

	req.DeviceUUID = metadata.GetDeviceUUID(r.Context())
	uid, ok := h.uid(w, r)
	if !ok {
		return
	}
	req.UserID = uid

	res, err := h.service.Logout(r.Context(), session, &req)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}

	response.WriteJson(w, res, nil)
}
