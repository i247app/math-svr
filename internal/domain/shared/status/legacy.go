package status

// Legacy status codes (6000–6099, 7000–7099). They are wire contract: the
// client reads them from "mstatus", so never renumber or reuse one.
// Never returned by any route today; kept registered (014 decision 5).
const (
	MessagingPublishFailed   StatusCode = 6001
	MessagingInvalidEnvelope StatusCode = 6002
	MessagingConfigInvalid   StatusCode = 6003
	MessagingConnectFailed   StatusCode = 6004
	CacheConnectFailed       StatusCode = 7001
	CacheConfigInvalid       StatusCode = 7002
	CacheOpFailed            StatusCode = 7003
	CacheSerializeFailed     StatusCode = 7004
)

func init() {
	Register(MessagingPublishFailed, "Failed to publish event", "Gửi sự kiện thất bại")
	Register(MessagingInvalidEnvelope, "Event envelope is invalid", "Gói sự kiện không hợp lệ")
	Register(MessagingConfigInvalid, "Messaging configuration is invalid", "Cấu hình messaging không hợp lệ")
	Register(MessagingConnectFailed, "Failed to connect to messaging broker", "Kết nối broker thất bại")
	Register(CacheConnectFailed, "Failed to connect to cache", "Kết nối cache thất bại")
	Register(CacheConfigInvalid, "Cache configuration is invalid", "Cấu hình cache không hợp lệ")
	Register(CacheOpFailed, "Cache operation failed", "Thao tác cache thất bại")
	Register(CacheSerializeFailed, "Failed to serialize cache value", "Không thể mã hoá giá trị cache")
}
