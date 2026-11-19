package chat

import (
	"time"

	"math-ai.com/math-ai/internal/domain/shared/audit"
)

// Message models one ma_chat_messages row.
//
// seqNo is the load-bearing field: a per-conversation monotonic counter
// allocated from the conversation row in the same transaction as the insert.
// It is what makes ordering correct when two messages land in the same
// microsecond or client clocks disagree, makes pagination stable while new
// messages arrive, turns unread into a subtraction, and lets a reconnecting
// client say "I have up to 41" so the server can replay 42 onward — which is
// how this feature covers the socket layer's lack of replay.
//
// There is deliberately no file/url field. Attachments live in their own table
// so that a message can carry several of them without changing this shape once
// the table already holds data.
type Message struct {
	audit.Audit

	messageId        int64
	conversationId   int64
	seqNo            int64
	senderProfileId  *int64
	senderUID        *int64
	messageType      string
	content          *string
	attachmentCount  int64
	replyToMessageId *int64
	systemEvent      *string
	systemPayload    *string
	metadata         *string
	clientMsgId      *string
	sentDt           time.Time
	editedDt         time.Time
	revokedDt        time.Time
	messageStatus    *string
}

func NewMessage() *Message { return &Message{} }

func (m *Message) MessageId() int64      { return m.messageId }
func (m *Message) SetMessageId(id int64) { m.messageId = id }

func (m *Message) ConversationId() int64      { return m.conversationId }
func (m *Message) SetConversationId(id int64) { m.conversationId = id }

func (m *Message) SeqNo() int64     { return m.seqNo }
func (m *Message) SetSeqNo(n int64) { m.seqNo = n }

func (m *Message) SenderProfileId() *int64      { return m.senderProfileId }
func (m *Message) SetSenderProfileId(id *int64) { m.senderProfileId = id }

func (m *Message) SenderUID() *int64      { return m.senderUID }
func (m *Message) SetSenderUID(id *int64) { m.senderUID = id }

func (m *Message) MessageType() string     { return m.messageType }
func (m *Message) SetMessageType(t string) { m.messageType = t }

func (m *Message) Content() *string     { return m.content }
func (m *Message) SetContent(c *string) { m.content = c }

func (m *Message) AttachmentCount() int64     { return m.attachmentCount }
func (m *Message) SetAttachmentCount(n int64) { m.attachmentCount = n }

func (m *Message) ReplyToMessageId() *int64      { return m.replyToMessageId }
func (m *Message) SetReplyToMessageId(id *int64) { m.replyToMessageId = id }

func (m *Message) SystemEvent() *string     { return m.systemEvent }
func (m *Message) SetSystemEvent(e *string) { m.systemEvent = e }

func (m *Message) SystemPayload() *string     { return m.systemPayload }
func (m *Message) SetSystemPayload(p *string) { m.systemPayload = p }

func (m *Message) Metadata() *string     { return m.metadata }
func (m *Message) SetMetadata(v *string) { m.metadata = v }

func (m *Message) ClientMsgId() *string     { return m.clientMsgId }
func (m *Message) SetClientMsgId(v *string) { m.clientMsgId = v }

func (m *Message) SentDt() time.Time     { return m.sentDt }
func (m *Message) SetSentDt(t time.Time) { m.sentDt = t }

func (m *Message) EditedDt() time.Time     { return m.editedDt }
func (m *Message) SetEditedDt(t time.Time) { m.editedDt = t }

func (m *Message) RevokedDt() time.Time     { return m.revokedDt }
func (m *Message) SetRevokedDt(t time.Time) { m.revokedDt = t }

func (m *Message) MessageStatus() *string     { return m.messageStatus }
func (m *Message) SetMessageStatus(s *string) { m.messageStatus = s }
