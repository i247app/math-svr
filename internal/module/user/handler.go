package user

import (
	"encoding/json"
	"net/http"

	"math-ai.com/math-ai/internal/application/dto/user"
	"math-ai.com/math-ai/internal/shared/response"
	"math-ai.com/math-ai/internal/shared/utils"
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
func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req user.CreateUserReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}

	res, err := h.userSvc.CreateUser(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}

	response.WriteJson(w, res, nil)
}

// GET /users/{id}
func (h *UserHandler) GetUserById(w http.ResponseWriter, r *http.Request) {
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
func (h *UserHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
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
func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
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
