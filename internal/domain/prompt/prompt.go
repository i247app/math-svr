package prompt

import (
	"math-ai.com/math-ai/internal/domain/shared/audit"
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
	audit.Audit

	promptId      int64
	promptCode    Code
	grade         *int    // nil = the code does not vary by grade
	language      *string // nil = the code does not vary by language
	systemPrompt  string
	promptVersion int
	description   *string

	promptStatus *string
}

func NewPrompt() *Prompt { return &Prompt{} }

func (p *Prompt) PromptId() int64      { return p.promptId }
func (p *Prompt) SetPromptId(id int64) { p.promptId = id }

func (p *Prompt) PromptCode() Code     { return p.promptCode }
func (p *Prompt) SetPromptCode(c Code) { p.promptCode = c }

func (p *Prompt) Grade() *int     { return p.grade }
func (p *Prompt) SetGrade(g *int) { p.grade = g }

func (p *Prompt) Language() *string     { return p.language }
func (p *Prompt) SetLanguage(l *string) { p.language = l }

func (p *Prompt) SystemPrompt() string     { return p.systemPrompt }
func (p *Prompt) SetSystemPrompt(s string) { p.systemPrompt = s }

func (p *Prompt) PromptVersion() int     { return p.promptVersion }
func (p *Prompt) SetPromptVersion(v int) { p.promptVersion = v }

func (p *Prompt) Description() *string     { return p.description }
func (p *Prompt) SetDescription(s *string) { p.description = s }

func (p *Prompt) PromptStatus() *string     { return p.promptStatus }
func (p *Prompt) SetPromptStatus(s *string) { p.promptStatus = s }
