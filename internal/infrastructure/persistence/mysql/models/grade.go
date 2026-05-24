package models

import (
	"time"
)

type GradeModel struct {
	Id           int64
	GradeId      string
	Label        string
	Description  string
	ImageKey     *string
	DisplayOrder int8
	Note         *string
	GradeStatus  *string
	Status       string
	CreateId     *string
	CreateDt     time.Time
	ModifyId     *string
	ModifyDt     time.Time
}
