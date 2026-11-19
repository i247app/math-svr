package models

type SemesterModel struct {
	Audit

	SemesterId     int64
	Name           string
	Description    string
	ImageKey       *string
	DisplayOrder   int8
	SemesterStatus *string
}
