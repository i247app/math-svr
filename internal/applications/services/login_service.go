package services

import (
	"context"
	"errors"
	"fmt"

	"math-ai.com/math-ai/internal/applications/dto"
	"math-ai.com/math-ai/internal/applications/utils"
	"math-ai.com/math-ai/internal/applications/validators"
	diRepo "math-ai.com/math-ai/internal/core/di/repositories"
	diSvc "math-ai.com/math-ai/internal/core/di/services"
	"math-ai.com/math-ai/internal/session"
	"math-ai.com/math-ai/internal/shared/constant/enum"
	"math-ai.com/math-ai/internal/shared/constant/status"
	err_svc "math-ai.com/math-ai/internal/shared/error"
	"math-ai.com/math-ai/internal/shared/logger"
	hasher "math-ai.com/math-ai/internal/shared/utils/hash"
)

type LoginService struct {
	validator       validators.ILoginValidator
	repo            diRepo.ILoginRepository
	userRepo        diRepo.IUserRepository
	responseBuilder *utils.ResponseBuilder
}

func NewLoginService(
	validator validators.ILoginValidator,
	repo diRepo.ILoginRepository,
	userRepo diRepo.IUserRepository,
	storageService diSvc.IStorageService,
) diSvc.ILoginService {
	return &LoginService{
		validator:       validator,
		repo:            repo,
		userRepo:        userRepo,
		responseBuilder: utils.NewResponseBuilder(storageService),
	}
}

func (s *LoginService) Login(ctx context.Context, sess *session.AppSession, req *dto.LoginRequest) (status.Code, *dto.LoginResponse, error) {
	logger := logger.GetLogger(ctx)

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
		return status.FAIL, nil, err
	}
	if user == nil {
		logger.Infof("User not found with login name: %s", req.LoginName)
		return status.LOGIN_WRONG_CREDENTIALS, nil, err_svc.ErrInvalidCredentials
	}

	// compare password
	err = hasher.DefaultHasher.Compare(req.RawPassword, user.Password())
	if err != nil {
		logger.Infof("Error comparing password: %v", err)
		return status.LOGIN_WRONG_CREDENTIALS, nil, err_svc.ErrInvalidCredentials
	}

	// Get auth token
	authTokenRaw, ok := sess.Get("token")
	if !ok {
		return status.FAIL, nil, errors.New("session token not found")
	}
	authToken := authTokenRaw.(string)

	logger.Info("Login successful, updating session data...")
	sess.Init(session.InitData{
		Source:    "login",
		IsSecure:  isSecure,
		UID:       user.ID(),
		Email:     user.Email(),
		LoginName: req.LoginName,
	})

	loginLog, err := s.repo.GetLoginLogByUIDAndDeviceUUID(ctx, user.ID(), req.DeviceUUID)
	if err != nil {
		logger.Infof("Error getting login log: %v", err)
		return status.FAIL, nil, err
	}

	if loginLog != nil {
		updateLoginLogDTO := dto.BuildLoginLogDomainForUpdate(loginLog.ID(), user.ID(), req.IpAddress, req.DeviceUUID, authToken, loginStatus)
		err = s.repo.UpdateLoginLog(ctx, updateLoginLogDTO)
		if err != nil {
			logger.Infof("Error updating login log: %v", err)
			return status.FAIL, nil, err
		}
	} else {
		createLoginLogDTO := dto.BuildLoginLogDomainForCreate(user.ID(), req.IpAddress, req.DeviceUUID, authToken, loginStatus)
		err = s.repo.StoreLoginLog(ctx, createLoginLogDTO)
		if err != nil {
			logger.Infof("Error storing login log: %v", err)
			return status.FAIL, nil, err
		}
	}

	res := &dto.LoginResponse{
		IsSecure:    isSecure,
		Needs2FA:    needs2FA,
		User:        s.responseBuilder.BuildUserResponse(ctx, user),
		LoginStatus: loginStatus,
	}

	return status.SUCCESS, res, nil
}

func (s *LoginService) Logout(ctx context.Context, sess *session.AppSession) (status.Code, error) {
	logger := logger.GetLogger(ctx)

	if sess == nil {
		logger.Error("Session is nil during logout")
		return status.FAIL, fmt.Errorf("failed to logout: session is nil")
	}

	// Invalidate session
	sess.MarkExpired()
	sess.MarkNotSecure()

	return status.SUCCESS, nil
}
