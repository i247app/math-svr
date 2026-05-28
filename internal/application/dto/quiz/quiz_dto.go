package quiz

import (
	"encoding/json"

	domain "math-ai.com/math-ai/internal/domain/quiz"
	"math-ai.com/math-ai/internal/shared/enum"
	"math-ai.com/math-ai/internal/shared/pagination"
)

// QuizAnswerChoice is one option in a multiple-choice question. The
// label is the alphabetical key ("A".."D") the student selects to
// answer; content is the displayed text.
type QuizAnswerChoice struct {
	Label   string `json:"label"`
	Content string `json:"content"`
}

// QuizQuestion is one MCQ item as produced by the bot. RightAnswer is
// the label of the correct option and is only included in responses
// once the quiz has been graded (otherwise it would leak the answer key).
type QuizQuestion struct {
	QuestionNumber int                `json:"question_number"`
	QuestionName   string             `json:"question_name"`
	Answers        []QuizAnswerChoice `json:"answers"`
	RightAnswer    string             `json:"right_answer,omitempty"`
}

// QuizStudentAnswer is the student's chosen label for a single question.
type QuizStudentAnswer struct {
	QuestionNumber int    `json:"question_number"`
	Label          string `json:"label"`
}

// QuizGradingResult is the bot's grading output. AIDetectGrade is only
// populated for ASSESSMENT (and reinforce-assessment); PRACTICE rounds
// omit it.
type QuizGradingResult struct {
	TotalQuestions  int     `json:"total_questions"`
	CorrectNumber   int     `json:"correct_number"`
	ScorePercentage int     `json:"score_percentage"`
	AIReview        string  `json:"ai_review"`
	AIDetectGrade   *string `json:"ai_detect_grade,omitempty"`
}

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
	ID             int64               `json:"id"`
	QuizID         string              `json:"quiz_id"`
	UserID         *string             `json:"user_id,omitempty"`
	ProfileID      *string             `json:"profile_id,omitempty"`
	Purpose        string              `json:"purpose"`
	TypeOfQuiz     *string             `json:"type_of_quiz,omitempty"`
	Title          *string             `json:"title,omitempty"`
	PreviousQuizID *string             `json:"previous_quiz_id,omitempty"`
	Questions      []QuizQuestion      `json:"questions,omitempty"`
	Answers        []QuizStudentAnswer `json:"answers,omitempty"`
	Grading        *QuizGradingResult  `json:"grading,omitempty"`
	QuizStatus     *string             `json:"quiz_status,omitempty"`
	CreateDt       string              `json:"create_dt"`
	ModifyDt       string              `json:"modify_dt"`
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
//   - ChapterDescriptions, when non-empty, OVERRIDE the chapters the
//     resolver would derive from the profile's (program, grade,
//     semester) triple. This is what lets anonymous callers (no
//     profile) still pin chapter focus, and lets profile-bearing
//     callers override the curriculum-derived list when they want to
//     scope the quiz to a subset of chapters.
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
	UserID              *string           `json:"user_id"`
	ProfileID           *string           `json:"profile_id,omitempty"`
	Purpose             string            `json:"purpose"`
	TypeOfQuiz          string            `json:"type_of_quiz,omitempty"`
	Language            enum.LanguageType `json:"language,omitempty"`
	ProgramLabel        string            `json:"program_label,omitempty"`
	GradeLabel          string            `json:"grade_label,omitempty"`
	SemesterLabel       string            `json:"semester_label,omitempty"`
	ChapterDescriptions []string          `json:"chapter_descriptions,omitempty"`
	NumQuestions        int               `json:"num_questions,omitempty"`
	PreviousQuizID      *string           `json:"previous_quiz_id,omitempty"`
}

type GenerateQuizRes struct {
	Quiz *QuizResponse `json:"quiz"`
}

type SubmitQuizAnswersReq struct {
	QuizID   string              `json:"quiz_id"`
	Language enum.LanguageType   `json:"language,omitempty"`
	Answers  []QuizStudentAnswer `json:"answers"`
}

type SubmitQuizAnswersRes struct {
	Quiz *QuizResponse `json:"quiz"`
}

type GetQuizByQuizIdReq struct {
	QuizID string `json:"quiz_id"`
}

type GetQuizByQuizIdRes struct {
	Quiz *QuizResponse `json:"quiz"`
}

// ListQuizzesReq filters the quiz list. ProfileID and UserID are both
// optional; at least one must be supplied. When both are supplied they
// are AND'd so the result is the intersection (a specific child of a
// specific parent).
type ListQuizzesReq struct {
	ProfileID *string `json:"profile_id,omitempty"`
	UserID    *string `json:"user_id,omitempty"`
	Purpose   *string `json:"purpose,omitempty"`
	Page      int     `json:"page,omitempty"`
	Size      int     `json:"size,omitempty"`
}

type ListQuizzesRes struct {
	Quizzes    []*QuizResponse        `json:"quizzes"`
	Pagination *pagination.Pagination `json:"pagination"`
}

type DeleteQuizReq struct {
	QuizID string `json:"quiz_id"`
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
		ID:             q.Id(),
		QuizID:         q.QuizId(),
		UserID:         q.UserId(),
		ProfileID:      q.ProfileId(),
		Purpose:        q.Purpose(),
		TypeOfQuiz:     q.TypeOfQuiz(),
		Title:          q.Title(),
		PreviousQuizID: q.PreviousQuizId(),
		QuizStatus:     q.QuizStatus(),
		CreateDt:       q.CreateDt().String(),
		ModifyDt:       q.ModifyDt().String(),
	}

	if questions := parseQuestions(q.Questions()); len(questions) > 0 {
		if !includeRightAnswers {
			for i := range questions {
				questions[i].RightAnswer = ""
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
	if q.AIReview() == nil {
		return nil
	}
	g := &QuizGradingResult{
		AIReview:      *q.AIReview(),
		AIDetectGrade: q.AIDetectGrade(),
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
