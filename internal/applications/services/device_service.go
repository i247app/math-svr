package services

import (
	"context"

	"math-ai.com/math-ai/internal/applications/dto"
	"math-ai.com/math-ai/internal/core/di/repositories"
	"math-ai.com/math-ai/internal/shared/constant/status"
)

type DeviceService struct {
	repo repositories.IDeviceRepository
}

func NewDeviceService(repo repositories.IDeviceRepository) *DeviceService {
	return &DeviceService{
		repo: repo,
	}
}

func (s *DeviceService) GetDeviceByDeviceUUID(ctx context.Context, deviceUUID string) (status.Code, *dto.DeviceResponse, error) {
	device, err := s.repo.GetDeviceByDeviceUUID(ctx, deviceUUID)
	if err != nil {
		return status.INTERNAL, nil, err
	}

	if device == nil {
		return status.NOT_FOUND, nil, nil
	}

	res := dto.DeviceResponseFromDomain(device)

	return status.SUCCESS, res, nil
}

func (s *DeviceService) CreateDevice(ctx context.Context, req *dto.CreateDeviceReq) (status.Code, error) {
	// statusCode, err := ValidationCreateDeviceReq(req)
	// if err != nil {
	// 	return statusCode, err
	// }

	deviceDomain := dto.BuildDeviceDomainForCreate(req)

	err := s.repo.StoreDevice(ctx, nil, deviceDomain)
	if err != nil {
		return status.INTERNAL, err
	}

	return status.SUCCESS, nil
}

func (s *DeviceService) UpdateDevice(ctx context.Context, req *dto.UpdateDeviceReq) (status.Code, error) {
	// statusCode, err := ValidationUpdateDeviceReq(req)
	// if err != nil {
	// 	return statusCode, err
	// }

	deviceDomain := dto.BuildDeviceDomainForUpdate(req)
	err := s.repo.UpdateDevice(ctx, deviceDomain)
	if err != nil {
		return status.INTERNAL, err
	}

	return status.SUCCESS, nil
}

func (s *DeviceService) MarkVerifiedDevice(ctx context.Context, uid string, deviceUUID string) (status.Code, error) {
	err := s.repo.MarkVerifiedDeviceByUIDAndDeviceUUID(ctx, uid, deviceUUID)
	if err != nil {
		return status.INTERNAL, err
	}

	return status.SUCCESS, nil
}
