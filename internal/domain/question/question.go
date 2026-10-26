package question

import (
	"math-ai.com/math-ai/internal/domain/shared/mtime"
)

// Question is one row of the question bank (ma_questions), written and
// reviewed by admins. It belongs to no exam: nothing here refers to
// ma_exam_pools.
//
// questionHash fingerprints the content (see migration 036) so the same
// question cannot be stored twice while live.
type Question struct {
	questionId       int64
	questionKind     Kind
	questionText     string
	questionImageKey *string
	options          []Option // nil for a kind with no options
	correctAnswer    CorrectAnswer
	explanation      *string
	questionGrade    int
	questionTopic    *string
	questionLevel    *int // 0..9, nil = not rated
	questionTags     []string
	questionSource   Source
	verifiedCount    int
	questionHash     string

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

func (q *Question) QuestionId() int64                { return q.questionId }
func (q *Question) SetQuestionId(id int64)           { q.questionId = id }
func (q *Question) QuestionKind() Kind               { return q.questionKind }
func (q *Question) SetQuestionKind(k Kind)           { q.questionKind = k }
func (q *Question) QuestionText() string             { return q.questionText }
func (q *Question) SetQuestionText(s string)         { q.questionText = s }
func (q *Question) QuestionImageKey() *string        { return q.questionImageKey }
func (q *Question) SetQuestionImageKey(s *string)    { q.questionImageKey = s }
func (q *Question) Options() []Option                { return q.options }
func (q *Question) SetOptions(o []Option)            { q.options = o }
func (q *Question) CorrectAnswer() CorrectAnswer     { return q.correctAnswer }
func (q *Question) SetCorrectAnswer(a CorrectAnswer) { q.correctAnswer = a }
func (q *Question) Explanation() *string             { return q.explanation }
func (q *Question) SetExplanation(s *string)         { q.explanation = s }
func (q *Question) QuestionGrade() int               { return q.questionGrade }
func (q *Question) SetQuestionGrade(g int)           { q.questionGrade = g }
func (q *Question) QuestionTopic() *string           { return q.questionTopic }
func (q *Question) SetQuestionTopic(s *string)       { q.questionTopic = s }
func (q *Question) QuestionLevel() *int              { return q.questionLevel }
func (q *Question) SetQuestionLevel(l *int)          { q.questionLevel = l }
func (q *Question) QuestionTags() []string           { return q.questionTags }
func (q *Question) SetQuestionTags(t []string)       { q.questionTags = t }
func (q *Question) QuestionSource() Source           { return q.questionSource }
func (q *Question) SetQuestionSource(s Source)       { q.questionSource = s }
func (q *Question) VerifiedCount() int               { return q.verifiedCount }
func (q *Question) SetVerifiedCount(n int)           { q.verifiedCount = n }
func (q *Question) QuestionHash() string             { return q.questionHash }
func (q *Question) SetQuestionHash(h string)         { q.questionHash = h }
func (q *Question) QuestionStatus() *string          { return q.questionStatus }
func (q *Question) SetQuestionStatus(s *string)      { q.questionStatus = s }
func (q *Question) RptFlg() *string                  { return q.rptFlg }
func (q *Question) SetRptFlg(s *string)              { q.rptFlg = s }
func (q *Question) Kwords() *string                  { return q.kwords }
func (q *Question) SetKwords(s *string)              { q.kwords = s }
func (q *Question) Note() *string                    { return q.note }
func (q *Question) SetNote(s *string)                { q.note = s }
func (q *Question) Status() string                   { return q.status }
func (q *Question) SetStatus(s string)               { q.status = s }
func (q *Question) CreateId() *int64                 { return q.createId }
func (q *Question) SetCreateId(id *int64)            { q.createId = id }
func (q *Question) CreateDt() mtime.MathTime         { return q.createDt }
func (q *Question) SetCreateDt(t mtime.MathTime)     { q.createDt = t }
func (q *Question) ModifyId() *int64                 { return q.modifyId }
func (q *Question) SetModifyId(id *int64)            { q.modifyId = id }
func (q *Question) ModifyDt() mtime.MathTime         { return q.modifyDt }
func (q *Question) SetModifyDt(t mtime.MathTime)     { q.modifyDt = t }
