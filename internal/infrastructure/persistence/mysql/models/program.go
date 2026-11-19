package models

type ProgramModel struct {
	Audit

	ProgramId     int64
	Label         string
	Description   string
	ImageKey      *string
	DisplayOrder  int8
	ProgramStatus *string
}
