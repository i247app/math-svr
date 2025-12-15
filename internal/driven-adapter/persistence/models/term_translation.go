package models

type TermTranslationModel struct {
	ID          string
	TermID      string
	Language    string
	Name        string
	Description *string
}
