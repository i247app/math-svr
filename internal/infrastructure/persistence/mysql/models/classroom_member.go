package models

import "time"

type ClassroomMemberModel struct {
	Audit

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
	InviteBy           *int64
	InviteDt           *time.Time
	MemberStatus       *string
}
