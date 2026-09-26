package grade

import (
	"encoding/json"
	"fmt"
	"net/http"

	dto "math-ai.com/math-ai/internal/application/dto/grade"
	"math-ai.com/math-ai/internal/shared/response"
	"math-ai.com/math-ai/internal/shared/utils"
)

const (
	MaxImageUploadSize = 10 << 20 // 10 MB
)

type GradeHandler struct {
	gradeSvc *Service
}

func NewGradeHandler(gradeSvc *Service) *GradeHandler {
	return &GradeHandler{gradeSvc: gradeSvc}
}

// POST /grades/list
func (h *GradeHandler) HandleListGrades(w http.ResponseWriter, r *http.Request) {
	var req dto.ListGradesReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}

	res, err := h.gradeSvc.ListGrades(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}

	response.WriteJson(w, res, nil)
}

// POST /grades/create
func (h *GradeHandler) HandleCreateGrade(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateGradeReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}

	res, err := h.gradeSvc.CreateGrade(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}

	response.WriteJson(w, res, nil)
}

// POST /grades/update
func (h *GradeHandler) HandleUpdateGrade(w http.ResponseWriter, r *http.Request) {
	var req dto.UpdateGradeReq
	contentType := r.Header.Get("Content-Type")

	if contentType == "application/json" {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			response.WriteJson(w, nil, fmt.Errorf("invalid parameters"))
			return
		}
	} else {
		if err := r.ParseMultipartForm(MaxImageUploadSize); err != nil {
			response.WriteJson(w, nil, fmt.Errorf("invalid form data"))
			return
		}
		req.GradeID = utils.StringToInt64(r.FormValue("grade_id"), 0)
		req.Label = utils.ToStringPtr(r.FormValue("label"))
		req.Description = utils.ToStringPtr(r.FormValue("description"))
		req.Note = utils.ToStringPtr(r.FormValue("note"))

		// Handle avatar file
		file, header, err := r.FormFile("image")
		if err == nil {
			defer file.Close()
			req.ImageFile = file
			req.ImageFilename = header.Filename
			req.ImageContentType = header.Header.Get("Content-Type")
		}
	}

	res, err := h.gradeSvc.UpdateGrade(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}

	response.WriteJson(w, res, nil)
}

// POST /grades/soft-delete
func (h *GradeHandler) HandleSoftDeleteGrade(w http.ResponseWriter, r *http.Request) {
	var req dto.DeleteGradeReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}

	res, err := h.gradeSvc.SoftDeleteGrade(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}

	response.WriteJson(w, res, nil)
}

// POST /grades/force-delete
func (h *GradeHandler) HandleForceDeleteGrade(w http.ResponseWriter, r *http.Request) {
	var req dto.DeleteGradeReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}

	res, err := h.gradeSvc.ForceDeleteGrade(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}

	response.WriteJson(w, res, nil)
}

// POST /grades/detail
func (h *GradeHandler) HandleGetGrade(w http.ResponseWriter, r *http.Request) {
	var req dto.GetGradeReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}

	res, err := h.gradeSvc.GetGrade(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}

	response.WriteJson(w, res, nil)
}
