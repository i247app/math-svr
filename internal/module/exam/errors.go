package exam

import "errors"

// Plain sentinel errors carried as the debug cause inside a MathError.
// They exist so the message a developer reads is written once, next to
// the rule it describes, instead of inline at every call site.
var (
	ErrUidNotFoundFromSession   = errors.New("uid not found from session")
	ErrSessionNotSecure         = errors.New("session is not secure; finish signing in first")
	ErrProfileIDRequired        = errors.New("profile_id is required")
	ErrProfileNotFound          = errors.New("profile not found")
	ErrProfileNotOwned          = errors.New("profile does not belong to this user")
	ErrGuestProfileNotOwned     = errors.New("a profile_id may only be stated by an authenticated caller")
	ErrGuestAssessmentOnly      = errors.New("a guest may only be handed an ASSESSMENT round")
	ErrGuestDailyLimit          = errors.New("this guest has reached the daily exam ceiling")
	ErrExamTypeRequired         = errors.New("exam_type is required")
	ErrExamTypeInvalid          = errors.New("exam_type must be one of ASSESSMENT, PRACTICE, GRADE")
	ErrPracticeJourneyRequired  = errors.New("esess_id is required for a PRACTICE exam")
	ErrStatsPracticeNotAJourney = errors.New("PRACTICE is not a journey; read it under its journey's practice field")
	ErrLevelOutOfRange          = errors.New("level must be between 0 and 9")
	ErrGradeOutOfRange          = errors.New("grade must be between 0 and 5")
	ErrGradeRequired            = errors.New("grade is required")
	ErrAttemptIDRequired        = errors.New("elink_id is required")
	ErrAnswersRequired          = errors.New("answers is required")
	ErrDuplicateAnswer          = errors.New("answers contain a duplicate question_number")
	ErrQuestionNumberInvalid    = errors.New("question_number must be positive")
	ErrAnswerLabelRequired      = errors.New("answer label is required")
	ErrJourneyIDRequired        = errors.New("esess_id is required")
	ErrJourneyStatusInvalid     = errors.New("status must be ACTIVE, COMPLETE or CANCEL")
	ErrDetailIDRequired         = errors.New("one of elink_id or esess_id is required")
	ErrDetailIDAmbiguous        = errors.New("send either elink_id or esess_id, not both")
	ErrBotAdapterNotConfigured  = errors.New("bot adapter is not configured")
	ErrModelReturnedNothing     = errors.New("model returned no questions")
	ErrInvalidTz                = errors.New("tz must be a numeric offset such as +07:00")
	ErrInvalidDateRange         = errors.New("from_dt and to_dt must both be set and span at most two years")
)
