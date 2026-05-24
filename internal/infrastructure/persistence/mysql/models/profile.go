package models

import (
	"time"
)

type ProfileModel struct {
	Id            int64
	ProfileId     string
	UserId        string
	Name          string
	AvatarKey     *string
	Dob           *time.Time
	ProgramId     *string
	GradeId       *string
	SemesterId    *string
	IsDefault     bool
	Note          *string
	ProfileStatus *string
	Status        string
	CreateId      *string
	CreateDt      time.Time
	ModifyId      *string
	ModifyDt      time.Time
}
