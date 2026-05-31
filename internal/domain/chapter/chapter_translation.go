package chapter

import (
	mtime "math-ai.com/math-ai/internal/domain/shared/time"
)

// ChapterTranslation is a per-language override of a Chapter's display
// fields. One row per (chapter_id, language) — the uniqueness invariant
// is enforced at the application layer because the migration does not
// declare a composite UNIQUE.
type ChapterTranslation struct {
	id                   int64
	chapterTranslationId int64
	chapterId            int64
	language             string
	label                string
	description          string
	note                 *string
	ctStatus             *string
	status               string
	createId             *int64
	createDt             mtime.MathTime
	modifyId             *int64
	modifyDt             mtime.MathTime
}

func NewChapterTranslation() *ChapterTranslation {
	return &ChapterTranslation{}
}

func (t *ChapterTranslation) Id() int64                   { return t.id }
func (t *ChapterTranslation) SetId(id int64)              { t.id = id }
func (t *ChapterTranslation) ChapterTranslationId() int64 { return t.chapterTranslationId }
func (t *ChapterTranslation) SetChapterTranslationId(id int64) {
	t.chapterTranslationId = id
}
func (t *ChapterTranslation) ChapterId() int64         { return t.chapterId }
func (t *ChapterTranslation) SetChapterId(id int64)    { t.chapterId = id }
func (t *ChapterTranslation) Language() string         { return t.language }
func (t *ChapterTranslation) SetLanguage(l string)     { t.language = l }
func (t *ChapterTranslation) Label() string            { return t.label }
func (t *ChapterTranslation) SetLabel(l string)        { t.label = l }
func (t *ChapterTranslation) Description() string      { return t.description }
func (t *ChapterTranslation) SetDescription(d string)  { t.description = d }
func (t *ChapterTranslation) Note() *string            { return t.note }
func (t *ChapterTranslation) SetNote(n *string)        { t.note = n }
func (t *ChapterTranslation) CtStatus() *string        { return t.ctStatus }
func (t *ChapterTranslation) SetCtStatus(s *string)    { t.ctStatus = s }
func (t *ChapterTranslation) Status() string           { return t.status }
func (t *ChapterTranslation) SetStatus(s string)       { t.status = s }
func (t *ChapterTranslation) CreateId() *int64         { return t.createId }
func (t *ChapterTranslation) SetCreateId(id *int64)    { t.createId = id }
func (t *ChapterTranslation) CreateDt() mtime.MathTime { return t.createDt }
func (t *ChapterTranslation) SetCreateDt(time mtime.MathTime) {
	t.createDt = time
}
func (t *ChapterTranslation) ModifyId() *int64         { return t.modifyId }
func (t *ChapterTranslation) SetModifyId(id *int64)    { t.modifyId = id }
func (t *ChapterTranslation) ModifyDt() mtime.MathTime { return t.modifyDt }
func (t *ChapterTranslation) SetModifyDt(time mtime.MathTime) {
	t.modifyDt = time
}
