package enum

// PresenceStateType is the realtime availability of a user account
// (ma_user_presence.presence_state).
//
// ONLINE means the user holds at least one live WebSocket — NOT that they
// have a valid session. Sessions here live 14 days, so a session-based
// definition would leave nearly every user permanently green and make the
// indicator meaningless. AWAY is reserved for a future idle timer and is
// never written today.
type PresenceStateType string

const (
	PresenceStateOnline  PresenceStateType = "ONLINE"
	PresenceStateAway    PresenceStateType = "AWAY"
	PresenceStateOffline PresenceStateType = "OFFLINE"
)

func (s PresenceStateType) String() string {
	return string(s)
}

func (s PresenceStateType) IsValid() bool {
	switch s {
	case PresenceStateOnline, PresenceStateAway, PresenceStateOffline:
		return true
	default:
		return false
	}
}
