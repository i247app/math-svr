package models

import (
	"time"
)

type ProfileModel struct {
	Audit

	ProfileId     int64
	ProfileCode   string
	Uid           int64
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
	ProfileStatus *string
}
