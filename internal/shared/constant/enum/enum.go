package enum

const (
	RoleAdmin ERole = "admin"
	RoleUser  ERole = "user"
	RoleGuest ERole = "guest"

	StatusActive   EStatus = "ACTIVE"
	StatusInactive EStatus = "INACTIVE"
	StatusBanned   EStatus = "BANNED"

	LoginStatusActive            ELoginStatus = "ACTIVE"
	LoginStatusTwoFactorRequired ELoginStatus = "2FA_REQUIRED"

	TypeQuizMath    ETypeQuiz = "MATH"
	TypeQuizScience ETypeQuiz = "SCIENCE"
	TypeQuizHistory ETypeQuiz = "HISTORY"

	TypeQuizPurpuseNew      ETypeQuizPurpuse = "NEW"
	TypeQuizPurpusePractice ETypeQuizPurpuse = "PRACTICE"
	TypeQuizPurpuseExam     ETypeQuizPurpuse = "EXAM"

	DefaultLang = "en"
)
