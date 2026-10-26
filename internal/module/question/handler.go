package question

import (
	"encoding/json"
	"errors"
	"net/http"

	dto "math-ai.com/math-ai/internal/application/dto/question"
	"math-ai.com/math-ai/internal/shared/response"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// POST /questions/create — admin only; stores a batch of questions, all or none.
func (h *Handler) HandleCreateQuestions(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateQuestionsReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	res, err := h.service.CreateQuestions(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	response.WriteJson(w, res, nil)
}

// POST /questions/list — admin only; pages the bank, answer key included.
func (h *Handler) HandleListQuestions(w http.ResponseWriter, r *http.Request) {
	var req dto.ListQuestionsReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	res, err := h.service.ListQuestions(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	response.WriteJson(w, res, nil)
}

// POST /questions/update — admin only; edits one question (absent fields kept).
func (h *Handler) HandleUpdateQuestion(w http.ResponseWriter, r *http.Request) {
	var req dto.UpdateQuestionReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	res, err := h.service.UpdateQuestion(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	response.WriteJson(w, res, nil)
}

// POST /questions/mark-status — admin only; sets one question's review state.
func (h *Handler) HandleMarkQuestionStatus(w http.ResponseWriter, r *http.Request) {
	var req dto.MarkQuestionStatusReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	res, err := h.service.MarkQuestionStatus(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	response.WriteJson(w, res, nil)
}

// POST /questions/verify — admin only; verified_count + 1.
func (h *Handler) HandleVerifyQuestion(w http.ResponseWriter, r *http.Request) {
	var req dto.VerifyQuestionReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	res, err := h.service.VerifyQuestion(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	response.WriteJson(w, res, nil)
}

// maxUploadBody caps the whole multipart body: the image plus room for the
// form framing. Beyond it the read fails instead of buffering more.
const maxUploadBody = MaxImageBytes + 1<<20

// POST /questions/images/upload — admin only; multipart/form-data with one
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
