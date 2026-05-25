package quiz

import (
	mtime "math-ai.com/math-ai/internal/domain/shared/time"
)

// Quiz is a single round of AI-generated questions tied to a specific
// (user, profile) pair. One row per round — reinforce/remedial rounds
// are new rows that link back via previousQuizId.
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
	quizId          string
	userId          *string
	profileId       *string
	quizType        string
	questions       *string
	answers         *string
	aiReview        *string
	aiDetectGrade   *string
	totalQuestions  *int
	correctNumber   *int
	scorePercentage *int
	previousQuizId  *string
	note            *string
	quizStatus      *string
	status          string
	createId        *string
	createDt        mtime.MathTime
	modifyId        *string
	modifyDt        mtime.MathTime
}

func NewQuiz() *Quiz {
	return &Quiz{}
}

func (q *Quiz) Id() int64                  { return q.id }
func (q *Quiz) SetId(id int64)             { q.id = id }
func (q *Quiz) QuizId() string             { return q.quizId }
func (q *Quiz) SetQuizId(id string)        { q.quizId = id }
func (q *Quiz) UserId() *string            { return q.userId }
func (q *Quiz) SetUserId(id *string)       { q.userId = id }
func (q *Quiz) ProfileId() *string         { return q.profileId }
func (q *Quiz) SetProfileId(id *string)    { q.profileId = id }
func (q *Quiz) QuizType() string           { return q.quizType }
func (q *Quiz) SetQuizType(t string)       { q.quizType = t }
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
func (q *Quiz) PreviousQuizId() *string    { return q.previousQuizId }
func (q *Quiz) SetPreviousQuizId(id *string) {
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
func (q *Quiz) CreateId() *string            { return q.createId }
func (q *Quiz) SetCreateId(id *string)       { q.createId = id }
func (q *Quiz) CreateDt() mtime.MathTime     { return q.createDt }
func (q *Quiz) SetCreateDt(t mtime.MathTime) { q.createDt = t }
func (q *Quiz) ModifyId() *string            { return q.modifyId }
func (q *Quiz) SetModifyId(id *string)       { q.modifyId = id }
func (q *Quiz) ModifyDt() mtime.MathTime     { return q.modifyDt }
func (q *Quiz) SetModifyDt(t mtime.MathTime) { q.modifyDt = t }
