package controller

import (
	"encoding/json"
	"fmt"
	"net/http"

	"math-ai.com/math-ai/internal/app/resources"
	"math-ai.com/math-ai/internal/applications/dto"
	di "math-ai.com/math-ai/internal/core/di/services"
	"math-ai.com/math-ai/internal/shared/constant/status"
	"math-ai.com/math-ai/internal/shared/utils/response"
)

const (
	// MaxUploadSize defines the maximum file size (100MB)
	MaxUploadSize = 100 << 20 // 100 MB
)

type StorageController struct {
	appResources *resources.AppResource
	service      di.IStorageService
}

func NewStorageController(appResources *resources.AppResource, service di.IStorageService) *StorageController {
	return &StorageController{
		appResources: appResources,
		service:      service,
	}
}

// HandleUpload handles file upload
// POST /storage/upload
func (s *StorageController) HandleUpload(w http.ResponseWriter, r *http.Request) {
	// Parse multipart form with max size
	if err := r.ParseMultipartForm(MaxUploadSize); err != nil {
		////logger.Errorf("Failed to parse multipart form: %v", err)
		response.WriteJson(w, r.Context(), nil, fmt.Errorf("file too large or invalid form data"), status.FAIL)
		return
	}

	// Get file from form
	file, header, err := r.FormFile("file")
	if err != nil {
		////logger.Errorf("Failed to get file from form: %v", err)
		response.WriteJson(w, r.Context(), nil, fmt.Errorf("no file provided"), status.FAIL)
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
	statusCode, uploadRes, err := s.service.HandleUpload(r.Context(), uploadReq)
	if err != nil {
		response.WriteJson(w, r.Context(), nil, err, statusCode)
		return
	}

	response.WriteJson(w, r.Context(), uploadRes, nil, statusCode)
}

// HandleDelete handles file deletion
// POST /storage/delete
func (s *StorageController) HandleDelete(w http.ResponseWriter, r *http.Request) {
	var req dto.DeleteFileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, r.Context(), nil, fmt.Errorf("invalid request body"), status.FAIL)
		return
	}

	if req.Key == "" {
		response.WriteJson(w, r.Context(), nil, fmt.Errorf("key is required"), status.FAIL)
		return
	}

	statusCode, err := s.service.HandleDelete(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, r.Context(), nil, err, statusCode)
		return
	}

	response.WriteJson(w, r.Context(), map[string]string{"message": "File deleted successfully"}, nil, statusCode)
}

// HandleGetPreviewURL handles getting preview URL from original URL
// POST /storage/preview-url
func (s *StorageController) HandleGetPreviewURL(w http.ResponseWriter, r *http.Request) {
	var req dto.GetPreviewURLRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, r.Context(), nil, fmt.Errorf("invalid request body"), status.FAIL)
		return
	}

	if req.URL == "" {
		response.WriteJson(w, r.Context(), nil, fmt.Errorf("url is required"), status.FAIL)
		return
	}

	statusCode, previewURL, err := s.service.GetPreviewURL(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, r.Context(), nil, err, statusCode)
		return
	}

	res := &dto.GetPreviewURLResponse{
		PreviewURL: previewURL,
	}

	response.WriteJson(w, r.Context(), res, nil, status.OK)
}
