package dto

import (
	domain "math-ai.com/math-ai/internal/core/domain/login"
	"math-ai.com/math-ai/internal/shared/constant/enum"
)

func BuildLoginLogDomainForCreate(uid, ipAddress, deviceUUID, token string, loginStatus enum.ELoginStatus) *domain.LoginLog {
	loginLogDomain := domain.NewLoginLogDomain()
	loginLogDomain.GenerateID()
	loginLogDomain.SetUID(uid)
	loginLogDomain.SetIPAddress(ipAddress)
	loginLogDomain.SetDeviceUUID(deviceUUID)
	loginLogDomain.SetToken(token)
	loginLogDomain.SetStatus(string(loginStatus))

	return loginLogDomain
}

func BuildLoginLogDomainForUpdate(id, uid, ipAddress, deviceUUID, token string, loginStatus enum.ELoginStatus) *domain.LoginLog {
	loginLogDomain := domain.NewLoginLogDomain()
	loginLogDomain.SetID(id)
	loginLogDomain.SetUID(uid)
	loginLogDomain.SetIPAddress(ipAddress)
	loginLogDomain.SetDeviceUUID(deviceUUID)
	loginLogDomain.SetToken(token)
	loginLogDomain.SetStatus(string(loginStatus))

	return loginLogDomain
}
