package question

import (
	"math-ai.com/math-ai/internal/domain/shared/mtime"
)

// Answer is one option of a bank question, in display order. There is no
// A/B/C/D label: a label is how one sitting happened to show the options,
// not part of the question.
type Answer struct {
	Content   string
	IsCorrect bool
}

// Question is one row of the question bank (ma_questions): one question of
// one ma_exam_pools set, identified by (examId, questionNumber) — the
// CANONICAL number inside that set, never a sitting's shuffled position.
//
// verifiedCount mirrors the pool's verified_count (0 = not verified by an
// admin, >0 = verified); the pool is the source of truth and every copy
// overwrites it.
type Question struct {
	questionId     int64
	examId         int64
	questionNumber int
	questionType   string
	questionName   string
	answers        []Answer
	questionTopic  *string
	questionGrade  int
	questionTags   []string
	verifiedCount  int

	questionStatus *string
	rptFlg         *string
	kwords         *string
	note           *string
	status         string
	createId       *int64
	createDt       mtime.MathTime
	modifyId       *int64
	modifyDt       mtime.MathTime
}

func NewQuestion() *Question { return &Question{} }

func (q *Question) QuestionId() int64            { return q.questionId }
func (q *Question) SetQuestionId(id int64)       { q.questionId = id }
func (q *Question) ExamId() int64                { return q.examId }
func (q *Question) SetExamId(id int64)           { q.examId = id }
func (q *Question) QuestionNumber() int          { return q.questionNumber }
func (q *Question) SetQuestionNumber(n int)      { q.questionNumber = n }
func (q *Question) QuestionType() string         { return q.questionType }
func (q *Question) SetQuestionType(t string)     { q.questionType = t }
func (q *Question) QuestionName() string         { return q.questionName }
func (q *Question) SetQuestionName(s string)     { q.questionName = s }
func (q *Question) Answers() []Answer            { return q.answers }
func (q *Question) SetAnswers(a []Answer)        { q.answers = a }
func (q *Question) QuestionTopic() *string       { return q.questionTopic }
func (q *Question) SetQuestionTopic(s *string)   { q.questionTopic = s }
func (q *Question) QuestionGrade() int           { return q.questionGrade }
func (q *Question) SetQuestionGrade(g int)       { q.questionGrade = g }
func (q *Question) QuestionTags() []string       { return q.questionTags }
func (q *Question) SetQuestionTags(t []string)   { q.questionTags = t }
func (q *Question) VerifiedCount() int           { return q.verifiedCount }
func (q *Question) SetVerifiedCount(n int)       { q.verifiedCount = n }
func (q *Question) QuestionStatus() *string      { return q.questionStatus }
func (q *Question) SetQuestionStatus(s *string)  { q.questionStatus = s }
func (q *Question) RptFlg() *string              { return q.rptFlg }
func (q *Question) SetRptFlg(s *string)          { q.rptFlg = s }
func (q *Question) Kwords() *string              { return q.kwords }
func (q *Question) SetKwords(s *string)          { q.kwords = s }
func (q *Question) Note() *string                { return q.note }
func (q *Question) SetNote(s *string)            { q.note = s }
func (q *Question) Status() string               { return q.status }
func (q *Question) SetStatus(s string)           { q.status = s }
func (q *Question) CreateId() *int64             { return q.createId }
func (q *Question) SetCreateId(id *int64)        { q.createId = id }
func (q *Question) CreateDt() mtime.MathTime     { return q.createDt }
func (q *Question) SetCreateDt(t mtime.MathTime) { q.createDt = t }
func (q *Question) ModifyId() *int64             { return q.modifyId }
func (q *Question) SetModifyId(id *int64)        { q.modifyId = id }
func (q *Question) ModifyDt() mtime.MathTime     { return q.modifyDt }
func (q *Question) SetModifyDt(t mtime.MathTime) { q.modifyDt = t }
