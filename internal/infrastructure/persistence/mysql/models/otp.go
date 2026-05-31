package models

import (
	"time"
)

type OtpModel struct {
	Id           int64
	OtpId        int64
	OtpType      string
	UserId       *int64
	Identifier   string
	DeviceUUID   *string
	DeviceName   *string
	OtpCode      string
	OtpCreateDt  *time.Time
	OtpExpireDt  *time.Time
	AttemptCount int
	Note         *string
	OtpStatus    *string
	Status       string
	CreateId     *int64
	CreateDt     time.Time
	ModifyId     *int64
	ModifyDt     time.Time
}
