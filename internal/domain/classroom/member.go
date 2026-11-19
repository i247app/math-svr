package classroom

import (
	"time"

	"math-ai.com/math-ai/internal/domain/shared/audit"
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
	audit.Audit

	memberId           int64
	classroomId        int64
	profileId          int64
	memberRole         string
	invitationId       *int64
	joinedDt           time.Time
	leftDt             time.Time
	removedByProfileId *int64
	removedDt          time.Time
	lastSeenDt         time.Time
	inviteBy           *int64
	inviteDt           time.Time
	memberStatus       *string
}

func NewMember() *Member {
	return &Member{}
}

func (m *Member) MemberId() int64      { return m.memberId }
func (m *Member) SetMemberId(id int64) { m.memberId = id }

func (m *Member) ClassroomId() int64      { return m.classroomId }
func (m *Member) SetClassroomId(id int64) { m.classroomId = id }

func (m *Member) ProfileId() int64      { return m.profileId }
func (m *Member) SetProfileId(id int64) { m.profileId = id }

func (m *Member) MemberRole() string     { return m.memberRole }
func (m *Member) SetMemberRole(v string) { m.memberRole = v }

func (m *Member) InvitationId() *int64     { return m.invitationId }
func (m *Member) SetInvitationId(v *int64) { m.invitationId = v }

func (m *Member) JoinedDt() time.Time     { return m.joinedDt }
func (m *Member) SetJoinedDt(t time.Time) { m.joinedDt = t }

func (m *Member) LeftDt() time.Time     { return m.leftDt }
func (m *Member) SetLeftDt(t time.Time) { m.leftDt = t }

func (m *Member) RemovedByProfileId() *int64     { return m.removedByProfileId }
func (m *Member) SetRemovedByProfileId(v *int64) { m.removedByProfileId = v }

func (m *Member) RemovedDt() time.Time     { return m.removedDt }
func (m *Member) SetRemovedDt(t time.Time) { m.removedDt = t }

func (m *Member) LastSeenDt() time.Time     { return m.lastSeenDt }
func (m *Member) SetLastSeenDt(t time.Time) { m.lastSeenDt = t }

func (m *Member) InviteBy() *int64     { return m.inviteBy }
func (m *Member) SetInviteBy(v *int64) { m.inviteBy = v }

func (m *Member) InviteDt() time.Time     { return m.inviteDt }
func (m *Member) SetInviteDt(t time.Time) { m.inviteDt = t }

func (m *Member) MemberStatus() *string     { return m.memberStatus }
func (m *Member) SetMemberStatus(v *string) { m.memberStatus = v }
