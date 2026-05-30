package models

import "time"

type ClassroomInvitationModel struct {
	Id                    int64
	InvitationId          string
	ClassroomId           string
	InviterProfileId      string
	InvitedProfileId      *string
	InviteeIdentifier     *string
	InviteeIdentifierType *string
	ProposedRole          string
	Token                 string
	Message               *string
	SentDt                time.Time
	ExpiresDt             *time.Time
	RespondedDt           *time.Time
	ResponseProfileId     *string
	CancelledByProfileId  *string
	Note                  *string
	InvitationStatus      *string
	Status                string
	CreateId              *string
	CreateDt              time.Time
	ModifyId              *string
	ModifyDt              time.Time
}
