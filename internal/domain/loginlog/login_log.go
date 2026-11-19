package loginlog

import (
	"math-ai.com/math-ai/internal/domain/shared/audit"
)

type LoginLog struct {
	audit.Audit

	loginLogId     int64
	uid            int64
	ipAddress      string
	deviceUUID     string
	token          string
	loginLogStatus *string
}

func NewLoginLog() *LoginLog {
	return &LoginLog{}
}

func (l *LoginLog) LoginLogId() int64              { return l.loginLogId }
func (l *LoginLog) SetLoginLogId(loginLogId int64) { l.loginLogId = loginLogId }

func (l *LoginLog) UID() int64       { return l.uid }
func (l *LoginLog) SetUID(uid int64) { l.uid = uid }

func (l *LoginLog) IpAddress() string             { return l.ipAddress }
func (l *LoginLog) SetIpAddress(ipAddress string) { l.ipAddress = ipAddress }

func (l *LoginLog) DeviceUUID() string              { return l.deviceUUID }
func (l *LoginLog) SetDeviceUUID(deviceUUID string) { l.deviceUUID = deviceUUID }

func (l *LoginLog) Token() string         { return l.token }
func (l *LoginLog) SetToken(token string) { l.token = token }

func (l *LoginLog) LoginLogStatus() *string                  { return l.loginLogStatus }
func (l *LoginLog) SetLoginLogStatus(loginLogStatus *string) { l.loginLogStatus = loginLogStatus }
