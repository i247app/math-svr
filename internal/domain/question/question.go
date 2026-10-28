package question

import (
	"math-ai.com/math-ai/internal/domain/questionpool"
	"math-ai.com/math-ai/internal/domain/shared/mtime"
)

// Option is one option of a generated question, in canonical order.
// Grading is by label.
type Option struct {
	ID   string
	Text string
}

// Question is one row of ma_questions: one element of an exam pool's
// questions_json, keyed by (examId, questionNumber). It belongs to the
// exam flow, not to the admin bank (ma_question_pools).
//
// questionKind is NULL for a single-choice question. correctAnswer is the
// key in the bank's shape (questionpool.CorrectAnswer, the same JSON as
// ma_question_pools.correct_answer): option ids for a choice question,
// the value for a true/false one, the accepted numbers for a numeric one
// (whose options are empty).
//
// The row is a copy: questions_json stays the source a sitting is served
// and graded from, and every write of that JSON rewrites the exam's rows in
// the same transaction.
type Question struct {
	questionId     int64
	examId         int64
	questionNumber int
	renderType     string
	questionKind   *string
	questionText   string
	options        []Option
	correctAnswer  questionpool.CorrectAnswer
	questionTopic  *string
	questionGrade  *int

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

func (q *Question) QuestionId() int64                             { return q.questionId }
func (q *Question) SetQuestionId(id int64)                        { q.questionId = id }
func (q *Question) ExamId() int64                                 { return q.examId }
func (q *Question) SetExamId(id int64)                            { q.examId = id }
func (q *Question) QuestionNumber() int                           { return q.questionNumber }
func (q *Question) SetQuestionNumber(n int)                       { q.questionNumber = n }
func (q *Question) RenderType() string                            { return q.renderType }
func (q *Question) SetRenderType(s string)                        { q.renderType = s }
func (q *Question) QuestionKind() *string                         { return q.questionKind }
func (q *Question) SetQuestionKind(k *string)                     { q.questionKind = k }
func (q *Question) QuestionText() string                          { return q.questionText }
func (q *Question) SetQuestionText(s string)                      { q.questionText = s }
func (q *Question) Options() []Option                             { return q.options }
func (q *Question) SetOptions(a []Option)                         { q.options = a }
func (q *Question) CorrectAnswer() questionpool.CorrectAnswer     { return q.correctAnswer }
func (q *Question) SetCorrectAnswer(a questionpool.CorrectAnswer) { q.correctAnswer = a }
func (q *Question) QuestionTopic() *string                        { return q.questionTopic }
func (q *Question) SetQuestionTopic(s *string)                    { q.questionTopic = s }
func (q *Question) QuestionGrade() *int                           { return q.questionGrade }
func (q *Question) SetQuestionGrade(g *int)                       { q.questionGrade = g }
func (q *Question) QuestionStatus() *string                       { return q.questionStatus }
func (q *Question) SetQuestionStatus(s *string)                   { q.questionStatus = s }
func (q *Question) RptFlg() *string                               { return q.rptFlg }
func (q *Question) SetRptFlg(s *string)                           { q.rptFlg = s }
func (q *Question) Kwords() *string                               { return q.kwords }
func (q *Question) SetKwords(s *string)                           { q.kwords = s }
func (q *Question) Note() *string                                 { return q.note }
func (q *Question) SetNote(s *string)                             { q.note = s }
func (q *Question) Status() string                                { return q.status }
func (q *Question) SetStatus(s string)                            { q.status = s }
func (q *Question) CreateId() *int64                              { return q.createId }
func (q *Question) SetCreateId(id *int64)                         { q.createId = id }
func (q *Question) CreateDt() mtime.MathTime                      { return q.createDt }
func (q *Question) SetCreateDt(t mtime.MathTime)                  { q.createDt = t }
func (q *Question) ModifyId() *int64                              { return q.modifyId }
func (q *Question) SetModifyId(id *int64)                         { q.modifyId = id }
func (q *Question) ModifyDt() mtime.MathTime                      { return q.modifyDt }
func (q *Question) SetModifyDt(t mtime.MathTime)                  { q.modifyDt = t }
