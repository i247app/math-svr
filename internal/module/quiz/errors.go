package quiz

import "errors"

// Module-scoped sentinel errors. See conventions.md §Errors.
var (
	ErrAnswersRequired                                           = errors.New("answers are required")
	ErrAtLeastOneOfProfileIDOrUserIDRequired                     = errors.New("at least one of profile_id or user_id is required")
	ErrLabelRequired                                             = errors.New("label is required")
	ErrLanguageMustBeVnOrEn                                      = errors.New("language must be 'vn' or 'en'")
	ErrPreviousQuizDoesNotBelongToThisProfile                    = errors.New("previous quiz does not belong to this profile")
	ErrPreviousQuizMustBeSubmittedBeforeGeneratingReinforceRound = errors.New("previous quiz must be submitted before generating a reinforce round")
	ErrPreviousQuizNotFound                                      = errors.New("previous quiz not found")
	ErrProfileNotFound                                           = errors.New("profile not found")
	ErrDefaultProfileNotFound                                    = errors.New("default profile not found")
	ErrPurposeRequired                                           = errors.New("purpose is required")
	ErrPurposeMustBeOneOfAssessmentPracticeExam                  = errors.New("purpose must be one of ASSESSMENT, PRACTICE, EXAM")
	ErrQuestionNumberMustBePositive                              = errors.New("question_number must be positive")
	ErrQuizHasAlreadyBeenSubmitted                               = errors.New("quiz has already been submitted")
	ErrQuizHasNoQuestionsToGrade                                 = errors.New("quiz has no questions to grade")
	ErrQuizNotFound                                              = errors.New("quiz not found")
	ErrQuizIDRequired                                            = errors.New("quiz_id is required")
	ErrQuizBotAdapterNotConfigured                               = errors.New("quiz: bot adapter is not configured")
	ErrQuizModelReturnedZeroQuestions                            = errors.New("quiz: model returned zero questions")
	ErrTypeOfQuizMustBeOneOfGeneralReinforcement                 = errors.New("type_of_quiz must be one of GENERAL, REINFORCEMENT")
	ErrUidNotFoundFromSession                                    = errors.New("uid not found from session")
)
