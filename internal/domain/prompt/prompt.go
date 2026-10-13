package prompt

import (
	"math-ai.com/math-ai/internal/domain/shared/mtime"
)

// Prompt is the STATIC half of one kind of LLM call: the system message,
// sent verbatim. The dynamic half (the user message) is built in code from
// the request. Which kinds exist, and which dimensions each one varies by,
// is fixed by the catalog (catalog.go); a row holds the text for one scope
// of one code.
//
// An admin rewrites systemPrompt in place — there is no history.
// promptVersion starts at 1 and grows by one per rewrite; exam generation
// puts it into the cache tag, so a rewritten prompt stops serving sets
// generated from the old text.
type Prompt struct {
	promptId      int64
	promptCode    Code
	grade         *int    // nil = the code does not vary by grade
	language      *string // nil = the code does not vary by language
	systemPrompt  string
	promptVersion int
	description   *string

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

func NewPrompt() *Prompt { return &Prompt{} }

func (p *Prompt) PromptId() int64              { return p.promptId }
func (p *Prompt) SetPromptId(id int64)         { p.promptId = id }
func (p *Prompt) PromptCode() Code             { return p.promptCode }
func (p *Prompt) SetPromptCode(c Code)         { p.promptCode = c }
func (p *Prompt) Grade() *int                  { return p.grade }
func (p *Prompt) SetGrade(g *int)              { p.grade = g }
func (p *Prompt) Language() *string            { return p.language }
func (p *Prompt) SetLanguage(l *string)        { p.language = l }
func (p *Prompt) SystemPrompt() string         { return p.systemPrompt }
func (p *Prompt) SetSystemPrompt(s string)     { p.systemPrompt = s }
func (p *Prompt) PromptVersion() int           { return p.promptVersion }
func (p *Prompt) SetPromptVersion(v int)       { p.promptVersion = v }
func (p *Prompt) Description() *string         { return p.description }
func (p *Prompt) SetDescription(s *string)     { p.description = s }
func (p *Prompt) PromptStatus() *string        { return p.promptStatus }
func (p *Prompt) SetPromptStatus(s *string)    { p.promptStatus = s }
func (p *Prompt) RptFlg() *string              { return p.rptFlg }
func (p *Prompt) SetRptFlg(s *string)          { p.rptFlg = s }
func (p *Prompt) Kwords() *string              { return p.kwords }
func (p *Prompt) SetKwords(s *string)          { p.kwords = s }
func (p *Prompt) Note() *string                { return p.note }
func (p *Prompt) SetNote(s *string)            { p.note = s }
func (p *Prompt) Status() string               { return p.status }
func (p *Prompt) SetStatus(s string)           { p.status = s }
func (p *Prompt) CreateId() *int64             { return p.createId }
func (p *Prompt) SetCreateId(id *int64)        { p.createId = id }
func (p *Prompt) CreateDt() mtime.MathTime     { return p.createDt }
func (p *Prompt) SetCreateDt(t mtime.MathTime) { p.createDt = t }
func (p *Prompt) ModifyId() *int64             { return p.modifyId }
func (p *Prompt) SetModifyId(id *int64)        { p.modifyId = id }
func (p *Prompt) ModifyDt() mtime.MathTime     { return p.modifyDt }
func (p *Prompt) SetModifyDt(t mtime.MathTime) { p.modifyDt = t }
