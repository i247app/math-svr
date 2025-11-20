package controller

import (
	"encoding/json"
	"fmt"
	"net/http"

	"math-ai.com/math-ai/internal/app/resources"
	"math-ai.com/math-ai/internal/applications/dto"
	di "math-ai.com/math-ai/internal/core/di/services"
	"math-ai.com/math-ai/internal/shared/constant/status"
	"math-ai.com/math-ai/internal/shared/utils/response"
)

type UserController struct {
	appResources *resources.AppResource
	service      di.IUserService
}

func NewUserController(appResources *resources.AppResource, service di.IUserService) *UserController {
	return &UserController{
		appResources: appResources,
		service:      service,
	}
}

// Get - /users/list
func (u *UserController) HandlerGetListUsers(w http.ResponseWriter, r *http.Request) {
	var req dto.ListUserRequest

	statusCode, users, pagination, err := u.service.ListUsers(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, r.Context(), nil, err, statusCode)
		return
	}

	res := &dto.ListUserResponse{
		Items:      users,
		Pagination: pagination,
	}

	response.WriteJson(w, r.Context(), res, nil, statusCode)
}

// Get - /users/{id}
func (u *UserController) HandlerGetUser(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("id")

	statusCode, user, err := u.service.GetUserByID(r.Context(), userID)
	if err != nil {
		response.WriteJson(w, r.Context(), nil, err, statusCode)
		return
	}

	res := &dto.GetUserResponse{
		User: user,
	}

	response.WriteJson(w, r.Context(), res, nil, statusCode)
}

// Get - /users/profile
func (u *UserController) HandlerGetProfile(w http.ResponseWriter, r *http.Request) {
	var userID string

	statusCode, user, err := u.service.GetUserByID(r.Context(), userID)
	if err != nil {
		response.WriteJson(w, r.Context(), nil, err, statusCode)
		return
	}

	res := &dto.GetUserResponse{
		User: user,
	}

	response.WriteJson(w, r.Context(), res, nil, statusCode)
}

// POST - /users/create
func (u *UserController) HandlerCreateUser(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, r.Context(), nil, fmt.Errorf("invalid parameters"), status.BAD_REQUEST)
		return
	}

	req.DeviceUUID = r.Header.Get("Device-UUID")
	req.DeviceName = r.Header.Get("Device-Name")

	statusCode, user, err := u.service.CreateUser(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, r.Context(), nil, err, statusCode)
		return
	}

	res := &dto.CreateUserResponse{
		User: user,
	}

	response.WriteJson(w, r.Context(), res, nil, statusCode)
}

// POST - /users/update
func (u *UserController) HandlerUpdateUser(w http.ResponseWriter, r *http.Request) {
	var req dto.UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, r.Context(), nil, fmt.Errorf("invalid parameters"), status.BAD_REQUEST)
		return
	}

	statusCode, user, err := u.service.UpdateUser(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, r.Context(), nil, err, statusCode)
		return
	}

	res := &dto.UpdateUserResponse{
		User: user,
	}

	response.WriteJson(w, r.Context(), res, nil, statusCode)
}

// POST - /users/delete
func (u *UserController) HandlerDeleteUser(w http.ResponseWriter, r *http.Request) {
	var req dto.DeleteUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, r.Context(), nil, fmt.Errorf("invalid parameters"), status.BAD_REQUEST)
		return
	}

	statusCode, err := u.service.DeleteUser(r.Context(), req.UserID)
	if err != nil {
		response.WriteJson(w, r.Context(), nil, err, statusCode)
		return
	}

	response.WriteJson(w, r.Context(), "Delete successfully", nil, statusCode)
}

// POST - /users/force-delete
func (u *UserController) HandlerForceDeleteUser(w http.ResponseWriter, r *http.Request) {
	var req dto.DeleteUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, r.Context(), nil, fmt.Errorf("invalid parameters"), status.BAD_REQUEST)
		return
	}

	statusCode, err := u.service.ForceDeleteUser(r.Context(), req.UserID)
	if err != nil {
		response.WriteJson(w, r.Context(), nil, err, statusCode)
		return
	}

	response.WriteJson(w, r.Context(), "Force delete successfully", nil, statusCode)
}
