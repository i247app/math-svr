package device

import (
	"math-ai.com/math-ai/internal/domain/device"
	mtime "math-ai.com/math-ai/internal/domain/shared/time"
)

// DeviceResponse is the wire shape of a device row. device_uuid is included so
// the client can correlate the response with the local identifier it
// originally sent; we never echo back the push token (treated like a secret).
type DeviceResponse struct {
	DeviceID   string         `json:"device_id"`
	UserID     *string        `json:"user_id,omitempty"`
	DeviceUUID string         `json:"device_uuid"`
	DeviceName string         `json:"device_name"`
	IsVerified bool           `json:"is_verified"`
	Note       *string        `json:"note,omitempty"`
	Status     string         `json:"status"`
	CreateDt   mtime.MathTime `json:"create_dt"`
	ModifyDt   mtime.MathTime `json:"modify_dt"`
}

func DomainToResponse(d *device.Device) *DeviceResponse {
	if d == nil {
		return nil
	}

	return &DeviceResponse{
		DeviceID:   d.DeviceId(),
		UserID:     d.UserId(),
		DeviceUUID: d.DeviceUUID(),
		DeviceName: d.DeviceName(),
		IsVerified: d.IsVerified(),
		Note:       d.Note(),
		Status:     d.Status(),
		CreateDt:   d.CreateDt(),
		ModifyDt:   d.ModifyDt(),
	}
}

func DomainListToResponse(ds []*device.Device) []*DeviceResponse {
	out := make([]*DeviceResponse, 0, len(ds))
	for _, d := range ds {
		out = append(out, DomainToResponse(d))
	}
	return out
}

type GetDeviceByIdReq struct {
	DeviceID string `json:"device_id"`
}

type GetDeviceByIdRes struct {
	Device *DeviceResponse `json:"device"`
}

type ListDevicesReq struct {
	UserID string `json:"user_id"`
}

type ListDevicesRes struct {
	Devices []*DeviceResponse `json:"devices"`
}

type UpdateDeviceReq struct {
	UserID          string  `json:"user_id"`
	DeviceID        string  `json:"device_id"`
	DeviceName      string  `json:"device_name,omitempty"`
	DevicePushToken *string `json:"device_push_token,omitempty"`
	Note            *string `json:"note,omitempty"`
}

type UpdateDeviceRes struct {
	Device *DeviceResponse `json:"device"`
}

type RevokeDeviceReq struct {
	UserID   string `json:"user_id"`
	DeviceID string `json:"device_id"`
}

type RevokeDeviceRes struct{}

type DeleteDeviceReq struct {
	UserID   string `json:"user_id"`
	DeviceID string `json:"device_id"`
}

type DeleteDeviceRes struct{}

// VerifyDeviceReq is the internal-only DTO consumed by the future 2FA flow.
// No HTTP route binds to this — the auth/2fa module will call the underlying
// command directly through the Service. Kept here as the canonical shape so
// the contract is discoverable.
type VerifyDeviceReq struct {
	UserID   string `json:"user_id"`
	DeviceID string `json:"device_id"`
}

type VerifyDeviceRes struct {
	Device *DeviceResponse `json:"device"`
}
