package models

import (
	"time"
)

type ProfileModel struct {
	ProfileId     int64
	ProfileCode   string
	UserId        int64
	Name          string
	Phone         *string
	Email         *string
	Role          *string
	IdentityCode  *string
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
	RptFlg        *string
	Kwords        *string
	Note          *string
	ProfileStatus *string
	Status        string
	CreateId      *int64
	CreateDt      time.Time
	ModifyId      *int64
	ModifyDt      time.Time
}
