package repositories

import (
	"context"
	"fmt"
	"strings"

	"math-ai.com/math-ai/internal/domain/program"
	mtime "math-ai.com/math-ai/internal/domain/shared/time"
	"math-ai.com/math-ai/internal/infrastructure/database"
	"math-ai.com/math-ai/internal/infrastructure/persistence/mysql/models"
	"math-ai.com/math-ai/internal/shared/enum"
	"math-ai.com/math-ai/internal/shared/pagination"
)

// Reference-data aggregate: read-only, JOIN'd against the translations table
// so label/description carry the requested language. Note that the base table
// column is misspelled `discription`; the repo isolates that quirk so the
// domain entity exposes a clean `Description()` getter.
const (
	programTable             = "ma_programs"
	programTranslationsTable = "ma_program_translations"

	// LEFT JOIN — programs without a row in the requested language still
	// surface using their base label/discription.
	programFromJoin = programTable + ` p
	LEFT JOIN ` + programTranslationsTable + ` t
	  ON t.program_id = p.program_id
	  AND t.language = ?
	  AND t.status = ?
	  AND t.deleted_dt IS NULL`

	programColumns = `p.id, p.program_id,
		COALESCE(t.label, p.label) AS label,
		COALESCE(t.description, p.discription) AS description,
		p.image_key, p.display_order, p.note,
		p.program_status, p.status,
		p.create_id, p.create_dt, p.modify_id, p.modify_dt`

	programActiveWhere = `p.status IN (?) AND p.deleted_dt IS NULL`
)

func programJoinArgs(lang enum.LanguageType) []any {
	return []any{lang, enum.StatusActive}
}

func programActiveArgs() []any {
	return []any{enum.StatusActive}
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

func (r *ProgramRepository) ListPrograms(ctx context.Context, params *program.ListProgramsParams) ([]*program.Program, *pagination.Pagination, error) {
	lang := params.Language
	if lang == "" {
		lang = enum.LanguageTypeVietnamese
	}

	var total int64
	countQuery := `SELECT COUNT(*) FROM ` + programFromJoin +
		` WHERE ` + programActiveWhere
	countArgs := append(programJoinArgs(lang), programActiveArgs()...)
	if err := r.db.QueryRow(ctx, countQuery, countArgs...).Scan(&total); err != nil {
		return nil, nil, fmt.Errorf("program repo count: %w", err)
	}

	pg := pagination.NewPagination(params.Page, params.Limit, total)

	listArgs := append(programJoinArgs(lang), programActiveArgs()...)
	listArgs = append(listArgs, pg.Size, pg.Skip)
	query := `SELECT ` + programColumns + ` FROM ` + programFromJoin +
		` WHERE ` + programActiveWhere +
		` ORDER BY p.display_order ASC LIMIT ? OFFSET ?`
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
func (r *ProgramRepository) ListProgramsByIds(ctx context.Context, ids []string, lang enum.LanguageType) ([]*program.Program, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	if lang == "" {
		lang = enum.LanguageTypeVietnamese
	}

	placeholders := make([]string, len(ids))
	args := append(programJoinArgs(lang), programActiveArgs()...)
	for i, id := range ids {
		placeholders[i] = "?"
		args = append(args, id)
	}

	query := `SELECT ` + programColumns + ` FROM ` + programFromJoin +
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
