package services

import (
	"context"

	"math-ai.com/math-ai/internal/applications/dto"
	"math-ai.com/math-ai/internal/applications/validators"
	"math-ai.com/math-ai/internal/core/di/repositories"
	"math-ai.com/math-ai/internal/session"
	"math-ai.com/math-ai/internal/shared/constant/enum"
	"math-ai.com/math-ai/internal/shared/constant/status"
	err_svc "math-ai.com/math-ai/internal/shared/error"
	hasher "math-ai.com/math-ai/internal/shared/utils/hash"
)

type LoginService struct {
	validator validators.ILoginValidator
	repo      repositories.ILoginRepository
	userRepo  repositories.IUserRepository
}

func NewLoginService(
	validator validators.ILoginValidator,
	repo repositories.ILoginRepository,
	userRepo repositories.IUserRepository,
) *LoginService {
	return &LoginService{
		validator: validator,
		repo:      repo,
		userRepo:  userRepo,
	}
}

func (s *LoginService) Login(ctx context.Context, sess *session.AppSession, req *dto.LoginRequest) (status.Code, *dto.LoginResponse, error) {
	// Validate request
	if statusCode, err := s.validator.ValidateLoginRequest(req); err != nil {
		return statusCode, nil, err
	}

	var (
		isSecure    bool              = true
		needs2FA    bool              = false
		loginStatus enum.ELoginStatus = enum.LoginStatusActive
	)

	// get user by login name
	user, err := s.userRepo.GetUserByLoginName(ctx, req.LoginName)
	if err != nil {
		return status.INTERNAL, nil, err
	}
	if user == nil {
		//logger.Info("User not found with login name: %s", req.LoginName)
		return status.LOGIN_WRONG_CREDENTIALS, nil, err_svc.ErrInvalidCredentials
	}

	// compare password
	err = hasher.DefaultHasher.Compare(req.RawPassword, user.Password())
	if err != nil {
		//logger.Info("Error comparing password: %v", err)
		return status.LOGIN_WRONG_CREDENTIALS, nil, err_svc.ErrInvalidCredentials
	}

	authToken := "1234"

	loginLog, err := s.repo.GetLoginLogByUIDAndDeviceUUID(ctx, user.ID(), req.DeviceUUID)
	if err != nil {
		//logger.Info("Error getting login log: %v", err)
		return status.INTERNAL, nil, err
	}

	if loginLog != nil {
		updateLoginLogDTO := dto.BuildLoginLogDomainForUpdate(loginLog.ID(), user.ID(), req.IpAddress, req.DeviceUUID, authToken, loginStatus)
		err = s.repo.UpdateLoginLog(ctx, updateLoginLogDTO)
		if err != nil {
			//logger.Info("Error updating login log: %v", err)
			return status.INTERNAL, nil, err
		}
	} else {
		createLoginLogDTO := dto.BuildLoginLogDomainForCreate(user.ID(), req.IpAddress, req.DeviceUUID, authToken, loginStatus)
		err = s.repo.StoreLoginLog(ctx, createLoginLogDTO)
		if err != nil {
			//logger.Info("Error storing login log: %v", err)
			return status.INTERNAL, nil, err
		}
	}

	res := &dto.LoginResponse{
		IsSecure:    isSecure,
		Needs2FA:    needs2FA,
		User:        dto.UserResponseFromDomain(user),
		LoginStatus: loginStatus,
	}

	return status.SUCCESS, res, nil
}
