package models

import (
	"time"
)

type LoginLogModel struct {
	Id             int64
	LoginLogId     string
	UserId         string
	IpAddress      string
	DeviceUUID     string
	Token          string
	Note           *string
	LoginLogStatus *string
	Status         string
	CreateId       *string
	CreateDt       time.Time
	ModifyId       *string
	ModifyDt       time.Time
}
