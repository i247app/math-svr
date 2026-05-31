package models

import (
	"time"
)

type SemesterModel struct {
	Id             int64
	SemesterId     int64
	Name           string
	Description    string
	ImageKey       *string
	DisplayOrder   int8
	Note           *string
	SemesterStatus *string
	Status         string
	CreateId       *int64
	CreateDt       time.Time
	ModifyId       *int64
	ModifyDt       time.Time
}
