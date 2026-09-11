package quiz

import (
	"encoding/json"

	"math-ai.com/math-ai/internal/application/dto/question"
	domain "math-ai.com/math-ai/internal/domain/quiz"
	"math-ai.com/math-ai/internal/shared/enum"
	"math-ai.com/math-ai/internal/shared/pagination"
)

// The MCQ contract moved to internal/application/dto/question so the
// exercise module (and, from the exam refactor on, the exam module) can
// share it without importing this package. These aliases keep every
// existing reference — and the wire shape — byte-for-byte unchanged.
type (
	QuizAnswerChoice  = question.AnswerChoice
	QuizQuestion      = question.Question
	QuizStudentAnswer = question.StudentAnswer
	QuizGradingResult = question.GradingResult
)

const (
	QuestionTypeArithmetic    = question.TypeArithmetic
	QuestionTypeCount         = question.TypeCount
	QuestionTypePickByIcon    = question.TypePickByIcon
	QuestionTypeIdentifyShape = question.TypeIdentifyShape
)

// QuizResponse is the wire shape returned by every quiz endpoint.
// Questions / Answers / Grading are nil-omitted so generated and
// submitted shapes share one envelope. UserID / ProfileID are nullable
// because anonymous quizzes (generated without a profile) carry neither.
//
// Purpose carries the persisted ma_quizzes.purpose value (ASSESSMENT /
// PRACTICE / EXAM); TypeOfQuiz carries ma_quizzes.type_of_quiz (GENERAL
// / REINFORCEMENT). TypeOfQuiz is *string because historical rows may
// be NULL — the column has a DB-level default but Go-side we don't
// assume it.
type QuizResponse struct {
	ID              int64               `json:"id"`
	QuizID          int64               `json:"quiz_id"`
	UserID          *int64              `json:"user_id,omitempty"`
	ProfileID       *int64              `json:"profile_id,omitempty"`
	Purpose         string              `json:"purpose"`
	TypeOfQuiz      *string             `json:"type_of_quiz,omitempty"`
	Title           *string             `json:"title,omitempty"`            // AI grade/level label, e.g. "Grade 1 - Level 1"
	ShortText       *string             `json:"short_text,omitempty"`       // AI short topic description
	AssessmentGrade *string             `json:"assessment_grade,omitempty"` // grade the AI calibrated the quiz to, set at generation
	PreviousQuizID  *int64              `json:"previous_quiz_id,omitempty"`
	Questions       []QuizQuestion      `json:"questions,omitempty"`
	Answers         []QuizStudentAnswer `json:"answers,omitempty"`
	Grading         *QuizGradingResult  `json:"grading,omitempty"`
	QuizStatus      *string             `json:"quiz_status,omitempty"`
	CreateDt        string              `json:"create_dt"`
	ModifyDt        string              `json:"modify_dt"`
}

// GenerateQuizReq carries the quiz-generation request. Both ProfileID
// and the academic context fields are optional:
//   - ProfileID, when set, supplies the user owner and the fallback
//     curriculum source; when absent the quiz is generated anonymously
//     (no profile or user owner) and is only retrievable by quiz_id.
//   - ProgramLabel / GradeLabel / SemesterLabel, when set, OVERRIDE
//     whatever the profile would resolve to. Missing labels fall back
//     to the profile's curriculum; if the profile has none, the bot
//     prompt adapts and still generates a reasonable elementary-level
//     quiz.
//   - ChapterDescriptions is the only source of chapter focus — the
//     server derives none of its own. Anonymous and profile-bearing
//     callers alike pin it here to scope the quiz to specific units;
//     leaving it empty just drops the chapter block from the prompt.
//
// Purpose accepts the persisted ma_quizzes.purpose vocabulary
// (ASSESSMENT / PRACTICE / EXAM). TypeOfQuiz is optional — when omitted
// it is inferred from PreviousQuizID (set ⇒ REINFORCEMENT, otherwise
// GENERAL); when supplied it is validated and used as-is.
//
// Labels (not IDs) are accepted so the client can supply ad-hoc context
// like "Grade 2 fractions review" without needing a curriculum row, and
// so the service does not have to do an extra round-trip to translate.
type GenerateQuizReq struct {
	UserID              *int64            `json:"user_id"`
	ProfileID           *int64            `json:"profile_id,omitempty"`
	Purpose             string            `json:"purpose"`
	TypeOfQuiz          string            `json:"type_of_quiz,omitempty"`
	Language            enum.LanguageType `json:"language,omitempty"`
	ProgramLabel        string            `json:"program_label,omitempty"`
	GradeLabel          string            `json:"grade_label,omitempty"`
	SemesterLabel       string            `json:"semester_label,omitempty"`
	ChapterDescriptions []string          `json:"chapters,omitempty"`
	NumQuestions        int               `json:"num_questions,omitempty"`
	PreviousQuizID      *int64            `json:"previous_quiz_id,omitempty"`
}

type GenerateQuizRes struct {
	Quiz *QuizResponse `json:"quiz"`
}

type SubmitQuizAnswersReq struct {
	QuizID   int64               `json:"quiz_id"`
	Language enum.LanguageType   `json:"language,omitempty"`
	Answers  []QuizStudentAnswer `json:"answers"`
}

type SubmitQuizAnswersRes struct {
	Quiz *QuizResponse `json:"quiz"`
}

type GetQuizByQuizIdReq struct {
	QuizID int64 `json:"quiz_id"`
}

type GetQuizByQuizIdRes struct {
	Quiz *QuizResponse `json:"quiz"`
}

// ListQuizzesReq filters the quiz list. ProfileID and UserID are both
// optional; at least one must be supplied. When both are supplied they
// are AND'd so the result is the intersection (a specific child of a
// specific parent).
type ListQuizzesReq struct {
	ProfileID *int64  `json:"profile_id,omitempty"`
	UserID    *int64  `json:"user_id,omitempty"`
	Purpose   *string `json:"purpose,omitempty"`
	Page      int     `json:"page,omitempty"`
	Size      int     `json:"size,omitempty"`
}

type ListQuizzesRes struct {
	Quizzes    []*QuizResponse        `json:"quizzes"`
	Pagination *pagination.Pagination `json:"pagination"`
}

type DeleteQuizReq struct {
	QuizID int64 `json:"quiz_id"`
}

type DeleteQuizRes struct{}

// DomainToResponse maps a domain Quiz into its wire shape. The
// includeRightAnswers flag controls whether the generation answer key
// is exposed; callers set it to true only for already-graded quizzes
// (review / history endpoints) so live quizzes don't leak the answer.
func DomainToResponse(q *domain.Quiz, includeRightAnswers bool) *QuizResponse {
	if q == nil {
		return nil
	}

	res := &QuizResponse{
		ID:              q.Id(),
		QuizID:          q.QuizId(),
		UserID:          q.UserId(),
		ProfileID:       q.ProfileId(),
		Purpose:         q.Purpose(),
		TypeOfQuiz:      q.TypeOfQuiz(),
		Title:           q.Title(),
		ShortText:       q.ShortText(),
		AssessmentGrade: q.AssessmentGrade(),
		PreviousQuizID:  q.PreviousQuizId(),
		QuizStatus:      q.QuizStatus(),
		CreateDt:        q.CreateDt().String(),
		ModifyDt:        q.ModifyDt().String(),
	}

	if questions := parseQuestions(q.Questions()); len(questions) > 0 {
		if !includeRightAnswers {
			for i := range questions {
				questions[i].RightAnswerLabel = ""
				questions[i].RightAnswerContent = ""
			}
		}
		res.Questions = questions
	}
	if answers := parseAnswers(q.Answers()); len(answers) > 0 {
		res.Answers = answers
	}
	if grading := parseGrading(q); grading != nil {
		res.Grading = grading
	}
	return res
}

// DomainListToResponse maps a slice of domain Quizzes; the answer key is
// included only for SUBMITTED entries so a graded history view shows the
// correct answers while in-flight quizzes do not leak them.
func DomainListToResponse(quizzes []*domain.Quiz) []*QuizResponse {
	result := make([]*QuizResponse, len(quizzes))
	for i, q := range quizzes {
		includeAnswers := false
		if s := q.QuizStatus(); s != nil && *s == string(enum.QuizStatusTypeSubmitted) {
			includeAnswers = true
		}
		result[i] = DomainToResponse(q, includeAnswers)
	}
	return result
}

func parseQuestions(raw *string) []QuizQuestion {
	if raw == nil || *raw == "" {
		return nil
	}
	var out []QuizQuestion
	if err := json.Unmarshal([]byte(*raw), &out); err != nil {
		return nil
	}
	return out
}

func parseAnswers(raw *string) []QuizStudentAnswer {
	if raw == nil || *raw == "" {
		return nil
	}
	var out []QuizStudentAnswer
	if err := json.Unmarshal([]byte(*raw), &out); err != nil {
		return nil
	}
	return out
}

func parseGrading(q *domain.Quiz) *QuizGradingResult {
	if q.Review() == nil {
		return nil
	}
	g := &QuizGradingResult{
		Review:          *q.Review(),
		AssessmentGrade: q.AssessmentGrade(),
	}
	if q.TotalQuestions() != nil {
		g.TotalQuestions = *q.TotalQuestions()
	}
	if q.CorrectNumber() != nil {
		g.CorrectNumber = *q.CorrectNumber()
	}
	if q.ScorePercentage() != nil {
		g.ScorePercentage = *q.ScorePercentage()
	}
	return g
}
