package services

import (
	"context"

	"math-ai.com/math-ai/internal/applications/dto"
	"math-ai.com/math-ai/internal/applications/validators"
	di "math-ai.com/math-ai/internal/core/di/repositories"
	"math-ai.com/math-ai/internal/shared/constant/status"
)

type DeviceService struct {
	validator validators.IDeviceValidator
	repo      di.IDeviceRepository
}

func NewDeviceService(
	validator validators.IDeviceValidator,
	repo di.IDeviceRepository,
) *DeviceService {
	return &DeviceService{
		validator: validator,
		repo:      repo,
	}
}

func (s *DeviceService) GetDeviceByDeviceUUID(ctx context.Context, deviceUUID string) (status.Code, *dto.DeviceResponse, error) {
	device, err := s.repo.GetDeviceByDeviceUUID(ctx, deviceUUID)
	if err != nil {
		return status.FAIL, nil, err
	}

	if device == nil {
		return status.NOT_FOUND, nil, nil
	}

	res := dto.DeviceResponseFromDomain(device)

	return status.SUCCESS, res, nil
}

func (s *DeviceService) CreateDevice(ctx context.Context, req *dto.CreateDeviceRequest) (status.Code, error) {
	// Validate request
	if statusCode, err := s.validator.ValidateCreateDeviceRequest(req); err != nil {
		return statusCode, err
	}

	deviceDomain := dto.BuildDeviceDomainForCreate(req)

	err := s.repo.StoreDevice(ctx, nil, deviceDomain)
	if err != nil {
		return status.FAIL, err
	}

	return status.SUCCESS, nil
}

func (s *DeviceService) UpdateDevice(ctx context.Context, req *dto.UpdateDeviceRequest) (status.Code, error) {
	// Validate request
	if statusCode, err := s.validator.ValidateUpdateDeviceRequest(req); err != nil {
		return statusCode, err
	}

	deviceDomain := dto.BuildDeviceDomainForUpdate(req)
	err := s.repo.UpdateDevice(ctx, deviceDomain)
	if err != nil {
		return status.FAIL, err
	}

	return status.SUCCESS, nil
}

func (s *DeviceService) MarkVerifiedDevice(ctx context.Context, uid string, deviceUUID string) (status.Code, error) {
	err := s.repo.MarkVerifiedDeviceByUIDAndDeviceUUID(ctx, uid, deviceUUID)
	if err != nil {
		return status.FAIL, err
	}

	return status.SUCCESS, nil
}
