package user

import (
	"encoding/json"
	"fmt"
	"net/http"

	"math-ai.com/math-ai/internal/application/dto/user"
	"math-ai.com/math-ai/internal/shared/response"
	"math-ai.com/math-ai/internal/shared/utils"
)

const (
	MaxAvatarUploadSize = 10 << 20 // 10 MB
)

type UserHandler struct {
	userSvc *Service
}

func NewUserHandler(userSvc *Service) *UserHandler {
	return &UserHandler{
		userSvc: userSvc,
	}
}

// POST /users/create
func (h *UserHandler) HandleCreateUser(w http.ResponseWriter, r *http.Request) {
	var req user.CreateUserReq
	contentType := r.Header.Get("Content-Type")

	if contentType == "application/json" {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			response.WriteJson(w, nil, fmt.Errorf("invalid parameters"))
			return
		}
	} else {
		if err := r.ParseMultipartForm(MaxAvatarUploadSize); err != nil {
			response.WriteJson(w, nil, fmt.Errorf("invalid form data"))
			return
		}

		// Parse form fields
		req.Name = r.FormValue("name")
		req.Phone = r.FormValue("phone")
		req.Email = r.FormValue("email")

		// Handle avatar file
		file, header, err := r.FormFile("avatar")
		if err == nil {
			defer file.Close()
			req.AvatarFile = file
			req.AvatarFilename = header.Filename
			req.AvatarContentType = header.Header.Get("Content-Type")
		}
	}

	res, err := h.userSvc.CreateUser(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}

	response.WriteJson(w, res, nil)
}

// GET /users/{id}
func (h *UserHandler) HandleGetUserById(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")

	userID, err := utils.StringToUUID(idStr)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}

	res, err := h.userSvc.GetUserById(r.Context(), &user.GetUserByUserIdReq{UserID: userID})
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}

	response.WriteJson(w, res, nil)
}

// POST /users/list
func (h *UserHandler) HandleListUsers(w http.ResponseWriter, r *http.Request) {
	// page := r.URL.Query().Get("page")
	// limit := r.URL.Query().Get("limit")

	// req := user.ListUsersReq{
	// 	Page:  utils.StringToInt64(page, 0),
	// 	Limit: utils.StringToInt64(limit, 0),
	// }

	var req user.ListUsersReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}

	res, err := h.userSvc.ListUsers(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}

	response.WriteJson(w, res, nil)
}

// POST /users/update
func (h *UserHandler) HandleUpdateUser(w http.ResponseWriter, r *http.Request) {
	var req user.UpdateUserReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}

	res, err := h.userSvc.UpdateUser(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}

	response.WriteJson(w, res, nil)
}
