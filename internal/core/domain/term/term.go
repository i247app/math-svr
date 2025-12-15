package domain

import (
	"time"

	"github.com/google/uuid"
	"math-ai.com/math-ai/internal/driven-adapter/persistence/models"
	"math-ai.com/math-ai/internal/shared/constant/enum"
)

type Term struct {
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

func NewTermDomain() *Term {
	return &Term{}
}

func (s *Term) ID() string {
	return s.id
}

func (s *Term) GenerateID() {
	s.id = uuid.New().String()
}

func (s *Term) SetID(id string) {
	s.id = id
}

func (s *Term) Name() string {
	return s.name
}

func (s *Term) SetName(name string) {
	s.name = name
}

func (s *Term) Description() *string {
	return s.description
}

func (s *Term) SetDescription(description *string) {
	s.description = description
}

func (s *Term) ImageKey() *string {
	return s.imageKey
}

func (s *Term) SetImageKey(imageKey *string) {
	s.imageKey = imageKey
}

func (s *Term) Status() string {
	return s.status
}

func (s *Term) SetStatus(status string) {
	if status == "" {
		status = string(enum.StatusActive)
	}
	s.status = status
}

func (g *Term) DisplayOrder() int8 {
	return g.displayOrder
}

func (g *Term) SetDisplayOrder(displayOrder int8) {
	g.displayOrder = displayOrder
}

func (s *Term) CreateID() *int64 {
	return s.createID
}

func (s *Term) SetCreateID(createID *int64) {
	s.createID = createID
}

func (s *Term) CreatedAt() time.Time {
	return s.createDT
}

func (s *Term) SetCreatedAt(createDT time.Time) {
	s.createDT = createDT
}

func (s *Term) ModifyID() *int64 {
	return s.modifyID
}

func (s *Term) SetModifyID(modifyID *int64) {
	s.modifyID = modifyID
}

func (s *Term) ModifiedAt() time.Time {
	return s.modifyDT
}

func (s *Term) SetModifiedAt(modifyDT time.Time) {
	s.modifyDT = modifyDT
}

func (s *Term) DeletedAt() *time.Time {
	return s.deletedDT
}

func (s *Term) SetDeletedAt(deletedDT *time.Time) {
	s.deletedDT = deletedDT
}

// BuildTermDomainFromModel builds a Term from a model
// Note: name and description come from translation data
func BuildTermDomainFromModel(model *models.TermModel) *Term {
	return &Term{
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
