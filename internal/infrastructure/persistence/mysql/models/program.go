package models

import (
	"time"
)

type ProgramModel struct {
	ProgramId     int64
	Label         string
	Description   string
	ImageKey      *string
	DisplayOrder  int8
	RptFlg        *string
	Kwords        *string
	Note          *string
	ProgramStatus *string
	Status        string
	CreateId      *int64
	CreateDt      time.Time
	ModifyId      *int64
	ModifyDt      time.Time
}
