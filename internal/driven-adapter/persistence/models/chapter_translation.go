package models

type ChapterTranslationModel struct {
	ID          string
	ChapterID   string
	Language    string
	Title       string
	Description *string
}
