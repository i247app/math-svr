package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"math-ai.com/math-ai/internal/domain/program"
	"math-ai.com/math-ai/internal/domain/shared/mtime"
	"math-ai.com/math-ai/internal/infrastructure/database"
	"math-ai.com/math-ai/internal/infrastructure/persistence/mysql/models"
	"math-ai.com/math-ai/internal/shared/enum"
	"math-ai.com/math-ai/internal/shared/pagination"
)

const (
	programTable = "ma_programs"

	programColumns = `p.id, p.program_id, p.label, p.description,
		p.image_key, p.display_order, p.note,
		p.program_status, p.status,
		p.create_id, p.create_dt, p.modify_id, p.modify_dt`

	programActiveWhere = `p.status = ? AND p.deleted_dt IS NULL
		AND (p.program_status IS NULL OR p.program_status != ?)`
)

func programActiveArgs() []any {
	return []any{enum.StatusActive, enum.StatusInactive}
}

type ProgramRepository struct {
	db database.Executor
}

func NewProgramRepository(db database.Executor) program.IRepository {
	return &ProgramRepository{db: db}
}

func scanProgram(s database.RowScanner) (*models.ProgramModel, error) {
	var m models.ProgramModel
	if err := s.Scan(&m.Id, &m.ProgramId, &m.Label, &m.Description, &m.ImageKey,
		&m.DisplayOrder, &m.Note, &m.ProgramStatus, &m.Status,
		&m.CreateId, &m.CreateDt, &m.ModifyId, &m.ModifyDt); err != nil {
		return nil, err
	}
	return &m, nil
}

// findOneBy is the single-row read helper. `where` is a package-controlled
// SQL fragment; args supply placeholders. programActiveWhere is prepended
// so every read excludes soft-deleted and inactive rows.
func (r *ProgramRepository) findOneBy(ctx context.Context, where string, args ...any) (*program.Program, error) {
	fullArgs := append(programActiveArgs(), args...)
	query := `SELECT ` + programColumns + ` FROM ` + programTable + ` p` +
		` WHERE ` + programActiveWhere + ` AND (` + where + `)`

	m, err := scanProgram(r.db.QueryRow(ctx, query, fullArgs...))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("program repo find (%s): %w", where, err)
	}
	return ModelToDomainProgram(m), nil
}

func (r *ProgramRepository) FindByProgramId(ctx context.Context, programId int64) (*program.Program, error) {
	return r.findOneBy(ctx, "p.program_id = ?", programId)
}

func (r *ProgramRepository) ListPrograms(ctx context.Context, params *program.ListProgramsParams) ([]*program.Program, *pagination.Pagination, error) {
	countQuery := `SELECT COUNT(*) FROM ` + programTable + ` p` +
		` WHERE ` + programActiveWhere
	var total int64
	if err := r.db.QueryRow(ctx, countQuery, programActiveArgs()...).Scan(&total); err != nil {
		return nil, nil, fmt.Errorf("program repo count: %w", err)
	}

	listArgs := programActiveArgs()
	query := `SELECT ` + programColumns + ` FROM ` + programTable + ` p` +
		` WHERE ` + programActiveWhere +
		` ORDER BY p.display_order ASC, p.id ASC`

	var pg *pagination.Pagination
	if !params.TakeAll {
		pg = pagination.NewPagination(params.Page, params.Limit, total)
		query += ` LIMIT ? OFFSET ?`
		listArgs = append(listArgs, pg.Size, pg.Skip)
	} else {
		pg = pagination.NewPagination(1, total, total)
	}

	rows, err := r.db.Query(ctx, query, listArgs...)
	if err != nil {
		return nil, nil, fmt.Errorf("program repo list: %w", err)
	}
	defer rows.Close()

	var programs []*program.Program
	for rows.Next() {
		m, err := scanProgram(rows)
		if err != nil {
			return nil, nil, fmt.Errorf("program repo scan row: %w", err)
		}
		programs = append(programs, ModelToDomainProgram(m))
	}
	if err := rows.Err(); err != nil {
		return nil, nil, fmt.Errorf("program repo rows iteration: %w", err)
	}

	return programs, pg, nil
}

// ListProgramsByIds resolves a batch of programs in a single round-trip.
// Caller (typically a service composing a parent aggregate response) is
// responsible for keying the result by ProgramId() — order is not preserved
// because the IN-clause makes no ordering guarantee.
func (r *ProgramRepository) ListProgramsByIds(ctx context.Context, ids []int64) ([]*program.Program, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	placeholders := make([]string, len(ids))
	args := programActiveArgs()
	for i, id := range ids {
		placeholders[i] = "?"
		args = append(args, id)
	}

	query := `SELECT ` + programColumns + ` FROM ` + programTable + ` p` +
		` WHERE ` + programActiveWhere +
		` AND p.program_id IN (` + strings.Join(placeholders, ",") + `)`

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("program repo list by ids: %w", err)
	}
	defer rows.Close()

	var programs []*program.Program
	for rows.Next() {
		m, err := scanProgram(rows)
		if err != nil {
			return nil, fmt.Errorf("program repo scan row: %w", err)
		}
		programs = append(programs, ModelToDomainProgram(m))
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("program repo rows iteration: %w", err)
	}
	return programs, nil
}

func (r *ProgramRepository) Create(ctx context.Context, p *program.Program) (*program.Program, error) {
	query := `
		INSERT INTO ` + programTable + `
			(program_id, label, description, image_key, display_order, note, program_status, create_dt, modify_dt)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	result, err := r.db.Exec(ctx, query,
		p.ProgramId(), p.Label(), p.Description(), p.ImageKey(),
		p.DisplayOrder(), p.Note(), p.ProgramStatus(), mtime.Now().Time, mtime.Now().Time)
	if err != nil {
		return nil, fmt.Errorf("program repo create: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("program repo last insert id: %w", err)
	}
	return r.findBareById(ctx, id)
}

// findBareById hydrates a program right after INSERT using its surrogate id.
func (r *ProgramRepository) findBareById(ctx context.Context, id int64) (*program.Program, error) {
	return r.findOneBy(ctx, "p.id = ?", id)
}

// Update applies a partial update using COALESCE(?, col) for every
// nullable / non-zero column. display_order is int8 — a literal zero is
// a legal value, so it is always written and not coalesced.
func (r *ProgramRepository) Update(ctx context.Context, p *program.Program) error {
	var label, description, imageKey any
	if p.Label() != "" {
		label = p.Label()
	}
	if p.Description() != "" {
		description = p.Description()
	}
	if p.ImageKey() != nil && *p.ImageKey() != "" {
		imageKey = *p.ImageKey()
	}

	query := `
		UPDATE ` + programTable + `
		SET label         = COALESCE(?, label),
			description   = COALESCE(?, description),
			image_key     = COALESCE(?, image_key),
			display_order = ?,
			note          = COALESCE(?, note),
			modify_dt     = ?
		WHERE program_id = ?
	`
	if _, err := r.db.Exec(ctx, query,
		label, description, imageKey,
		p.DisplayOrder(), p.Note(), mtime.Now().Time, p.ProgramId()); err != nil {
		return fmt.Errorf("program repo update: %w", err)
	}
	return nil
}

func (r *ProgramRepository) SoftDeleteByProgramId(ctx context.Context, programId int64) error {
	query := `
		UPDATE ` + programTable + `
		SET program_status = ?,
			status         = ?,
			deleted_dt     = ?,
			modify_dt      = ?
		WHERE program_id = ?
	`
	now := mtime.Now().Time
	if _, err := r.db.Exec(ctx, query,
		enum.StatusInactive, enum.StatusInactive, now, now, programId); err != nil {
		return fmt.Errorf("program repo soft delete: %w", err)
	}
	return nil
}

func (r *ProgramRepository) ForceDeleteByProgramId(ctx context.Context, programId int64) error {
	query := `
		DELETE FROM ` + programTable + `
		WHERE program_id = ?
	`
	if _, err := r.db.Exec(ctx, query, programId); err != nil {
		return fmt.Errorf("program repo force delete: %w", err)
	}
	return nil
}

func ModelToDomainProgram(m *models.ProgramModel) *program.Program {
	p := program.NewProgram()
	p.SetId(m.Id)
	p.SetProgramId(m.ProgramId)
	p.SetLabel(m.Label)
	p.SetDescription(m.Description)
	p.SetImageKey(m.ImageKey)
	p.SetDisplayOrder(m.DisplayOrder)
	p.SetNote(m.Note)
	p.SetProgramStatus(m.ProgramStatus)
	p.SetStatus(m.Status)
	p.SetCreateId(m.CreateId)
	p.SetCreateDt(mtime.MathTime{Time: m.CreateDt})
	p.SetModifyId(m.ModifyId)
	p.SetModifyDt(mtime.MathTime{Time: m.ModifyDt})
	return p
}
