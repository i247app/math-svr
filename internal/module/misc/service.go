package misc

import (
	"context"

	dto "math-ai.com/math-ai/internal/application/dto/misc"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/mtime"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/infrastructure/logger"
)

// ClearDataRepository wipes all user-generated data and resets the matching
// external-id counters. Implemented by the MySQL maintenance repository so the
// module never touches raw SQL.
type ClearDataRepository interface {
	ClearData(ctx context.Context) (tablesCleared []string, seqsReset []string, err error)
}

type Service struct {
	maintenanceRepo ClearDataRepository
}

func NewService(maintenanceRepo ClearDataRepository) *Service {
	return &Service{
		maintenanceRepo: maintenanceRepo,
	}
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

// ClearData wipes every user-generated table and resets the matching id
// counters, keeping only the seeded reference/curriculum data. This is the API
// equivalent of `make clear-data-local` / `make clear-data-ec2`.
func (s *Service) ClearData(ctx context.Context) (*dto.ClearDataRes, error) {
	log := logger.From(ctx)
	log.Warn("misc.clear_data.start")

	tables, seqs, err := s.maintenanceRepo.ClearData(ctx)
	if err != nil {
		return nil, errs.NewError(ctx, status.INTERNAL_SERVER_ERROR, nil, err)
	}

	log.Warn("misc.clear_data.done", "tables_cleared", len(tables), "seqs_reset", len(seqs))

	return &dto.ClearDataRes{
		TablesCleared: tables,
		SeqsReset:     seqs,
	}, nil
}
