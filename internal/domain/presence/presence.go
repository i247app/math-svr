// Package presence models a user account's realtime availability — the
// online/offline dot shown next to a classroom member.
//
// Keyed by user_id rather than profile_id because a WebSocket belongs to an
// account, not to one of the profiles that account holds. Read paths that
// start from a profile (a classroom member list) map profile → user through
// ma_profiles.user_id before looking presence up.
package presence

import "math-ai.com/math-ai/internal/domain/shared/mtime"

// Presence models one ma_user_presence row.
//
// connectionCount, not a boolean, is what makes the state correct when one
// person has the app open on a phone and a tablet: the row only flips to
// OFFLINE when the last connection goes away.
type Presence struct {
	id              int64
	userId          int64
	presenceState   string
	connectionCount int64
	lastOnlineDt    mtime.MathTime
	lastSeenDt      mtime.MathTime
	lastDeviceUuid  *string
	lastPlatform    *string
	note            *string
	status          string
	createId        *int64
	createDt        mtime.MathTime
	modifyId        *int64
	modifyDt        mtime.MathTime
}

func NewPresence() *Presence { return &Presence{} }

func (p *Presence) Id() int64                    { return p.id }
func (p *Presence) SetId(id int64)               { p.id = id }
func (p *Presence) UserId() int64                { return p.userId }
func (p *Presence) SetUserId(id int64)           { p.userId = id }
func (p *Presence) PresenceState() string        { return p.presenceState }
func (p *Presence) SetPresenceState(s string)    { p.presenceState = s }
func (p *Presence) ConnectionCount() int64       { return p.connectionCount }
func (p *Presence) SetConnectionCount(c int64)   { p.connectionCount = c }
func (p *Presence) LastOnlineDt() mtime.MathTime { return p.lastOnlineDt }
func (p *Presence) SetLastOnlineDt(t mtime.MathTime) {
	p.lastOnlineDt = t
}
func (p *Presence) LastSeenDt() mtime.MathTime { return p.lastSeenDt }
func (p *Presence) SetLastSeenDt(t mtime.MathTime) {
	p.lastSeenDt = t
}
func (p *Presence) LastDeviceUuid() *string     { return p.lastDeviceUuid }
func (p *Presence) SetLastDeviceUuid(u *string) { p.lastDeviceUuid = u }
func (p *Presence) LastPlatform() *string       { return p.lastPlatform }
func (p *Presence) SetLastPlatform(v *string)   { p.lastPlatform = v }
func (p *Presence) Note() *string               { return p.note }
func (p *Presence) SetNote(n *string)           { p.note = n }
func (p *Presence) Status() string              { return p.status }
func (p *Presence) SetStatus(v string)          { p.status = v }
func (p *Presence) CreateId() *int64            { return p.createId }
func (p *Presence) SetCreateId(id *int64)       { p.createId = id }
func (p *Presence) CreateDt() mtime.MathTime    { return p.createDt }
func (p *Presence) SetCreateDt(t mtime.MathTime) {
	p.createDt = t
}
func (p *Presence) ModifyId() *int64      { return p.modifyId }
func (p *Presence) SetModifyId(id *int64) { p.modifyId = id }
func (p *Presence) ModifyDt() mtime.MathTime {
	return p.modifyDt
}
func (p *Presence) SetModifyDt(t mtime.MathTime) { p.modifyDt = t }

// IsOnline reports whether the account currently holds a live connection.
// Read paths should call this rather than comparing presenceState strings,
// so an added AWAY state does not silently read as offline everywhere.
func (p *Presence) IsOnline() bool { return p.connectionCount > 0 }
