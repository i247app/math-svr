package models

import (
	"time"

	"github.com/google/uuid"
)

type GradeModel struct {
	Id           int64
	GradeId      uuid.UUID
	Label        string
	Description  string
	ImageKey     *string
	DisplayOrder int8
	Note         *string
	GradeStatus  *string
	Status       string
	CreateId     *uuid.UUID
	CreateDt     time.Time
	ModifyId     *uuid.UUID
	ModifyDt     time.Time
}
