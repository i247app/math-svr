package auth

import (
	"context"
	"errors"

	command "math-ai.com/math-ai/internal/application/command/auth"
	dto "math-ai.com/math-ai/internal/application/dto/auth"
	dtoUser "math-ai.com/math-ai/internal/application/dto/user"
	"math-ai.com/math-ai/internal/application/transaction"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
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
		IPAddress:       metadata.GetIPAddress(ctx),
		DevicePushToken: metadata.GetDevicePushToken(ctx),
	})
	if err != nil {
		return nil, err
	}

	userRes, err := s.userSvc.GetUserById(ctx, &dtoUser.GetUserByUserIdReq{UserID: result.UserID})
	if err != nil {
		return nil, err
	}
	if userRes == nil || userRes.User == nil {
		return nil, errs.NewError(ctx, status.USER_NOT_FOUND, nil, errors.New("user not found after login"))
	}

	return &dto.LoginRes{
		User: userRes.User,
	}, nil
}

func (s *Service) Logout(ctx context.Context, req *dto.LogoutReq) (*dto.LogoutRes, error) {
	if err := ValidateLogout(ctx, req); err != nil {
		return nil, err
	}

	if err := s.logoutCmd.Handle(ctx, command.LogoutCommand{
		Token: metadata.GetDevicePushToken(ctx),
	}); err != nil {
		return nil, err
	}

	return &dto.LogoutRes{}, nil
}
