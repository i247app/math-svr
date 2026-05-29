package profile

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	dto "math-ai.com/math-ai/internal/application/dto/profile"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/infrastructure/metadata"
	"math-ai.com/math-ai/internal/shared/response"
	"math-ai.com/math-ai/internal/shared/utils"
)

const (
	MaxAvatarUploadSize = 10 << 20 // 10 MB
)

type ProfileHandler struct {
	profileSvc *Service
}

func NewProfileHandler(profileSvc *Service) *ProfileHandler {
	return &ProfileHandler{profileSvc: profileSvc}
}

// POST /profiles/create
func (h *ProfileHandler) HandleCreateProfile(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateProfileReq
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
		req.Name = r.FormValue("name")
		req.Role = r.FormValue("role")
		req.IsDefault = utils.StringToBool(r.FormValue("is_default"), false)
		req.Dob = utils.ToStringPtr(r.FormValue("dob"))
		req.SchoolID = utils.ToStringPtr(r.FormValue("school_id"))
		req.GradeID = utils.ToStringPtr(r.FormValue("grade_id"))
		req.ProgramID = utils.ToStringPtr(r.FormValue("program_id"))
		req.SemesterID = utils.ToStringPtr(r.FormValue("semester_id"))

		// Handle avatar file
		file, header, err := r.FormFile("avatar")
		if err == nil {
			defer file.Close()
			req.AvatarFile = file
			req.AvatarFilename = header.Filename
			req.AvatarContentType = header.Header.Get("Content-Type")
		}
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
		Language:  metadata.GetClientLanguage(r.Context()).ToEnumLanguage(),
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

		req.ProfileID = r.FormValue("profile_id")
		req.Name = utils.ToStringPtr(r.FormValue("name"))
		req.Role = utils.ToStringPtr(r.FormValue("role"))
		req.IsDefault = utils.StringToBoolPtr(r.FormValue("is_default"))
		req.Dob = utils.ToStringPtr(r.FormValue("dob"))
		req.SchoolID = utils.ToStringPtr(r.FormValue("school_id"))
		req.GradeID = utils.ToStringPtr(r.FormValue("grade_id"))
		req.ProgramID = utils.ToStringPtr(r.FormValue("program_id"))
		req.SemesterID = utils.ToStringPtr(r.FormValue("semester_id"))

		// Handle avatar file
		file, header, err := r.FormFile("avatar")
		if err == nil {
			defer file.Close()
			req.AvatarFile = file
			req.AvatarFilename = header.Filename
			req.AvatarContentType = header.Header.Get("Content-Type")
		}
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

// POST /profiles/force-delete
func (h *ProfileHandler) HandleForceDeleteProfile(w http.ResponseWriter, r *http.Request) {
	var req dto.DeleteProfileReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}

	res, err := h.profileSvc.ForceDeleteProfile(r.Context(), &req)
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

	r.Body = http.MaxBytesReader(w, r.Body, MaxAvatarUploadSize)
	if err := r.ParseMultipartForm(MaxAvatarUploadSize); err != nil {
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

// POST /profiles/assign-school — body: { profile_id, school_id }
func (h *ProfileHandler) HandleAssignSchool(w http.ResponseWriter, r *http.Request) {
	var req dto.AssignSchoolReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}

	res, err := h.profileSvc.AssignSchool(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	response.WriteJson(w, res, nil)
}

// POST /profiles/remove-school — body: { profile_id }
func (h *ProfileHandler) HandleRemoveSchool(w http.ResponseWriter, r *http.Request) {
	var req dto.RemoveSchoolReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}

	res, err := h.profileSvc.RemoveSchool(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	response.WriteJson(w, res, nil)
}
