package models

import (
	"time"

	"github.com/google/uuid"
)

type LoginLogModel struct {
	Id             int64
	LoginLogId     uuid.UUID
	UserId         uuid.UUID
	IpAddress      string
	DeviceUUID     string
	Token          string
	Note           *string
	LoginLogStatus *string
	Status         string
	CreateId       *uuid.UUID
	CreateDt       time.Time
	ModifyId       *uuid.UUID
	ModifyDt       time.Time
}
