package models

type LessonTranslationModel struct {
	ID       string
	LessonID string
	Language string
	Title    string
	Content  *string
}
