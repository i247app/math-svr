package exam

import (
	"math-ai.com/math-ai/internal/domain/shared/mtime"
)

// ExamPool is one set of questions produced by the model, and nothing else.
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
type ExamPool struct {
	examId int64

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
	// verifiedCount: 0 = not verified by an admin, >0 = verified (see
	// IExamPoolRepository.MarkVerified / ReplaceQuestions).
	verifiedCount int

	rptFlg *string

	kwords *string

	note       *string
	examStatus *string
	status     string
	createId   *int64
	createDt   mtime.MathTime
	modifyId   *int64
	modifyDt   mtime.MathTime
}

func NewExamPool() *ExamPool { return &ExamPool{} }

func (a *ExamPool) ExamId() int64                { return a.examId }
func (a *ExamPool) SetExamId(id int64)           { a.examId = id }
func (a *ExamPool) ReqExamType() string          { return a.reqExamType }
func (a *ExamPool) SetReqExamType(t string)      { a.reqExamType = t }
func (a *ExamPool) ReqGrade() int                { return a.reqGrade }
func (a *ExamPool) SetReqGrade(g int)            { a.reqGrade = g }
func (a *ExamPool) ReqLevel() *int               { return a.reqLevel }
func (a *ExamPool) SetReqLevel(l *int)           { a.reqLevel = l }
func (a *ExamPool) ReqNumQues() int              { return a.reqNumQues }
func (a *ExamPool) SetReqNumQues(n int)          { a.reqNumQues = n }
func (a *ExamPool) ReqSemester() *string         { return a.reqSemester }
func (a *ExamPool) SetReqSemester(s *string)     { a.reqSemester = s }
func (a *ExamPool) ReqProgram() *string          { return a.reqProgram }
func (a *ExamPool) SetReqProgram(s *string)      { a.reqProgram = s }
func (a *ExamPool) ReqExtras() *string           { return a.reqExtras }
func (a *ExamPool) SetReqExtras(s *string)       { a.reqExtras = s }
func (a *ExamPool) AiTitle() *string             { return a.aiTitle }
func (a *ExamPool) SetAiTitle(s *string)         { a.aiTitle = s }
func (a *ExamPool) AiShortText() *string         { return a.aiShortText }
func (a *ExamPool) SetAiShortText(s *string)     { a.aiShortText = s }
func (a *ExamPool) AiQuestionsJson() string      { return a.aiQuestionsJson }
func (a *ExamPool) SetAiQuestionsJson(s string)  { a.aiQuestionsJson = s }
func (a *ExamPool) VerifiedCount() int           { return a.verifiedCount }
func (a *ExamPool) SetVerifiedCount(n int)       { a.verifiedCount = n }
func (a *ExamPool) Note() *string                { return a.note }
func (a *ExamPool) SetNote(s *string)            { a.note = s }
func (a *ExamPool) RptFlg() *string              { return a.rptFlg }
func (a *ExamPool) SetRptFlg(v *string)          { a.rptFlg = v }
func (a *ExamPool) Kwords() *string              { return a.kwords }
func (a *ExamPool) SetKwords(v *string)          { a.kwords = v }
func (a *ExamPool) ExamStatus() *string          { return a.examStatus }
func (a *ExamPool) SetExamStatus(s *string)      { a.examStatus = s }
func (a *ExamPool) Status() string               { return a.status }
func (a *ExamPool) SetStatus(s string)           { a.status = s }
func (a *ExamPool) CreateId() *int64             { return a.createId }
func (a *ExamPool) SetCreateId(id *int64)        { a.createId = id }
func (a *ExamPool) CreateDt() mtime.MathTime     { return a.createDt }
func (a *ExamPool) SetCreateDt(t mtime.MathTime) { a.createDt = t }
func (a *ExamPool) ModifyId() *int64             { return a.modifyId }
func (a *ExamPool) SetModifyId(id *int64)        { a.modifyId = id }
func (a *ExamPool) ModifyDt() mtime.MathTime     { return a.modifyDt }
func (a *ExamPool) SetModifyDt(t mtime.MathTime) { a.modifyDt = t }
