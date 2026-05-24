package models

import (
	"time"
)

type DeviceModel struct {
	Id              int64
	DeviceId        string
	UserId          *string
	DeviceUUID      string
	DeviceName      string
	DevicePushToken *string
	IsVerified      bool
	Note            *string
	DeviceStatus    *string
	Status          string
	CreateId        *string
	CreateDt        time.Time
	ModifyId        *string
	ModifyDt        time.Time
}
