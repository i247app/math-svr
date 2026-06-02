package device

import (
	"context"

	command "math-ai.com/math-ai/internal/application/command/device"
	dto "math-ai.com/math-ai/internal/application/dto/device"
	query "math-ai.com/math-ai/internal/application/query/device"
	"math-ai.com/math-ai/internal/application/transaction"
	domain "math-ai.com/math-ai/internal/domain/device"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/infrastructure/logger"
)

// Service is the device module's outward surface. Two callers exist:
//   - HTTP handlers in this package (list / get / update / revoke / delete).
//   - The future 2FA module, which calls VerifyDevice after a successful
//     OTP/authenticator challenge.
//
// EnsureDeviceForLogin is the seam consumed by the auth login command. It
// runs against the supplied transactional repo (the command is already inside
// a UoW), looks up the (user, device_uuid) pair, auto-creates the row on
// first sight, and returns the resulting device entity so the caller can
// branch on is_verified.
type Service struct {
	getDeviceByIdQuery       *query.GetDeviceByIdQueryHandler
	listDevicesByUserIdQuery *query.ListDevicesByUserIdQueryHandler
	markVerifiedCmd          *command.MarkDeviceVerifiedCommandHandler
	updateDeviceCmd          *command.UpdateDeviceCommandHandler
	revokeDeviceCmd          *command.RevokeDeviceCommandHandler
	softDeleteDeviceCmd      *command.SoftDeleteDeviceCommandHandler
	repo                     domain.IRepository
}

func NewService(repo domain.IRepository, uow transaction.UnitOfWork) *Service {
	return &Service{
		getDeviceByIdQuery:       query.NewGetDeviceByIdQueryHandler(repo),
		listDevicesByUserIdQuery: query.NewListDevicesByUserIdQueryHandler(repo),
		markVerifiedCmd:          command.NewMarkDeviceVerifiedCommandHandler(uow),
		updateDeviceCmd:          command.NewUpdateDeviceCommandHandler(uow),
		revokeDeviceCmd:          command.NewRevokeDeviceCommandHandler(uow),
		softDeleteDeviceCmd:      command.NewSoftDeleteDeviceCommandHandler(uow),
		repo:                     repo,
	}
}

func (s *Service) GetDeviceById(ctx context.Context, req *dto.GetDeviceByIdReq) (*dto.GetDeviceByIdRes, error) {
	if err := ValidateGetDevice(ctx, req); err != nil {
		return nil, err
	}
	d, err := s.getDeviceByIdQuery.Handle(ctx, query.GetDeviceByIdQuery{DeviceID: req.DeviceID})
	if err != nil {
		return nil, err
	}
	if d == nil {
		return nil, errs.NewError(ctx, status.DEVICE_NOT_FOUND, nil, ErrDeviceNotFound)
	}
	return &dto.GetDeviceByIdRes{Device: dto.DomainToResponse(d)}, nil
}

func (s *Service) ListDevicesByUserId(ctx context.Context, req *dto.ListDevicesReq) (*dto.ListDevicesRes, error) {
	if err := ValidateListDevices(ctx, req); err != nil {
		return nil, err
	}
	devices, err := s.listDevicesByUserIdQuery.Handle(ctx, query.ListDevicesByUserIdQuery{UserID: req.UserID})
	if err != nil {
		return nil, err
	}
	return &dto.ListDevicesRes{Devices: dto.DomainListToResponse(devices)}, nil
}

func (s *Service) UpdateDevice(ctx context.Context, req *dto.UpdateDeviceReq) (*dto.UpdateDeviceRes, error) {
	if err := ValidateUpdateDevice(ctx, req); err != nil {
		return nil, err
	}
	updated, err := s.updateDeviceCmd.Handle(ctx, command.UpdateDeviceCommand{
		UserID:          req.UserID,
		DeviceID:        req.DeviceID,
		DeviceName:      req.DeviceName,
		DevicePushToken: req.DevicePushToken,
		Note:            req.Note,
	})
	if err != nil {
		return nil, err
	}
	logger.From(ctx).Info("device.updated", "device_id", req.DeviceID, "user_id", req.UserID)
	return &dto.UpdateDeviceRes{Device: dto.DomainToResponse(updated)}, nil
}

func (s *Service) RevokeDevice(ctx context.Context, req *dto.RevokeDeviceReq) (*dto.RevokeDeviceRes, error) {
	if err := ValidateRevokeDevice(ctx, req); err != nil {
		return nil, err
	}
	if err := s.revokeDeviceCmd.Handle(ctx, command.RevokeDeviceCommand{
		UserID:   req.UserID,
		DeviceID: req.DeviceID,
	}); err != nil {
		return nil, err
	}
	logger.From(ctx).Info("device.revoked", "device_id", req.DeviceID, "user_id", req.UserID)
	return &dto.RevokeDeviceRes{}, nil
}

func (s *Service) SoftDeleteDevice(ctx context.Context, req *dto.DeleteDeviceReq) (*dto.DeleteDeviceRes, error) {
	if err := ValidateDeleteDevice(ctx, req); err != nil {
		return nil, err
	}
	if err := s.softDeleteDeviceCmd.Handle(ctx, command.SoftDeleteDeviceCommand{
		UserID:   req.UserID,
		DeviceID: req.DeviceID,
	}); err != nil {
		return nil, err
	}
	logger.From(ctx).Info("device.soft_deleted", "device_id", req.DeviceID, "user_id", req.UserID)
	return &dto.DeleteDeviceRes{}, nil
}

// VerifyDevice is the seam the future 2FA module calls after a successful
// out-of-band challenge. It's exposed on the device service (not the auth
// service) so the device trust state is owned by one module. The 2FA module
// owns the challenge, this module owns the trust decision.
func (s *Service) VerifyDevice(ctx context.Context, req *dto.VerifyDeviceReq) (*dto.VerifyDeviceRes, error) {
	if err := ValidateVerifyDevice(ctx, req); err != nil {
		return nil, err
	}
	if err := s.markVerifiedCmd.Handle(ctx, command.MarkDeviceVerifiedCommand{
		UserID:   req.UserID,
		DeviceID: req.DeviceID,
	}); err != nil {
		return nil, err
	}
	d, err := s.getDeviceByIdQuery.Handle(ctx, query.GetDeviceByIdQuery{DeviceID: req.DeviceID})
	if err != nil {
		return nil, err
	}
	logger.From(ctx).Info("device.verified", "device_id", req.DeviceID, "user_id", req.UserID)
	return &dto.VerifyDeviceRes{Device: dto.DomainToResponse(d)}, nil
}
