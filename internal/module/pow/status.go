package pow

import "math-ai.com/math-ai/internal/domain/shared/status"

// Status codes of this aggregate (13800–13899). They are wire contract: the
// client reads them from "mstatus", so never renumber or reuse one.
// Moved from domain/shared/status (docs/features/014-status-registry).
const (
	// Proof of work — 13800-13899. Gates /exams/generate against bots.
	PowChallengeFailed   status.StatusCode = 13800 // could not mint a challenge
	PowChallengeNotFound status.StatusCode = 13801 // verify without a pending challenge (never issued, or already used)
	PowChallengeExpired  status.StatusCode = 13802
	PowInvalidProof      status.StatusCode = 13803
	PowRequired          status.StatusCode = 13804 // a gated route was called without a fresh, unused proof
	PowInvalidRequest    status.StatusCode = 13805
)

func init() {
	status.Register(PowChallengeFailed, "Could not create a verification challenge. Please try again", "Không tạo được thử thách xác minh. Vui lòng thử lại")
	status.Register(PowChallengeNotFound, "No verification challenge in progress. Please request a new one", "Không có thử thách xác minh nào. Vui lòng yêu cầu thử thách mới")
	status.Register(PowChallengeExpired, "The verification challenge has expired. Please request a new one", "Thử thách xác minh đã hết hạn. Vui lòng yêu cầu thử thách mới")
	status.Register(PowInvalidProof, "Verification failed. Please request a new challenge", "Xác minh không thành công. Vui lòng yêu cầu thử thách mới")
	status.Register(PowRequired, "Please complete verification before continuing", "Vui lòng hoàn tất xác minh trước khi tiếp tục")
	status.Register(PowInvalidRequest, "Invalid verification request", "Yêu cầu xác minh không hợp lệ")
}
