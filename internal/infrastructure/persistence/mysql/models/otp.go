package models

import (
	"time"
)

type OtpModel struct {
	OtpId         int64
	OtpType       string
	UserId        *int64
	Identifier    string
	DeviceUUID    *string
	DeviceName    *string
	OtpCode       string
	OtpCreateDt   *time.Time
	OtpExpireDt   *time.Time
	OtpVerifiedDt *time.Time
	AttemptCount  int
	RptFlg        *string
	Kwords        *string
	Note          *string
	OtpStatus     *string
	Status        string
	CreateId      *int64
	CreateDt      time.Time
	ModifyId      *int64
	ModifyDt      time.Time
}
