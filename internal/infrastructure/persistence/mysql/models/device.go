package models

import (
	"time"
)

type DeviceModel struct {
	Id              int64
	DeviceId        int64
	UserId          *int64
	DeviceUUID      string
	DeviceName      string
	DevicePushToken *string
	IsVerified      bool
	TrustDt         *time.Time
	Note            *string
	DeviceStatus    *string
	Status          string
	CreateId        *int64
	CreateDt        time.Time
	ModifyId        *int64
	ModifyDt        time.Time
}
