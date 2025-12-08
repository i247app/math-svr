package dto

import (
	domain "math-ai.com/math-ai/internal/core/domain/login"
	"math-ai.com/math-ai/internal/shared/constant/enum"
)

type LoginLog struct {
	ID         string `json:"id"`
	UID        string `json:"uid"`
	IpAddress  string `json:"ip_address"`
	DeviceUUID string `json:"device_uuid"`
	Token      string `json:"token"`
}

type LoginRequest struct {
	LoginName   string `json:"login_name"`
	RawPassword string `json:"password"`

	DeviceUUID string `json:"device_uuid,omitempty"`
	DeviceName string `json:"device_name,omitempty"`
	IpAddress  string
}

type LoginResponse struct {
	User        *UserResponse     `json:"user"`
	Needs2FA    bool              `json:"needs_2fa"`
	IsSecure    bool              `json:"is_secure"`
	LoginStatus enum.ELoginStatus `json:"login_status"`
}

func BuildLoginDomain(uid string, password string) *domain.Login {
	loginDomain := domain.NewLoginDomain()
	loginDomain.GenerateID()
	loginDomain.SetUID(uid)
	loginDomain.SetHassPass(password)

	return loginDomain
}
