package models

import "math-ai.com/math-ai/internal/shared/utils/time"

type DeviceModel struct {
	ID              string
	UID             *string
	DeviceUuid      string
	DeviceName      string
	DevicePushToken *string
	IsVerified      bool
	Note            *string
	DeviceStatus    string
	Status          string
	CreateID        *int64
	CreateDT        time.MathTime
	ModifyID        *int64
	ModifyDT        time.MathTime
	DeletedDT       *time.MathTime
}
