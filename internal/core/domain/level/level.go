package domain

import (
	"time"

	"github.com/google/uuid"
	"math-ai.com/math-ai/internal/driven-adapter/persistence/models"
	"math-ai.com/math-ai/internal/shared/constant/enum"
)

type Level struct {
	id          string
	label       string
	description string
	status      string
	createID    *int64
	createDT    time.Time
	modifyID    *int64
	modifyDT    time.Time
	deletedDT   *time.Time
}

func NewLevelDomain() *Level {
	return &Level{}
}

func (l *Level) ID() string {
	return l.id
}

func (l *Level) GenerateID() {
	l.id = uuid.New().String()
}

func (l *Level) SetID(id string) {
	l.id = id
}

func (l *Level) Label() string {
	return l.label
}

func (l *Level) SetLabel(label string) {
	l.label = label
}

func (l *Level) Description() string {
	return l.description
}

func (l *Level) SetDescription(description string) {
	l.description = description
}

func (l *Level) Status() string {
	return l.status
}

func (l *Level) SetStatus(status string) {
	if status == "" {
		status = string(enum.StatusActive)
	}
	l.status = status
}

func (l *Level) CreateID() *int64 {
	return l.createID
}

func (l *Level) SetCreateID(createID *int64) {
	l.createID = createID
}

func (l *Level) CreatedAt() time.Time {
	return l.createDT
}

func (l *Level) SetCreatedAt(createDT time.Time) {
	l.createDT = createDT
}

func (l *Level) ModifyID() *int64 {
	return l.modifyID
}

func (l *Level) SetModifyID(modifyID *int64) {
	l.modifyID = modifyID
}

func (l *Level) ModifiedAt() time.Time {
	return l.modifyDT
}

func (l *Level) SetModifiedAt(modifyDT time.Time) {
	l.modifyDT = modifyDT
}

func (l *Level) DeletedAt() *time.Time {
	return l.deletedDT
}

func (l *Level) SetDeletedAt(deletedDT *time.Time) {
	l.deletedDT = deletedDT
}

func BuildLevelDomainFromModel(model *models.LevelModel) *Level {
	return &Level{
		id:          model.ID,
		label:       model.Label,
		description: model.Description,
		status:      model.Status,
		createID:    model.CreateID,
		createDT:    model.CreateDT,
		modifyID:    model.ModifyID,
		modifyDT:    model.ModifyDT,
		deletedDT:   model.DeletedDT,
	}
}
