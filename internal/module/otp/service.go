package otp

import (
	"context"

	"github.com/google/uuid"
	"math-ai.com/math-ai/internal/adapter/otp_delivery"
	command "math-ai.com/math-ai/internal/application/command/otp"
	dto "math-ai.com/math-ai/internal/application/dto/otp"
	userDto "math-ai.com/math-ai/internal/application/dto/user"
	"math-ai.com/math-ai/internal/module/user"

	query "math-ai.com/math-ai/internal/application/query/otp"
	"math-ai.com/math-ai/internal/application/transaction"
	domain "math-ai.com/math-ai/internal/domain/otp"
	"math-ai.com/math-ai/internal/infrastructure/logger"
	"math-ai.com/math-ai/internal/infrastructure/metadata"
	"math-ai.com/math-ai/internal/shared/enum"
	"math-ai.com/math-ai/internal/shared/utils"
)

// Service composes the OTP command/query handlers behind a single surface.
// Two kinds of callers exist:
//   - HTTP handlers in this package (send / verify / resend / revoke).
//   - Other modules that need OTP plumbing — e.g. an auth/2fa orchestrator
//     could call Verify() directly to keep the response shape uniform.
type Service struct {
	userSvc      *user.Service
	sendCmd      *command.SendOtpCommandHandler
	verifyCmd    *command.VerifyOtpCommandHandler
	revokeCmd    *command.RevokeOtpCommandHandler
	getByIdQuery *query.GetOtpByIdQueryHandler
	repo         domain.IRepository
}

func NewService(
	userSvc *user.Service,
	repo domain.IRepository,
	uow transaction.UnitOfWork,
	delivery *otp_delivery.Adapter,
) *Service {
	return &Service{
		userSvc:      userSvc,
		sendCmd:      command.NewSendOtpCommandHandler(uow, delivery),
		verifyCmd:    command.NewVerifyOtpCommandHandler(uow),
		revokeCmd:    command.NewRevokeOtpCommandHandler(uow),
		getByIdQuery: query.NewGetOtpByIdQueryHandler(repo),
		repo:         repo,
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
		channel = enum.OtpChannelSMS
		userResp, _ := s.userSvc.GetUserByPhone(ctx, &userDto.GetUserByPhoneReq{
			Phone: req.Identifier,
		})
		user = userResp.User
	}

	var userId *uuid.UUID
	switch req.OtpType {
	case string(enum.OtpTypeLogin2FA):
		if user != nil {
			userId = &user.UserID
		}
	case string(enum.OtpTypeRegister):
		userId = nil
	default:
	}

	deviceUUID := metadata.GetDeviceID(ctx)
	deviceName := metadata.GetDeviceName(ctx)

	result, err := s.sendCmd.Handle(ctx, command.SendOtpCommand{
		OtpType:    enum.OtpType(req.OtpType),
		Identifier: req.Identifier,
		UserID:     userId,
		DeviceUUID: &deviceUUID,
		DeviceName: &deviceName,
		Channel:    channel,
		// Language: lang,
	})
	if err != nil {
		return nil, err
	}

	log.Infof("otp sent %s:", result.Code)

	return &dto.SendOtpRes{
		OtpID:     result.OtpID,
		ExpiresAt: result.ExpiresAt,
		Channel:   string(result.Channel),
		Code:      result.Code,
	}, nil
}

func (s *Service) Verify(ctx context.Context, req *dto.VerifyOtpReq) (*dto.VerifyOtpRes, error) {
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

	log.Infof("otp verified %s:", result.OtpID)

	return &dto.VerifyOtpRes{
		OtpID:    result.OtpID,
		Verified: true,
		// UserID:     result.UserID,
		// DeviceUUID: result.DeviceUUID,
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
