package models

import "time"

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
	CreateDT        time.Time
	ModifyID        *int64
	ModifyDT        time.Time
	DeletedDT       *time.Time
}
