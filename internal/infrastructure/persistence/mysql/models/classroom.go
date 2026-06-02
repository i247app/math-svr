package models

import "time"

type ClassroomModel struct {
	Id                     int64
	ClassroomId            int64
	OwnerProfileId         int64
	Name                   string
	Description            *string
	SchoolId               *int64
	GradeId                *int64
	ClassroomCode          *string
	ClassroomCodeExpiresDt *time.Time
	MaxMembers             *int64
	MemberCount            int64
	StudentCount           int64
	TeacherCount           int64
	CoverKey               *string
	Note                   *string
	ClassroomStatus        *string
	Status                 string
	CreateId               *int64
	CreateDt               time.Time
	ModifyId               *int64
	ModifyDt               time.Time
}
