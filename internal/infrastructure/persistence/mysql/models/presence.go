package models

import "time"

type PresenceModel struct {
	UserId          int64
	PresenceState   string
	ConnectionCount int64
	LastOnlineDt    *time.Time
	LastSeenDt      *time.Time
	LastDeviceUuid  *string
	LastPlatform    *string
	RptFlg          *string
	Kwords          *string
	Note            *string
	Status          string
	CreateId        *int64
	CreateDt        time.Time
	ModifyId        *int64
	ModifyDt        time.Time
}
