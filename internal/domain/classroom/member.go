package classroom

import (
	mtime "math-ai.com/math-ai/internal/domain/shared/time"
)

// Member models ma_classroom_members — the (classroom, profile) join row
// with role and lifecycle status. The UNIQUE (classroom_id, profile_id)
// constraint means rejoin updates the same row: member_status flips
// LEFT/REMOVED → ACTIVE, joined_dt is refreshed, and left_dt/removed_*
// are cleared. invitationId is the back-reference to the invitation
// that produced the row (null for the owner and for code-joiners).
type Member struct {
	id                   int64
	memberId             string
	classroomId          string
	profileId            string
	memberRole           string
	invitationId         *string
	joinedDt             mtime.MathTime
	leftDt               mtime.MathTime
	removedByProfileId   *string
	removedDt            mtime.MathTime
	lastSeenDt           mtime.MathTime
	note                 *string
	memberStatus         *string
	status               string
	createId             *string
	createDt             mtime.MathTime
	modifyId             *string
	modifyDt             mtime.MathTime
}

func NewMember() *Member {
	return &Member{}
}

func (m *Member) Id() int64                                { return m.id }
func (m *Member) SetId(id int64)                           { m.id = id }
func (m *Member) MemberId() string                         { return m.memberId }
func (m *Member) SetMemberId(id string)                    { m.memberId = id }
func (m *Member) ClassroomId() string                      { return m.classroomId }
func (m *Member) SetClassroomId(id string)                 { m.classroomId = id }
func (m *Member) ProfileId() string                        { return m.profileId }
func (m *Member) SetProfileId(id string)                   { m.profileId = id }
func (m *Member) MemberRole() string                       { return m.memberRole }
func (m *Member) SetMemberRole(v string)                   { m.memberRole = v }
func (m *Member) InvitationId() *string                    { return m.invitationId }
func (m *Member) SetInvitationId(v *string)                { m.invitationId = v }
func (m *Member) JoinedDt() mtime.MathTime                 { return m.joinedDt }
func (m *Member) SetJoinedDt(t mtime.MathTime)             { m.joinedDt = t }
func (m *Member) LeftDt() mtime.MathTime                   { return m.leftDt }
func (m *Member) SetLeftDt(t mtime.MathTime)               { m.leftDt = t }
func (m *Member) RemovedByProfileId() *string              { return m.removedByProfileId }
func (m *Member) SetRemovedByProfileId(v *string)          { m.removedByProfileId = v }
func (m *Member) RemovedDt() mtime.MathTime                { return m.removedDt }
func (m *Member) SetRemovedDt(t mtime.MathTime)            { m.removedDt = t }
func (m *Member) LastSeenDt() mtime.MathTime               { return m.lastSeenDt }
func (m *Member) SetLastSeenDt(t mtime.MathTime)           { m.lastSeenDt = t }
func (m *Member) Note() *string                            { return m.note }
func (m *Member) SetNote(n *string)                        { m.note = n }
func (m *Member) MemberStatus() *string                    { return m.memberStatus }
func (m *Member) SetMemberStatus(v *string)                { m.memberStatus = v }
func (m *Member) Status() string                           { return m.status }
func (m *Member) SetStatus(v string)                       { m.status = v }
func (m *Member) CreateId() *string                        { return m.createId }
func (m *Member) SetCreateId(id *string)                   { m.createId = id }
func (m *Member) CreateDt() mtime.MathTime                 { return m.createDt }
func (m *Member) SetCreateDt(t mtime.MathTime)             { m.createDt = t }
func (m *Member) ModifyId() *string                        { return m.modifyId }
func (m *Member) SetModifyId(id *string)                   { m.modifyId = id }
func (m *Member) ModifyDt() mtime.MathTime                 { return m.modifyDt }
func (m *Member) SetModifyDt(t mtime.MathTime)             { m.modifyDt = t }
