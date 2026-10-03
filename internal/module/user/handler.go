package user

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"math-ai.com/math-ai/internal/application/dto/user"
	"math-ai.com/math-ai/internal/application/resource"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/infrastructure/session"
	sctx "math-ai.com/math-ai/internal/shared/context"
	"math-ai.com/math-ai/internal/shared/response"
	"math-ai.com/math-ai/internal/shared/utils"
)

const (
	MaxAvatarUploadSize = 10 << 20 // 10 MB
)

type UserHandler struct {
	appResource *resource.Resource
	service     *Service
}

func NewUserHandler(appResource *resource.Resource, service *Service) *UserHandler {
	return &UserHandler{
		appResource: appResource,
		service:     service,
	}
}

// POST /users/create
func (h *UserHandler) HandleCreateUser(w http.ResponseWriter, r *http.Request) {
	req, err := decodeCreateUserReq(r)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	defer closeAvatarFile(req)

	sess, err := h.appResource.GetRequestSession(r)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}

	res, err := h.service.CreateUser(r.Context(), sess, req)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}

	response.WriteJson(w, res, nil)
}

// POST /users/admin/create
func (h *UserHandler) HandleAdminCreateUser(w http.ResponseWriter, r *http.Request) {
	req, err := decodeCreateUserReq(r)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	defer closeAvatarFile(req)

	// A key caller has no account (and may have no session at all), so
	// there is no acting uid to record; an ADMIN session caller has one.
	var adminUID *int64
	if sctx.GetAdminVia(r.Context()) != sctx.AdminViaAPIKey {
		sess, err := h.appResource.GetRequestSession(r)
		if err != nil {
			response.WriteJson(w, nil, err)
			return
		}
		uid, ok := sess.UID()
		if !ok {
			response.WriteJson(w, nil, errs.NewUnauthorizedError(r.Context(), session.ErrUidNotFoundFromSession))
			return
		}
		adminUID = &uid
	}

	res, err := h.service.AdminCreateUser(r.Context(), adminUID, req)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}

	response.WriteJson(w, res, nil)
}

// closeAvatarFile closes the multipart avatar decodeCreateUserReq opened.
func closeAvatarFile(req *user.CreateUserReq) {
	if c, ok := req.AvatarFile.(io.Closer); ok {
		_ = c.Close()
	}
}

// decodeCreateUserReq reads a create-user body: JSON, or multipart when
// it carries an avatar file.
func decodeCreateUserReq(r *http.Request) (*user.CreateUserReq, error) {
	var req user.CreateUserReq
	if r.Header.Get("Content-Type") == "application/json" {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			return nil, fmt.Errorf("invalid parameters")
		}
		return &req, nil
	}

	if err := r.ParseMultipartForm(MaxAvatarUploadSize); err != nil {
		return nil, fmt.Errorf("invalid form data")
	}
	req.Name = r.FormValue("name")
	req.Phone = r.FormValue("phone")
	req.Email = r.FormValue("email")
	req.Role = r.FormValue("role")
	req.Password = r.FormValue("password")
	// The multipart text part "avatar" carries a string reference
	// (URL or S3 key). The file part "avatar" carries an upload.
	// FormValue and FormFile read from disjoint maps so the same
	// name coexists; the validator rejects sending both.
	req.Avatar = r.FormValue("avatar_key")

	// The caller closes the file (closeAvatarFile) once the service is done.
	if file, header, err := r.FormFile("avatar"); err == nil {
		req.AvatarFile = file
		req.AvatarFilename = header.Filename
		req.AvatarContentType = header.Header.Get("Content-Type")
	}
	return &req, nil
}

// POST /users/create/guest
//
// Public on purpose: it is the one route a visitor with no account can
// call to GET one. The body may be empty — the guest is identified by
// metadata.device_uuid, which every request already carries — and the
// session it returns is unsecure, so it opens nothing but /exams/*.
//
// A missing or malformed body is not an error here: there is nothing in
// it the server needs, so an undecodable body is treated as an empty one
// rather than refused.
func (h *UserHandler) HandleCreateGuest(w http.ResponseWriter, r *http.Request) {
	var req user.CreateGuestReq
	_ = json.NewDecoder(r.Body).Decode(&req)

	sess, err := h.appResource.GetRequestSession(r)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}

	res, err := h.service.CreateGuest(r.Context(), sess, &req)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}

	response.WriteJson(w, res, nil)
}

// POST /users/detail
func (h *UserHandler) HandleGetUserById(w http.ResponseWriter, r *http.Request) {
	var req user.GetUserByUidReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}

	if req.UID == 0 {
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
		req.UID = uid
	}

	res, err := h.service.GetUserById(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}

	response.WriteJson(w, res, nil)
}

// POST /users/identifier-available
func (h *UserHandler) HandleIdentifierAvailable(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req user.CheckIdentifierReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}

	if h.appResource != nil {
		req.EmailOTPEnable = h.appResource.Env.EmailOTPEnable
		req.PhoneOTPEnable = h.appResource.Env.PhoneOTPEnable
	}

	res, err := h.service.CheckIdentifierAvailable(ctx, &req)
	if err != nil {
		response.WriteJson(w, res, err)
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

	res, err := h.service.GetUserById(r.Context(), &user.GetUserByUidReq{UID: uid})
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

	res, err := h.service.ListUsers(r.Context(), &req)
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
		req.UID = utils.StringToInt64(r.FormValue("uid"), 0)
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

	res, err := h.service.UpdateUser(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}

	response.WriteJson(w, res, nil)
}

// POST /users/upload-avatar — multipart form with fields:
//
//	uid  string (uuid)
//	file     file
//
// Mirrors /profiles/upload-avatar so the mobile client can use one
// uploader for both endpoints. uid is required (no implicit
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

	uidStr := r.FormValue("uid")
	if uidStr == "" {
		response.WriteJson(w, nil,
			errs.NewError(ctx, status.USER_NOT_FOUND, nil,
				ErrUIDFormFieldRequired))
		return
	}
	uid := utils.StringToInt64(uidStr, 0)

	file, header, err := r.FormFile("file")
	if err != nil {
		response.WriteJson(w, nil,
			errs.NewError(ctx, status.USER_AVATAR_INVALID_FILE, nil, err))
		return
	}
	defer file.Close()

	res, err := h.service.UploadAvatar(
		ctx,
		uid,
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
