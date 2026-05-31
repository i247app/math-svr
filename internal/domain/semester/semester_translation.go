package semester

import (
	mtime "math-ai.com/math-ai/internal/domain/shared/time"
)

// SemesterTranslation is a per-language override of a Semester's display
// fields. The DB enforces UNIQUE(semester_id, language); validation at
// the application layer rejects duplicates earlier so the failure
// surfaces as a clean status code rather than a MySQL constraint.
type SemesterTranslation struct {
	id                    int64
	semesterTranslationId int64
	semesterId            int64
	language              string
	name                  string
	description           string
	note                  *string
	stStatus              *string
	status                string
	createId              *int64
	createDt              mtime.MathTime
	modifyId              *int64
	modifyDt              mtime.MathTime
}

func NewSemesterTranslation() *SemesterTranslation {
	return &SemesterTranslation{}
}

func (t *SemesterTranslation) Id() int64                         { return t.id }
func (t *SemesterTranslation) SetId(id int64)                    { t.id = id }
func (t *SemesterTranslation) SemesterTranslationId() int64      { return t.semesterTranslationId }
func (t *SemesterTranslation) SetSemesterTranslationId(id int64) { t.semesterTranslationId = id }
func (t *SemesterTranslation) SemesterId() int64                 { return t.semesterId }
func (t *SemesterTranslation) SetSemesterId(id int64)            { t.semesterId = id }
func (t *SemesterTranslation) Language() string                  { return t.language }
func (t *SemesterTranslation) SetLanguage(l string)              { t.language = l }
func (t *SemesterTranslation) Name() string                      { return t.name }
func (t *SemesterTranslation) SetName(n string)                  { t.name = n }
func (t *SemesterTranslation) Description() string               { return t.description }
func (t *SemesterTranslation) SetDescription(d string)           { t.description = d }
func (t *SemesterTranslation) Note() *string                     { return t.note }
func (t *SemesterTranslation) SetNote(n *string)                 { t.note = n }
func (t *SemesterTranslation) StStatus() *string                 { return t.stStatus }
func (t *SemesterTranslation) SetStStatus(s *string)             { t.stStatus = s }
func (t *SemesterTranslation) Status() string                    { return t.status }
func (t *SemesterTranslation) SetStatus(s string)                { t.status = s }
func (t *SemesterTranslation) CreateId() *int64                  { return t.createId }
func (t *SemesterTranslation) SetCreateId(id *int64)             { t.createId = id }
func (t *SemesterTranslation) CreateDt() mtime.MathTime          { return t.createDt }
func (t *SemesterTranslation) SetCreateDt(time mtime.MathTime)   { t.createDt = time }
func (t *SemesterTranslation) ModifyId() *int64                  { return t.modifyId }
func (t *SemesterTranslation) SetModifyId(id *int64)             { t.modifyId = id }
func (t *SemesterTranslation) ModifyDt() mtime.MathTime          { return t.modifyDt }
func (t *SemesterTranslation) SetModifyDt(time mtime.MathTime)   { t.modifyDt = time }
