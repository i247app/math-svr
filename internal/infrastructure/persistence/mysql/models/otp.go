package models

import (
	"time"

	"github.com/google/uuid"
)

type OtpModel struct {
	Id           int64
	OtpId        uuid.UUID
	OtpType      string
	UserId       *uuid.UUID
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
	CreateId     *uuid.UUID
	CreateDt     time.Time
	ModifyId     *uuid.UUID
	ModifyDt     time.Time
}
