package permission

import (
	"encoding/json"
	"fmt"
	"mime"
	"mime/multipart"
	"net/http"
	"strconv"

	dto "math-ai.com/math-ai/internal/application/dto/role"
	"math-ai.com/math-ai/internal/application/resource"
	"math-ai.com/math-ai/internal/shared/response"
	"math-ai.com/math-ai/internal/shared/utils"
)

// MaxRoleImageUploadSize caps the in-memory part of a multipart body, the
// same 10 MB the user avatar and banner uploads use.
const MaxRoleImageUploadSize = 10 << 20

type RoleHandler struct {
	appResource *resource.Resource
	service     *Service
}

func NewRoleHandler(appResource *resource.Resource, service *Service) *RoleHandler {
	return &RoleHandler{
		appResource: appResource,
		service:     service,
	}
}

// actorID pulls the authenticated user's id off the session so create/update
// can stamp create_id/modify_id. Nil when no session is bound (the routes are
// auth-gated, so this is defensive only).
func (h *RoleHandler) actorID(r *http.Request) *int64 {
	sess, err := h.appResource.GetRequestSession(r)
	if err != nil || sess == nil {
		return nil
	}
	uid, ok := sess.UID()
	if !ok {
		return nil
	}
	return &uid
}

// POST /roles/create — JSON, or multipart/form-data with the text fields
// role_code, role_name, description, note, role_status plus an optional
// file part "role_image".
func (h *RoleHandler) HandleCreateRole(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateRoleReq
	if isMultipart(r) {
		if err := r.ParseMultipartForm(MaxRoleImageUploadSize); err != nil {
			response.WriteJson(w, nil, fmt.Errorf("invalid form data"))
			return
		}
		req.RoleCode = r.FormValue("role_code")
		req.RoleName = r.FormValue("role_name")
		// Absent/empty optional fields stay nil — an empty role_status would
		// fail validation instead of falling back to ACTIVE.
		req.Description = utils.ToStringPtr(r.FormValue("description"))
		req.Note = utils.ToStringPtr(r.FormValue("note"))
		req.RoleStatus = utils.ToStringPtr(r.FormValue("role_status"))

		file, header, err := r.FormFile("role_image")
		if err == nil {
			defer file.Close()
			req.RoleImageFile = file
			req.RoleImageFilename = header.Filename
			req.RoleImageContentType = contentTypeOf(header)
		}
	} else if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}

	res, err := h.service.CreateRole(r.Context(), &req, h.actorID(r))
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	response.WriteJson(w, res, nil)
}

// POST /roles/update — JSON, or multipart/form-data with role_id and the
// optional text fields role_name, description, note, role_status plus an
// optional file part "role_image" that replaces the current image, or
// remove_role_image=true to delete it.
func (h *RoleHandler) HandleUpdateRole(w http.ResponseWriter, r *http.Request) {
	var req dto.UpdateRoleReq
	if isMultipart(r) {
		if err := r.ParseMultipartForm(MaxRoleImageUploadSize); err != nil {
			response.WriteJson(w, nil, fmt.Errorf("invalid form data"))
			return
		}
		req.RoleID = utils.StringToInt64(r.FormValue("role_id"), 0)
		// Empty = "leave unchanged", matching an omitted JSON field.
		req.RoleName = utils.ToStringPtr(r.FormValue("role_name"))
		req.Description = utils.ToStringPtr(r.FormValue("description"))
		req.Note = utils.ToStringPtr(r.FormValue("note"))
		req.RoleStatus = utils.ToStringPtr(r.FormValue("role_status"))
		if v := r.FormValue("remove_role_image"); v != "" {
			remove, err := strconv.ParseBool(v)
			if err != nil {
				response.WriteJson(w, nil, fmt.Errorf("invalid remove_role_image"))
				return
			}
			req.RemoveRoleImage = remove
		}

		file, header, err := r.FormFile("role_image")
		if err == nil {
			defer file.Close()
			req.RoleImageFile = file
			req.RoleImageFilename = header.Filename
			req.RoleImageContentType = contentTypeOf(header)
		}
	} else if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}

	res, err := h.service.UpdateRole(r.Context(), &req, h.actorID(r))
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	response.WriteJson(w, res, nil)
}

// POST /roles/soft-delete
func (h *RoleHandler) HandleSoftDeleteRole(w http.ResponseWriter, r *http.Request) {
	var req dto.DeleteRoleReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	res, err := h.service.SoftDeleteRole(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	response.WriteJson(w, res, nil)
}

// POST /roles/force-delete
func (h *RoleHandler) HandleForceDeleteRole(w http.ResponseWriter, r *http.Request) {
	var req dto.DeleteRoleReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	res, err := h.service.ForceDeleteRole(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	response.WriteJson(w, res, nil)
}

// POST /roles/detail
func (h *RoleHandler) HandleGetRole(w http.ResponseWriter, r *http.Request) {
	var req dto.GetRoleReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	res, err := h.service.GetRole(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	response.WriteJson(w, res, nil)
}

// POST /roles/list — also handles search via the optional `search` filter.
func (h *RoleHandler) HandleListRoles(w http.ResponseWriter, r *http.Request) {
	var req dto.ListRolesReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	res, err := h.service.ListRoles(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	response.WriteJson(w, res, nil)
}

// isMultipart reports whether the body is multipart/form-data. Anything else
// is decoded as JSON, so a JSON client keeps working whatever parameters
// (e.g. "; charset=utf-8") it adds to its Content-Type.
func isMultipart(r *http.Request) bool {
	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	return err == nil && mediaType == "multipart/form-data"
}

func contentTypeOf(header *multipart.FileHeader) string {
	return header.Header.Get("Content-Type")
}
