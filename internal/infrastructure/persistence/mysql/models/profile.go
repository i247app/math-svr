package models

import (
	"time"
)

type ProfileModel struct {
	Id            int64
	ProfileId     int64
	UserId        int64
	Name          string
	Role          string
	AvatarKey     *string
	Dob           *time.Time
	SchoolId      *int64
	ProgramId     *int64
	GradeId       *int64
	SemesterId    *int64
	IsDefault     bool
	IdType        *string
	TeacherId     *string
	StudentId     *string
	Note          *string
	ProfileStatus *string
	Status        string
	CreateId      *int64
	CreateDt      time.Time
	ModifyId      *int64
	ModifyDt      time.Time
}
