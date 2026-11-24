package game

import "errors"

var (
	ErrUIDNotFoundFromSession = errors.New("uid not found from session")
	ErrProfileIDRequired      = errors.New("profile_id is required")
	ErrProfileNotFound        = errors.New("profile not found")
	ErrProfileNotOwned        = errors.New("profile does not belong to this user")
	ErrGuestNotAllowed        = errors.New("a guest may not play games")
	ErrGsessIDRequired        = errors.New("gsess_id is required")
	ErrGradeOutOfRange        = errors.New("grade must be between 0 and 5")
	ErrLevelOutOfRange        = errors.New("level must be between 0 and 9")
	ErrNumQuestionsOutOfRange = errors.New("num_questions must be between 1 and 50")
)
