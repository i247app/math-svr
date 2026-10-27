package questionpool

import (
	"math-ai.com/math-ai/internal/domain/shared/mtime"
)

// Question is one row of the question bank (ma_question_pools), written and
// reviewed by admins. It belongs to no exam: nothing here refers to
// ma_exam_pools.
//
// questionHash fingerprints the content (see migration 036) so the same
// question cannot be stored twice while live.
type QuestionPool struct {
	questionPoolId   int64
	questionKind     Kind
	questionText     string
	questionImageKey *string
	options          []Option // nil for a kind with no options
	correctAnswer    CorrectAnswer
	explanation      *string
	questionGrade    int
	questionTopic    *string
	questionLevel    *int // 0..9, nil = not rated
	questionSubLevel *int // 0..9 inside questionLevel, nil = not rated
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

func NewQuestionPool() *QuestionPool { return &QuestionPool{} }

func (q *QuestionPool) QuestionPoolId() int64            { return q.questionPoolId }
func (q *QuestionPool) SetQuestionPoolId(id int64)       { q.questionPoolId = id }
func (q *QuestionPool) QuestionKind() Kind               { return q.questionKind }
func (q *QuestionPool) SetQuestionKind(k Kind)           { q.questionKind = k }
func (q *QuestionPool) QuestionText() string             { return q.questionText }
func (q *QuestionPool) SetQuestionText(s string)         { q.questionText = s }
func (q *QuestionPool) QuestionImageKey() *string        { return q.questionImageKey }
func (q *QuestionPool) SetQuestionImageKey(s *string)    { q.questionImageKey = s }
func (q *QuestionPool) Options() []Option                { return q.options }
func (q *QuestionPool) SetOptions(o []Option)            { q.options = o }
func (q *QuestionPool) CorrectAnswer() CorrectAnswer     { return q.correctAnswer }
func (q *QuestionPool) SetCorrectAnswer(a CorrectAnswer) { q.correctAnswer = a }
func (q *QuestionPool) Explanation() *string             { return q.explanation }
func (q *QuestionPool) SetExplanation(s *string)         { q.explanation = s }
func (q *QuestionPool) QuestionGrade() int               { return q.questionGrade }
func (q *QuestionPool) SetQuestionGrade(g int)           { q.questionGrade = g }
func (q *QuestionPool) QuestionTopic() *string           { return q.questionTopic }
func (q *QuestionPool) SetQuestionTopic(s *string)       { q.questionTopic = s }
func (q *QuestionPool) QuestionLevel() *int              { return q.questionLevel }
func (q *QuestionPool) SetQuestionLevel(l *int)          { q.questionLevel = l }
func (q *QuestionPool) QuestionSubLevel() *int           { return q.questionSubLevel }
func (q *QuestionPool) SetQuestionSubLevel(l *int)       { q.questionSubLevel = l }
func (q *QuestionPool) QuestionTags() []string           { return q.questionTags }
func (q *QuestionPool) SetQuestionTags(t []string)       { q.questionTags = t }
func (q *QuestionPool) QuestionSource() Source           { return q.questionSource }
func (q *QuestionPool) SetQuestionSource(s Source)       { q.questionSource = s }
func (q *QuestionPool) VerifiedCount() int               { return q.verifiedCount }
func (q *QuestionPool) SetVerifiedCount(n int)           { q.verifiedCount = n }
func (q *QuestionPool) QuestionHash() string             { return q.questionHash }
func (q *QuestionPool) SetQuestionHash(h string)         { q.questionHash = h }
func (q *QuestionPool) QuestionStatus() *string          { return q.questionStatus }
func (q *QuestionPool) SetQuestionStatus(s *string)      { q.questionStatus = s }
func (q *QuestionPool) RptFlg() *string                  { return q.rptFlg }
func (q *QuestionPool) SetRptFlg(s *string)              { q.rptFlg = s }
func (q *QuestionPool) Kwords() *string                  { return q.kwords }
func (q *QuestionPool) SetKwords(s *string)              { q.kwords = s }
func (q *QuestionPool) Note() *string                    { return q.note }
func (q *QuestionPool) SetNote(s *string)                { q.note = s }
func (q *QuestionPool) Status() string                   { return q.status }
func (q *QuestionPool) SetStatus(s string)               { q.status = s }
func (q *QuestionPool) CreateId() *int64                 { return q.createId }
func (q *QuestionPool) SetCreateId(id *int64)            { q.createId = id }
func (q *QuestionPool) CreateDt() mtime.MathTime         { return q.createDt }
func (q *QuestionPool) SetCreateDt(t mtime.MathTime)     { q.createDt = t }
func (q *QuestionPool) ModifyId() *int64                 { return q.modifyId }
func (q *QuestionPool) SetModifyId(id *int64)            { q.modifyId = id }
func (q *QuestionPool) ModifyDt() mtime.MathTime         { return q.modifyDt }
func (q *QuestionPool) SetModifyDt(t mtime.MathTime)     { q.modifyDt = t }
