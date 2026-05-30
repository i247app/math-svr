package models

import (
	"time"
)

type ProfileModel struct {
	Id            int64
	ProfileId     string
	UserId        string
	Name          string
	Role          string
	AvatarKey     *string
	Dob           *time.Time
	SchoolId      *string
	ProgramId     *string
	GradeId       *string
	SemesterId    *string
	IsDefault     bool
	IdType        *string
	TeacherId     *string
	StudentId     *string
	Note          *string
	ProfileStatus *string
	Status        string
	CreateId      *string
	CreateDt      time.Time
	ModifyId      *string
	ModifyDt      time.Time
}
