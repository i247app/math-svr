package domain

import (
	"github.com/google/uuid"
	"math-ai.com/math-ai/internal/driven-adapter/persistence/models"
)

type TermTranslation struct {
	id          string
	termID      string
	language    string
	name        string
	description *string
}

func NewTermTranslation() *TermTranslation {
	return &TermTranslation{}
}

func (st *TermTranslation) ID() string {
	return st.id
}

func (st *TermTranslation) GenerateID() {
	st.id = uuid.New().String()
}

func (st *TermTranslation) SetID(id string) {
	st.id = id
}

func (st *TermTranslation) TermID() string {
	return st.termID
}

func (st *TermTranslation) SetTermID(termID string) {
	st.termID = termID
}

func (st *TermTranslation) Language() string {
	return st.language
}

func (st *TermTranslation) SetLanguage(language string) {
	st.language = language
}

func (st *TermTranslation) Name() string {
	return st.name
}

func (st *TermTranslation) SetName(name string) {
	st.name = name
}

func (st *TermTranslation) Description() *string {
	return st.description
}

func (st *TermTranslation) SetDescription(description *string) {
	st.description = description
}

func BuildTermTranslationFromModel(model *models.TermTranslationModel) *TermTranslation {
	return &TermTranslation{
		id:          model.ID,
		termID:      model.TermID,
		language:    model.Language,
		name:        model.Name,
		description: model.Description,
	}
}
