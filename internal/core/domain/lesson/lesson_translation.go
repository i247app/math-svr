package domain

import (
	"github.com/google/uuid"
	"math-ai.com/math-ai/internal/driven-adapter/persistence/models"
)

type LessonTranslation struct {
	id       string
	lessonID string
	language string
	title    string
	content  *string
}

func NewLessonTranslation() *LessonTranslation {
	return &LessonTranslation{}
}

func (lt *LessonTranslation) ID() string {
	return lt.id
}

func (lt *LessonTranslation) GenerateID() {
	lt.id = uuid.New().String()
}

func (lt *LessonTranslation) SetID(id string) {
	lt.id = id
}

func (lt *LessonTranslation) LessonID() string {
	return lt.lessonID
}

func (lt *LessonTranslation) SetLessonID(lessonID string) {
	lt.lessonID = lessonID
}

func (lt *LessonTranslation) Language() string {
	return lt.language
}

func (lt *LessonTranslation) SetLanguage(language string) {
	lt.language = language
}

func (lt *LessonTranslation) Title() string {
	return lt.title
}

func (lt *LessonTranslation) SetTitle(title string) {
	lt.title = title
}

func (lt *LessonTranslation) Content() *string {
	return lt.content
}

func (lt *LessonTranslation) SetContent(content *string) {
	lt.content = content
}

func BuildLessonTranslationFromModel(model *models.LessonTranslationModel) *LessonTranslation {
	return &LessonTranslation{
		id:       model.ID,
		lessonID: model.LessonID,
		language: model.Language,
		title:    model.Title,
		content:  model.Content,
	}
}
