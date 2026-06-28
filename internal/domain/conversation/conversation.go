package conversation

import (
	mtime "math-ai.com/math-ai/internal/domain/shared/time"
)

// Conversation models ma_ai_conversations — one AI chat "session". The
// external conversationId is what the client carries between turns to keep
// context; the server re-sends the windowed history on each turn, so the
// conversation itself holds no message state beyond messageCount (used to
// derive the next message seq_no atomically inside the UoW).
//
// userId is the owner (from the session); profileId is optional. title and
// purpose are nullable and unused in v1 (reserved for AI auto-titling).
type Conversation struct {
	id                 int64
	conversationId     int64
	userId             int64
	profileId          *int64
	title              *string
	purpose            *string
	messageCount       int64
	note               *string
	conversationStatus *string
	status             string
	createId           *int64
	createDt           mtime.MathTime
	modifyId           *int64
	modifyDt           mtime.MathTime
}

func NewConversation() *Conversation {
	return &Conversation{}
}

func (c *Conversation) Id() int64                  { return c.id }
func (c *Conversation) SetId(id int64)             { c.id = id }
func (c *Conversation) ConversationId() int64      { return c.conversationId }
func (c *Conversation) SetConversationId(id int64) { c.conversationId = id }
func (c *Conversation) UserId() int64              { return c.userId }
func (c *Conversation) SetUserId(id int64)         { c.userId = id }
func (c *Conversation) ProfileId() *int64          { return c.profileId }
func (c *Conversation) SetProfileId(id *int64)     { c.profileId = id }
func (c *Conversation) Title() *string             { return c.title }
func (c *Conversation) SetTitle(t *string)         { c.title = t }
func (c *Conversation) Purpose() *string           { return c.purpose }
func (c *Conversation) SetPurpose(p *string)       { c.purpose = p }
func (c *Conversation) MessageCount() int64        { return c.messageCount }
func (c *Conversation) SetMessageCount(n int64)    { c.messageCount = n }
func (c *Conversation) Note() *string              { return c.note }
func (c *Conversation) SetNote(n *string)          { c.note = n }
func (c *Conversation) ConversationStatus() *string {
	return c.conversationStatus
}
func (c *Conversation) SetConversationStatus(v *string) { c.conversationStatus = v }
func (c *Conversation) Status() string                  { return c.status }
func (c *Conversation) SetStatus(v string)              { c.status = v }
func (c *Conversation) CreateId() *int64                { return c.createId }
func (c *Conversation) SetCreateId(id *int64)           { c.createId = id }
func (c *Conversation) CreateDt() mtime.MathTime        { return c.createDt }
func (c *Conversation) SetCreateDt(t mtime.MathTime)    { c.createDt = t }
func (c *Conversation) ModifyId() *int64                { return c.modifyId }
func (c *Conversation) SetModifyId(id *int64)           { c.modifyId = id }
func (c *Conversation) ModifyDt() mtime.MathTime        { return c.modifyDt }
func (c *Conversation) SetModifyDt(t mtime.MathTime)    { c.modifyDt = t }
