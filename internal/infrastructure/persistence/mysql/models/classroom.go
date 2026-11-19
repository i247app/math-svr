package models

import "time"

type ClassroomModel struct {
	Audit

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
	ClassroomStatus        *string
}
