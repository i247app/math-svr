package models

import "time"

type ClassroomMemberModel struct {
	Id                  int64
	MemberId            string
	ClassroomId         string
	ProfileId           string
	MemberRole          string
	InvitationId        *string
	JoinedDt            *time.Time
	LeftDt              *time.Time
	RemovedByProfileId  *string
	RemovedDt           *time.Time
	LastSeenDt          *time.Time
	Note                *string
	MemberStatus        *string
	Status              string
	CreateId            *string
	CreateDt            time.Time
	ModifyId            *string
	ModifyDt            time.Time
}
