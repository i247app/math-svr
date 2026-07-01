package misc

import (
	"context"

	dto "math-ai.com/math-ai/internal/application/dto/misc"
	mtime "math-ai.com/math-ai/internal/domain/shared/time"
)

type Service struct {
}

func NewService() *Service {
	return &Service{}
}

func (s *Service) LogsTimeFormat(ctx context.Context, req *dto.LogTimeFormatReq) (*dto.LogTimeFormatRes, error) {
	apptime, err := mtime.ParseFromString(req.Time)
	if err != nil {
		return nil, err
	}

	res := &dto.LogTimeFormatRes{
		DefaultFormat:       apptime.Time.Format(mtime.DefaultFormat),
		DateOnlyFormat:      apptime.Time.Format(mtime.DateOnly),
		CoreDateTimeFormat:  apptime.Time.Format(mtime.CoreDateTimeFormat),
		DateTimeFormat20FSP: apptime.Time.Format(mtime.DateTimeFormat20FSP),
		Format17FSP:         apptime.Time.Format(mtime.Format17FSP),
		Format20FSP:         apptime.Time.Format(mtime.Format20FSP),
	}

	return res, nil
}
