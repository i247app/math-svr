package models

import "math-ai.com/math-ai/internal/shared/utils/time"

type SemesterModel struct {
	ID           string
	Name         string
	Description  *string
	ImageKey     *string
	Status       string
	DisplayOrder int8
	CreateID     *int64
	CreateDT     time.MathTime
	ModifyID     *int64
	ModifyDT     time.MathTime
	DeletedDT    *time.MathTime
}
