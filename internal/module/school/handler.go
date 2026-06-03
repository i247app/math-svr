package school

import (
	"encoding/json"
	"net/http"

	dto "math-ai.com/math-ai/internal/application/dto/school"
	"math-ai.com/math-ai/internal/application/resource"
	"math-ai.com/math-ai/internal/shared/response"
	"math-ai.com/math-ai/internal/shared/utils"
)

type SchoolHandler struct {
	appResource *resource.Resource
	schoolSvc   *Service
}

func NewSchoolHandler(appResource *resource.Resource, schoolSvc *Service) *SchoolHandler {
	return &SchoolHandler{
		appResource: appResource,
		schoolSvc:   schoolSvc,
	}
}

// POST /schools/create
func (h *SchoolHandler) HandleCreateSchool(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateSchoolReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}

	res, err := h.schoolSvc.CreateSchool(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}

	response.WriteJson(w, res, nil)
}

// POST /schools/update
func (h *SchoolHandler) HandleUpdateSchool(w http.ResponseWriter, r *http.Request) {
	var req dto.UpdateSchoolReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}

	res, err := h.schoolSvc.UpdateSchool(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}

	response.WriteJson(w, res, nil)
}

// POST /schools/soft-delete
func (h *SchoolHandler) HandleSoftDeleteSchool(w http.ResponseWriter, r *http.Request) {
	var req dto.DeleteSchoolReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}

	res, err := h.schoolSvc.SoftDeleteSchool(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}

	response.WriteJson(w, res, nil)
}

// POST /schools/force-delete
func (h *SchoolHandler) HandleForceDeleteSchool(w http.ResponseWriter, r *http.Request) {
	var req dto.DeleteSchoolReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}

	res, err := h.schoolSvc.ForceDeleteSchool(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}

	response.WriteJson(w, res, nil)
}

// GET /schools/{id}
func (h *SchoolHandler) HandleGetSchool(w http.ResponseWriter, r *http.Request) {
	req := dto.GetSchoolReq{SchoolID: utils.StringToInt64(r.PathValue("id"), 0)}

	res, err := h.schoolSvc.GetSchool(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}

	response.WriteJson(w, res, nil)
}

// POST /schools/list — also handles search via the optional `search` filter.
func (h *SchoolHandler) HandleListSchools(w http.ResponseWriter, r *http.Request) {
	var req dto.ListSchoolsReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}

	res, err := h.schoolSvc.ListSchools(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}

	response.WriteJson(w, res, nil)
}
