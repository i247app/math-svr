package questionpool

import (
	"encoding/json"
	"errors"
	"net/http"

	dto "math-ai.com/math-ai/internal/application/dto/questionpool"
	"math-ai.com/math-ai/internal/shared/response"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// POST /question-pools/create — admin only; stores a batch of questions, all or none.
func (h *Handler) HandleCreateQuestionPools(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateQuestionPoolsReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	res, err := h.service.CreateQuestionPools(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	response.WriteJson(w, res, nil)
}

// POST /question-pools/list — admin only; pages the bank, answer key included.
func (h *Handler) HandleListQuestionPools(w http.ResponseWriter, r *http.Request) {
	var req dto.ListQuestionPoolsReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	res, err := h.service.ListQuestionPools(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	response.WriteJson(w, res, nil)
}

// POST /question-pools/update — admin only; edits one question (absent fields kept).
func (h *Handler) HandleUpdateQuestionPool(w http.ResponseWriter, r *http.Request) {
	var req dto.UpdateQuestionPoolReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	res, err := h.service.UpdateQuestionPool(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	response.WriteJson(w, res, nil)
}

// POST /question-pools/mark-status — admin only; sets one question's review state.
func (h *Handler) HandleMarkQuestionPoolStatus(w http.ResponseWriter, r *http.Request) {
	var req dto.MarkQuestionPoolStatusReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	res, err := h.service.MarkQuestionPoolStatus(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	response.WriteJson(w, res, nil)
}

// POST /question-pools/verify — admin only; verified_count + 1.
func (h *Handler) HandleVerifyQuestionPool(w http.ResponseWriter, r *http.Request) {
	var req dto.VerifyQuestionPoolReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	res, err := h.service.VerifyQuestionPool(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	response.WriteJson(w, res, nil)
}

// POST /question-pools/soft-delete — admin only; hides one question.
func (h *Handler) HandleSoftDeleteQuestionPool(w http.ResponseWriter, r *http.Request) {
	var req dto.DeleteQuestionPoolReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	res, err := h.service.SoftDeleteQuestionPool(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	response.WriteJson(w, res, nil)
}

// POST /question-pools/force-delete — admin only; removes one question's row.
func (h *Handler) HandleForceDeleteQuestionPool(w http.ResponseWriter, r *http.Request) {
	var req dto.DeleteQuestionPoolReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	res, err := h.service.ForceDeleteQuestionPool(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	response.WriteJson(w, res, nil)
}

// maxUploadBody caps the whole multipart body: the image plus room for the
// form framing. Beyond it the read fails instead of buffering more.
const maxUploadBody = MaxImageBytes + 1<<20

// POST /question-pools/images/upload — admin only; multipart/form-data with one
// file part "image". Returns {image_key, image_url}.
func (h *Handler) HandleUploadImage(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadBody)
	if err := r.ParseMultipartForm(maxUploadBody); err != nil {
		var tooBig *http.MaxBytesError
		if errors.As(err, &tooBig) {
			err = errors.Join(ErrImageTooLarge, err)
		}
		response.WriteJson(w, nil, h.service.imageFormError(r.Context(), err))
		return
	}
	defer r.MultipartForm.RemoveAll()

	var file ImageFile
	part, header, err := r.FormFile("image")
	if err == nil {
		defer part.Close()
		file = ImageFile{Reader: part, Filename: header.Filename, ContentType: header.Header.Get("Content-Type"), Size: header.Size}
	}
	res, err := h.service.UploadImage(r.Context(), file)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	response.WriteJson(w, res, nil)
}
