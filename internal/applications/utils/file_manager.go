package utils

import (
	"context"
	"fmt"
	"io"

	"math-ai.com/math-ai/internal/applications/dto"
	di "math-ai.com/math-ai/internal/core/di/services"
	"math-ai.com/math-ai/internal/shared/constant/status"
)

// FileManager provides generic file management operations for any module
// Handles upload, update, and deletion of files (avatars, icons, documents, etc.)
type FileManager struct {
	storageService di.IStorageService
}

// NewFileManager creates a new FileManager instance
func NewFileManager(storageService di.IStorageService) *FileManager {
	return &FileManager{
		storageService: storageService,
	}
}

// UploadFile uploads a file to storage and returns the key
// folder parameter specifies the S3 folder (e.g., "user", "grade", "profile")
func (f *FileManager) UploadFile(ctx context.Context, file io.Reader, filename, contentType, folder string) (*string, status.Code, error) {
	if file == nil {
		return nil, status.SUCCESS, nil
	}

	statusCode, res, err := f.storageService.HandleUpload(ctx, &dto.UploadFileRequest{
		File:        file,
		Filename:    filename,
		ContentType: contentType,
		Folder:      folder,
	})
	if err != nil {
		return nil, statusCode, fmt.Errorf("failed to upload file to %s: %w", folder, err)
	}

	return &res.Key, status.SUCCESS, nil
}

// DeleteFile removes a file from storage
// Safe to call with nil or empty key - will be no-op
func (f *FileManager) DeleteFile(ctx context.Context, fileKey *string) {
	if fileKey == nil || *fileKey == "" {
		return
	}

	_, err := f.storageService.HandleDelete(ctx, &dto.DeleteFileRequest{
		Key: *fileKey,
	})

	// Don't return error - file cleanup is not critical
	// Errors are already logged in storage service
	_ = err
}

// UpdateFile handles file upload with automatic cleanup of old file
// Returns the new file key to set on the entity
// - If newFile is provided: uploads new file and deletes old file
// - If deleteFile is true: deletes the current file and returns empty string
// - Otherwise: returns nil (no change)
func (f *FileManager) UpdateFile(
	ctx context.Context,
	currentFileKey *string,
	newFile io.Reader,
	newFilename string,
	newContentType string,
	deleteFile bool,
	folder string,
) (*string, status.Code, error) {
	// Handle new file upload
	if newFile != nil {
		// Upload new file
		newKey, statusCode, err := f.UploadFile(ctx, newFile, newFilename, newContentType, folder)
		if err != nil {
			return nil, statusCode, err
		}

		// Delete old file if exists
		f.DeleteFile(ctx, currentFileKey)

		return newKey, status.SUCCESS, nil
	}

	// Handle file deletion
	if deleteFile {
		f.DeleteFile(ctx, currentFileKey)
		emptyString := ""
		return &emptyString, status.SUCCESS, nil
	}

	// No change
	return nil, status.SUCCESS, nil
}

// UploadWithCleanup uploads a new file and deletes the old one atomically
// Useful when you need to replace an existing file
func (f *FileManager) UploadWithCleanup(
	ctx context.Context,
	oldFileKey *string,
	newFile io.Reader,
	newFilename string,
	newContentType string,
	folder string,
) (*string, status.Code, error) {
	// Upload new file first
	newKey, statusCode, err := f.UploadFile(ctx, newFile, newFilename, newContentType, folder)
	if err != nil {
		return nil, statusCode, err
	}

	// Delete old file after successful upload
	f.DeleteFile(ctx, oldFileKey)

	return newKey, status.SUCCESS, nil
}
