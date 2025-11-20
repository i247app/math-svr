package domain

import (
	"github.com/google/uuid"
	"math-ai.com/math-ai/internal/driven-adapter/persistence/models"
)

type Device struct {
	id              string
	uid             *string
	deviceUuid      string
	deviceName      string
	devicePushToken *string
	isVerified      bool
}

func NewDeviceDomain() *Device {
	return &Device{}
}

func (d *Device) ID() string {
	return d.id
}

func (u *Device) GenerateID() {
	u.id = uuid.New().String()
}

func (d *Device) SetID(id string) {
	d.id = id
}

func (d *Device) UID() *string {
	return d.uid
}

func (d *Device) SetUID(uid *string) {
	d.uid = uid
}

func (d *Device) DeviceUuid() string {
	return d.deviceUuid
}

func (d *Device) SetDeviceUuid(deviceUuid string) {
	d.deviceUuid = deviceUuid
}

func (d *Device) DeviceName() string {
	return d.deviceName
}

func (d *Device) SetDeviceName(deviceName string) {
	d.deviceName = deviceName
}

func (d *Device) DevicePushToken() *string {
	return d.devicePushToken
}

func (d *Device) SetDevicePushToken(devicePushToken *string) {
	d.devicePushToken = devicePushToken
}

func (d *Device) IsVerified() bool {
	return d.isVerified
}

func (d *Device) SetIsVerified(isVerified bool) {
	d.isVerified = isVerified
}

func BuildDeviceDomainFromModel(model *models.DeviceModel) *Device {
	return &Device{
		id:              model.ID,
		uid:             model.UID,
		deviceUuid:      model.DeviceUuid,
		deviceName:      model.DeviceName,
		devicePushToken: model.DevicePushToken,
		isVerified:      model.IsVerified,
	}
}
