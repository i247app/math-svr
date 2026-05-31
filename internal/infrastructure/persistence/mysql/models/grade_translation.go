package models

import "time"

type GradeTranslationModel struct {
	Id                 int64
	GradeTranslationId int64
	GradeId            int64
	Language           string
	Label              string
	Description        string
	Note               *string
	GtStatus           *string
	Status             string
	CreateId           *int64
	CreateDt           time.Time
	ModifyId           *int64
	ModifyDt           time.Time
}
