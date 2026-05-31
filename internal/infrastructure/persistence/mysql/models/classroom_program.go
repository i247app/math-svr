package models

import "time"

type ClassroomProgramModel struct {
	Id                 int64
	ClassroomProgramId string
	ClassroomId        string
	ProgramId          string
	Note               *string
	Status             string
	CreateId           *string
	CreateDt           time.Time
	ModifyId           *string
	ModifyDt           time.Time
}
