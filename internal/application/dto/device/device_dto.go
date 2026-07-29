package device

import (
	"math-ai.com/math-ai/internal/domain/device"
)

// DeviceResponse is the wire shape of a device row. device_uuid is included so
// the client can correlate the response with the local identifier it
// originally sent; we never echo back the push token (treated like a secret).
type DeviceResponse struct {
	DeviceID   int64   `json:"device_id"`
	UserID     *int64  `json:"user_id,omitempty"`
	DeviceUUID string  `json:"device_uuid"`
	DeviceName string  `json:"device_name"`
	Platform   string  `json:"platform"`
	IsVerified bool    `json:"is_verified"`
	Note       *string `json:"note,omitempty"`
	Status     string  `json:"status"`
	CreateDt   string  `json:"create_dt"`
	ModifyDt   string  `json:"modify_dt"`
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
		Platform:   d.Platform(),
		IsVerified: d.IsVerified(),
		Note:       d.Note(),
		Status:     d.Status(),
		CreateDt:   d.CreateDt().String(),
		ModifyDt:   d.ModifyDt().String(),
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
	DeviceID int64 `json:"device_id"`
}

type GetDeviceByIdRes struct {
	Device *DeviceResponse `json:"device"`
}

type ListDevicesReq struct {
	UserID int64 `json:"user_id"`
	// IsVerified is optional. Omitted (or null) → no filter, current
	// behavior preserved. true/false → only devices whose is_verified
	// matches exactly.
	IsVerified *bool `json:"is_verified,omitempty"`
}

type ListDevicesRes struct {
	Devices []*DeviceResponse `json:"devices"`
}

type UpdateDeviceReq struct {
	UserID          int64   `json:"user_id"`
	DeviceID        int64   `json:"device_id"`
	DeviceName      string  `json:"device_name,omitempty"`
	DevicePushToken *string `json:"device_push_token,omitempty"`
	Note            *string `json:"note,omitempty"`
}

type UpdateDeviceRes struct {
	Device *DeviceResponse `json:"device"`
}

type RevokeDeviceReq struct {
	UserID    int64  `json:"user_id"`
	DevicUUID string `json:"device_uuid"`
}

type RevokeDeviceRes struct{}

type DeleteDeviceReq struct {
	UserID   int64 `json:"user_id"`
	DeviceID int64 `json:"device_id"`
}

type DeleteDeviceRes struct{}

// VerifyDeviceReq is the internal-only DTO consumed by the future 2FA flow.
// No HTTP route binds to this — the auth/2fa module will call the underlying
// command directly through the Service. Kept here as the canonical shape so
// the contract is discoverable.
type VerifyDeviceReq struct {
	UserID          int64   `json:"user_id"`
	DeviceUUID      string  `json:"device_uuid"`
	DeviceName      string  `json:"device_name,omitempty"`
	Platform        string  `json:"platform,omitempty"`
	DevicePushToken *string `json:"device_push_token,omitempty"`
}

type VerifyDeviceRes struct {
	Device *DeviceResponse `json:"device"`
}
