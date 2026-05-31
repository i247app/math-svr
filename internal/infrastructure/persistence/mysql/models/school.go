package models

import "time"

type SchoolModel struct {
	Id           int64
	SchoolId     int64
	Name         string
	Description  *string
	ImageKey     *string
	District     *string
	Province     *string
	Note         *string
	SchoolStatus *string
	Status       string
	CreateId     *int64
	CreateDt     time.Time
	ModifyId     *int64
	ModifyDt     time.Time
}
