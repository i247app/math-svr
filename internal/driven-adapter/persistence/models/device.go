package models

import "time"

type DeviceModel struct {
	ID              string
	UID             *string
	DeviceUuid      string
	DeviceName      string
	DevicePushToken *string
	IsVerified      bool
	Status          string
	CreateID        *int64
	CreateDT        time.Time
	ModifyID        *int64
	ModifyDT        time.Time
}
