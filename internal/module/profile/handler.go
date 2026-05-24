package profile

import (
	"encoding/json"
	"errors"
	"net/http"

	dto "math-ai.com/math-ai/internal/application/dto/profile"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/shared/enum"
	"math-ai.com/math-ai/internal/shared/response"
)

// maxAvatarUploadBytes caps an avatar multipart request before we open
// the file. 8 MiB is generous for a user-facing image.
const maxAvatarUploadBytes = 8 << 20

type ProfileHandler struct {
	profileSvc *Service
}

func NewProfileHandler(profileSvc *Service) *ProfileHandler {
	return &ProfileHandler{profileSvc: profileSvc}
}

// POST /profiles/create
func (h *ProfileHandler) HandleCreateProfile(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateProfileReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}

	res, err := h.profileSvc.CreateProfile(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	response.WriteJson(w, res, nil)
}

// GET /profiles/{id}?language=vn|en
func (h *ProfileHandler) HandleGetProfileById(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")

	res, err := h.profileSvc.GetProfileById(r.Context(), &dto.GetProfileByIdReq{
		ProfileID: idStr,
		Language:  enum.LanguageType(r.URL.Query().Get("language")),
	})
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	response.WriteJson(w, res, nil)
}

// POST /profiles/list
func (h *ProfileHandler) HandleListProfiles(w http.ResponseWriter, r *http.Request) {
	var req dto.ListProfilesReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}

	res, err := h.profileSvc.ListProfilesByUserId(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	response.WriteJson(w, res, nil)
}

// POST /profiles/update
func (h *ProfileHandler) HandleUpdateProfile(w http.ResponseWriter, r *http.Request) {
	var req dto.UpdateProfileReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}

	res, err := h.profileSvc.UpdateProfile(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	response.WriteJson(w, res, nil)
}

// POST /profiles/soft-delete
func (h *ProfileHandler) HandleSoftDeleteProfile(w http.ResponseWriter, r *http.Request) {
	var req dto.DeleteProfileReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}

	res, err := h.profileSvc.SoftDeleteProfile(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	response.WriteJson(w, res, nil)
}

// POST /profiles/upload-avatar — multipart form with fields:
//
//	profile_id  string (uuid)
//	file        file
func (h *ProfileHandler) HandleUploadAvatar(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	r.Body = http.MaxBytesReader(w, r.Body, maxAvatarUploadBytes)
	if err := r.ParseMultipartForm(maxAvatarUploadBytes); err != nil {
		response.WriteJson(w, nil,
			errs.NewError(ctx, status.PROFILE_AVATAR_INVALID_FILE, nil, err))
		return
	}

	profileIDStr := r.FormValue("profile_id")
	if profileIDStr == "" {
		response.WriteJson(w, nil,
			errs.NewError(ctx, status.PROFILE_NOT_FOUND, nil,
				errors.New("profile_id form field is required")))
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		response.WriteJson(w, nil,
			errs.NewError(ctx, status.PROFILE_AVATAR_INVALID_FILE, nil, err))
		return
	}
	defer file.Close()

	res, err := h.profileSvc.UploadAvatar(
		ctx,
		profileIDStr,
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
