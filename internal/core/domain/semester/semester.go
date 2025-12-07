package domain

import (
	"github.com/google/uuid"
	"math-ai.com/math-ai/internal/driven-adapter/persistence/models"
)

type Semester struct {
	id          string
	name        string // From translation
	description string // From translation
}

func NewSemester() *Semester {
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

func (s *Semester) Description() string {
	return s.description
}

func (s *Semester) SetDescription(description string) {
	s.description = description
}

func BuildSemesterFromModel(model *models.SemesterModel) *Semester {
	return &Semester{
		id: model.ID,
	}
}
