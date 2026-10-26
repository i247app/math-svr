package question

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path"
	"strings"

	"math-ai.com/math-ai/internal/adapter/storage"
	dto "math-ai.com/math-ai/internal/application/dto/question"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/infrastructure/logger"
	sctx "math-ai.com/math-ai/internal/shared/context"
)

// MaxImageBytes caps one question image. A child's screen needs far less;
// the cap keeps an admin from storing a camera original.
const MaxImageBytes = 5 << 20

// imageFolder is the S3 folder; storage makes keys "question-images/…".
const imageFolder = "question-images"

// imageTypes are the accepted images, by extension → content type. SVG is
// deliberately absent: it can carry script and is served to browsers.
var imageTypes = map[string]string{
	".jpg":  "image/jpeg",
	".jpeg": "image/jpeg",
	".png":  "image/png",
	".gif":  "image/gif",
	".webp": "image/webp",
}

var (
	ErrImageMissing  = errors.New("image file part is required")
	ErrImageTooLarge = errors.New("request body is over the image size limit")
	ErrStorageOff    = errors.New("question: storage adapter is not configured")
)

// imageFormError maps a body that is not a usable multipart form. The admin
// gate still comes first, so an unauthorised caller learns nothing.
func (s *Service) imageFormError(ctx context.Context, err error) error {
	if gateErr := requireAdminGate(ctx); gateErr != nil {
		return gateErr
	}
	if errors.Is(err, ErrImageTooLarge) {
		return errs.NewError(ctx, status.QUESTION_IMAGE_INVALID_FILE, nil, err)
	}
	return errs.NewError(ctx, status.QUESTION_IMAGE_MISSING, nil, fmt.Errorf("body is not multipart/form-data with an image part: %w", err))
}

// ImageFile is the multipart "image" part. Size is the part's size as the
// client declared it; the read itself is capped at MaxImageBytes as well.
type ImageFile struct {
	Reader      io.Reader
	Filename    string
	ContentType string
	Size        int64
}

// UploadImage stores one image under question-images/ and returns its key
// (what create takes in question_image_key / options[].image_key) and a
// presigned URL to preview it. The type is checked three ways — extension,
// declared content type, and the bytes themselves — so a renamed file is
// refused.
func (s *Service) UploadImage(ctx context.Context, file ImageFile) (*dto.UploadImageRes, error) {
	if err := requireAdminGate(ctx); err != nil {
		return nil, err
	}
	if file.Reader == nil {
		return nil, errs.NewError(ctx, status.QUESTION_IMAGE_MISSING, nil, ErrImageMissing)
	}
	if s.storage == nil {
		return nil, errs.NewError(ctx, status.STORAGE_CONFIG_INVALID, nil, ErrStorageOff)
	}
	data, contentType, err := readImage(file)
	if err != nil {
		return nil, errs.NewError(ctx, status.QUESTION_IMAGE_INVALID_FILE, nil, err)
	}

	uploaded, err := s.storage.HandleUpload(ctx, &storage.UploadFileRequest{
		File:        bytes.NewReader(data),
		Filename:    file.Filename,
		ContentType: contentType,
		Folder:      imageFolder,
	})
	if err != nil {
		return nil, errs.NewError(ctx, status.QUESTION_IMAGE_UPLOAD_FAILED, nil, err)
	}
	if uploaded == nil || !strings.HasPrefix(uploaded.Key, ImageKeyPrefix) {
		return nil, errs.NewError(ctx, status.QUESTION_IMAGE_UPLOAD_FAILED, nil,
			fmt.Errorf("question: storage returned key %q, want one under %s", keyOf(uploaded), ImageKeyPrefix))
	}
	logger.From(ctx).Infof("question.image.uploaded admin_uid=%d key=%s bytes=%d", sctx.UID(ctx), uploaded.Key, len(data))
	return &dto.UploadImageRes{ImageKey: uploaded.Key, ImageUrl: s.presign(ctx, 0, &uploaded.Key)}, nil
}

// readImage reads at most MaxImageBytes+1 bytes and returns them with the
// content type they really are, or why the file is refused.
func readImage(file ImageFile) ([]byte, string, error) {
	want, ok := imageTypes[strings.ToLower(path.Ext(file.Filename))]
	if !ok {
		return nil, "", fmt.Errorf("extension of %q is not one of .jpg .jpeg .png .gif .webp", file.Filename)
	}
	declared := strings.ToLower(strings.TrimSpace(file.ContentType))
	if declared != "" && declared != "application/octet-stream" && declared != want && !(declared == "image/jpg" && want == "image/jpeg") {
		return nil, "", fmt.Errorf("declared type %q does not match the extension", file.ContentType)
	}
	if file.Size > MaxImageBytes {
		return nil, "", fmt.Errorf("image is %d bytes, at most %d", file.Size, MaxImageBytes)
	}
	data, err := io.ReadAll(io.LimitReader(file.Reader, MaxImageBytes+1))
	if err != nil {
		return nil, "", fmt.Errorf("read image: %w", err)
	}
	if len(data) > MaxImageBytes {
		return nil, "", fmt.Errorf("image is over %d bytes", MaxImageBytes)
	}
	if len(data) == 0 {
		return nil, "", errors.New("image is empty")
	}
	// The bytes decide, not the name: http.DetectContentType sniffs the
	// magic number (WebP included).
	if sniffed := http.DetectContentType(data); sniffed != want {
		return nil, "", fmt.Errorf("content is %s, not %s", sniffed, want)
	}
	return data, want, nil
}

func keyOf(r *storage.UploadFileResponse) string {
	if r == nil {
		return ""
	}
	return r.Key
}
