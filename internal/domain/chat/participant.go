package chat

import (
	"time"

	"math-ai.com/math-ai/internal/domain/shared/audit"
)

// Participant models one ma_chat_participants row — a profile's membership of
// one thread, plus every piece of state that is private to that profile.
//
// It carries BOTH identities on purpose. profileId is the acting identity
// (classroom membership is keyed by profile, and it is the profile's name and
// avatar that render). uid is the delivery target — WebSocket topics are
// user:{uid} and push tokens hang off devices. Denormalising uid here keeps
// the send path, the hottest query in the feature, from joining ma_profiles.
//
// Read state is a watermark, not one row per read message: lastReadSeqNo
// answers both "what have I not read" and "has the other person read message
// 57" from a single column, where receipts would cost one row per member per
// message.
type Participant struct {
	audit.Audit

	participantId      int64
	conversationId     int64
	profileId          int64
	uid                int64
	participantRole    string
	lastReadSeqNo      int64
	lastReadMessageId  *int64
	lastReadDt         time.Time
	lastDeliveredSeqNo int64
	unreadCount        int64
	isMuted            bool
	mutedUntilDt       time.Time
	isPinned           bool
	clearedBeforeSeqNo int64
	joinedDt           time.Time
	leftDt             time.Time
	invitedByProfileId *int64
	participantStatus  *string
}

func NewParticipant() *Participant { return &Participant{} }

func (p *Participant) ParticipantId() int64      { return p.participantId }
func (p *Participant) SetParticipantId(id int64) { p.participantId = id }

func (p *Participant) ConversationId() int64      { return p.conversationId }
func (p *Participant) SetConversationId(id int64) { p.conversationId = id }

func (p *Participant) ProfileId() int64      { return p.profileId }
func (p *Participant) SetProfileId(id int64) { p.profileId = id }

func (p *Participant) Uid() int64      { return p.uid }
func (p *Participant) SetUid(id int64) { p.uid = id }

func (p *Participant) ParticipantRole() string     { return p.participantRole }
func (p *Participant) SetParticipantRole(r string) { p.participantRole = r }

func (p *Participant) LastReadSeqNo() int64     { return p.lastReadSeqNo }
func (p *Participant) SetLastReadSeqNo(n int64) { p.lastReadSeqNo = n }

func (p *Participant) LastReadMessageId() *int64      { return p.lastReadMessageId }
func (p *Participant) SetLastReadMessageId(id *int64) { p.lastReadMessageId = id }

func (p *Participant) LastReadDt() time.Time     { return p.lastReadDt }
func (p *Participant) SetLastReadDt(t time.Time) { p.lastReadDt = t }

func (p *Participant) LastDeliveredSeqNo() int64     { return p.lastDeliveredSeqNo }
func (p *Participant) SetLastDeliveredSeqNo(n int64) { p.lastDeliveredSeqNo = n }

func (p *Participant) UnreadCount() int64     { return p.unreadCount }
func (p *Participant) SetUnreadCount(n int64) { p.unreadCount = n }

func (p *Participant) IsMuted() bool     { return p.isMuted }
func (p *Participant) SetIsMuted(v bool) { p.isMuted = v }

func (p *Participant) MutedUntilDt() time.Time     { return p.mutedUntilDt }
func (p *Participant) SetMutedUntilDt(t time.Time) { p.mutedUntilDt = t }

func (p *Participant) IsPinned() bool     { return p.isPinned }
func (p *Participant) SetIsPinned(v bool) { p.isPinned = v }

func (p *Participant) ClearedBeforeSeqNo() int64     { return p.clearedBeforeSeqNo }
func (p *Participant) SetClearedBeforeSeqNo(n int64) { p.clearedBeforeSeqNo = n }

func (p *Participant) JoinedDt() time.Time     { return p.joinedDt }
func (p *Participant) SetJoinedDt(t time.Time) { p.joinedDt = t }

func (p *Participant) LeftDt() time.Time     { return p.leftDt }
func (p *Participant) SetLeftDt(t time.Time) { p.leftDt = t }

func (p *Participant) InvitedByProfileId() *int64      { return p.invitedByProfileId }
func (p *Participant) SetInvitedByProfileId(id *int64) { p.invitedByProfileId = id }

func (p *Participant) ParticipantStatus() *string     { return p.participantStatus }
func (p *Participant) SetParticipantStatus(s *string) { p.participantStatus = s }
