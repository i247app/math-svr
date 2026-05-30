package profile

import (
	"encoding/json"
	"fmt"
	"net/http"

	dto "math-ai.com/math-ai/internal/application/dto/profile"
	"math-ai.com/math-ai/internal/shared/response"
)

// POST /profile/upload-static-file
func (h *ProfileHandler) HandleUploadStaticFile(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(MaxAvatarUploadSize); err != nil {
		////logger.Errorf("Failed to parse multipart form: %v", err)
		response.WriteJson(w, nil, fmt.Errorf("file too large or invalid form data"))
		return
	}

	// Get file from form
	file, header, err := r.FormFile("file")
	if err != nil {
		////logger.Errorf("Failed to get file from form: %v", err)
		response.WriteJson(w, nil, fmt.Errorf("no file provided"))
		return
	}
	defer file.Close()

	// Get optional folder parameter
	folder := r.FormValue("folder")

	// Get content type from header
	contentType := header.Header.Get("Content-Type")

	// Create upload request
	uploadReq := &dto.UploadFileRequest{
		File:        file,
		Filename:    header.Filename,
		ContentType: contentType,
		Folder:      folder,
	}

	// Upload file
	uploadRes, err := h.profileSvc.UploadAvatarStatic(r.Context(), uploadReq)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}

	response.WriteJson(w, uploadRes, nil)
}

// POST /profile/delete-static-file
func (h *ProfileHandler) HandleDeleteStaticFile(w http.ResponseWriter, r *http.Request) {
	var req dto.DeleteFileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, fmt.Errorf("invalid request body"))
		return
	}

	if req.Key == "" {
		response.WriteJson(w, nil, fmt.Errorf("key is required"))
		return
	}

	res, err := h.profileSvc.DeleteFile(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}

	response.WriteJson(w, res, nil)
}
