package models

import (
	"time"
)

type OtpModel struct {
	Audit

	OtpId         int64
	OtpType       string
	Uid           *int64
	Identifier    string
	DeviceUUID    *string
	DeviceName    *string
	OtpCode       string
	OtpCreateDt   *time.Time
	OtpExpireDt   *time.Time
	OtpVerifiedDt *time.Time
	AttemptCount  int
	OtpStatus     *string
}
