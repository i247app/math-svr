package banner

import "math-ai.com/math-ai/internal/domain/shared/status"

// Status codes of this aggregate (13400–13499). They are wire contract: the
// client reads them from "mstatus", so never renumber or reuse one.
// Moved from domain/shared/status (docs/features/014-status-registry).
const (
	// Banner — 13400-13499
	BannerNotFound          status.StatusCode = 13400
	BannerMissingID         status.StatusCode = 13401
	BannerMissingMediaURL   status.StatusCode = 13402
	BannerInvalidMediaType  status.StatusCode = 13403
	BannerInvalidStatus     status.StatusCode = 13404
	BannerTitleTooLong      status.StatusCode = 13405
	BannerShortTextTooLong  status.StatusCode = 13406
	BannerMediaURLTooLong   status.StatusCode = 13407
	BannerButtonTextTooLong status.StatusCode = 13408
	BannerButtonLinkTooLong status.StatusCode = 13409
	BannerNoteTooLong       status.StatusCode = 13410
	// Banner media upload
	BannerMediaInvalidReference status.StatusCode = 13411
	BannerMediaInvalidFile      status.StatusCode = 13412
	BannerMediaUploadFailed     status.StatusCode = 13413
)

func init() {
	status.Register(BannerNotFound, "Banner not found", "Không tìm thấy banner")
	status.Register(BannerMissingID, "banner_id is required", "Thiếu banner_id")
	status.Register(BannerMissingMediaURL, "media_url_key is required for image/video banners", "media_url_key là bắt buộc đối với banner ảnh/video")
	status.Register(BannerInvalidMediaType, "Invalid media type", "Loại media không hợp lệ")
	status.Register(BannerInvalidStatus, "Invalid banner status", "Trạng thái banner không hợp lệ")
	status.Register(BannerTitleTooLong, "Title is too long", "Tiêu đề quá dài")
	status.Register(BannerShortTextTooLong, "Short text is too long", "Đoạn mô tả ngắn quá dài")
	status.Register(BannerMediaURLTooLong, "media_url_key is too long", "media_url_key quá dài")
	status.Register(BannerButtonTextTooLong, "Button text is too long", "Nội dung nút quá dài")
	status.Register(BannerButtonLinkTooLong, "Button link URL is too long", "Đường dẫn nút quá dài")
	status.Register(BannerNoteTooLong, "Note is too long", "Ghi chú quá dài")
	status.Register(BannerMediaInvalidReference, "Invalid banner media reference", "Tham chiếu media của banner không hợp lệ")
	status.Register(BannerMediaInvalidFile, "Invalid banner media file", "Tệp media của banner không hợp lệ")
	status.Register(BannerMediaUploadFailed, "Failed to upload banner media", "Tải media của banner lên thất bại")
}
