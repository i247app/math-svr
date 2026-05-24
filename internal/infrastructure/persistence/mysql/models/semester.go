package models

import (
	"time"
)

type SemesterModel struct {
	Id             int64
	SemesterId     string
	Name           string
	Description    string
	ImageKey       *string
	DisplayOrder   int8
	Note           *string
	SemesterStatus *string
	Status         string
	CreateId       *string
	CreateDt       time.Time
	ModifyId       *string
	ModifyDt       time.Time
}
