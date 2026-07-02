package conversation

import (
	"math-ai.com/math-ai/internal/domain/shared/mtime"
)

// Message models ma_ai_messages — one turn in a conversation. role is
// stored as a plain string (faithful to the VARCHAR column); the module
// layer validates it against enum.ConversationRoleType. seqNo is the
// strictly-increasing ordinal within a conversation, derived from the
// parent's messageCount inside the UoW so concurrent turns cannot collide.
type Message struct {
	id             int64
	messageId      int64
	conversationId int64
	role           string
	content        string
	seqNo          int64
	note           *string
	status         string
	createId       *int64
	createDt       mtime.MathTime
	modifyId       *int64
	modifyDt       mtime.MathTime
}

func NewMessage() *Message {
	return &Message{}
}

func (m *Message) Id() int64                  { return m.id }
func (m *Message) SetId(id int64)             { m.id = id }
func (m *Message) MessageId() int64           { return m.messageId }
func (m *Message) SetMessageId(id int64)      { m.messageId = id }
func (m *Message) ConversationId() int64      { return m.conversationId }
func (m *Message) SetConversationId(id int64) { m.conversationId = id }
func (m *Message) Role() string               { return m.role }
func (m *Message) SetRole(r string)           { m.role = r }
func (m *Message) Content() string            { return m.content }
func (m *Message) SetContent(c string)        { m.content = c }
func (m *Message) SeqNo() int64               { return m.seqNo }
func (m *Message) SetSeqNo(n int64)           { m.seqNo = n }
func (m *Message) Note() *string              { return m.note }
func (m *Message) SetNote(n *string)          { m.note = n }
func (m *Message) Status() string             { return m.status }
func (m *Message) SetStatus(v string)         { m.status = v }
func (m *Message) CreateId() *int64           { return m.createId }
func (m *Message) SetCreateId(id *int64)      { m.createId = id }
func (m *Message) CreateDt() mtime.MathTime   { return m.createDt }
func (m *Message) SetCreateDt(t mtime.MathTime) {
	m.createDt = t
}
func (m *Message) ModifyId() *int64         { return m.modifyId }
func (m *Message) SetModifyId(id *int64)    { m.modifyId = id }
func (m *Message) ModifyDt() mtime.MathTime { return m.modifyDt }
func (m *Message) SetModifyDt(t mtime.MathTime) {
	m.modifyDt = t
}
