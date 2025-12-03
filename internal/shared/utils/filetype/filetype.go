package filetype

import (
	"fmt"
	"mime"
	"path/filepath"
	"strings"
)

// FileType represents a category of files
type FileType string

const (
	FileTypeImage    FileType = "image"
	FileTypeVideo    FileType = "video"
	FileTypeAudio    FileType = "audio"
	FileTypeDocument FileType = "document"
	FileTypeOther    FileType = "other"
)

// AllowedMimeTypes defines the allowed MIME types for file uploads
var AllowedMimeTypes = map[string]FileType{
	// Images
	"image/jpeg":    FileTypeImage,
	"image/jpg":     FileTypeImage,
	"image/png":     FileTypeImage,
	"image/gif":     FileTypeImage,
	"image/webp":    FileTypeImage,
	"image/svg+xml": FileTypeImage,
	"image/bmp":     FileTypeImage,
	"image/tiff":    FileTypeImage,

	// Videos
	"video/mp4":        FileTypeVideo,
	"video/mpeg":       FileTypeVideo,
	"video/quicktime":  FileTypeVideo,
	"video/x-msvideo":  FileTypeVideo,
	"video/x-matroska": FileTypeVideo,
	"video/webm":       FileTypeVideo,

	// Audio
	"audio/mpeg": FileTypeAudio,
	"audio/mp3":  FileTypeAudio,
	"audio/wav":  FileTypeAudio,
	"audio/ogg":  FileTypeAudio,
	"audio/webm": FileTypeAudio,
	"audio/aac":  FileTypeAudio,

	// Documents
	"application/pdf":    FileTypeDocument,
	"application/msword": FileTypeDocument,
	"application/vnd.openxmlformats-officedocument.wordprocessingml.document": FileTypeDocument,
	"application/vnd.ms-excel": FileTypeDocument,
	"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet": FileTypeDocument,
	"text/plain": FileTypeDocument,
	"text/csv":   FileTypeDocument,
}

// AllowedExtensions defines the allowed file extensions
var AllowedExtensions = map[string]FileType{
	// Images
	".jpg":  FileTypeImage,
	".jpeg": FileTypeImage,
	".png":  FileTypeImage,
	".gif":  FileTypeImage,
	".webp": FileTypeImage,
	".svg":  FileTypeImage,
	".bmp":  FileTypeImage,
	".tiff": FileTypeImage,
	".tif":  FileTypeImage,

	// Videos
	".mp4":  FileTypeVideo,
	".mpeg": FileTypeVideo,
	".mpg":  FileTypeVideo,
	".mov":  FileTypeVideo,
	".avi":  FileTypeVideo,
	".mkv":  FileTypeVideo,
	".webm": FileTypeVideo,

	// Audio
	".mp3": FileTypeAudio,
	".wav": FileTypeAudio,
	".ogg": FileTypeAudio,
	".aac": FileTypeAudio,

	// Documents
	".pdf":  FileTypeDocument,
	".doc":  FileTypeDocument,
	".docx": FileTypeDocument,
	".xls":  FileTypeDocument,
	".xlsx": FileTypeDocument,
	".txt":  FileTypeDocument,
	".csv":  FileTypeDocument,
}

// ValidateFile checks if a file is allowed based on filename and content type
func ValidateFile(filename string, contentType string) error {
	// Normalize content type
	contentType = strings.ToLower(strings.TrimSpace(contentType))

	// Check content type if provided
	if contentType != "" {
		if _, allowed := AllowedMimeTypes[contentType]; !allowed {
			return fmt.Errorf("file type '%s' is not allowed", contentType)
		}
	}

	// Check file extension
	ext := strings.ToLower(filepath.Ext(filename))
	if ext == "" {
		return fmt.Errorf("file must have an extension")
	}

	if _, allowed := AllowedExtensions[ext]; !allowed {
		return fmt.Errorf("file extension '%s' is not allowed", ext)
	}

	return nil
}

// GetFileType returns the file type based on filename and content type
func GetFileType(filename string, contentType string) FileType {
	// Try content type first
	if contentType != "" {
		if fileType, ok := AllowedMimeTypes[strings.ToLower(contentType)]; ok {
			return fileType
		}
	}

	// Fall back to extension
	ext := strings.ToLower(filepath.Ext(filename))
	if fileType, ok := AllowedExtensions[ext]; ok {
		return fileType
	}

	return FileTypeOther
}

// GetContentTypeFromFilename returns the MIME type based on filename
func GetContentTypeFromFilename(filename string) string {
	ext := filepath.Ext(filename)
	contentType := mime.TypeByExtension(ext)

	if contentType == "" {
		return "application/octet-stream"
	}

	return contentType
}

// IsImage checks if the file is an image
func IsImage(filename string, contentType string) bool {
	return GetFileType(filename, contentType) == FileTypeImage
}

// IsVideo checks if the file is a video
func IsVideo(filename string, contentType string) bool {
	return GetFileType(filename, contentType) == FileTypeVideo
}

// IsAudio checks if the file is audio
func IsAudio(filename string, contentType string) bool {
	return GetFileType(filename, contentType) == FileTypeAudio
}

// IsDocument checks if the file is a document
func IsDocument(filename string, contentType string) bool {
	return GetFileType(filename, contentType) == FileTypeDocument
}
