package notification

import (
	"math-ai.com/math-ai/internal/domain/shared/mtime"
)

// Notification models ma_notifications — one persisted in-app notification
// for a single recipient (user_id). It is both an inbox/feed record (title,
// shortText, isRead, category) and the source payload for a push delivery via
// FCM.
//
// actionType / actionData drive client-side deep-linking; actionData holds an
// opaque JSON blob (mirrors how quiz questions/answers are stored — parsed at
// the application layer, never relationally modelled here).
type Notification struct {
	id                 int64
	notificationId     int64 // external id (minted via Seq.Next)
	userId             int64 // recipient user_id
	title              string
	shortText          string
	category           *string // INFO, WARNING, ERROR
	isRead             bool
	actionType         *string
	actionData         *string // opaque JSON
	priority           *string // LOW, NORMAL, HIGH
	note               *string
	notificationStatus *string // ACTIVE, ARCHIVED, DELETED
	status             string
	createId           *int64
	createDt           mtime.MathTime
	modifyId           *int64
	modifyDt           mtime.MathTime
}

func NewNotification() *Notification { return &Notification{} }

func (n *Notification) Id() int64                       { return n.id }
func (n *Notification) SetId(id int64)                  { n.id = id }
func (n *Notification) NotificationId() int64           { return n.notificationId }
func (n *Notification) SetNotificationId(id int64)      { n.notificationId = id }
func (n *Notification) UserId() int64                   { return n.userId }
func (n *Notification) SetUserId(userId int64)          { n.userId = userId }
func (n *Notification) Title() string                   { return n.title }
func (n *Notification) SetTitle(t string)               { n.title = t }
func (n *Notification) ShortText() string               { return n.shortText }
func (n *Notification) SetShortText(s string)           { n.shortText = s }
func (n *Notification) Category() *string               { return n.category }
func (n *Notification) SetCategory(c *string)           { n.category = c }
func (n *Notification) IsRead() bool                    { return n.isRead }
func (n *Notification) SetIsRead(v bool)                { n.isRead = v }
func (n *Notification) ActionType() *string             { return n.actionType }
func (n *Notification) SetActionType(a *string)         { n.actionType = a }
func (n *Notification) ActionData() *string             { return n.actionData }
func (n *Notification) SetActionData(d *string)         { n.actionData = d }
func (n *Notification) Priority() *string               { return n.priority }
func (n *Notification) SetPriority(p *string)           { n.priority = p }
func (n *Notification) Note() *string                   { return n.note }
func (n *Notification) SetNote(note *string)            { n.note = note }
func (n *Notification) NotificationStatus() *string     { return n.notificationStatus }
func (n *Notification) SetNotificationStatus(s *string) { n.notificationStatus = s }
func (n *Notification) Status() string                  { return n.status }
func (n *Notification) SetStatus(s string)              { n.status = s }
func (n *Notification) CreateId() *int64                { return n.createId }
func (n *Notification) SetCreateId(id *int64)           { n.createId = id }
func (n *Notification) CreateDt() mtime.MathTime        { return n.createDt }
func (n *Notification) SetCreateDt(t mtime.MathTime)    { n.createDt = t }
func (n *Notification) ModifyId() *int64                { return n.modifyId }
func (n *Notification) SetModifyId(id *int64)           { n.modifyId = id }
func (n *Notification) ModifyDt() mtime.MathTime        { return n.modifyDt }
func (n *Notification) SetModifyDt(t mtime.MathTime)    { n.modifyDt = t }
