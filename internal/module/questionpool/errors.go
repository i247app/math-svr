package questionpool

import "errors"

var (
	ErrImageMissing  = errors.New("image file part is required")
	ErrImageTooLarge = errors.New("request body is over the image size limit")
	ErrStorageOff    = errors.New("question_pool: storage provider is not configured")

	ErrImportFileMissing  = errors.New("file part is required")
	ErrImportFileTooLarge = errors.New("request body is over the import file size limit")

	ErrAdminGateMissing = errors.New("question_pool: route did not pass the admin gate")
)
