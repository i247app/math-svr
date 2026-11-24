package game

import "math-ai.com/math-ai/internal/domain/shared/status"

// Status codes of this aggregate (14700–14799). They are wire contract: the
// client reads them from "mstatus", so never renumber or reuse one.
const (
	GameSessionNotFound          status.StatusCode = 14700
	GameSessionNotOwned          status.StatusCode = 14701
	GameMissingProfileID         status.StatusCode = 14702
	GameProfileNotFound          status.StatusCode = 14703
	GameProfileNotOwned          status.StatusCode = 14704
	GameMissingSessionID         status.StatusCode = 14705
	GameInvalidGrade             status.StatusCode = 14706
	GameInvalidNumQuestions      status.StatusCode = 14707
	GameGuestNotAllowed          status.StatusCode = 14708
	GameQuestionBankInsufficient status.StatusCode = 14709
	GameInvalidAnswers           status.StatusCode = 14710
	GameAlreadySubmitted         status.StatusCode = 14711
	GameGenerationFailed         status.StatusCode = 14712
	GameGradingFailed            status.StatusCode = 14713
	GameInvalidLevel             status.StatusCode = 14714
)

func init() {
	status.Register(GameSessionNotFound, "Game not found", "Không tìm thấy ván chơi")
	status.Register(GameSessionNotOwned, "This game belongs to another profile", "Ván chơi này thuộc về hồ sơ khác")
	status.Register(GameMissingProfileID, "profile_id is required", "Thiếu profile_id")
	status.Register(GameProfileNotFound, "Profile not found", "Không tìm thấy hồ sơ")
	status.Register(GameProfileNotOwned, "This profile belongs to another user", "Hồ sơ này thuộc về người dùng khác")
	status.Register(GameMissingSessionID, "gsess_id is required", "Thiếu gsess_id")
	status.Register(GameInvalidGrade, "Grade must be between 0 and 5", "Lớp phải từ 0 đến 5")
	status.Register(GameInvalidNumQuestions, "The number of questions must be between 1 and 50", "Số câu hỏi phải từ 1 đến 50")
	status.Register(GameGuestNotAllowed, "Please register to play games", "Vui lòng đăng ký tài khoản để chơi game")
	status.Register(GameQuestionBankInsufficient, "There are not enough questions for this game", "Không đủ câu hỏi cho ván chơi này")
	status.Register(GameInvalidAnswers, "Invalid answers", "Câu trả lời không hợp lệ")
	status.Register(GameAlreadySubmitted, "This game has already been submitted", "Ván chơi này đã được nộp")
	status.Register(GameGenerationFailed, "Failed to create the game", "Tạo ván chơi thất bại")
	status.Register(GameGradingFailed, "Failed to grade the game", "Chấm điểm ván chơi thất bại")
	status.Register(GameInvalidLevel, "Level must be between 0 and 9", "Level phải từ 0 đến 9")
}
