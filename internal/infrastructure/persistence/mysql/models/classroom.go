package models

import "time"

type ClassroomModel struct {
	Id                  int64
	ClassroomId         string
	OwnerProfileId      string
	Name                string
	Description         *string
	SchoolId            *string
	GradeId             *string
	InviteCode          *string
	InviteCodeExpiresDt *time.Time
	MaxMembers          *int64
	MemberCount         int64
	StudentCount        int64
	TeacherCount        int64
	CoverKey            *string
	Note                *string
	ClassroomStatus     *string
	Status              string
	CreateId            *string
	CreateDt            time.Time
	ModifyId            *string
	ModifyDt            time.Time
}
