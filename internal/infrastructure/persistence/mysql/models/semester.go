package models

import (
	"time"

	"github.com/google/uuid"
)

type SemesterModel struct {
	Id             int64
	SemesterId     uuid.UUID
	Name           string
	Description    string
	ImageKey       *string
	DisplayOrder   int8
	Note           *string
	SemesterStatus *string
	Status         string
	CreateId       *uuid.UUID
	CreateDt       time.Time
	ModifyId       *uuid.UUID
	ModifyDt       time.Time
}
