package models

import (
	"time"
)

type GradeModel struct {
	ID           string
	Label        string
	Description  *string
	IconURL      *string
	Status       string
	DisplayOrder int8
	CreateID     *int64
	CreateDT     time.Time
	ModifyID     *int64
	ModifyDT     time.Time
	DeletedDT    *time.Time
	// Label and Description moved to GradeTranslationModel
}
