package domain

import (
	"github.com/google/uuid"
	"math-ai.com/math-ai/internal/driven-adapter/persistence/models"
	"math-ai.com/math-ai/internal/shared/utils/time"
)

type Device struct {
	id              string
	uid             *string
	deviceUuid      string
	deviceName      string
	devicePushToken *string
	isVerified      bool
	note            *string
	deviceStatus    string
	status          string
	createID        *int64
	createDT        time.MathTime
	modifyID        *int64
	modifyDT        time.MathTime
	deletedDT       *time.MathTime
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

func (d *Device) Note() *string {
	return d.note
}

func (d *Device) SetNote(note *string) {
	d.note = note
}

func (d *Device) DeviceStatus() string {
	return d.deviceStatus
}

func (d *Device) SetDeviceStatus(deviceStatus string) {
	d.deviceStatus = deviceStatus
}

func (d *Device) Status() string {
	return d.status
}

func (d *Device) SetStatus(status string) {
	d.status = status
}

func (d *Device) CreatedBy() *int64 {
	return d.createID
}

func (d *Device) SetCreatedBy(createID *int64) {
	d.createID = createID
}

func (d *Device) CreateDT() time.MathTime {
	return d.createDT
}

func (d *Device) SetCreateDT(createDT time.MathTime) {
	d.createDT = createDT
}

func (d *Device) ModifiedBy() *int64 {
	return d.modifyID
}

func (d *Device) SetModifiedBy(modifyID *int64) {
	d.modifyID = modifyID
}

func (d *Device) ModifyDT() time.MathTime {
	return d.modifyDT
}

func (d *Device) SetModifyDT(modifyDT time.MathTime) {
	d.modifyDT = modifyDT
}

func (d *Device) DeletedDT() *time.MathTime {
	return d.deletedDT
}

func (d *Device) SetDeletedDT(deletedDT *time.MathTime) {
	d.deletedDT = deletedDT
}

func BuildDeviceDomainFromModel(model *models.DeviceModel) *Device {
	return &Device{
		id:              model.ID,
		uid:             model.UID,
		deviceUuid:      model.DeviceUuid,
		deviceName:      model.DeviceName,
		devicePushToken: model.DevicePushToken,
		isVerified:      model.IsVerified,
		note:            model.Note,
		deviceStatus:    model.DeviceStatus,
		status:          model.Status,
		createID:        model.CreateID,
		createDT:        model.CreateDT,
		modifyID:        model.ModifyID,
		modifyDT:        model.ModifyDT,
		deletedDT:       model.DeletedDT,
	}
}
