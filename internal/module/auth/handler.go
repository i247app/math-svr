package auth

import (
	"encoding/json"
	"net/http"

	dto "math-ai.com/math-ai/internal/application/dto/auth"
	"math-ai.com/math-ai/internal/application/resource"
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

// POST /auth/login
func (h *AuthHandler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	var req dto.LoginReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}

	res, err := h.service.Login(r.Context(), &req)
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

	res, err := h.service.LoginWithOTP(r.Context(), &req)
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

// POST /auth/login-resume
func (h *AuthHandler) HandleLoginResume(w http.ResponseWriter, r *http.Request) {
	// Get session
	session, err := h.appResource.GetRequestSession(r)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}

	res, err := h.service.LoginResume(r.Context(), session)
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

	res, err := h.service.Logout(r.Context(), session, &req)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}

	response.WriteJson(w, res, nil)
}
