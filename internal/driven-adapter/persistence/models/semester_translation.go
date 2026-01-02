package models

import "math-ai.com/math-ai/internal/shared/utils/time"

type SemesterTranslationModel struct {
	ID          string
	SemesterID  string
	Language    string
	Name        string
	Description *string
	Note        *string
	STStatus    string
	Status      string
	CreateID    *int64
	CreateDT    time.MathTime
	ModifyID    *int64
	ModifyDT    time.MathTime
	DeletedDT   *time.MathTime
}
