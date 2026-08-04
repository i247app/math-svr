package otp

import (
	"context"
	"fmt"

	notifAdapter "math-ai.com/math-ai/internal/adapter/notification"
	"math-ai.com/math-ai/internal/adapter/otp_delivery"
	command "math-ai.com/math-ai/internal/application/command/otp"
	deviceDTO "math-ai.com/math-ai/internal/application/dto/device"
	notifDto "math-ai.com/math-ai/internal/application/dto/notification"
	dto "math-ai.com/math-ai/internal/application/dto/otp"
	userDto "math-ai.com/math-ai/internal/application/dto/user"
	"math-ai.com/math-ai/internal/module/device"
	"math-ai.com/math-ai/internal/module/notification"
	"math-ai.com/math-ai/internal/module/user"

	query "math-ai.com/math-ai/internal/application/query/otp"
	"math-ai.com/math-ai/internal/application/transaction"
	domain "math-ai.com/math-ai/internal/domain/otp"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/infrastructure/logger"
	"math-ai.com/math-ai/internal/infrastructure/metadata"
	"math-ai.com/math-ai/internal/infrastructure/session"
	"math-ai.com/math-ai/internal/shared/enum"
	"math-ai.com/math-ai/internal/shared/utils"
)

// Service composes the OTP command/query handlers behind a single surface.
// Two kinds of callers exist:
//   - HTTP handlers in this package (send / verify / resend / revoke).
//   - Other modules that need OTP plumbing — e.g. an auth/2fa orchestrator
//     could call Verify() directly to keep the response shape uniform.
type Service struct {
	userSvc         *user.Service
	deviceSvc       *device.Service
	notificationSvc *notification.Service
	sendCmd         *command.SendOtpCommandHandler
	verifyCmd       *command.VerifyOtpCommandHandler
	revokeCmd       *command.RevokeOtpCommandHandler
	getByIdQuery    *query.GetOtpByIdQueryHandler
	repo            domain.IRepository
}

func NewService(
	userSvc *user.Service,
	deviceSvc *device.Service,
	notificationSvc *notification.Service,
	repo domain.IRepository,
	uow transaction.UnitOfWork,
	delivery *otp_delivery.Adapter,
	pushAdapter *notifAdapter.Adapter,
) *Service {
	return &Service{
		userSvc:         userSvc,
		deviceSvc:       deviceSvc,
		notificationSvc: notificationSvc,
		sendCmd:         command.NewSendOtpCommandHandler(uow, delivery, pushAdapter),
		verifyCmd:       command.NewVerifyOtpCommandHandler(uow),
		revokeCmd:       command.NewRevokeOtpCommandHandler(uow),
		getByIdQuery:    query.NewGetOtpByIdQueryHandler(repo),
		repo:            repo,
	}
}

func (s *Service) Send(ctx context.Context, req *dto.SendOtpReq) (*dto.SendOtpRes, error) {
	log := logger.From(ctx)
	if err := ValidateSendOtp(ctx, req); err != nil {
		return nil, err
	}

	var user *userDto.UserResponse
	var channel enum.OtpChannel
	if utils.ValidateEmail(req.Identifier) {
		channel = enum.OtpChannelEmail
		userResp, _ := s.userSvc.GetUserByEmail(ctx, &userDto.GetUserByEmailReq{
			Email: req.Identifier,
		})
		user = userResp.User
	} else if utils.ValidatePhone(req.Identifier) {
		normalizePhone, err := utils.NormalizePhone(req.Identifier)
		if err != nil {
			return nil, errs.NewError(ctx, status.FAIL, nil, fmt.Errorf("failed to normalize phone: %w", err))
		}
		channel = enum.OtpChannelSMS
		userResp, _ := s.userSvc.GetUserByPhone(ctx, &userDto.GetUserByPhoneReq{
			Phone: normalizePhone,
		})
		user = userResp.User
	}

	var userId *int64
	switch req.OtpType {
	case string(enum.OtpTypeLogin2FA):
		if user == nil {
			return nil, errs.NewError(ctx, status.USER_NOT_FOUND, nil, ErrUserNotFound)
		}

		userId = &user.UserID
	case string(enum.OtpTypeRegister):
		if user != nil {
			return nil, errs.NewError(ctx, status.USER_ALREADY_EXISTS, nil, ErrUserAlreadyExists)
		}
		userId = nil
	default:
	}

	deviceUUID := metadata.GetDeviceID(ctx)
	deviceName := metadata.GetDeviceName(ctx)

	result, err := s.sendCmd.Handle(ctx, command.SendOtpCommand{
		OtpType:        enum.OtpType(req.OtpType),
		Identifier:     req.Identifier,
		UserID:         userId,
		DeviceUUID:     &deviceUUID,
		DeviceName:     &deviceName,
		Channel:        channel,
		TargetDeviceID: req.TargetDeviceID,
	})
	if err != nil {
		return nil, err
	}

	log.Infof("otp sent %s:", result.OTPCode)

	// Trusted-device push 2FA: alert the account owner that a login was
	// requested, independent of the OTP itself — this row never carries the
	// code (see send_otp_command.go), it's a security notice only. Best
	// effort: a notice failure must not fail the OTP send the user is
	// actively waiting on.
	if req.TargetDeviceID != nil && userId != nil && s.notificationSvc != nil {
		requestingDevice := metadata.GetDeviceName(ctx)
		if requestingDevice == "" {
			requestingDevice = "một thiết bị"
		}
		category := enum.NotificationCategoryTypeWarning.String()
		_, nerr := s.notificationSvc.SendNotification(ctx, &notifDto.SendNotificationReq{
			UserID:    *userId,
			Title:     "Cảnh báo đăng nhập",
			ShortText: fmt.Sprintf("Có yêu cầu đăng nhập mới từ %s.Mã OTP của bạn là %s. Nếu không phải bạn, vui lòng đổi không được để lộ mã otp.", requestingDevice, result.OTPCode),
			Category:  &category,
		})
		if nerr != nil {
			log.Warnf("otp.login_security_notice_failed user_id=%d err=%v", *userId, nerr)
		}
	}

	return &dto.SendOtpRes{
		ExpiresAt: result.ExpiresAt.String(),
		Channel:   string(result.Channel),
		OTPCode:   result.OTPCode,
		OtpType:   result.OTPType,
	}, nil
}

func (s *Service) Verify(ctx context.Context, sess *session.AppSession, req *dto.VerifyOtpReq) (*dto.VerifyOtpRes, error) {
	log := logger.From(ctx)

	if err := ValidateVerifyOtp(ctx, req); err != nil {
		return nil, err
	}

	result, err := s.verifyCmd.Handle(ctx, command.VerifyOtpCommand{
		OtpType:    enum.OtpType(req.OtpType),
		Identifier: req.Identifier,
		Code:       req.OtpCode,
	})
	if err != nil {
		return nil, err
	}

	var user *userDto.UserResponse
	if result != nil && result.UserID != nil {
		userRes, err := s.userSvc.GetUserById(ctx, &userDto.GetUserByUserIdReq{UserID: *result.UserID})
		if err != nil {
			return nil, err
		}
		if userRes == nil || userRes.User == nil {
			return &dto.VerifyOtpRes{
				OtpType:  req.OtpType,
				Verified: true,
				User:     user,
			}, nil
		}

		// Update session
		switch req.OtpType {
		case string(enum.OtpTypeLogin2FA):
			deviceUUID := metadata.GetDeviceID(ctx)
			log.Info("Mark device as trusted")
			_, err := s.deviceSvc.VerifyDevice(ctx, &deviceDTO.VerifyDeviceReq{
				UserID:          *result.UserID,
				DeviceUUID:      deviceUUID,
				DeviceName:      metadata.GetDeviceName(ctx),
				Platform:        metadata.GetPlatform(ctx),
				DevicePushToken: utils.ToStringPtr(metadata.GetDevicePushToken(ctx)),
			})
			if err != nil {
				return nil, err
			}

			log.Info("Login successful, updating session data...")
			sessionData := session.InitData{
				Source:    "login",
				IsSecure:  true,
				UID:       userRes.User.UserID,
				LoginName: req.Identifier,
			}

			if userRes.User.Email != nil {
				sessionData.Email = *userRes.User.Email
			}

			sess.Init(sessionData)
		}
		user = userRes.User
	}

	return &dto.VerifyOtpRes{
		Verified: true,
		OtpType:  req.OtpType,
		User:     user,
	}, nil
}

func (s *Service) Revoke(ctx context.Context, req *dto.RevokeOtpReq) (*dto.RevokeOtpRes, error) {
	if err := ValidateRevokeOtp(ctx, req); err != nil {
		return nil, err
	}

	if err := s.revokeCmd.Handle(ctx, command.RevokeOtpCommand{
		OtpType:    enum.OtpType(req.OtpType),
		Identifier: req.Identifier,
	}); err != nil {
		return nil, err
	}
	return &dto.RevokeOtpRes{}, nil
}
