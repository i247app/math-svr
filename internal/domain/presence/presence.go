// Package presence models a user account's realtime availability — the
// online/offline dot shown next to a classroom member.
//
// Keyed by uid rather than profile_id because a WebSocket belongs to an
// account, not to one of the profiles that account holds. Read paths that
// start from a profile (a classroom member list) map profile → user through
// ma_profiles.uid before looking presence up.
package presence

import (
	"time"

	"math-ai.com/math-ai/internal/domain/shared/audit"
)

// Presence models one ma_user_presence row.
//
// connectionCount, not a boolean, is what makes the state correct when one
// person has the app open on a phone and a tablet: the row only flips to
// OFFLINE when the last connection goes away.
type Presence struct {
	audit.Audit

	uid             int64
	presenceState   string
	connectionCount int64
	lastOnlineDt    time.Time
	lastSeenDt      time.Time
	lastDeviceUuid  *string
	lastPlatform    *string
}

func NewPresence() *Presence { return &Presence{} }

func (p *Presence) UID() int64      { return p.uid }
func (p *Presence) SetUID(id int64) { p.uid = id }

func (p *Presence) PresenceState() string     { return p.presenceState }
func (p *Presence) SetPresenceState(s string) { p.presenceState = s }

func (p *Presence) ConnectionCount() int64     { return p.connectionCount }
func (p *Presence) SetConnectionCount(c int64) { p.connectionCount = c }

func (p *Presence) LastOnlineDt() time.Time     { return p.lastOnlineDt }
func (p *Presence) SetLastOnlineDt(t time.Time) { p.lastOnlineDt = t }

func (p *Presence) LastSeenDt() time.Time     { return p.lastSeenDt }
func (p *Presence) SetLastSeenDt(t time.Time) { p.lastSeenDt = t }

func (p *Presence) LastDeviceUuid() *string     { return p.lastDeviceUuid }
func (p *Presence) SetLastDeviceUuid(u *string) { p.lastDeviceUuid = u }

func (p *Presence) LastPlatform() *string     { return p.lastPlatform }
func (p *Presence) SetLastPlatform(v *string) { p.lastPlatform = v }

// IsOnline reports whether the account currently holds a live connection.
// Read paths should call this rather than comparing presenceState strings,
// so an added AWAY state does not silently read as offline everywhere.
func (p *Presence) IsOnline() bool { return p.connectionCount > 0 }
