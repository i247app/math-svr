package models

import (
	"time"
)

type ProgramModel struct {
	Id            int64
	ProgramId     int64
	Label         string
	Description   string
	ImageKey      *string
	DisplayOrder  int8
	Note          *string
	ProgramStatus *string
	Status        string
	CreateId      *int64
	CreateDt      time.Time
	ModifyId      *int64
	ModifyDt      time.Time
}
