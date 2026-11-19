package models

type SchoolModel struct {
	Audit

	SchoolId     int64
	Name         string
	Description  *string
	ImageKey     *string
	District     *string
	Province     *string
	SchoolStatus *string
}
