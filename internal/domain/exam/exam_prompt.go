package exam

import (
	"math-ai.com/math-ai/internal/domain/shared/mtime"
)

// ExamPrompt is the system prompt an exam generation sends for one grade.
// There is exactly one row per grade (0 = Mẫu giáo … 5); an admin rewrites
// it in place, so there is no history.
//
// systemPrompt is used verbatim — nothing in it is filled in at runtime.
// promptVersion starts at 1 and grows by one per rewrite; it is part of the
// exam cache tag, which is what stops a rewritten prompt from serving sets
// generated from the old text.
type ExamPrompt struct {
	promptId      int64
	grade         int
	systemPrompt  string
	promptVersion int

	promptStatus *string
	rptFlg       *string
	kwords       *string
	note         *string
	status       string
	createId     *int64
	createDt     mtime.MathTime
	modifyId     *int64
	modifyDt     mtime.MathTime
}

func NewExamPrompt() *ExamPrompt { return &ExamPrompt{} }

func (p *ExamPrompt) PromptId() int64              { return p.promptId }
func (p *ExamPrompt) SetPromptId(id int64)         { p.promptId = id }
func (p *ExamPrompt) Grade() int                   { return p.grade }
func (p *ExamPrompt) SetGrade(g int)               { p.grade = g }
func (p *ExamPrompt) SystemPrompt() string         { return p.systemPrompt }
func (p *ExamPrompt) SetSystemPrompt(s string)     { p.systemPrompt = s }
func (p *ExamPrompt) PromptVersion() int           { return p.promptVersion }
func (p *ExamPrompt) SetPromptVersion(v int)       { p.promptVersion = v }
func (p *ExamPrompt) PromptStatus() *string        { return p.promptStatus }
func (p *ExamPrompt) SetPromptStatus(s *string)    { p.promptStatus = s }
func (p *ExamPrompt) RptFlg() *string              { return p.rptFlg }
func (p *ExamPrompt) SetRptFlg(s *string)          { p.rptFlg = s }
func (p *ExamPrompt) Kwords() *string              { return p.kwords }
func (p *ExamPrompt) SetKwords(s *string)          { p.kwords = s }
func (p *ExamPrompt) Note() *string                { return p.note }
func (p *ExamPrompt) SetNote(s *string)            { p.note = s }
func (p *ExamPrompt) Status() string               { return p.status }
func (p *ExamPrompt) SetStatus(s string)           { p.status = s }
func (p *ExamPrompt) CreateId() *int64             { return p.createId }
func (p *ExamPrompt) SetCreateId(id *int64)        { p.createId = id }
func (p *ExamPrompt) CreateDt() mtime.MathTime     { return p.createDt }
func (p *ExamPrompt) SetCreateDt(t mtime.MathTime) { p.createDt = t }
func (p *ExamPrompt) ModifyId() *int64             { return p.modifyId }
func (p *ExamPrompt) SetModifyId(id *int64)        { p.modifyId = id }
func (p *ExamPrompt) ModifyDt() mtime.MathTime     { return p.modifyDt }
func (p *ExamPrompt) SetModifyDt(t mtime.MathTime) { p.modifyDt = t }
