package questionpool

import (
	"context"
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

// maxUploadBody caps the whole multipart body of an image upload: the
// image plus room for the form framing. Beyond it the read fails instead of
// buffering more.
const maxUploadBody = MaxImageBytes + 1<<20

// maxImportBody is the same for validate-file / import-file.
const maxImportBody = MaxImportFileBytes + 1<<20

// parseMultipart caps the body at maxBody and parses the form; tooLarge
// joins the error when the cap was hit. The caller removes the form's temp
// files (r.MultipartForm.RemoveAll) on success.
func parseMultipart(w http.ResponseWriter, r *http.Request, maxBody int64, tooLarge error) error {
	r.Body = http.MaxBytesReader(w, r.Body, maxBody)
	if err := r.ParseMultipartForm(maxBody); err != nil {
		var tooBig *http.MaxBytesError
		if errors.As(err, &tooBig) {
			err = errors.Join(tooLarge, err)
		}
		return err
	}
	return nil
}

// POST /question-pools/images/upload — admin only; multipart/form-data with one
// file part "image". Returns {image_key, image_url}.
func (h *Handler) HandleUploadImage(w http.ResponseWriter, r *http.Request) {
	if err := parseMultipart(w, r, maxUploadBody, ErrImageTooLarge); err != nil {
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

// POST /question-pools/validate-file — admin only; multipart/form-data with
// one file part "file" (a .json array of questions). Writes nothing;
// returns {report}.
func (h *Handler) HandleValidateFile(w http.ResponseWriter, r *http.Request) {
	h.handleImport(w, r, h.service.ValidateFile)
}

// POST /question-pools/import-file — admin only; same body as validate-file.
// Stores every valid, new question of the file, or none; returns {report}.
func (h *Handler) HandleImportFile(w http.ResponseWriter, r *http.Request) {
	h.handleImport(w, r, h.service.ImportFile)
}

func (h *Handler) handleImport(w http.ResponseWriter, r *http.Request,
	run func(context.Context, ImportFile) (*dto.ImportQuestionPoolsRes, error)) {
	if err := parseMultipart(w, r, maxImportBody, ErrImportFileTooLarge); err != nil {
		response.WriteJson(w, nil, h.service.importFormError(r.Context(), err))
		return
	}
	defer r.MultipartForm.RemoveAll()

	var file ImportFile
	part, header, err := r.FormFile("file")
	if err == nil {
		defer part.Close()
		file = ImportFile{Reader: part, Filename: header.Filename, Size: header.Size}
	}
	res, err := run(r.Context(), file)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	response.WriteJson(w, res, nil)
}
