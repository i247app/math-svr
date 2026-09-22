package exam

import (
	"math-ai.com/math-ai/internal/domain/shared/mtime"
)

// AiExam is one set of questions produced by the model, and nothing else.
// It belongs to no user: the req* fields record what was asked for and the
// ai* fields record what came back, which is exactly what makes the row
// reusable. Two children asking for the same thing can be served the same
// row instead of paying for a second generation.
//
// reqExtras is the cache tag — a normalised join of the req* values built
// by one function in the module layer. It is intentionally not unique, so
// several variants can share a tag and a child is not handed the identical
// exam twice.
//
// aiQuestionsJson is stored verbatim as the model returned it, after the
// server has normalised question_grade. Parsing happens on read, in the
// application layer.
type AiExam struct {
	id       int64
	aiExamId int64

	reqExamType string
	reqGrade    int
	reqLevel    *int // mirrors nullable req_level; always NULL until a level rule exists
	reqNumQues  int
	reqSemester *string
	reqProgram  *string
	reqExtras   *string

	aiTitle         *string
	aiShortText     *string
	aiQuestionsJson string

	rptFlg *string

	kwords *string

	note         *string
	aiExamStatus *string
	status       string
	createId     *int64
	createDt     mtime.MathTime
	modifyId     *int64
	modifyDt     mtime.MathTime
}

func NewAiExam() *AiExam { return &AiExam{} }

func (a *AiExam) Id() int64                    { return a.id }
func (a *AiExam) SetId(id int64)               { a.id = id }
func (a *AiExam) AiExamId() int64              { return a.aiExamId }
func (a *AiExam) SetAiExamId(id int64)         { a.aiExamId = id }
func (a *AiExam) ReqExamType() string          { return a.reqExamType }
func (a *AiExam) SetReqExamType(t string)      { a.reqExamType = t }
func (a *AiExam) ReqGrade() int                { return a.reqGrade }
func (a *AiExam) SetReqGrade(g int)            { a.reqGrade = g }
func (a *AiExam) ReqLevel() *int               { return a.reqLevel }
func (a *AiExam) SetReqLevel(l *int)           { a.reqLevel = l }
func (a *AiExam) ReqNumQues() int              { return a.reqNumQues }
func (a *AiExam) SetReqNumQues(n int)          { a.reqNumQues = n }
func (a *AiExam) ReqSemester() *string         { return a.reqSemester }
func (a *AiExam) SetReqSemester(s *string)     { a.reqSemester = s }
func (a *AiExam) ReqProgram() *string          { return a.reqProgram }
func (a *AiExam) SetReqProgram(s *string)      { a.reqProgram = s }
func (a *AiExam) ReqExtras() *string           { return a.reqExtras }
func (a *AiExam) SetReqExtras(s *string)       { a.reqExtras = s }
func (a *AiExam) AiTitle() *string             { return a.aiTitle }
func (a *AiExam) SetAiTitle(s *string)         { a.aiTitle = s }
func (a *AiExam) AiShortText() *string         { return a.aiShortText }
func (a *AiExam) SetAiShortText(s *string)     { a.aiShortText = s }
func (a *AiExam) AiQuestionsJson() string      { return a.aiQuestionsJson }
func (a *AiExam) SetAiQuestionsJson(s string)  { a.aiQuestionsJson = s }
func (a *AiExam) Note() *string                { return a.note }
func (a *AiExam) SetNote(s *string)            { a.note = s }
func (a *AiExam) RptFlg() *string              { return a.rptFlg }
func (a *AiExam) SetRptFlg(v *string)          { a.rptFlg = v }
func (a *AiExam) Kwords() *string              { return a.kwords }
func (a *AiExam) SetKwords(v *string)          { a.kwords = v }
func (a *AiExam) AiExamStatus() *string        { return a.aiExamStatus }
func (a *AiExam) SetAiExamStatus(s *string)    { a.aiExamStatus = s }
func (a *AiExam) Status() string               { return a.status }
func (a *AiExam) SetStatus(s string)           { a.status = s }
func (a *AiExam) CreateId() *int64             { return a.createId }
func (a *AiExam) SetCreateId(id *int64)        { a.createId = id }
func (a *AiExam) CreateDt() mtime.MathTime     { return a.createDt }
func (a *AiExam) SetCreateDt(t mtime.MathTime) { a.createDt = t }
func (a *AiExam) ModifyId() *int64             { return a.modifyId }
func (a *AiExam) SetModifyId(id *int64)        { a.modifyId = id }
func (a *AiExam) ModifyDt() mtime.MathTime     { return a.modifyDt }
func (a *AiExam) SetModifyDt(t mtime.MathTime) { a.modifyDt = t }
