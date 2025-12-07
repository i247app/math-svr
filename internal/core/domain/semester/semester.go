package domain

import (
	"time"

	"github.com/google/uuid"
	"math-ai.com/math-ai/internal/driven-adapter/persistence/models"
	"math-ai.com/math-ai/internal/shared/constant/enum"
)

type Semester struct {
	id           string
	name         string
	description  *string
	imageKey     *string
	status       string
	displayOrder int8
	createID     *int64
	createDT     time.Time
	modifyID     *int64
	modifyDT     time.Time
	deletedDT    *time.Time
}

func NewSemesterDomain() *Semester {
	return &Semester{}
}

func (s *Semester) ID() string {
	return s.id
}

func (s *Semester) GenerateID() {
	s.id = uuid.New().String()
}

func (s *Semester) SetID(id string) {
	s.id = id
}

func (s *Semester) Name() string {
	return s.name
}

func (s *Semester) SetName(name string) {
	s.name = name
}

func (s *Semester) Description() *string {
	return s.description
}

func (s *Semester) SetDescription(description *string) {
	s.description = description
}

func (s *Semester) ImageKey() *string {
	return s.imageKey
}

func (s *Semester) SetImageKey(imageKey *string) {
	s.imageKey = imageKey
}

func (s *Semester) Status() string {
	return s.status
}

func (s *Semester) SetStatus(status string) {
	if status == "" {
		status = string(enum.StatusActive)
	}
	s.status = status
}

func (g *Semester) DisplayOrder() int8 {
	return g.displayOrder
}

func (g *Semester) SetDisplayOrder(displayOrder int8) {
	g.displayOrder = displayOrder
}

func (s *Semester) CreateID() *int64 {
	return s.createID
}

func (s *Semester) SetCreateID(createID *int64) {
	s.createID = createID
}

func (s *Semester) CreatedAt() time.Time {
	return s.createDT
}

func (s *Semester) SetCreatedAt(createDT time.Time) {
	s.createDT = createDT
}

func (s *Semester) ModifyID() *int64 {
	return s.modifyID
}

func (s *Semester) SetModifyID(modifyID *int64) {
	s.modifyID = modifyID
}

func (s *Semester) ModifiedAt() time.Time {
	return s.modifyDT
}

func (s *Semester) SetModifiedAt(modifyDT time.Time) {
	s.modifyDT = modifyDT
}

func (s *Semester) DeletedAt() *time.Time {
	return s.deletedDT
}

func (s *Semester) SetDeletedAt(deletedDT *time.Time) {
	s.deletedDT = deletedDT
}

// BuildSemesterDomainFromModel builds a Semester from a model
// Note: name and description come from translation data
func BuildSemesterDomainFromModel(model *models.SemesterModel) *Semester {
	return &Semester{
		id:           model.ID,
		name:         model.Name,
		description:  model.Description,
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
