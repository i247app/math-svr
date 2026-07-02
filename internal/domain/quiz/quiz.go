package quiz

import (
	"math-ai.com/math-ai/internal/domain/shared/mtime"
)

// Quiz is a single round of AI-generated questions tied to a specific
// (user, profile) pair. One row per round — reinforce/remedial rounds
// are new rows that link back via previousQuizId.
//
// purpose categorises the *intent* of the quiz (ASSESSMENT, PRACTICE,
// EXAM) — what was previously persisted in the legacy `type` column.
// typeOfQuiz categorises the *learning shape* (GENERAL = new knowledge,
// REINFORCEMENT = consolidating / reviewing past mistakes). Both are
// persisted; they evolve independently.
//
// questions and answers are opaque JSON blobs: questions follow the
// generation prompt schema, answers follow {question_number: label}.
// They are stored verbatim from / to the LLM to avoid leaky parsing
// invariants — application-layer code parses on read.
//
// aiReview / aiDetectGrade are populated only after submission; both are
// nullable on the schema (see migration 012) to allow generate-then-grade
// in separate steps.
type Quiz struct {
	id              int64
	quizId          int64
	userId          *int64
	profileId       *int64
	purpose         string
	typeOfQuiz      *string
	title           *string // AI-generated grade/level label, e.g. "Grade 1 - Level 1"
	shortText       *string // AI-generated short topic description, e.g. "Các số trong phạm vi 20"
	questions       *string
	answers         *string
	aiReview        *string
	aiDetectGrade   *string
	totalQuestions  *int
	correctNumber   *int
	scorePercentage *int
	previousQuizId  *int64
	note            *string
	quizStatus      *string
	status          string
	createId        *int64
	createDt        mtime.MathTime
	modifyId        *int64
	modifyDt        mtime.MathTime
}

func NewQuiz() *Quiz {
	return &Quiz{}
}

func (q *Quiz) Id() int64                  { return q.id }
func (q *Quiz) SetId(id int64)             { q.id = id }
func (q *Quiz) QuizId() int64              { return q.quizId }
func (q *Quiz) SetQuizId(id int64)         { q.quizId = id }
func (q *Quiz) UserId() *int64             { return q.userId }
func (q *Quiz) SetUserId(id *int64)        { q.userId = id }
func (q *Quiz) ProfileId() *int64          { return q.profileId }
func (q *Quiz) SetProfileId(id *int64)     { q.profileId = id }
func (q *Quiz) Purpose() string            { return q.purpose }
func (q *Quiz) SetPurpose(p string)        { q.purpose = p }
func (q *Quiz) TypeOfQuiz() *string        { return q.typeOfQuiz }
func (q *Quiz) SetTypeOfQuiz(t *string)    { q.typeOfQuiz = t }
func (q *Quiz) Title() *string             { return q.title }
func (q *Quiz) SetTitle(t *string)         { q.title = t }
func (q *Quiz) ShortText() *string         { return q.shortText }
func (q *Quiz) SetShortText(s *string)     { q.shortText = s }
func (q *Quiz) Questions() *string         { return q.questions }
func (q *Quiz) SetQuestions(s *string)     { q.questions = s }
func (q *Quiz) Answers() *string           { return q.answers }
func (q *Quiz) SetAnswers(s *string)       { q.answers = s }
func (q *Quiz) AIReview() *string          { return q.aiReview }
func (q *Quiz) SetAIReview(s *string)      { q.aiReview = s }
func (q *Quiz) AIDetectGrade() *string     { return q.aiDetectGrade }
func (q *Quiz) SetAIDetectGrade(s *string) { q.aiDetectGrade = s }
func (q *Quiz) TotalQuestions() *int       { return q.totalQuestions }
func (q *Quiz) SetTotalQuestions(n *int)   { q.totalQuestions = n }
func (q *Quiz) CorrectNumber() *int        { return q.correctNumber }
func (q *Quiz) SetCorrectNumber(n *int)    { q.correctNumber = n }
func (q *Quiz) ScorePercentage() *int      { return q.scorePercentage }
func (q *Quiz) SetScorePercentage(n *int)  { q.scorePercentage = n }
func (q *Quiz) PreviousQuizId() *int64     { return q.previousQuizId }
func (q *Quiz) SetPreviousQuizId(id *int64) {
	if id == nil {
		q.previousQuizId = nil
		return
	}
	q.previousQuizId = id
}
func (q *Quiz) Note() *string                { return q.note }
func (q *Quiz) SetNote(s *string)            { q.note = s }
func (q *Quiz) QuizStatus() *string          { return q.quizStatus }
func (q *Quiz) SetQuizStatus(s *string)      { q.quizStatus = s }
func (q *Quiz) Status() string               { return q.status }
func (q *Quiz) SetStatus(s string)           { q.status = s }
func (q *Quiz) CreateId() *int64             { return q.createId }
func (q *Quiz) SetCreateId(id *int64)        { q.createId = id }
func (q *Quiz) CreateDt() mtime.MathTime     { return q.createDt }
func (q *Quiz) SetCreateDt(t mtime.MathTime) { q.createDt = t }
func (q *Quiz) ModifyId() *int64             { return q.modifyId }
func (q *Quiz) SetModifyId(id *int64)        { q.modifyId = id }
func (q *Quiz) ModifyDt() mtime.MathTime     { return q.modifyDt }
func (q *Quiz) SetModifyDt(t mtime.MathTime) { q.modifyDt = t }
