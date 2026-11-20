package job

import "math-ai.com/math-ai/internal/domain/shared/status"

// Status codes of this aggregate (12000–12099). They are wire contract: the
// client reads them from "mstatus", so never renumber or reuse one.
// Moved from domain/shared/status (docs/features/014-status-registry).
const (
	JobNotFound             status.StatusCode = 12001
	JobAlreadyPaused        status.StatusCode = 12002
	JobNotPaused            status.StatusCode = 12003
	JobInFlight             status.StatusCode = 12004
	JobRuntimeUnavailable   status.StatusCode = 12005
	JobMissingName          status.StatusCode = 12006
	JobTriggerFailed        status.StatusCode = 12007
	TaskHandlerNotFound     status.StatusCode = 12008
	TaskQueueFull           status.StatusCode = 12009
	TaskMissingName         status.StatusCode = 12010
	TaskEnqueueFailed       status.StatusCode = 12011
	JobInvalidSchedule      status.StatusCode = 12012
	JobUpdateScheduleFailed status.StatusCode = 12013
)

func init() {
	status.Register(JobNotFound, "Job not found", "Không tìm thấy job")
	status.Register(JobAlreadyPaused, "Job is already paused", "Job đã ở trạng thái tạm dừng")
	status.Register(JobNotPaused, "Job is not paused", "Job không ở trạng thái tạm dừng")
	status.Register(JobInFlight, "Job is already running an execution", "Job đang chạy một lần thực thi khác")
	status.Register(JobRuntimeUnavailable, "Job runtime is not started or is shutting down", "Hệ thống job chưa khởi động hoặc đang dừng")
	status.Register(JobMissingName, "Job name is required", "Tên job là bắt buộc")
	status.Register(JobTriggerFailed, "Failed to trigger job", "Kích hoạt job thất bại")
	status.Register(TaskHandlerNotFound, "Task handler is not registered", "Không tìm thấy task handler được đăng ký")
	status.Register(TaskQueueFull, "Task queue is full; please retry shortly", "Hàng đợi task đã đầy; vui lòng thử lại sau")
	status.Register(TaskMissingName, "Task name is required", "Tên task là bắt buộc")
	status.Register(TaskEnqueueFailed, "Failed to enqueue task", "Đưa task vào hàng đợi thất bại")
	status.Register(JobInvalidSchedule, "Invalid job schedule", "Lịch chạy job không hợp lệ")
	status.Register(JobUpdateScheduleFailed, "Failed to update job schedule", "Cập nhật lịch chạy job thất bại")
}
