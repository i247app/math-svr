package program

import (
	"math-ai.com/math-ai/internal/domain/shared/audit"
)

type Program struct {
	audit.Audit

	programId     int64
	label         string
	description   string
	imageKey      *string
	displayOrder  int8
	programStatus *string
}

func NewProgram() *Program {
	return &Program{}
}

func (p *Program) ProgramId() int64             { return p.programId }
func (p *Program) SetProgramId(programId int64) { p.programId = programId }

func (p *Program) Label() string         { return p.label }
func (p *Program) SetLabel(label string) { p.label = label }

func (p *Program) Description() string               { return p.description }
func (p *Program) SetDescription(description string) { p.description = description }

func (p *Program) ImageKey() *string            { return p.imageKey }
func (p *Program) SetImageKey(imageKey *string) { p.imageKey = imageKey }

func (p *Program) DisplayOrder() int8                { return p.displayOrder }
func (p *Program) SetDisplayOrder(displayOrder int8) { p.displayOrder = displayOrder }

func (p *Program) ProgramStatus() *string                 { return p.programStatus }
func (p *Program) SetProgramStatus(programStatus *string) { p.programStatus = programStatus }
