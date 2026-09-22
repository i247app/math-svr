package misc

import (
	"context"
	"errors"
	"fmt"
	"strings"

	dto "math-ai.com/math-ai/internal/application/dto/misc"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/mtime"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/infrastructure/logger"
)

// ClearDataRepository wipes user-generated data and resets the matching
// external-id counters. Implemented by the MySQL maintenance repository so the
// module never touches raw SQL.
type ClearDataRepository interface {
	ClearData(ctx context.Context) (tablesCleared []string, seqsReset []string, err error)
	// ClearDataKeepingUsers wipes everything except rows owned by the
	// whitelisted users (by uid). Sequences are not reset in this mode.
	ClearDataKeepingUsers(ctx context.Context, keepUids []int64) (tablesProcessed []string, err error)
	ClearDataTables(ctx context.Context, tables []string) (tablesCleared []string, seqsReset []string, err error)
	// ClearableTables is the allow-list of table names the table-scoped clear
	// may wipe; the service validates a request against it.
	ClearableTables() []string
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

// ClearData wipes every user-generated table, keeping only the seeded
// reference/curriculum data. This is the API equivalent of
// `make clear-data-local` / `make clear-data-ec2`.
//
// When req.WhitelistId carries uids, those users and everything they own are
// preserved (deletes instead of truncates; sequences are left untouched to
// avoid colliding with the surviving ids). An empty/omitted whitelist keeps the
// original full-wipe behaviour and resets the sequences.
func (s *Service) ClearData(ctx context.Context, req *dto.ClearDataReq) (*dto.ClearDataRes, error) {
	log := logger.From(ctx)

	if req != nil && len(req.WhitelistId) > 0 {
		keep := dedupeInt64(req.WhitelistId)
		log.Warn("misc.clear_data.start", "keep_uids", len(keep))

		tables, err := s.maintenanceRepo.ClearDataKeepingUsers(ctx, keep)
		if err != nil {
			return nil, errs.NewError(ctx, status.INTERNAL_SERVER_ERROR, nil, err)
		}

		log.Warn("misc.clear_data.done", "tables_processed", len(tables), "kept_uids", len(keep))

		return &dto.ClearDataRes{
			TablesCleared: tables,
			SeqsReset:     []string{},
			KeptUids:      keep,
		}, nil
	}

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

// dedupeInt64 returns xs with duplicates removed, preserving first-seen order.
func dedupeInt64(xs []int64) []int64 {
	seen := make(map[int64]struct{}, len(xs))
	out := make([]int64, 0, len(xs))
	for _, x := range xs {
		if _, ok := seen[x]; ok {
			continue
		}
		seen[x] = struct{}{}
		out = append(out, x)
	}
	return out
}

// ClearDataTables wipes only the tables named in the request, rejecting any
// name outside the clear-data allow-list before touching the database. Like
// ClearData it keeps seeded reference/curriculum data intact.
func (s *Service) ClearDataTables(ctx context.Context, req *dto.ClearDataTablesReq) (*dto.ClearDataRes, error) {
	if req == nil || len(req.Tables) == 0 {
		return nil, errs.NewError(ctx, status.FAIL, nil,
			errors.New("misc clear-data: tables is required"))
	}

	clearable := s.maintenanceRepo.ClearableTables()
	allowed := make(map[string]struct{}, len(clearable))
	for _, t := range clearable {
		allowed[t] = struct{}{}
	}
	var invalid []string
	for _, t := range req.Tables {
		if _, ok := allowed[t]; !ok {
			invalid = append(invalid, t)
		}
	}
	if len(invalid) > 0 {
		return nil, errs.NewError(ctx, status.FAIL, nil,
			fmt.Errorf("misc clear-data: not clearable: %s (allowed: %s)",
				strings.Join(invalid, ", "), strings.Join(clearable, ", ")))
	}

	log := logger.From(ctx)
	log.Warn("misc.clear_data_tables.start", "requested", len(req.Tables))

	tables, seqs, err := s.maintenanceRepo.ClearDataTables(ctx, req.Tables)
	if err != nil {
		return nil, errs.NewError(ctx, status.INTERNAL_SERVER_ERROR, nil, err)
	}

	log.Warn("misc.clear_data_tables.done", "tables_cleared", len(tables), "seqs_reset", len(seqs))

	return &dto.ClearDataRes{
		TablesCleared: tables,
		SeqsReset:     seqs,
	}, nil
}
