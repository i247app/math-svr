package models

import (
	"time"
)

type OtpModel struct {
	Id           int64
	OtpId        string
	OtpType      string
	UserId       *string
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
	CreateId     *string
	CreateDt     time.Time
	ModifyId     *string
	ModifyDt     time.Time
}
