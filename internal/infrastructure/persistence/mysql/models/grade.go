package models

import (
	"time"
)

type GradeModel struct {
	Id           int64
	GradeId      int64
	Label        string
	Description  string
	ImageKey     *string
	DisplayOrder int8
	Note         *string
	GradeStatus  *string
	Status       string
	CreateId     *int64
	CreateDt     time.Time
	ModifyId     *int64
	ModifyDt     time.Time
}
