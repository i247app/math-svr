package auth

import (
	"context"
	"errors"
	"fmt"

	command "math-ai.com/math-ai/internal/application/command/auth"
	dto "math-ai.com/math-ai/internal/application/dto/auth"
	dtoOtp "math-ai.com/math-ai/internal/application/dto/otp"
	dtoUser "math-ai.com/math-ai/internal/application/dto/user"
	"math-ai.com/math-ai/internal/application/transaction"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/infrastructure/logger"
	"math-ai.com/math-ai/internal/infrastructure/metadata"
	"math-ai.com/math-ai/internal/infrastructure/session"
	"math-ai.com/math-ai/internal/module/otp"
	"math-ai.com/math-ai/internal/module/user"
	"math-ai.com/math-ai/internal/shared/enum"
)

type Service struct {
	userSvc   *user.Service
	otpSvc    *otp.Service
	loginCmd  *command.LoginCommandHandler
	logoutCmd *command.LogoutCommandHandler
}

func NewService(userSvc *user.Service, otpSvc *otp.Service, uow transaction.UnitOfWork) *Service {
	return &Service{
		userSvc:   userSvc,
		otpSvc:    otpSvc,
		loginCmd:  command.NewLoginCommandHandler(uow),
		logoutCmd: command.NewLogoutCommandHandler(uow),
	}
}

func (s *Service) Login(ctx context.Context, sess *session.AppSession, req *dto.LoginReq) (*dto.LoginRes, error) {
	log := logger.From(ctx)
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
		}, errs.NewError(ctx, status.NO_DATA, nil, errors.New("user not found"))
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

	log.Info("Login successful, updating session data...")
	sessionData := session.InitData{
		Source:    "login",
		IsSecure:  true,
		UID:       userRes.User.UserID,
		LoginName: req.Phone,
	}

	if userRes.User.Email != nil {
		sessionData.Email = *userRes.User.Email
	}

	sess.Init(sessionData)

	return &dto.LoginRes{
		TwoFactorRequired: result.TwoFactorRequired,
		User:              userRes.User,
	}, nil
}

func (s *Service) LoginResume(ctx context.Context, sess *session.AppSession) (*dto.LoginRes, error) {
	if !sess.IsValid() {
		return nil, errs.NewError(ctx, status.UNAUTHORIZED, nil, errors.New("session is not valid"))
	}

	uid, ok := sess.UID()
	if !ok {
		// return nil, m_status.UNAUTHORIZED, fmt.Errorf("failed to resume session: uid not found in session")
		return nil, errs.NewError(ctx, status.UNAUTHORIZED, nil, fmt.Errorf("uid not found in session"))
	}

	userRes, err := s.userSvc.GetUserById(ctx, &dtoUser.GetUserByUserIdReq{UserID: uid})
	if err != nil {
		return nil, err
	}
	if userRes == nil || userRes.User == nil {
		return &dto.LoginRes{
			User: nil,
		}, nil
	}

	return &dto.LoginRes{
		User: userRes.User,
	}, nil
}

func (s *Service) LoginWithOTP(ctx context.Context, req *dto.LoginReq) (*dto.LoginWithOTPRes, error) {
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
		return &dto.LoginWithOTPRes{
			User: nil,
		}, errs.NewError(ctx, status.NO_DATA, nil, errors.New("user not found"))
	}

	userRes, err := s.userSvc.GetUserById(ctx, &dtoUser.GetUserByUserIdReq{UserID: result.UserID})
	if err != nil {
		return nil, err
	}
	if userRes == nil || userRes.User == nil {
		return &dto.LoginWithOTPRes{
			User: nil,
		}, nil
	}

	otpCreated, err := s.otpSvc.Send(ctx, &dtoOtp.SendOtpReq{
		OtpType:    string(enum.OtpTypeLogin2FA),
		Identifier: req.Phone,
		UserID:     &userRes.User.UserID,
	})
	if err != nil {
		return nil, err
	}

	return &dto.LoginWithOTPRes{
		User:      userRes.User,
		OTPCode:   otpCreated.OTPCode,
		OtpType:   otpCreated.OtpType,
		ExpiresAt: otpCreated.ExpiresAt,
	}, nil
}

func (s *Service) Logout(ctx context.Context, sess *session.AppSession, req *dto.LogoutReq) (*dto.LogoutRes, error) {
	if err := ValidateLogout(ctx, req); err != nil {
		return nil, err
	}

	// if err := s.logoutCmd.Handle(ctx, command.LogoutCommand{
	// 	DeviceUUID: metadata.GetDeviceID(ctx),
	// }); err != nil {
	// 	return nil, err
	// }

	sess.MarkNotSecure()

	return &dto.LogoutRes{}, nil
}
