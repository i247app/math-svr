package models

import "time"

type SemesterTranslationModel struct {
	Id                    int64
	SemesterTranslationId string
	SemesterId            string
	Language              string
	Name                  string
	Description           string
	Note                  *string
	StStatus              *string
	Status                string
	CreateId              *string
	CreateDt              time.Time
	ModifyId              *string
	ModifyDt              time.Time
}
