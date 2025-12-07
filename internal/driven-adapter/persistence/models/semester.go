package models

import "time"

type SemesterModel struct {
	ID           string
	Name         string
	Description  *string
	ImageKey     *string
	Status       string
	DisplayOrder int8
	CreateID     *int64
	CreateDT     time.Time
	ModifyID     *int64
	ModifyDT     time.Time
	DeletedDT    *time.Time
}
