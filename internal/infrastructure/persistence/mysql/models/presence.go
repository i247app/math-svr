package models

import "time"

type PresenceModel struct {
	Audit

	Uid             int64
	PresenceState   string
	ConnectionCount int64
	LastOnlineDt    *time.Time
	LastSeenDt      *time.Time
	LastDeviceUuid  *string
	LastPlatform    *string
}
