package domain

import (
	"time"

	"github.com/google/uuid"
	"math-ai.com/math-ai/internal/driven-adapter/persistence/models"
)

type LoginLog struct {
	id         string
	uid        string
	ipaddress  string
	deviceUUID string
	token      string
	status     string
	createDT   time.Time
	modifyDT   time.Time
	deletedDT  *time.Time
}

func NewLoginLogDomain() *LoginLog {
	return &LoginLog{}
}

func (l *LoginLog) ID() string {
	return l.id
}

func (l *LoginLog) SetID(id string) {
	l.id = id
}

func (l *LoginLog) GenerateID() {
	l.id = uuid.New().String()
}

func (l *LoginLog) UID() string {
	return l.uid
}

func (l *LoginLog) SetUID(uid string) {
	l.uid = uid
}

func (l *LoginLog) IPAddress() string {
	return l.ipaddress
}

func (l *LoginLog) SetIPAddress(ipaddress string) {
	l.ipaddress = ipaddress
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

func (l *LoginLog) Status() string {
	return l.status
}

func (l *LoginLog) SetStatus(status string) {
	l.status = status
}

func (l *LoginLog) CreateDT() time.Time {
	return l.createDT
}

func (l *LoginLog) ModifyDT() time.Time {
	return l.modifyDT
}

func (l *LoginLog) SetModifyDT(modifyDT time.Time) {
	l.modifyDT = modifyDT
}

func (l *LoginLog) DeletedDT() *time.Time {
	return l.deletedDT
}

func BuildLoginLogFromModel(model *models.LoginLogModel) *LoginLog {
	return &LoginLog{
		id:         model.ID,
		uid:        model.UID,
		ipaddress:  model.IPaddress,
		deviceUUID: model.DeviceUUID,
		token:      model.Token,
		status:     model.Status,
		createDT:   model.CreateDT,
		modifyDT:   model.ModifyDT,
		deletedDT:  model.DeletedDT,
	}
}
