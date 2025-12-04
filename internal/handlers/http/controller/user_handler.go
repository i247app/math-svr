package controller

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"math-ai.com/math-ai/internal/app/resources"
	"math-ai.com/math-ai/internal/applications/dto"
	di "math-ai.com/math-ai/internal/core/di/services"
	"math-ai.com/math-ai/internal/shared/constant/enum"
	"math-ai.com/math-ai/internal/shared/constant/status"
	"math-ai.com/math-ai/internal/shared/utils/response"
)

const (
	MaxAvatarUploadSize = 10 << 20 // 10 MB
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

// POST - /users/create
func (u *UserController) HandlerCreateUser(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateUserRequest

	// Check content type - support both JSON and multipart form
	contentType := r.Header.Get("Content-Type")

	if contentType == "application/json" {
		// JSON request (no avatar)
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			response.WriteJson(w, r.Context(), nil, fmt.Errorf("invalid parameters"), status.BAD_REQUEST)
			return
		}
	} else {
		// Multipart form request (with avatar)
		if err := r.ParseMultipartForm(MaxAvatarUploadSize); err != nil {
			////logger.Errorf("Failed to parse multipart form: %v", err)
			response.WriteJson(w, r.Context(), nil, fmt.Errorf("invalid form data"), status.BAD_REQUEST)
			return
		}

		// Parse form fields
		req.Name = r.FormValue("name")
		req.Phone = r.FormValue("phone")
		req.Email = r.FormValue("email")
		req.Password = r.FormValue("password")

		// Parse role
		if roleStr := r.FormValue("role"); roleStr != "" {
			req.Role = enum.ERole(roleStr)
		}

		// Parse DOB
		if dobStr := r.FormValue("dob"); dobStr != "" {
			req.Dob = &dobStr
		}

		// Handle avatar file
		file, header, err := r.FormFile("avatar")
		if err == nil {
			defer file.Close()
			req.AvatarFile = file
			req.AvatarFilename = header.Filename
			req.AvatarContentType = header.Header.Get("Content-Type")
		}
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

	// Check content type - support both JSON and multipart form
	contentType := r.Header.Get("Content-Type")

	if contentType == "application/json" {
		// JSON request (no avatar)
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			response.WriteJson(w, r.Context(), nil, fmt.Errorf("invalid parameters"), status.BAD_REQUEST)
			return
		}
	} else {
		// Multipart form request (with avatar)
		if err := r.ParseMultipartForm(MaxAvatarUploadSize); err != nil {
			////logger.Errorf("Failed to parse multipart form: %v", err)
			response.WriteJson(w, r.Context(), nil, fmt.Errorf("invalid form data"), status.BAD_REQUEST)
			return
		}

		// Required field
		req.UID = r.FormValue("uid")

		// Optional fields
		if name := r.FormValue("name"); name != "" {
			req.Name = &name
		}
		if phone := r.FormValue("phone"); phone != "" {
			req.Phone = &phone
		}
		if email := r.FormValue("email"); email != "" {
			req.Email = &email
		}
		if dob := r.FormValue("dob"); dob != "" {
			req.Dob = &dob
		}
		if roleStr := r.FormValue("role"); roleStr != "" {
			role := enum.ERole(roleStr)
			req.Role = &role
		}
		if statusStr := r.FormValue("status"); statusStr != "" {
			stat := enum.EStatus(statusStr)
			req.Status = &stat
		}
		if grade := r.FormValue("grade"); grade != "" {
			req.Grade = &grade
		}
		if level := r.FormValue("level"); level != "" {
			req.Level = &level
		}

		// Parse delete_avatar flag
		if deleteAvatarStr := r.FormValue("delete_avatar"); deleteAvatarStr != "" {
			if deleteAvatar, err := strconv.ParseBool(deleteAvatarStr); err == nil {
				req.DeleteAvatar = deleteAvatar
			}
		}

		// Handle avatar file
		file, header, err := r.FormFile("avatar")
		if err == nil {
			defer file.Close()
			req.AvatarFile = file
			req.AvatarFilename = header.Filename
			req.AvatarContentType = header.Header.Get("Content-Type")
		}
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

	statusCode, err := u.service.DeleteUser(r.Context(), &req)
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

	statusCode, err := u.service.ForceDeleteUser(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, r.Context(), nil, err, statusCode)
		return
	}

	response.WriteJson(w, r.Context(), "Force delete successfully", nil, statusCode)
}
