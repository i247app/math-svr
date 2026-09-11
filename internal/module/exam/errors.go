package exam

import "errors"

// Plain sentinel errors carried as the debug cause inside a MathError.
// They exist so the message a developer reads is written once, next to
// the rule it describes, instead of inline at every call site.
var (
	ErrUidNotFoundFromSession  = errors.New("uid not found from session")
	ErrProfileIDRequired       = errors.New("profile_id is required")
	ErrProfileNotFound         = errors.New("profile not found")
	ErrProfileNotOwned         = errors.New("profile does not belong to this user")
	ErrExamTypeRequired        = errors.New("exam_type is required")
	ErrExamTypeInvalid         = errors.New("exam_type must be one of ASSESSMENT, PRACTICE, EXAM")
	ErrGradeOutOfRange         = errors.New("grade must be between 0 and 5")
	ErrAttemptIDRequired       = errors.New("user_ai_exam_id is required")
	ErrAnswersRequired         = errors.New("answers is required")
	ErrDuplicateAnswer         = errors.New("answers contain a duplicate question_number")
	ErrQuestionNumberInvalid   = errors.New("question_number must be positive")
	ErrAnswerLabelRequired     = errors.New("answer label is required")
	ErrBotAdapterNotConfigured = errors.New("bot adapter is not configured")
	ErrModelReturnedNothing    = errors.New("model returned no questions")
	ErrInvalidTz               = errors.New("tz must be a numeric offset such as +07:00")
	ErrInvalidDateRange        = errors.New("from_dt and to_dt must both be set and span at most two years")
)
