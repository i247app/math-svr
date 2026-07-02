package grade

import (
	"math-ai.com/math-ai/internal/domain/shared/mtime"
)

type Grade struct {
	id           int64
	gradeId      int64
	label        string
	description  string
	imageKey     *string
	displayOrder int8
	note         *string
	gradeStatus  *string
	status       string
	createId     *int64
	createDt     mtime.MathTime
	modifyId     *int64
	modifyDt     mtime.MathTime

	// translations is hydrated only by detail reads; list reads keep it
	// nil to avoid an N+1 fetch per grade. Nil and empty are equivalent
	// at the API boundary.
	translations []*GradeTranslation
}

func NewGrade() *Grade {
	return &Grade{}
}

func (g *Grade) Translations() []*GradeTranslation { return g.translations }
func (g *Grade) SetTranslations(t []*GradeTranslation) {
	g.translations = t
}

func (g *Grade) Id() int64 {
	return g.id
}

func (g *Grade) SetId(id int64) {
	g.id = id
}

func (g *Grade) GradeId() int64 {
	return g.gradeId
}

func (g *Grade) SetGradeId(gradeId int64) {
	g.gradeId = gradeId
}

func (g *Grade) Label() string {
	return g.label
}

func (g *Grade) SetLabel(label string) {
	g.label = label
}

func (g *Grade) Description() string {
	return g.description
}

func (g *Grade) SetDescription(description string) {
	g.description = description
}

func (g *Grade) ImageKey() *string {
	return g.imageKey
}

func (g *Grade) SetImageKey(imageKey *string) {
	g.imageKey = imageKey
}

func (g *Grade) DisplayOrder() int8 {
	return g.displayOrder
}

func (g *Grade) SetDisplayOrder(displayOrder int8) {
	g.displayOrder = displayOrder
}

func (g *Grade) Note() *string {
	return g.note
}

func (g *Grade) SetNote(note *string) {
	g.note = note
}

func (g *Grade) GradeStatus() *string {
	return g.gradeStatus
}

func (g *Grade) SetGradeStatus(gradeStatus *string) {
	g.gradeStatus = gradeStatus
}

func (g *Grade) Status() string {
	return g.status
}

func (g *Grade) SetStatus(status string) {
	g.status = status
}

func (g *Grade) CreateId() *int64 {
	return g.createId
}

func (g *Grade) SetCreateId(createId *int64) {
	g.createId = createId
}

func (g *Grade) CreateDt() mtime.MathTime {
	return g.createDt
}

func (g *Grade) SetCreateDt(createDt mtime.MathTime) {
	g.createDt = createDt
}

func (g *Grade) ModifyId() *int64 {
	return g.modifyId
}

func (g *Grade) SetModifyId(modifyId *int64) {
	g.modifyId = modifyId
}

func (g *Grade) ModifyDt() mtime.MathTime {
	return g.modifyDt
}

func (g *Grade) SetModifyDt(modifyDt mtime.MathTime) {
	g.modifyDt = modifyDt
}
