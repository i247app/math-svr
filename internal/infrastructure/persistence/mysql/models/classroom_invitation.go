package models

import "time"

type ClassroomInvitationModel struct {
	Id                    int64
	InvitationId          int64
	ClassroomId           int64
	InviterProfileId      int64
	InvitedProfileId      *int64
	InviteeIdentifier     *string
	InviteeIdentifierType *string
	ProposedRole          string
	Token                 string
	Message               *string
	SentDt                time.Time
	ExpiresDt             *time.Time
	RespondedDt           *time.Time
	ResponseProfileId     *int64
	CancelledByProfileId  *int64
	Note                  *string
	InvitationStatus      *string
	Status                string
	CreateId              *int64
	CreateDt              time.Time
	ModifyId              *int64
	ModifyDt              time.Time
}
