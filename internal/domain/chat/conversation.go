package chat

import "math-ai.com/math-ai/internal/domain/shared/mtime"

// Conversation models one ma_chat_conversations row — a thread of any shape.
//
// lastSeqNo is the per-conversation allocator behind Message.seqNo; it is
// advanced by IRepository.NextSeqNo inside a transaction, never by setting it
// on the entity.
//
// The lastMessage* fields are a denormalised preview so the conversation-list
// screen renders in one query instead of one query per row. They are written
// by UpdateLastMessage in the same transaction as the message insert.
type Conversation struct {
	id                         int64
	conversationId             int64
	conversationType           string
	classroomId                *int64
	dmKey                      *string
	title                      *string
	avatarKey                  *string
	ownerProfileId             *int64
	participantCount           int64
	lastSeqNo                  int64
	messageCount               int64
	lastMessageId              *int64
	lastMessageSeqNo           *int64
	lastMessageType            *string
	lastMessagePreview         *string
	lastMessageSenderProfileId *int64
	lastMessageDt              mtime.MathTime
	note                       *string
	conversationStatus         *string
	status                     string
	createId                   *int64
	createDt                   mtime.MathTime
	modifyId                   *int64
	modifyDt                   mtime.MathTime
}

func NewConversation() *Conversation { return &Conversation{} }

func (c *Conversation) Id() int64                    { return c.id }
func (c *Conversation) SetId(id int64)               { c.id = id }
func (c *Conversation) ConversationId() int64        { return c.conversationId }
func (c *Conversation) SetConversationId(id int64)   { c.conversationId = id }
func (c *Conversation) ConversationType() string     { return c.conversationType }
func (c *Conversation) SetConversationType(t string) { c.conversationType = t }
func (c *Conversation) ClassroomId() *int64          { return c.classroomId }
func (c *Conversation) SetClassroomId(id *int64)     { c.classroomId = id }
func (c *Conversation) DmKey() *string               { return c.dmKey }
func (c *Conversation) SetDmKey(k *string)           { c.dmKey = k }
func (c *Conversation) Title() *string               { return c.title }
func (c *Conversation) SetTitle(t *string)           { c.title = t }
func (c *Conversation) AvatarKey() *string           { return c.avatarKey }
func (c *Conversation) SetAvatarKey(k *string)       { c.avatarKey = k }
func (c *Conversation) OwnerProfileId() *int64       { return c.ownerProfileId }
func (c *Conversation) SetOwnerProfileId(id *int64)  { c.ownerProfileId = id }
func (c *Conversation) ParticipantCount() int64      { return c.participantCount }
func (c *Conversation) SetParticipantCount(n int64)  { c.participantCount = n }
func (c *Conversation) LastSeqNo() int64             { return c.lastSeqNo }
func (c *Conversation) SetLastSeqNo(n int64)         { c.lastSeqNo = n }
func (c *Conversation) MessageCount() int64          { return c.messageCount }
func (c *Conversation) SetMessageCount(n int64)      { c.messageCount = n }
func (c *Conversation) LastMessageId() *int64        { return c.lastMessageId }
func (c *Conversation) SetLastMessageId(id *int64)   { c.lastMessageId = id }
func (c *Conversation) LastMessageSeqNo() *int64     { return c.lastMessageSeqNo }
func (c *Conversation) SetLastMessageSeqNo(n *int64) { c.lastMessageSeqNo = n }
func (c *Conversation) LastMessageType() *string     { return c.lastMessageType }
func (c *Conversation) SetLastMessageType(t *string) { c.lastMessageType = t }
func (c *Conversation) LastMessagePreview() *string  { return c.lastMessagePreview }
func (c *Conversation) SetLastMessagePreview(p *string) {
	c.lastMessagePreview = p
}
func (c *Conversation) LastMessageSenderProfileId() *int64 {
	return c.lastMessageSenderProfileId
}
func (c *Conversation) SetLastMessageSenderProfileId(id *int64) {
	c.lastMessageSenderProfileId = id
}
func (c *Conversation) LastMessageDt() mtime.MathTime { return c.lastMessageDt }
func (c *Conversation) SetLastMessageDt(t mtime.MathTime) {
	c.lastMessageDt = t
}
func (c *Conversation) Note() *string               { return c.note }
func (c *Conversation) SetNote(n *string)           { c.note = n }
func (c *Conversation) ConversationStatus() *string { return c.conversationStatus }
func (c *Conversation) SetConversationStatus(s *string) {
	c.conversationStatus = s
}
func (c *Conversation) Status() string               { return c.status }
func (c *Conversation) SetStatus(s string)           { c.status = s }
func (c *Conversation) CreateId() *int64             { return c.createId }
func (c *Conversation) SetCreateId(id *int64)        { c.createId = id }
func (c *Conversation) CreateDt() mtime.MathTime     { return c.createDt }
func (c *Conversation) SetCreateDt(t mtime.MathTime) { c.createDt = t }
func (c *Conversation) ModifyId() *int64             { return c.modifyId }
func (c *Conversation) SetModifyId(id *int64)        { c.modifyId = id }
func (c *Conversation) ModifyDt() mtime.MathTime     { return c.modifyDt }
func (c *Conversation) SetModifyDt(t mtime.MathTime) { c.modifyDt = t }
