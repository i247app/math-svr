package program

import (
	"math-ai.com/math-ai/internal/domain/shared/mtime"
)

type Program struct {
	id            int64
	programId     int64
	label         string
	description   string
	imageKey      *string
	displayOrder  int8
	note          *string
	programStatus *string
	status        string
	createId      *int64
	createDt      mtime.MathTime
	modifyId      *int64
	modifyDt      mtime.MathTime
}

func NewProgram() *Program {
	return &Program{}
}

func (p *Program) Id() int64 {
	return p.id
}

func (p *Program) SetId(id int64) {
	p.id = id
}

func (p *Program) ProgramId() int64 {
	return p.programId
}

func (p *Program) SetProgramId(programId int64) {
	p.programId = programId
}

func (p *Program) Label() string {
	return p.label
}

func (p *Program) SetLabel(label string) {
	p.label = label
}

func (p *Program) Description() string {
	return p.description
}

func (p *Program) SetDescription(description string) {
	p.description = description
}

func (p *Program) ImageKey() *string {
	return p.imageKey
}

func (p *Program) SetImageKey(imageKey *string) {
	p.imageKey = imageKey
}

func (p *Program) DisplayOrder() int8 {
	return p.displayOrder
}

func (p *Program) SetDisplayOrder(displayOrder int8) {
	p.displayOrder = displayOrder
}

func (p *Program) Note() *string {
	return p.note
}

func (p *Program) SetNote(note *string) {
	p.note = note
}

func (p *Program) ProgramStatus() *string {
	return p.programStatus
}

func (p *Program) SetProgramStatus(programStatus *string) {
	p.programStatus = programStatus
}

func (p *Program) Status() string {
	return p.status
}

func (p *Program) SetStatus(status string) {
	p.status = status
}

func (p *Program) CreateId() *int64 {
	return p.createId
}

func (p *Program) SetCreateId(createId *int64) {
	p.createId = createId
}

func (p *Program) CreateDt() mtime.MathTime {
	return p.createDt
}

func (p *Program) SetCreateDt(createDt mtime.MathTime) {
	p.createDt = createDt
}

func (p *Program) ModifyId() *int64 {
	return p.modifyId
}

func (p *Program) SetModifyId(modifyId *int64) {
	p.modifyId = modifyId
}

func (p *Program) ModifyDt() mtime.MathTime {
	return p.modifyDt
}

func (p *Program) SetModifyDt(modifyDt mtime.MathTime) {
	p.modifyDt = modifyDt
}
