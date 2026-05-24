package models

import (
	"time"
)

type ProgramModel struct {
	Id            int64
	ProgramId     string
	Label         string
	Description   string
	ImageKey      *string
	DisplayOrder  int8
	Note          *string
	ProgramStatus *string
	Status        string
	CreateId      *string
	CreateDt      time.Time
	ModifyId      *string
	ModifyDt      time.Time
}
