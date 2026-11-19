package models

type LoginLogModel struct {
	Audit

	LoginLogId     int64
	Uid            int64
	IpAddress      string
	DeviceUUID     string
	Token          string
	LoginLogStatus *string
}
