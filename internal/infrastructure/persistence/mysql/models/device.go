package models

import (
	"time"

	"github.com/google/uuid"
)

type DeviceModel struct {
	Id              int64
	DeviceId        uuid.UUID
	UserId          *uuid.UUID
	DeviceUUID      string
	DeviceName      string
	DevicePushToken *string
	IsVerified      bool
	Note            *string
	DeviceStatus    *string
	Status          string
	CreateId        *uuid.UUID
	CreateDt        time.Time
	ModifyId        *uuid.UUID
	ModifyDt        time.Time
}
