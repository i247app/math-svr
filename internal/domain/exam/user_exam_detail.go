package exam

import (
	"math-ai.com/math-ai/internal/domain/shared/mtime"
)

// UserExamDetail is one question the child ACTUALLY ANSWERED. A skipped
// question has no row at all, which is why "how many did they answer" is a
// row count here and never a NULL check.
//
// Rows accumulate under userExamId for the life of the journey; grouping
// by userAiExamId splits them back into individual sittings. reqExamType
// is copied from the sitting so the journey's ASSESSMENT rows and its
// PRACTICE rows — which share userExamId — can be told apart without a
// join: every per-journey read filters on the pair.
//
// The question fields are a snapshot taken at submit time out of the
// AiExam's questions JSON. That freezes what the child was actually shown
// — the AiExam row is shared cache and may be superseded — and keeps a
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
type UserExamDetail struct {
	id               int64
	userExamDetailId int64
	userAiExamId     int64
	userExamId       int64
	aiExamId         int64
	reqExamType      string

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

	note         *string
	detailStatus *string
	status       string
	createId     *int64
	createDt     mtime.MathTime
	modifyId     *int64
	modifyDt     mtime.MathTime
}

func NewUserExamDetail() *UserExamDetail { return &UserExamDetail{} }

func (d *UserExamDetail) Id() int64                       { return d.id }
func (d *UserExamDetail) SetId(id int64)                  { d.id = id }
func (d *UserExamDetail) UserExamDetailId() int64         { return d.userExamDetailId }
func (d *UserExamDetail) SetUserExamDetailId(id int64)    { d.userExamDetailId = id }
func (d *UserExamDetail) UserAiExamId() int64             { return d.userAiExamId }
func (d *UserExamDetail) SetUserAiExamId(id int64)        { d.userAiExamId = id }
func (d *UserExamDetail) UserExamId() int64               { return d.userExamId }
func (d *UserExamDetail) SetUserExamId(id int64)          { d.userExamId = id }
func (d *UserExamDetail) AiExamId() int64                 { return d.aiExamId }
func (d *UserExamDetail) ReqExamType() string             { return d.reqExamType }
func (d *UserExamDetail) SetReqExamType(t string)         { d.reqExamType = t }
func (d *UserExamDetail) SetAiExamId(id int64)            { d.aiExamId = id }
func (d *UserExamDetail) QuestionNumber() int             { return d.questionNumber }
func (d *UserExamDetail) SetQuestionNumber(n int)         { d.questionNumber = n }
func (d *UserExamDetail) QuestionType() *string           { return d.questionType }
func (d *UserExamDetail) SetQuestionType(s *string)       { d.questionType = s }
func (d *UserExamDetail) QuestionName() *string           { return d.questionName }
func (d *UserExamDetail) SetQuestionName(s *string)       { d.questionName = s }
func (d *UserExamDetail) QuestionTopic() *string          { return d.questionTopic }
func (d *UserExamDetail) SetQuestionTopic(s *string)      { d.questionTopic = s }
func (d *UserExamDetail) QuestionGrade() *int             { return d.questionGrade }
func (d *UserExamDetail) SetQuestionGrade(g *int)         { d.questionGrade = g }
func (d *UserExamDetail) QuestionLevel() *int             { return d.questionLevel }
func (d *UserExamDetail) SetQuestionLevel(l *int)         { d.questionLevel = l }
func (d *UserExamDetail) RightAnswerLabel() *string       { return d.rightAnswerLabel }
func (d *UserExamDetail) SetRightAnswerLabel(s *string)   { d.rightAnswerLabel = s }
func (d *UserExamDetail) RightAnswerContent() *string     { return d.rightAnswerContent }
func (d *UserExamDetail) SetRightAnswerContent(s *string) { d.rightAnswerContent = s }
func (d *UserExamDetail) SelectedLabel() string           { return d.selectedLabel }
func (d *UserExamDetail) SetSelectedLabel(s string)       { d.selectedLabel = s }
func (d *UserExamDetail) SelectedContent() *string        { return d.selectedContent }
func (d *UserExamDetail) SetSelectedContent(s *string)    { d.selectedContent = s }
func (d *UserExamDetail) IsCorrect() bool                 { return d.isCorrect }
func (d *UserExamDetail) SetIsCorrect(b bool)             { d.isCorrect = b }
func (d *UserExamDetail) Note() *string                   { return d.note }
func (d *UserExamDetail) SetNote(s *string)               { d.note = s }
func (d *UserExamDetail) RptFlg() *string                 { return d.rptFlg }
func (d *UserExamDetail) SetRptFlg(v *string)             { d.rptFlg = v }
func (d *UserExamDetail) Kwords() *string                 { return d.kwords }
func (d *UserExamDetail) SetKwords(v *string)             { d.kwords = v }
func (d *UserExamDetail) DetailStatus() *string           { return d.detailStatus }
func (d *UserExamDetail) SetDetailStatus(s *string)       { d.detailStatus = s }
func (d *UserExamDetail) Status() string                  { return d.status }
func (d *UserExamDetail) SetStatus(s string)              { d.status = s }
func (d *UserExamDetail) CreateId() *int64                { return d.createId }
func (d *UserExamDetail) SetCreateId(id *int64)           { d.createId = id }
func (d *UserExamDetail) CreateDt() mtime.MathTime        { return d.createDt }
func (d *UserExamDetail) SetCreateDt(t mtime.MathTime)    { d.createDt = t }
func (d *UserExamDetail) ModifyId() *int64                { return d.modifyId }
func (d *UserExamDetail) SetModifyId(id *int64)           { d.modifyId = id }
func (d *UserExamDetail) ModifyDt() mtime.MathTime        { return d.modifyDt }
func (d *UserExamDetail) SetModifyDt(t mtime.MathTime)    { d.modifyDt = t }
