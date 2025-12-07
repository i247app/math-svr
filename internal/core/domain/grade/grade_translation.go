package domain

import (
	"github.com/google/uuid"
	"math-ai.com/math-ai/internal/driven-adapter/persistence/models"
)

type GradeTranslation struct {
	id          string
	gradeID     string
	language    string
	label       string
	description string
}

func NewGradeTranslation() *GradeTranslation {
	return &GradeTranslation{}
}

func (gt *GradeTranslation) ID() string {
	return gt.id
}

func (gt *GradeTranslation) GenerateID() {
	gt.id = uuid.New().String()
}

func (gt *GradeTranslation) SetID(id string) {
	gt.id = id
}

func (gt *GradeTranslation) GradeID() string {
	return gt.gradeID
}

func (gt *GradeTranslation) SetGradeID(gradeID string) {
	gt.gradeID = gradeID
}

func (gt *GradeTranslation) Language() string {
	return gt.language
}

func (gt *GradeTranslation) SetLanguage(language string) {
	gt.language = language
}

func (gt *GradeTranslation) Label() string {
	return gt.label
}

func (gt *GradeTranslation) SetLabel(label string) {
	gt.label = label
}

func (gt *GradeTranslation) Description() string {
	return gt.description
}

func (gt *GradeTranslation) SetDescription(description string) {
	gt.description = description
}

func BuildGradeTranslationFromModel(model *models.GradeTranslationModel) *GradeTranslation {
	return &GradeTranslation{
		id:          model.ID,
		gradeID:     model.GradeID,
		language:    model.Language,
		label:       model.Label,
		description: model.Description,
	}
}
