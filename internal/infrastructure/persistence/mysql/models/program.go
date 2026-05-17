package models

import (
	"time"

	"github.com/google/uuid"
)

type ProgramModel struct {
	Id            int64
	ProgramId     uuid.UUID
	Label         string
	Description   string
	ImageKey      *string
	DisplayOrder  int8
	Note          *string
	ProgramStatus *string
	Status        string
	CreateId      *uuid.UUID
	CreateDt      time.Time
	ModifyId      *uuid.UUID
	ModifyDt      time.Time
}
