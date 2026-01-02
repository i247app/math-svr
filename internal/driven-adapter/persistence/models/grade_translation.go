package models

import "math-ai.com/math-ai/internal/shared/utils/time"

type GradeTranslationModel struct {
	ID          string
	GradeID     string
	Language    string
	Label       string
	Description *string
	Note        *string
	GTStatus    string
	Status      string
	CreateID    *int64
	CreateDT    time.MathTime
	ModifyID    *int64
	ModifyDT    time.MathTime
	DeletedDT   *time.MathTime
}
