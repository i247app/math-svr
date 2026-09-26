package models

import (
	"time"
)

type DeviceModel struct {
	DeviceId        int64
	UserId          *int64
	DeviceUUID      string
	DeviceName      string
	Platform        string
	DevicePushToken *string
	IsVerified      bool
	TrustDt         *time.Time
	RptFlg          *string
	Kwords          *string
	Note            *string
	DeviceStatus    *string
	Status          string
	CreateId        *int64
	CreateDt        time.Time
	ModifyId        *int64
	ModifyDt        time.Time
}
