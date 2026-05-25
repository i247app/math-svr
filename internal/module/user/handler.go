package user

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"math-ai.com/math-ai/internal/application/dto/user"
	"math-ai.com/math-ai/internal/application/resource"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/shared/response"
	"math-ai.com/math-ai/internal/shared/utils"
)

const (
	MaxAvatarUploadSize = 10 << 20 // 10 MB
)

type UserHandler struct {
	appResource *resource.Resource
	userSvc     *Service
}

func NewUserHandler(appResource *resource.Resource, userSvc *Service) *UserHandler {
	return &UserHandler{
		appResource: appResource,
		userSvc:     userSvc,
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

	sess, err := h.appResource.GetRequestSession(r)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}

	res, err := h.userSvc.CreateUser(r.Context(), sess, &req)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}

	response.WriteJson(w, res, nil)
}

// GET /users/{id}
func (h *UserHandler) HandleGetUserById(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")

	res, err := h.userSvc.GetUserById(r.Context(), &user.GetUserByUserIdReq{UserID: idStr})
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}

	response.WriteJson(w, res, nil)
}

// Get /users/me
func (h *UserHandler) HandleGetUserMe(w http.ResponseWriter, r *http.Request) {
	session, err := h.appResource.GetRequestSession(r)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}

	uid, ok := session.UID()
	if !ok {
		response.WriteJson(w, nil, fmt.Errorf("invalid session"))
		return
	}

	res, err := h.userSvc.GetUserById(r.Context(), &user.GetUserByUserIdReq{UserID: uid})
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}

	response.WriteJson(w, res, nil)
}

// POST /users/list
func (h *UserHandler) HandleListUsers(w http.ResponseWriter, r *http.Request) {
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
		req.UserID = r.FormValue("user_id")
		req.Name = utils.ToStringPtr(r.FormValue("name"))
		req.Phone = utils.ToStringPtr(r.FormValue("phone"))
		req.Email = utils.ToStringPtr(r.FormValue("email"))

		// Handle avatar file
		file, header, err := r.FormFile("avatar")
		if err == nil {
			defer file.Close()
			req.AvatarFile = file
			req.AvatarFilename = header.Filename
			req.AvatarContentType = header.Header.Get("Content-Type")
		}
	}

	res, err := h.userSvc.UpdateUser(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}

	response.WriteJson(w, res, nil)
}

// POST /users/upload-avatar — multipart form with fields:
//
//	user_id  string (uuid)
//	file     file
//
// Mirrors /profiles/upload-avatar so the mobile client can use one
// uploader for both endpoints. user_id is required (no implicit
// "current session" — the parent might be uploading on behalf of a
// distinct account in admin flows).
func (h *UserHandler) HandleUploadAvatar(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	r.Body = http.MaxBytesReader(w, r.Body, MaxAvatarUploadSize)
	if err := r.ParseMultipartForm(MaxAvatarUploadSize); err != nil {
		response.WriteJson(w, nil,
			errs.NewError(ctx, status.USER_AVATAR_INVALID_FILE, nil, err))
		return
	}

	userIDStr := r.FormValue("user_id")
	if userIDStr == "" {
		response.WriteJson(w, nil,
			errs.NewError(ctx, status.USER_NOT_FOUND, nil,
				errors.New("user_id form field is required")))
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		response.WriteJson(w, nil,
			errs.NewError(ctx, status.USER_AVATAR_INVALID_FILE, nil, err))
		return
	}
	defer file.Close()

	res, err := h.userSvc.UploadAvatar(
		ctx,
		userIDStr,
		header.Filename,
		header.Header.Get("Content-Type"),
		file,
	)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	response.WriteJson(w, res, nil)
}
