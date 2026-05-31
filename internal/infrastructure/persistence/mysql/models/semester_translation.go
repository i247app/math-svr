package models

import "time"

type SemesterTranslationModel struct {
	Id                    int64
	SemesterTranslationId int64
	SemesterId            int64
	Language              string
	Name                  string
	Description           string
	Note                  *string
	StStatus              *string
	Status                string
	CreateId              *int64
	CreateDt              time.Time
	ModifyId              *int64
	ModifyDt              time.Time
}
