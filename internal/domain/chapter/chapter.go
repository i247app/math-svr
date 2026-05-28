package chapter

import (
	mtime "math-ai.com/math-ai/internal/domain/shared/time"
)

// Chapter is a curriculum unit owned by a (program, grade) pair. The base
// row carries the canonical label/description in the seeding language;
// per-language overrides live in ChapterTranslation rows joined on
// chapter_id. The DB column is misspelled `discription`; the domain
// surfaces a clean Description() so the typo never leaks past the repo.
type Chapter struct {
	id            int64
	chapterId     string
	programId     string
	gradeId       string
	semesterId    string
	label         string
	description   string
	displayOrder  int8
	note          *string
	chapterStatus *string
	status        string
	createId      *string
	createDt      mtime.MathTime
	modifyId      *string
	modifyDt      mtime.MathTime

	// translations is hydrated only by detail reads; list reads keep it
	// nil to avoid an N+1 fetch per chapter. Nil and empty are equivalent
	// at the API boundary.
	translations []*ChapterTranslation
}

func NewChapter() *Chapter {
	return &Chapter{}
}

func (c *Chapter) Id() int64                  { return c.id }
func (c *Chapter) SetId(id int64)             { c.id = id }
func (c *Chapter) ChapterId() string          { return c.chapterId }
func (c *Chapter) SetChapterId(id string)     { c.chapterId = id }
func (c *Chapter) ProgramId() string          { return c.programId }
func (c *Chapter) SetProgramId(id string)     { c.programId = id }
func (c *Chapter) GradeId() string            { return c.gradeId }
func (c *Chapter) SetGradeId(id string)       { c.gradeId = id }
func (c *Chapter) SemesterId() string         { return c.semesterId }
func (c *Chapter) SetSemesterId(id string)    { c.semesterId = id }
func (c *Chapter) Label() string              { return c.label }
func (c *Chapter) SetLabel(l string)          { c.label = l }
func (c *Chapter) Description() string        { return c.description }
func (c *Chapter) SetDescription(d string)    { c.description = d }
func (c *Chapter) DisplayOrder() int8         { return c.displayOrder }
func (c *Chapter) SetDisplayOrder(o int8)     { c.displayOrder = o }
func (c *Chapter) Note() *string              { return c.note }
func (c *Chapter) SetNote(n *string)          { c.note = n }
func (c *Chapter) ChapterStatus() *string     { return c.chapterStatus }
func (c *Chapter) SetChapterStatus(s *string) { c.chapterStatus = s }
func (c *Chapter) Status() string             { return c.status }
func (c *Chapter) SetStatus(s string)         { c.status = s }
func (c *Chapter) CreateId() *string          { return c.createId }
func (c *Chapter) SetCreateId(id *string)     { c.createId = id }
func (c *Chapter) CreateDt() mtime.MathTime   { return c.createDt }
func (c *Chapter) SetCreateDt(t mtime.MathTime) {
	c.createDt = t
}
func (c *Chapter) ModifyId() *string          { return c.modifyId }
func (c *Chapter) SetModifyId(id *string)     { c.modifyId = id }
func (c *Chapter) ModifyDt() mtime.MathTime   { return c.modifyDt }
func (c *Chapter) SetModifyDt(t mtime.MathTime) {
	c.modifyDt = t
}

func (c *Chapter) Translations() []*ChapterTranslation { return c.translations }
func (c *Chapter) SetTranslations(t []*ChapterTranslation) {
	c.translations = t
}
