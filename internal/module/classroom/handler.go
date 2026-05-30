package classroom

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	dto "math-ai.com/math-ai/internal/application/dto/classroom"
	"math-ai.com/math-ai/internal/application/resource"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/shared/response"
	"math-ai.com/math-ai/internal/shared/utils"
)

const (
	MaxAvatarUploadSize = 10 << 20 // 10 MB
)

type ClassroomHandler struct {
	appResource  *resource.Resource
	classroomSvc *Service
}

func NewClassroomHandler(appResource *resource.Resource, classroomSvc *Service) *ClassroomHandler {
	return &ClassroomHandler{
		appResource:  appResource,
		classroomSvc: classroomSvc,
	}
}

// sessionUID pulls the authenticated user's id off the session and
// returns "" + a MathError if no session is attached. All mutating
// routes funnel through this so the §0 Q1 contract (caller's profile
// must belong to the session user) is enforced uniformly.
func (h *ClassroomHandler) sessionUID(r *http.Request) (string, error) {
	sess, err := h.appResource.GetRequestSession(r)
	if err != nil {
		return "", err
	}
	uid, ok := sess.UID()
	if !ok {
		return "", errs.NewError(r.Context(), status.UNAUTHORIZED, nil,
			errors.New("uid not found from session"))
	}
	return uid, nil
}

// POST /classrooms/create
func (h *ClassroomHandler) HandleCreateClassroom(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateClassroomReq
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
		req.Name = r.FormValue("name")
		req.Description = utils.ToStringPtr(r.FormValue("description"))
		req.ProgramID = utils.ToStringPtr(r.FormValue("program_id"))
		req.GradeID = utils.ToStringPtr(r.FormValue("grade_id"))
		maxNumbers := r.FormValue("max_members")
		if maxNumbers != "" {
			num, err := utils.StringToInt64Err(maxNumbers)
			if err == nil {
				req.MaxMembers = &num
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

	uid, err := h.sessionUID(r)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	res, err := h.classroomSvc.CreateClassroom(r.Context(), &req, uid)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	response.WriteJson(w, res, nil)
}

// POST /classrooms/update
func (h *ClassroomHandler) HandleUpdateClassroom(w http.ResponseWriter, r *http.Request) {
	var req dto.UpdateClassroomReq
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
		req.ClassroomID = r.FormValue("classroom_id")
		req.Name = utils.ToStringPtr(r.FormValue("name"))
		req.Description = utils.ToStringPtr(r.FormValue("description"))
		req.ProgramID = utils.ToStringPtr(r.FormValue("program_id"))
		req.GradeID = utils.ToStringPtr(r.FormValue("grade_id"))
		maxNumbers := r.FormValue("max_members")
		if maxNumbers != "" {
			num, err := utils.StringToInt64Err(maxNumbers)
			if err == nil {
				req.MaxMembers = &num
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

	uid, err := h.sessionUID(r)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	res, err := h.classroomSvc.UpdateClassroom(r.Context(), &req, uid)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	response.WriteJson(w, res, nil)
}

// POST /classrooms/list
func (h *ClassroomHandler) HandleListClassrooms(w http.ResponseWriter, r *http.Request) {
	var req dto.ListClassroomsReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	uid, err := h.sessionUID(r)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	res, err := h.classroomSvc.ListClassrooms(r.Context(), &req, uid)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	response.WriteJson(w, res, nil)
}

// GET /classrooms/{id}?profile_id=...
// profile_id arrives as a query parameter for the GET path; POST callers
// can use /classrooms/get (not yet exposed — GET is the canonical read).
func (h *ClassroomHandler) HandleGetClassroom(w http.ResponseWriter, r *http.Request) {
	req := dto.GetClassroomReq{
		ProfileID:   r.URL.Query().Get("profile_id"),
		ClassroomID: r.PathValue("id"),
	}
	uid, err := h.sessionUID(r)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	res, err := h.classroomSvc.GetClassroom(r.Context(), &req, uid)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	response.WriteJson(w, res, nil)
}

// POST /classrooms/archive
func (h *ClassroomHandler) HandleArchiveClassroom(w http.ResponseWriter, r *http.Request) {
	var req dto.ArchiveClassroomReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	uid, err := h.sessionUID(r)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	res, err := h.classroomSvc.ArchiveClassroom(r.Context(), &req, uid)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	response.WriteJson(w, res, nil)
}

// POST /classrooms/restore
func (h *ClassroomHandler) HandleRestoreClassroom(w http.ResponseWriter, r *http.Request) {
	var req dto.RestoreClassroomReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	uid, err := h.sessionUID(r)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	res, err := h.classroomSvc.RestoreClassroom(r.Context(), &req, uid)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	response.WriteJson(w, res, nil)
}

// POST /classrooms/soft-delete
func (h *ClassroomHandler) HandleSoftDeleteClassroom(w http.ResponseWriter, r *http.Request) {
	var req dto.DeleteClassroomReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	uid, err := h.sessionUID(r)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	res, err := h.classroomSvc.SoftDeleteClassroom(r.Context(), &req, uid)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	response.WriteJson(w, res, nil)
}

// POST /classrooms/force-delete
func (h *ClassroomHandler) HandleForceDeleteClassroom(w http.ResponseWriter, r *http.Request) {
	var req dto.DeleteClassroomReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	uid, err := h.sessionUID(r)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	res, err := h.classroomSvc.ForceDeleteClassroom(r.Context(), &req, uid)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	response.WriteJson(w, res, nil)
}

// POST /classrooms/join-by-code
func (h *ClassroomHandler) HandleJoinByCode(w http.ResponseWriter, r *http.Request) {
	var req dto.JoinByCodeReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	uid, err := h.sessionUID(r)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	res, err := h.classroomSvc.JoinClassroomByCode(r.Context(), &req, uid)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	response.WriteJson(w, res, nil)
}

// POST /classrooms/leave
func (h *ClassroomHandler) HandleLeaveClassroom(w http.ResponseWriter, r *http.Request) {
	var req dto.LeaveClassroomReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	uid, err := h.sessionUID(r)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	res, err := h.classroomSvc.LeaveClassroom(r.Context(), &req, uid)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	response.WriteJson(w, res, nil)
}

// POST /classrooms/members/remove
func (h *ClassroomHandler) HandleRemoveMember(w http.ResponseWriter, r *http.Request) {
	var req dto.RemoveMemberReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	uid, err := h.sessionUID(r)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	res, err := h.classroomSvc.RemoveMember(r.Context(), &req, uid)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	response.WriteJson(w, res, nil)
}

// POST /classrooms/members/update-role
func (h *ClassroomHandler) HandleUpdateMemberRole(w http.ResponseWriter, r *http.Request) {
	var req dto.UpdateMemberRoleReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	uid, err := h.sessionUID(r)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	res, err := h.classroomSvc.UpdateMemberRole(r.Context(), &req, uid)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	response.WriteJson(w, res, nil)
}

// POST /classrooms/transfer-ownership
func (h *ClassroomHandler) HandleTransferOwnership(w http.ResponseWriter, r *http.Request) {
	var req dto.TransferOwnershipReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	uid, err := h.sessionUID(r)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	res, err := h.classroomSvc.TransferOwnership(r.Context(), &req, uid)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	response.WriteJson(w, res, nil)
}

// POST /classrooms/members/list
func (h *ClassroomHandler) HandleListMembers(w http.ResponseWriter, r *http.Request) {
	var req dto.ListMembersReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	uid, err := h.sessionUID(r)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	res, err := h.classroomSvc.ListMembers(r.Context(), &req, uid)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	response.WriteJson(w, res, nil)
}
