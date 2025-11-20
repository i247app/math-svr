package dto

import domain "math-ai.com/math-ai/internal/core/domain/device"

type CreateDeviceReq struct {
	UID             *string `json:"uid"`
	DeviceUUID      string  `json:"device_uuid"`
	DeviceName      string  `json:"device_name"`
	DevicePushToken *string `json:"device_push_token"`
	IsVerified      bool    `json:"is_verified"`
}

type UpdateDeviceReq struct {
	ID              string  `json:"id"`
	UID             *string `json:"uid"`
	DeviceUUID      *string `json:"device_uuid"`
	DeviceName      *string `json:"device_name"`
	DevicePushToken *string `json:"device_push_token"`
	IsVerified      *bool   `json:"is_verified"`
}

type DeviceResponse struct {
	ID              string  `json:"id"`
	UID             *string `json:"uid"`
	DeviceUUID      string  `json:"device_uuid"`
	DeviceName      string  `json:"device_name"`
	DevicePushToken *string `json:"device_push_token"`
	IsVerified      bool    `json:"is_verified"`
}

func BuildDeviceDomainForCreate(dto *CreateDeviceReq) *domain.Device {
	deviceDomain := domain.NewDeviceDomain()
	deviceDomain.GenerateID()
	deviceDomain.SetUID(dto.UID)
	deviceDomain.SetDeviceUuid(dto.DeviceUUID)
	deviceDomain.SetDeviceName(dto.DeviceName)
	deviceDomain.SetDevicePushToken(dto.DevicePushToken)
	deviceDomain.SetIsVerified(dto.IsVerified)

	return deviceDomain
}

func BuildDeviceDomainForUpdate(dto *UpdateDeviceReq) *domain.Device {
	deviceDomain := domain.NewDeviceDomain()
	deviceDomain.SetID(dto.ID)
	deviceDomain.SetUID(dto.UID)

	if dto.DeviceUUID != nil {
		deviceDomain.SetDeviceUuid(*dto.DeviceUUID)
	}

	if dto.DeviceName != nil {
		deviceDomain.SetDeviceName(*dto.DeviceName)
	}

	deviceDomain.SetDevicePushToken(dto.DevicePushToken)

	if dto.IsVerified != nil {
		deviceDomain.SetIsVerified(*dto.IsVerified)
	}

	return deviceDomain
}

func DeviceResponseFromDomain(deviceDomain *domain.Device) *DeviceResponse {
	return &DeviceResponse{
		ID:              deviceDomain.ID(),
		UID:             deviceDomain.UID(),
		DeviceUUID:      deviceDomain.DeviceUuid(),
		DeviceName:      deviceDomain.DeviceName(),
		DevicePushToken: deviceDomain.DevicePushToken(),
		IsVerified:      deviceDomain.IsVerified(),
	}
}
