package classroom

import (
	"math-ai.com/math-ai/internal/domain/shared/mtime"
)

// Member models ma_classroom_members — the (classroom, profile) join
// row that doubles as the invitation record. member_status is the
// source of truth for both invitation and membership lifecycle
// (PENDING → ACTIVE/REJECTED, ACTIVE → LEFT/REMOVED). The UNIQUE
// (classroom_id, profile_id) constraint means re-invitation and
// rejoin both reactivate the same row instead of inserting a
// duplicate. inviteBy + inviteDt capture who invited the profile and
// when; both are NULL for the owner row (created via CreateClassroom)
// and for code-joiners. invitationId is the legacy back-reference to
// ma_classroom_invitations and is no longer written — it remains on
// the entity only so historical rows hydrate without scan errors.
type Member struct {
	id                 int64
	memberId           int64
	classroomId        int64
	profileId          int64
	memberRole         string
	invitationId       *int64
	joinedDt           mtime.MathTime
	leftDt             mtime.MathTime
	removedByProfileId *int64
	removedDt          mtime.MathTime
	lastSeenDt         mtime.MathTime
	note               *string
	inviteBy           *int64
	inviteDt           mtime.MathTime
	memberStatus       *string
	status             string
	createId           *int64
	createDt           mtime.MathTime
	modifyId           *int64
	modifyDt           mtime.MathTime
}

func NewMember() *Member {
	return &Member{}
}

func (m *Member) Id() int64                      { return m.id }
func (m *Member) SetId(id int64)                 { m.id = id }
func (m *Member) MemberId() int64                { return m.memberId }
func (m *Member) SetMemberId(id int64)           { m.memberId = id }
func (m *Member) ClassroomId() int64             { return m.classroomId }
func (m *Member) SetClassroomId(id int64)        { m.classroomId = id }
func (m *Member) ProfileId() int64               { return m.profileId }
func (m *Member) SetProfileId(id int64)          { m.profileId = id }
func (m *Member) MemberRole() string             { return m.memberRole }
func (m *Member) SetMemberRole(v string)         { m.memberRole = v }
func (m *Member) InvitationId() *int64           { return m.invitationId }
func (m *Member) SetInvitationId(v *int64)       { m.invitationId = v }
func (m *Member) JoinedDt() mtime.MathTime       { return m.joinedDt }
func (m *Member) SetJoinedDt(t mtime.MathTime)   { m.joinedDt = t }
func (m *Member) LeftDt() mtime.MathTime         { return m.leftDt }
func (m *Member) SetLeftDt(t mtime.MathTime)     { m.leftDt = t }
func (m *Member) RemovedByProfileId() *int64     { return m.removedByProfileId }
func (m *Member) SetRemovedByProfileId(v *int64) { m.removedByProfileId = v }
func (m *Member) RemovedDt() mtime.MathTime      { return m.removedDt }
func (m *Member) SetRemovedDt(t mtime.MathTime)  { m.removedDt = t }
func (m *Member) LastSeenDt() mtime.MathTime     { return m.lastSeenDt }
func (m *Member) SetLastSeenDt(t mtime.MathTime) { m.lastSeenDt = t }
func (m *Member) Note() *string                  { return m.note }
func (m *Member) SetNote(n *string)              { m.note = n }
func (m *Member) InviteBy() *int64               { return m.inviteBy }
func (m *Member) SetInviteBy(v *int64)           { m.inviteBy = v }
func (m *Member) InviteDt() mtime.MathTime       { return m.inviteDt }
func (m *Member) SetInviteDt(t mtime.MathTime)   { m.inviteDt = t }
func (m *Member) MemberStatus() *string          { return m.memberStatus }
func (m *Member) SetMemberStatus(v *string)      { m.memberStatus = v }
func (m *Member) Status() string                 { return m.status }
func (m *Member) SetStatus(v string)             { m.status = v }
func (m *Member) CreateId() *int64               { return m.createId }
func (m *Member) SetCreateId(id *int64)          { m.createId = id }
func (m *Member) CreateDt() mtime.MathTime       { return m.createDt }
func (m *Member) SetCreateDt(t mtime.MathTime)   { m.createDt = t }
func (m *Member) ModifyId() *int64               { return m.modifyId }
func (m *Member) SetModifyId(id *int64)          { m.modifyId = id }
func (m *Member) ModifyDt() mtime.MathTime       { return m.modifyDt }
func (m *Member) SetModifyDt(t mtime.MathTime)   { m.modifyDt = t }
