package models

import (
	"time"
)

type DeviceModel struct {
	Audit

	DeviceId        int64
	Uid             *int64
	DeviceUUID      string
	DeviceName      string
	Platform        string
	DevicePushToken *string
	IsVerified      bool
	TrustDt         *time.Time
	DeviceStatus    *string
}
