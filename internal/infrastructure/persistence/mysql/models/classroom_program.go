package models

type ClassroomProgramModel struct {
	Audit

	ClassroomProgramId int64
	ClassroomId        int64
	ProgramId          int64
}
