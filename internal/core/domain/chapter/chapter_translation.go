package domain

import (
	"github.com/google/uuid"
	"math-ai.com/math-ai/internal/driven-adapter/persistence/models"
)

type ChapterTranslation struct {
	id          string
	chapterID   string
	language    string
	title       string
	description *string
}

func NewChapterTranslation() *ChapterTranslation {
	return &ChapterTranslation{}
}

func (ct *ChapterTranslation) ID() string {
	return ct.id
}

func (ct *ChapterTranslation) GenerateID() {
	ct.id = uuid.New().String()
}

func (ct *ChapterTranslation) SetID(id string) {
	ct.id = id
}

func (ct *ChapterTranslation) ChapterID() string {
	return ct.chapterID
}

func (ct *ChapterTranslation) SetChapterID(chapterID string) {
	ct.chapterID = chapterID
}

func (ct *ChapterTranslation) Language() string {
	return ct.language
}

func (ct *ChapterTranslation) SetLanguage(language string) {
	ct.language = language
}

func (ct *ChapterTranslation) Title() string {
	return ct.title
}

func (ct *ChapterTranslation) SetTitle(title string) {
	ct.title = title
}

func (ct *ChapterTranslation) Description() *string {
	return ct.description
}

func (ct *ChapterTranslation) SetDescription(description *string) {
	ct.description = description
}

func BuildChapterTranslationFromModel(model *models.ChapterTranslationModel) *ChapterTranslation {
	return &ChapterTranslation{
		id:          model.ID,
		chapterID:   model.ChapterID,
		language:    model.Language,
		title:       model.Title,
		description: model.Description,
	}
}
