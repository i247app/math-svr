package notification

import (
	"math-ai.com/math-ai/internal/domain/shared/audit"
)

// Notification models ma_notifications — one persisted in-app notification
// for a single recipient (uid). It is both an inbox/feed record (title,
// shortText, isRead, category) and the source payload for a push delivery via
// FCM.
//
// actionType / actionData drive client-side deep-linking; actionData holds an
// opaque JSON blob (mirrors how exercise questions/answers are stored — parsed at
// the application layer, never relationally modelled here).
type Notification struct {
	audit.Audit

	notificationId     int64 // external id (minted via Seq.Next)
	uid                int64 // recipient uid
	title              string
	shortText          string
	category           *string // INFO, WARNING, ERROR
	isRead             bool
	actionType         *string
	actionData         *string // opaque JSON
	priority           *string // LOW, NORMAL, HIGH
	notificationStatus *string // ACTIVE, ARCHIVED, DELETED
}

func NewNotification() *Notification { return &Notification{} }

func (n *Notification) NotificationId() int64      { return n.notificationId }
func (n *Notification) SetNotificationId(id int64) { n.notificationId = id }

func (n *Notification) UID() int64       { return n.uid }
func (n *Notification) SetUID(uid int64) { n.uid = uid }

func (n *Notification) Title() string     { return n.title }
func (n *Notification) SetTitle(t string) { n.title = t }

func (n *Notification) ShortText() string     { return n.shortText }
func (n *Notification) SetShortText(s string) { n.shortText = s }

func (n *Notification) Category() *string     { return n.category }
func (n *Notification) SetCategory(c *string) { n.category = c }

func (n *Notification) IsRead() bool     { return n.isRead }
func (n *Notification) SetIsRead(v bool) { n.isRead = v }

func (n *Notification) ActionType() *string     { return n.actionType }
func (n *Notification) SetActionType(a *string) { n.actionType = a }

func (n *Notification) ActionData() *string     { return n.actionData }
func (n *Notification) SetActionData(d *string) { n.actionData = d }

func (n *Notification) Priority() *string     { return n.priority }
func (n *Notification) SetPriority(p *string) { n.priority = p }

func (n *Notification) NotificationStatus() *string     { return n.notificationStatus }
func (n *Notification) SetNotificationStatus(s *string) { n.notificationStatus = s }
