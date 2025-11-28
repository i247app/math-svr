package dto

import domain "math-ai.com/math-ai/internal/core/domain/device"

type CreateDeviceRequest struct {
	UID             *string `json:"uid"`
	DeviceUUID      string  `json:"device_uuid"`
	DeviceName      string  `json:"device_name"`
	DevicePushToken *string `json:"device_push_token"`
	IsVerified      bool    `json:"is_verified"`
}

type UpdateDeviceRequest struct {
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

func BuildDeviceDomainForCreate(req *CreateDeviceRequest) *domain.Device {
	deviceDomain := domain.NewDeviceDomain()
	deviceDomain.GenerateID()
	deviceDomain.SetUID(req.UID)
	deviceDomain.SetDeviceUuid(req.DeviceUUID)
	deviceDomain.SetDeviceName(req.DeviceName)
	deviceDomain.SetDevicePushToken(req.DevicePushToken)
	deviceDomain.SetIsVerified(req.IsVerified)

	return deviceDomain
}

func BuildDeviceDomainForUpdate(req *UpdateDeviceRequest) *domain.Device {
	deviceDomain := domain.NewDeviceDomain()
	deviceDomain.SetID(req.ID)
	deviceDomain.SetUID(req.UID)

	if req.DeviceUUID != nil {
		deviceDomain.SetDeviceUuid(*req.DeviceUUID)
	}

	if req.DeviceName != nil {
		deviceDomain.SetDeviceName(*req.DeviceName)
	}

	deviceDomain.SetDevicePushToken(req.DevicePushToken)

	if req.IsVerified != nil {
		deviceDomain.SetIsVerified(*req.IsVerified)
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
