package models

import "time"

type GradeTranslationModel struct {
	Id                 int64
	GradeTranslationId string
	GradeId            string
	Language           string
	Label              string
	Description        string
	Note               *string
	GtStatus           *string
	Status             string
	CreateId           *string
	CreateDt           time.Time
	ModifyId           *string
	ModifyDt           time.Time
}
