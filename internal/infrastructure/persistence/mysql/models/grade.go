package models

type GradeModel struct {
	Audit

	GradeId      int64
	Label        string
	Description  string
	ImageKey     *string
	DisplayOrder int8
	GradeStatus  *string
}
