package classroom

import (
	mtime "math-ai.com/math-ai/internal/domain/shared/time"
)

// Invitation models ma_classroom_invitations. Exactly one of
// invitedProfileId / inviteeIdentifier should be set per invitation.
// token is the secret embedded in the invite link; uniqueness is
// enforced at the DB layer. responseProfileId captures *which* of
// the user's profiles claimed the invitation (only meaningful when
// the invitee was identified by EMAIL/PHONE — a user with multiple
// profiles must pick one at acceptance time).
type Invitation struct {
	id                    int64
	invitationId          int64
	classroomId           int64
	inviterProfileId      int64
	invitedProfileId      *int64
	inviteeIdentifier     *string
	inviteeIdentifierType *string
	proposedRole          string
	token                 string
	message               *string
	sentDt                mtime.MathTime
	expiresDt             mtime.MathTime
	respondedDt           mtime.MathTime
	responseProfileId     *int64
	cancelledByProfileId  *int64
	note                  *string
	invitationStatus      *string
	status                string
	createId              *int64
	createDt              mtime.MathTime
	modifyId              *int64
	modifyDt              mtime.MathTime
}

func NewInvitation() *Invitation {
	return &Invitation{}
}

func (i *Invitation) Id() int64                          { return i.id }
func (i *Invitation) SetId(id int64)                     { i.id = id }
func (i *Invitation) InvitationId() int64                { return i.invitationId }
func (i *Invitation) SetInvitationId(id int64)           { i.invitationId = id }
func (i *Invitation) ClassroomId() int64                 { return i.classroomId }
func (i *Invitation) SetClassroomId(id int64)            { i.classroomId = id }
func (i *Invitation) InviterProfileId() int64            { return i.inviterProfileId }
func (i *Invitation) SetInviterProfileId(id int64)       { i.inviterProfileId = id }
func (i *Invitation) InvitedProfileId() *int64           { return i.invitedProfileId }
func (i *Invitation) SetInvitedProfileId(v *int64)       { i.invitedProfileId = v }
func (i *Invitation) InviteeIdentifier() *string         { return i.inviteeIdentifier }
func (i *Invitation) SetInviteeIdentifier(v *string)     { i.inviteeIdentifier = v }
func (i *Invitation) InviteeIdentifierType() *string     { return i.inviteeIdentifierType }
func (i *Invitation) SetInviteeIdentifierType(v *string) { i.inviteeIdentifierType = v }
func (i *Invitation) ProposedRole() string               { return i.proposedRole }
func (i *Invitation) SetProposedRole(v string)           { i.proposedRole = v }
func (i *Invitation) Token() string                      { return i.token }
func (i *Invitation) SetToken(v string)                  { i.token = v }
func (i *Invitation) Message() *string                   { return i.message }
func (i *Invitation) SetMessage(v *string)               { i.message = v }
func (i *Invitation) SentDt() mtime.MathTime             { return i.sentDt }
func (i *Invitation) SetSentDt(t mtime.MathTime)         { i.sentDt = t }
func (i *Invitation) ExpiresDt() mtime.MathTime          { return i.expiresDt }
func (i *Invitation) SetExpiresDt(t mtime.MathTime)      { i.expiresDt = t }
func (i *Invitation) RespondedDt() mtime.MathTime        { return i.respondedDt }
func (i *Invitation) SetRespondedDt(t mtime.MathTime)    { i.respondedDt = t }
func (i *Invitation) ResponseProfileId() *int64          { return i.responseProfileId }
func (i *Invitation) SetResponseProfileId(v *int64)      { i.responseProfileId = v }
func (i *Invitation) CancelledByProfileId() *int64       { return i.cancelledByProfileId }
func (i *Invitation) SetCancelledByProfileId(v *int64)   { i.cancelledByProfileId = v }
func (i *Invitation) Note() *string                      { return i.note }
func (i *Invitation) SetNote(v *string)                  { i.note = v }
func (i *Invitation) InvitationStatus() *string          { return i.invitationStatus }
func (i *Invitation) SetInvitationStatus(v *string)      { i.invitationStatus = v }
func (i *Invitation) Status() string                     { return i.status }
func (i *Invitation) SetStatus(v string)                 { i.status = v }
func (i *Invitation) CreateId() *int64                   { return i.createId }
func (i *Invitation) SetCreateId(id *int64)              { i.createId = id }
func (i *Invitation) CreateDt() mtime.MathTime           { return i.createDt }
func (i *Invitation) SetCreateDt(t mtime.MathTime)       { i.createDt = t }
func (i *Invitation) ModifyId() *int64                   { return i.modifyId }
func (i *Invitation) SetModifyId(id *int64)              { i.modifyId = id }
func (i *Invitation) ModifyDt() mtime.MathTime           { return i.modifyDt }
func (i *Invitation) SetModifyDt(t mtime.MathTime)       { i.modifyDt = t }
