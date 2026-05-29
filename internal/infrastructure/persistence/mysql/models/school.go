package models

import "time"

type SchoolModel struct {
	Id           int64
	SchoolId     string
	Name         string
	Description  *string
	ImageKey     *string
	District     *string
	Province     *string
	Note         *string
	SchoolStatus *string
	Status       string
	CreateId     *string
	CreateDt     time.Time
	ModifyId     *string
	ModifyDt     time.Time
}
