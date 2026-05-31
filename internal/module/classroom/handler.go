package classroom

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	dto "math-ai.com/math-ai/internal/application/dto/classroom"
	"math-ai.com/math-ai/internal/application/resource"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	mtime "math-ai.com/math-ai/internal/domain/shared/time"
	"math-ai.com/math-ai/internal/infrastructure/logger"
	"math-ai.com/math-ai/internal/shared/response"
	"math-ai.com/math-ai/internal/shared/utils"
)

// hasProgramIDsForm returns true if the multipart payload contains a
// program_ids key at all (even with an empty value). The Update path
// uses this to distinguish "leave links alone" (nil) from "clear links"
// (non-nil empty slice), since multipart cannot express JSON null.
func hasProgramIDsForm(r *http.Request) bool {
	if r.MultipartForm == nil {
		return false
	}
	_, repeated := r.MultipartForm.Value["program_ids"]
	_, single := r.MultipartForm.Value["program_id"]
	return repeated || single
}

// parseProgramIDsFromForm reads the multipart payload's program_ids
// entries (repeated keys) and, as a fallback, a single comma-separated
// program_ids value. Blanks are dropped; deduplication is left to the
// validator so it can surface CLASSROOM_PROGRAM_DUPLICATE consistently
// between JSON and multipart callers.
func parseProgramIDsFromForm(r *http.Request) []int64 {
	out := make([]int64, 0)
	if r.MultipartForm != nil {
		for _, v := range r.MultipartForm.Value["program_ids"] {
			for _, part := range strings.Split(v, ",") {
				if trimmed := strings.TrimSpace(part); trimmed != "" {
					out = append(out, utils.StringToInt64(trimmed, 0))
				}
			}
		}
		// Legacy "program_id" single field still parsed so existing
		// clients keep working through the migration window.
		for _, v := range r.MultipartForm.Value["program_id"] {
			if trimmed := strings.TrimSpace(v); trimmed != "" {
				out = append(out, utils.StringToInt64(trimmed, 0))
			}
		}
	}
	return out
}

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
func (h *ClassroomHandler) sessionUID(r *http.Request) (int64, error) {
	sess, err := h.appResource.GetRequestSession(r)
	if err != nil {
		return 0, err
	}
	uid, ok := sess.UID()
	if !ok {
		return 0, errs.NewError(r.Context(), status.UNAUTHORIZED, nil,
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
		req.ProfileID = utils.StringToInt64(r.FormValue("profile_id"), 0)
		req.Name = r.FormValue("name")
		req.Description = utils.ToStringPtr(r.FormValue("description"))
		req.SchoolID = utils.StringToInt64Ptr(r.FormValue("school_id"))
		req.ProgramIDs = parseProgramIDsFromForm(r)
		req.GradeID = utils.StringToInt64Ptr(r.FormValue("grade_id"))
		req.InviteCode = utils.ToStringPtr(r.FormValue("invite_code"))
		if expires := r.FormValue("invite_code_expires_dt"); expires != "" {
			if parsed, err := mtime.Parse(expires); err == nil {
				req.InviteCodeExpiresDt = parsed
			} else {
				logger.From(r.Context()).Warnf("classroom.create invite_code_expires_dt parse failed value=%s err=%v", expires, err)
			}
		}
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
		req.ProfileID = utils.StringToInt64(r.FormValue("profile_id"), 0)
		req.ClassroomID = utils.StringToInt64(r.FormValue("classroom_id"), 0)
		req.Name = utils.ToStringPtr(r.FormValue("name"))
		req.Description = utils.ToStringPtr(r.FormValue("description"))
		req.SchoolID = utils.StringToInt64Ptr(r.FormValue("school_id"))
		if hasProgramIDsForm(r) {
			ids := parseProgramIDsFromForm(r)
			req.ProgramIDs = &ids
		}
		req.GradeID = utils.StringToInt64Ptr(r.FormValue("grade_id"))
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
		ProfileID:   utils.StringToInt64(r.URL.Query().Get("profile_id"), 0),
		ClassroomID: utils.StringToInt64(r.PathValue("id"), 0),
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

// POST /classrooms/invitations/send
func (h *ClassroomHandler) HandleSendInvitation(w http.ResponseWriter, r *http.Request) {
	var req dto.SendInvitationReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	uid, err := h.sessionUID(r)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	res, err := h.classroomSvc.SendInvitation(r.Context(), &req, uid)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	response.WriteJson(w, res, nil)
}

// POST /classrooms/invitations/my-pending
func (h *ClassroomHandler) HandleListMyPendingInvitations(w http.ResponseWriter, r *http.Request) {
	var req dto.ListMyPendingInvitationsReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	uid, err := h.sessionUID(r)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	res, err := h.classroomSvc.ListMyPendingInvitations(r.Context(), &req, uid)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	response.WriteJson(w, res, nil)
}

// POST /classrooms/invitations/list
func (h *ClassroomHandler) HandleListClassroomInvitations(w http.ResponseWriter, r *http.Request) {
	var req dto.ListClassroomInvitationsReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	uid, err := h.sessionUID(r)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	res, err := h.classroomSvc.ListClassroomInvitations(r.Context(), &req, uid)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	response.WriteJson(w, res, nil)
}

// POST /classrooms/invitations/accept
func (h *ClassroomHandler) HandleAcceptInvitation(w http.ResponseWriter, r *http.Request) {
	var req dto.AcceptInvitationReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	uid, err := h.sessionUID(r)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	res, err := h.classroomSvc.AcceptInvitation(r.Context(), &req, uid)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	response.WriteJson(w, res, nil)
}

// POST /classrooms/invitations/reject
func (h *ClassroomHandler) HandleRejectInvitation(w http.ResponseWriter, r *http.Request) {
	var req dto.RejectInvitationReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	uid, err := h.sessionUID(r)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	res, err := h.classroomSvc.RejectInvitation(r.Context(), &req, uid)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	response.WriteJson(w, res, nil)
}

// POST /classrooms/invitations/cancel
func (h *ClassroomHandler) HandleCancelInvitation(w http.ResponseWriter, r *http.Request) {
	var req dto.CancelInvitationReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	uid, err := h.sessionUID(r)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	res, err := h.classroomSvc.CancelInvitation(r.Context(), &req, uid)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	response.WriteJson(w, res, nil)
}
