package grade

import (
	"math-ai.com/math-ai/internal/domain/shared/audit"
)

type Grade struct {
	audit.Audit

	gradeId      int64
	label        string
	description  string
	imageKey     *string
	displayOrder int8
	gradeStatus  *string
}

func NewGrade() *Grade {
	return &Grade{}
}

func (g *Grade) GradeId() int64           { return g.gradeId }
func (g *Grade) SetGradeId(gradeId int64) { g.gradeId = gradeId }

func (g *Grade) Label() string         { return g.label }
func (g *Grade) SetLabel(label string) { g.label = label }

func (g *Grade) Description() string               { return g.description }
func (g *Grade) SetDescription(description string) { g.description = description }

func (g *Grade) ImageKey() *string            { return g.imageKey }
func (g *Grade) SetImageKey(imageKey *string) { g.imageKey = imageKey }

func (g *Grade) DisplayOrder() int8                { return g.displayOrder }
func (g *Grade) SetDisplayOrder(displayOrder int8) { g.displayOrder = displayOrder }

func (g *Grade) GradeStatus() *string               { return g.gradeStatus }
func (g *Grade) SetGradeStatus(gradeStatus *string) { g.gradeStatus = gradeStatus }
