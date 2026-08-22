package models

import "time"

type PresenceModel struct {
	Id              int64
	UserId          int64
	PresenceState   string
	ConnectionCount int64
	LastOnlineDt    *time.Time
	LastSeenDt      *time.Time
	LastDeviceUuid  *string
	LastPlatform    *string
	Note            *string
	Status          string
	CreateId        *int64
	CreateDt        time.Time
	ModifyId        *int64
	ModifyDt        time.Time
}
