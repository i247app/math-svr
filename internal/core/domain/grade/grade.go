package domain

import (
	"time"

	"github.com/google/uuid"
	"math-ai.com/math-ai/internal/driven-adapter/persistence/models"
	"math-ai.com/math-ai/internal/shared/constant/enum"
)

type Grade struct {
	id           string
	label        string
	description  string
	imageKey     *string
	status       string
	displayOrder int8
	createID     *int64
	createDT     time.Time
	modifyID     *int64
	modifyDT     time.Time
	deletedDT    *time.Time
}

func NewGradeDomain() *Grade {
	return &Grade{}
}

func (g *Grade) ID() string {
	return g.id
}

func (g *Grade) GenerateID() {
	g.id = uuid.New().String()
}

func (g *Grade) SetID(id string) {
	g.id = id
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

func (g *Grade) Status() string {
	return g.status
}

func (g *Grade) SetStatus(status string) {
	if status == "" {
		status = string(enum.StatusActive)
	}
	g.status = status
}

func (g *Grade) DisplayOrder() int8 {
	return g.displayOrder
}

func (g *Grade) SetDisplayOrder(displayOrder int8) {
	g.displayOrder = displayOrder
}

func (g *Grade) CreateID() *int64 {
	return g.createID
}

func (g *Grade) SetCreateID(createID *int64) {
	g.createID = createID
}

func (g *Grade) CreatedAt() time.Time {
	return g.createDT
}

func (g *Grade) SetCreatedAt(createDT time.Time) {
	g.createDT = createDT
}

func (g *Grade) ModifyID() *int64 {
	return g.modifyID
}

func (g *Grade) SetModifyID(modifyID *int64) {
	g.modifyID = modifyID
}

func (g *Grade) ModifiedAt() time.Time {
	return g.modifyDT
}

func (g *Grade) SetModifiedAt(modifyDT time.Time) {
	g.modifyDT = modifyDT
}

func (g *Grade) DeletedAt() *time.Time {
	return g.deletedDT
}

func (g *Grade) SetDeletedAt(deletedDT *time.Time) {
	g.deletedDT = deletedDT
}

// BuildGradeDomainFromModel builds a Grade from a model
// Note: label and description must be set separately from translation data
func BuildGradeDomainFromModel(model *models.GradeModel) *Grade {
	return &Grade{
		id:           model.ID,
		imageKey:     model.ImageKey,
		status:       model.Status,
		displayOrder: model.DisplayOrder,
		createID:     model.CreateID,
		createDT:     model.CreateDT,
		modifyID:     model.ModifyID,
		modifyDT:     model.ModifyDT,
		deletedDT:    model.DeletedDT,
	}
}
