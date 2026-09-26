package loginlog

import (
	"math-ai.com/math-ai/internal/domain/shared/mtime"
)

type LoginLog struct {
	loginLogId     int64
	userId         int64
	ipAddress      string
	deviceUUID     string
	token          string
	rptFlg         *string
	kwords         *string
	note           *string
	loginLogStatus *string
	status         string
	createId       *int64
	createDt       mtime.MathTime
	modifyId       *int64
	modifyDt       mtime.MathTime
}

func NewLoginLog() *LoginLog {
	return &LoginLog{}
}

func (l *LoginLog) LoginLogId() int64 {
	return l.loginLogId
}

func (l *LoginLog) SetLoginLogId(loginLogId int64) {
	l.loginLogId = loginLogId
}

func (l *LoginLog) UserId() int64 {
	return l.userId
}

func (l *LoginLog) SetUserId(userId int64) {
	l.userId = userId
}

func (l *LoginLog) IpAddress() string {
	return l.ipAddress
}

func (l *LoginLog) SetIpAddress(ipAddress string) {
	l.ipAddress = ipAddress
}

func (l *LoginLog) DeviceUUID() string {
	return l.deviceUUID
}

func (l *LoginLog) SetDeviceUUID(deviceUUID string) {
	l.deviceUUID = deviceUUID
}

func (l *LoginLog) Token() string {
	return l.token
}

func (l *LoginLog) SetToken(token string) {
	l.token = token
}

func (l *LoginLog) Note() *string {
	return l.note
}

func (l *LoginLog) SetNote(note *string) {
	l.note = note
}

// RptFlg is the client-side report flag (rpt_flg). Nil means "not reported".
func (l *LoginLog) RptFlg() *string {
	return l.rptFlg
}

func (l *LoginLog) SetRptFlg(rptFlg *string) {
	l.rptFlg = rptFlg
}

// Kwords holds search keywords (kwords) for a future text-search index.
func (l *LoginLog) Kwords() *string {
	return l.kwords
}

func (l *LoginLog) SetKwords(kwords *string) {
	l.kwords = kwords
}

func (l *LoginLog) LoginLogStatus() *string {
	return l.loginLogStatus
}

func (l *LoginLog) SetLoginLogStatus(loginLogStatus *string) {
	l.loginLogStatus = loginLogStatus
}

func (l *LoginLog) Status() string {
	return l.status
}

func (l *LoginLog) SetStatus(status string) {
	l.status = status
}

func (l *LoginLog) CreateId() *int64 {
	return l.createId
}

func (l *LoginLog) SetCreateId(createId *int64) {
	l.createId = createId
}

func (l *LoginLog) CreateDt() mtime.MathTime {
	return l.createDt
}

func (l *LoginLog) SetCreateDt(createDt mtime.MathTime) {
	l.createDt = createDt
}

func (l *LoginLog) ModifyId() *int64 {
	return l.modifyId
}

func (l *LoginLog) SetModifyId(modifyId *int64) {
	l.modifyId = modifyId
}

func (l *LoginLog) ModifyDt() mtime.MathTime {
	return l.modifyDt
}

func (l *LoginLog) SetModifyDt(modifyDt mtime.MathTime) {
	l.modifyDt = modifyDt
}
