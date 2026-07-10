package user

import (
	"encoding/json"
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

// multipartTextValue returns (value, true) if the named text part was
// present in the multipart form, and ("", false) otherwise. Used to
// distinguish "field absent" (leave avatar_key alone) from "field
// present and empty" (reject as invalid reference) in update flows.
func multipartTextValue(r *http.Request, name string) (string, bool) {
	if r.MultipartForm == nil || r.MultipartForm.Value == nil {
		return "", false
	}
	vs, ok := r.MultipartForm.Value[name]
	if !ok || len(vs) == 0 {
		return "", false
	}
	return vs[0], true
}

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
		req.Role = r.FormValue("role")
		// The multipart text part "avatar" carries a string reference
		// (URL or S3 key). The file part "avatar" carries an upload.
		// FormValue and FormFile read from disjoint maps so the same
		// name coexists; the validator rejects sending both.
		req.Avatar = r.FormValue("avatar_key")

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

// POST /users/detail
func (h *UserHandler) HandleGetUserById(w http.ResponseWriter, r *http.Request) {
	var req user.GetUserByUserIdReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}

	res, err := h.userSvc.GetUserById(r.Context(), &req)
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
		req.UserID = utils.StringToInt64(r.FormValue("user_id"), 0)
		req.Name = utils.ToStringPtr(r.FormValue("name"))
		req.Phone = utils.ToStringPtr(r.FormValue("phone"))
		req.Email = utils.ToStringPtr(r.FormValue("email"))
		req.Role = utils.ToStringPtr(r.FormValue("role"))
		req.Avatar = utils.ToStringPtr(r.FormValue("avatar_key"))

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
				ErrUserIDFormFieldRequired))
		return
	}
	userID := utils.StringToInt64(userIDStr, 0)

	file, header, err := r.FormFile("file")
	if err != nil {
		response.WriteJson(w, nil,
			errs.NewError(ctx, status.USER_AVATAR_INVALID_FILE, nil, err))
		return
	}
	defer file.Close()

	res, err := h.userSvc.UploadAvatar(
		ctx,
		userID,
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
