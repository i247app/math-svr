package program

import (
	mtime "math-ai.com/math-ai/internal/domain/shared/time"
)

// ProgramTranslation is a per-language override of a Program's display
// fields. Unlike grade/semester, the migration does NOT declare a
// composite UNIQUE on (program_id, language), so the (one-row-per-
// language) invariant is enforced purely at the application layer.
type ProgramTranslation struct {
	id                   int64
	programTranslationId int64
	programId            int64
	language             string
	label                string
	description          string
	note                 *string
	ptStatus             *string
	status               string
	createId             *int64
	createDt             mtime.MathTime
	modifyId             *int64
	modifyDt             mtime.MathTime
}

func NewProgramTranslation() *ProgramTranslation {
	return &ProgramTranslation{}
}

func (t *ProgramTranslation) Id() int64                        { return t.id }
func (t *ProgramTranslation) SetId(id int64)                   { t.id = id }
func (t *ProgramTranslation) ProgramTranslationId() int64      { return t.programTranslationId }
func (t *ProgramTranslation) SetProgramTranslationId(id int64) { t.programTranslationId = id }
func (t *ProgramTranslation) ProgramId() int64                 { return t.programId }
func (t *ProgramTranslation) SetProgramId(id int64)            { t.programId = id }
func (t *ProgramTranslation) Language() string                 { return t.language }
func (t *ProgramTranslation) SetLanguage(l string)             { t.language = l }
func (t *ProgramTranslation) Label() string                    { return t.label }
func (t *ProgramTranslation) SetLabel(l string)                { t.label = l }
func (t *ProgramTranslation) Description() string              { return t.description }
func (t *ProgramTranslation) SetDescription(d string)          { t.description = d }
func (t *ProgramTranslation) Note() *string                    { return t.note }
func (t *ProgramTranslation) SetNote(n *string)                { t.note = n }
func (t *ProgramTranslation) PtStatus() *string                { return t.ptStatus }
func (t *ProgramTranslation) SetPtStatus(s *string)            { t.ptStatus = s }
func (t *ProgramTranslation) Status() string                   { return t.status }
func (t *ProgramTranslation) SetStatus(s string)               { t.status = s }
func (t *ProgramTranslation) CreateId() *int64                 { return t.createId }
func (t *ProgramTranslation) SetCreateId(id *int64)            { t.createId = id }
func (t *ProgramTranslation) CreateDt() mtime.MathTime         { return t.createDt }
func (t *ProgramTranslation) SetCreateDt(time mtime.MathTime)  { t.createDt = time }
func (t *ProgramTranslation) ModifyId() *int64                 { return t.modifyId }
func (t *ProgramTranslation) SetModifyId(id *int64)            { t.modifyId = id }
func (t *ProgramTranslation) ModifyDt() mtime.MathTime         { return t.modifyDt }
func (t *ProgramTranslation) SetModifyDt(time mtime.MathTime)  { t.modifyDt = time }
