package models

import (
	"time"
)

type LoginLogModel struct {
	Id             int64
	LoginLogId     int64
	UserId         int64
	IpAddress      string
	DeviceUUID     string
	Token          string
	Note           *string
	LoginLogStatus *string
	Status         string
	CreateId       *int64
	CreateDt       time.Time
	ModifyId       *int64
	ModifyDt       time.Time
}
