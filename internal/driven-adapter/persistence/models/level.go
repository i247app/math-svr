package models

import (
	"time"
)

type LevelModel struct {
	ID           string
	Label        string
	Description  string
	Status       string
	DisplayOrder int8
	CreateID     *int64
	CreateDT     time.Time
	ModifyID     *int64
	ModifyDT     time.Time
	DeletedDT    *time.Time
}
