package models

import (
	"time"
)

type GradeModel struct {
	ID          string
	Label       string
	Description string
	Status      string
	CreateID    *int64
	CreateDT    time.Time
	ModifyID    *int64
	ModifyDT    time.Time
	DeletedDT   *time.Time
}
