package chapter

import (
	"encoding/json"
	"net/http"

	dto "math-ai.com/math-ai/internal/application/dto/chapter"
	"math-ai.com/math-ai/internal/application/resource"
	"math-ai.com/math-ai/internal/shared/enum"
	"math-ai.com/math-ai/internal/shared/response"
	"math-ai.com/math-ai/internal/shared/utils"
)

type ChapterHandler struct {
	appResource *resource.Resource
	chapterSvc  *Service
}

func NewChapterHandler(appResource *resource.Resource, chapterSvc *Service) *ChapterHandler {
	return &ChapterHandler{
		appResource: appResource,
		chapterSvc:  chapterSvc,
	}
}

// POST /chapters/create
func (h *ChapterHandler) HandleCreateChapter(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateChapterReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}

	res, err := h.chapterSvc.CreateChapter(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	response.WriteJson(w, res, nil)
}

// POST /chapters/update
func (h *ChapterHandler) HandleUpdateChapter(w http.ResponseWriter, r *http.Request) {
	var req dto.UpdateChapterReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}

	res, err := h.chapterSvc.UpdateChapter(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	response.WriteJson(w, res, nil)
}

// POST /chapters/delete
func (h *ChapterHandler) HandleSoftDeleteChapter(w http.ResponseWriter, r *http.Request) {
	var req dto.DeleteChapterReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}

	res, err := h.chapterSvc.SoftDeleteChapter(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	response.WriteJson(w, res, nil)
}

// POST /chapters/force-delete
func (h *ChapterHandler) HandleForceDeleteChapter(w http.ResponseWriter, r *http.Request) {
	var req dto.DeleteChapterReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}

	res, err := h.chapterSvc.ForceDeleteChapter(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	response.WriteJson(w, res, nil)
}

// GET /chapters/{id}
//
// Reads the optional `language` query parameter so callers can request a
// specific locale without sending a JSON body. Falls back to the project
// default (vn) when omitted.
func (h *ChapterHandler) HandleGetChapter(w http.ResponseWriter, r *http.Request) {
	req := dto.GetChapterReq{
		ChapterID: utils.StringToInt64(r.PathValue("id"), 0),
		Language:  enum.LanguageType(r.URL.Query().Get("language")),
	}

	res, err := h.chapterSvc.GetChapter(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	response.WriteJson(w, res, nil)
}

// POST /chapters/list
func (h *ChapterHandler) HandleListChapters(w http.ResponseWriter, r *http.Request) {
	var req dto.ListChaptersReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}

	res, err := h.chapterSvc.ListChapters(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	response.WriteJson(w, res, nil)
}
