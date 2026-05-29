package grade

import (
	mtime "math-ai.com/math-ai/internal/domain/shared/time"
)

// GradeTranslation is a per-language override of a Grade's display
// fields. The DB enforces UNIQUE(grade_id, language); validation at
// the application layer rejects duplicates earlier so the failure
// surfaces as a clean status code rather than a MySQL constraint.
type GradeTranslation struct {
	id                 int64
	gradeTranslationId string
	gradeId            string
	language           string
	label              string
	description        string
	note               *string
	gtStatus           *string
	status             string
	createId           *string
	createDt           mtime.MathTime
	modifyId           *string
	modifyDt           mtime.MathTime
}

func NewGradeTranslation() *GradeTranslation {
	return &GradeTranslation{}
}

func (t *GradeTranslation) Id() int64                       { return t.id }
func (t *GradeTranslation) SetId(id int64)                  { t.id = id }
func (t *GradeTranslation) GradeTranslationId() string      { return t.gradeTranslationId }
func (t *GradeTranslation) SetGradeTranslationId(id string) { t.gradeTranslationId = id }
func (t *GradeTranslation) GradeId() string                 { return t.gradeId }
func (t *GradeTranslation) SetGradeId(id string)            { t.gradeId = id }
func (t *GradeTranslation) Language() string                { return t.language }
func (t *GradeTranslation) SetLanguage(l string)            { t.language = l }
func (t *GradeTranslation) Label() string                   { return t.label }
func (t *GradeTranslation) SetLabel(l string)               { t.label = l }
func (t *GradeTranslation) Description() string             { return t.description }
func (t *GradeTranslation) SetDescription(d string)         { t.description = d }
func (t *GradeTranslation) Note() *string                   { return t.note }
func (t *GradeTranslation) SetNote(n *string)               { t.note = n }
func (t *GradeTranslation) GtStatus() *string               { return t.gtStatus }
func (t *GradeTranslation) SetGtStatus(s *string)           { t.gtStatus = s }
func (t *GradeTranslation) Status() string                  { return t.status }
func (t *GradeTranslation) SetStatus(s string)              { t.status = s }
func (t *GradeTranslation) CreateId() *string               { return t.createId }
func (t *GradeTranslation) SetCreateId(id *string)          { t.createId = id }
func (t *GradeTranslation) CreateDt() mtime.MathTime        { return t.createDt }
func (t *GradeTranslation) SetCreateDt(time mtime.MathTime) { t.createDt = time }
func (t *GradeTranslation) ModifyId() *string               { return t.modifyId }
func (t *GradeTranslation) SetModifyId(id *string)          { t.modifyId = id }
func (t *GradeTranslation) ModifyDt() mtime.MathTime        { return t.modifyDt }
func (t *GradeTranslation) SetModifyDt(time mtime.MathTime) { t.modifyDt = time }
