package auth

import (
	"context"

	dto "math-ai.com/math-ai/internal/application/dto/auth"
	dtoUser "math-ai.com/math-ai/internal/application/dto/user"
	"math-ai.com/math-ai/internal/module/user"
)

type Service struct {
	userSvc *user.Service
}

func NewService(userSvc *user.Service) *Service {
	return &Service{userSvc: userSvc}
}

func (s *Service) Login(ctx context.Context, req *dto.LoginReq) (*dto.LoginRes, error) {
	res, err := s.userSvc.GetUserByPhone(ctx, &dtoUser.GetUserByPhoneReq{Phone: req.Phone})
	if err != nil {
		return nil, err
	}

	svcRes := &dto.LoginRes{
		User: res.User,
	}
	return svcRes, nil
}
