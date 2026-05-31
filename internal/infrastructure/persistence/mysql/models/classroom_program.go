package models

import "time"

type ClassroomProgramModel struct {
	Id                 int64
	ClassroomProgramId int64
	ClassroomId        int64
	ProgramId          int64
	Note               *string
	Status             string
	CreateId           *int64
	CreateDt           time.Time
	ModifyId           *int64
	ModifyDt           time.Time
}
