package models

import (
	"math-ai.com/math-ai/internal/shared/utils/time"
)

type GradeModel struct {
	ID           string
	Label        string
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
