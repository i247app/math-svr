package models

type SemesterTranslationModel struct {
	ID          string
	SemesterID  string
	Language    string
	Name        string
	Description *string
}
