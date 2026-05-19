package auth

import (
	"context"

	command "math-ai.com/math-ai/internal/application/command/auth"
	dto "math-ai.com/math-ai/internal/application/dto/auth"
	dtoUser "math-ai.com/math-ai/internal/application/dto/user"
	"math-ai.com/math-ai/internal/application/transaction"
	"math-ai.com/math-ai/internal/infrastructure/metadata"
	"math-ai.com/math-ai/internal/module/user"
)

type Service struct {
	userSvc   *user.Service
	loginCmd  *command.LoginCommandHandler
	logoutCmd *command.LogoutCommandHandler
}

func NewService(userSvc *user.Service, uow transaction.UnitOfWork) *Service {
	return &Service{
		userSvc:   userSvc,
		loginCmd:  command.NewLoginCommandHandler(uow),
		logoutCmd: command.NewLogoutCommandHandler(uow),
	}
}

func (s *Service) Login(ctx context.Context, req *dto.LoginReq) (*dto.LoginRes, error) {
	if err := ValidateLogin(ctx, req); err != nil {
		return nil, err
	}

	result, err := s.loginCmd.Handle(ctx, command.LoginCommand{
		Phone:           req.Phone,
		DeviceUUID:      metadata.GetDeviceID(ctx),
		DeviceName:      metadata.GetDeviceName(ctx),
		IPAddress:       metadata.GetIPAddress(ctx),
		DevicePushToken: metadata.GetDevicePushToken(ctx),
	})
	if err != nil {
		return nil, err
	}
	if result == nil {
		return &dto.LoginRes{
			User: nil,
		}, nil
	}

	userRes, err := s.userSvc.GetUserById(ctx, &dtoUser.GetUserByUserIdReq{UserID: result.UserID})
	if err != nil {
		return nil, err
	}
	if userRes == nil || userRes.User == nil {
		return &dto.LoginRes{
			User: nil,
		}, nil
	}

	return &dto.LoginRes{
		TwoFactorRequired: result.TwoFactorRequired,
		User:              userRes.User,
	}, nil
}

func (s *Service) Logout(ctx context.Context, req *dto.LogoutReq) (*dto.LogoutRes, error) {
	if err := ValidateLogout(ctx, req); err != nil {
		return nil, err
	}

	if err := s.logoutCmd.Handle(ctx, command.LogoutCommand{
		// UserID:     metadata.GetUserID(ctx),
		DeviceUUID: metadata.GetDeviceID(ctx),
	}); err != nil {
		return nil, err
	}

	return &dto.LogoutRes{}, nil
}
