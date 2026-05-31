package models

import "time"

type ClassroomMemberModel struct {
	Id                 int64
	MemberId           int64
	ClassroomId        int64
	ProfileId          int64
	MemberRole         string
	InvitationId       *int64
	JoinedDt           *time.Time
	LeftDt             *time.Time
	RemovedByProfileId *int64
	RemovedDt          *time.Time
	LastSeenDt         *time.Time
	Note               *string
	InviteBy           *int64
	InviteDt           *time.Time
	MemberStatus       *string
	Status             string
	CreateId           *int64
	CreateDt           time.Time
	ModifyId           *int64
	ModifyDt           time.Time
}
