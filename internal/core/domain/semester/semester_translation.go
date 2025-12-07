package domain

import (
	"github.com/google/uuid"
	"math-ai.com/math-ai/internal/driven-adapter/persistence/models"
)

type SemesterTranslation struct {
	id          string
	semesterID  string
	language    string
	name        string
	description *string
}

func NewSemesterTranslation() *SemesterTranslation {
	return &SemesterTranslation{}
}

func (st *SemesterTranslation) ID() string {
	return st.id
}

func (st *SemesterTranslation) GenerateID() {
	st.id = uuid.New().String()
}

func (st *SemesterTranslation) SetID(id string) {
	st.id = id
}

func (st *SemesterTranslation) SemesterID() string {
	return st.semesterID
}

func (st *SemesterTranslation) SetSemesterID(semesterID string) {
	st.semesterID = semesterID
}

func (st *SemesterTranslation) Language() string {
	return st.language
}

func (st *SemesterTranslation) SetLanguage(language string) {
	st.language = language
}

func (st *SemesterTranslation) Name() string {
	return st.name
}

func (st *SemesterTranslation) SetName(name string) {
	st.name = name
}

func (st *SemesterTranslation) Description() *string {
	return st.description
}

func (st *SemesterTranslation) SetDescription(description *string) {
	st.description = description
}

func BuildSemesterTranslationFromModel(model *models.SemesterTranslationModel) *SemesterTranslation {
	return &SemesterTranslation{
		id:          model.ID,
		semesterID:  model.SemesterID,
		language:    model.Language,
		name:        model.Name,
		description: model.Description,
	}
}
