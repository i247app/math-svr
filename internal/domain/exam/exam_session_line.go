package exam

import (
	"math-ai.com/math-ai/internal/domain/shared/mtime"
)

// ExamSessionLine is one question the child ACTUALLY ANSWERED. A skipped
// question has no row at all, which is why "how many did they answer" is a
// row count here and never a NULL check.
//
// Rows accumulate under esessId for the life of the journey; grouping
// by elinkId splits them back into individual sittings. reqExamType
// is copied from the sitting so the journey's ASSESSMENT rows and its
// PRACTICE rows — which share esessId — can be told apart without a
// join: every per-journey read filters on the pair.
//
// The question fields are a snapshot taken at submit time out of the
// ExamPool's questions JSON. That freezes what the child was actually shown
// — the ExamPool row is shared cache and may be superseded — and keeps a
// review screen from parsing a LONGTEXT blob per row.
//
// questionGrade is the grade the QUESTION targets, which for an ASSESSMENT
// probe sits one band above the child's requested grade. It is the signal
// a promotion rule reads: "did they get the harder ones right?"
//
// questionLevel mirrors the nullable question_level column and is always
// NULL today: no rule defines a level, so neither the prompt nor the
// submit path produces one. The field stays so the entity keeps matching
// the table 1:1 and the column is ready the day a rule lands.
type ExamSessionLine struct {
	esessLnId   int64
	elinkId     int64
	esessId     int64
	examId      int64
	reqExamType string

	questionNumber     int
	questionType       *string
	questionName       *string
	questionTopic      *string
	questionGrade      *int
	questionLevel      *int
	rightAnswerLabel   *string
	rightAnswerContent *string

	selectedLabel   string
	selectedContent *string
	isCorrect       bool

	rptFlg *string

	kwords *string

	note          *string
	esessLnStatus *string
	status        string
	createId      *int64
	createDt      mtime.MathTime
	modifyId      *int64
	modifyDt      mtime.MathTime
}

func NewExamSessionLine() *ExamSessionLine { return &ExamSessionLine{} }

func (d *ExamSessionLine) EsessLnId() int64                { return d.esessLnId }
func (d *ExamSessionLine) SetEsessLnId(id int64)           { d.esessLnId = id }
func (d *ExamSessionLine) ElinkId() int64                  { return d.elinkId }
func (d *ExamSessionLine) SetElinkId(id int64)             { d.elinkId = id }
func (d *ExamSessionLine) EsessId() int64                  { return d.esessId }
func (d *ExamSessionLine) SetEsessId(id int64)             { d.esessId = id }
func (d *ExamSessionLine) ExamId() int64                   { return d.examId }
func (d *ExamSessionLine) ReqExamType() string             { return d.reqExamType }
func (d *ExamSessionLine) SetReqExamType(t string)         { d.reqExamType = t }
func (d *ExamSessionLine) SetExamId(id int64)              { d.examId = id }
func (d *ExamSessionLine) QuestionNumber() int             { return d.questionNumber }
func (d *ExamSessionLine) SetQuestionNumber(n int)         { d.questionNumber = n }
func (d *ExamSessionLine) QuestionType() *string           { return d.questionType }
func (d *ExamSessionLine) SetQuestionType(s *string)       { d.questionType = s }
func (d *ExamSessionLine) QuestionName() *string           { return d.questionName }
func (d *ExamSessionLine) SetQuestionName(s *string)       { d.questionName = s }
func (d *ExamSessionLine) QuestionTopic() *string          { return d.questionTopic }
func (d *ExamSessionLine) SetQuestionTopic(s *string)      { d.questionTopic = s }
func (d *ExamSessionLine) QuestionGrade() *int             { return d.questionGrade }
func (d *ExamSessionLine) SetQuestionGrade(g *int)         { d.questionGrade = g }
func (d *ExamSessionLine) QuestionLevel() *int             { return d.questionLevel }
func (d *ExamSessionLine) SetQuestionLevel(l *int)         { d.questionLevel = l }
func (d *ExamSessionLine) RightAnswerLabel() *string       { return d.rightAnswerLabel }
func (d *ExamSessionLine) SetRightAnswerLabel(s *string)   { d.rightAnswerLabel = s }
func (d *ExamSessionLine) RightAnswerContent() *string     { return d.rightAnswerContent }
func (d *ExamSessionLine) SetRightAnswerContent(s *string) { d.rightAnswerContent = s }
func (d *ExamSessionLine) SelectedLabel() string           { return d.selectedLabel }
func (d *ExamSessionLine) SetSelectedLabel(s string)       { d.selectedLabel = s }
func (d *ExamSessionLine) SelectedContent() *string        { return d.selectedContent }
func (d *ExamSessionLine) SetSelectedContent(s *string)    { d.selectedContent = s }
func (d *ExamSessionLine) IsCorrect() bool                 { return d.isCorrect }
func (d *ExamSessionLine) SetIsCorrect(b bool)             { d.isCorrect = b }
func (d *ExamSessionLine) Note() *string                   { return d.note }
func (d *ExamSessionLine) SetNote(s *string)               { d.note = s }
func (d *ExamSessionLine) RptFlg() *string                 { return d.rptFlg }
func (d *ExamSessionLine) SetRptFlg(v *string)             { d.rptFlg = v }
func (d *ExamSessionLine) Kwords() *string                 { return d.kwords }
func (d *ExamSessionLine) SetKwords(v *string)             { d.kwords = v }
func (d *ExamSessionLine) EsessLnStatus() *string          { return d.esessLnStatus }
func (d *ExamSessionLine) SetEsessLnStatus(s *string)      { d.esessLnStatus = s }
func (d *ExamSessionLine) Status() string                  { return d.status }
func (d *ExamSessionLine) SetStatus(s string)              { d.status = s }
func (d *ExamSessionLine) CreateId() *int64                { return d.createId }
func (d *ExamSessionLine) SetCreateId(id *int64)           { d.createId = id }
func (d *ExamSessionLine) CreateDt() mtime.MathTime        { return d.createDt }
func (d *ExamSessionLine) SetCreateDt(t mtime.MathTime)    { d.createDt = t }
func (d *ExamSessionLine) ModifyId() *int64                { return d.modifyId }
func (d *ExamSessionLine) SetModifyId(id *int64)           { d.modifyId = id }
func (d *ExamSessionLine) ModifyDt() mtime.MathTime        { return d.modifyDt }
func (d *ExamSessionLine) SetModifyDt(t mtime.MathTime)    { d.modifyDt = t }
