package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"math-ai.com/math-ai/internal/domain/classroom"
	mtime "math-ai.com/math-ai/internal/domain/shared/time"
	"math-ai.com/math-ai/internal/infrastructure/database"
	"math-ai.com/math-ai/internal/infrastructure/persistence/mysql/models"
	"math-ai.com/math-ai/internal/shared/enum"
)

const (
	classroomProgramTable = "ma_classroom_programs"

	classroomProgramColumns = `cp.id, cp.classroom_program_id, cp.classroom_id, cp.program_id,
		cp.note, cp.status, cp.create_id, cp.create_dt, cp.modify_id, cp.modify_dt`

	// classroomProgramActiveWhere mirrors the convention used by every
	// other repo: filter out system-inactive and soft-deleted rows.
	// Because pairs are hard-deleted in practice, deleted_dt should
	// always be NULL — the filter is here for shape consistency.
	classroomProgramActiveWhere = `cp.status = ? AND cp.deleted_dt IS NULL`
)

func classroomProgramActiveArgs() []any {
	return []any{enum.StatusActive}
}

type ClassroomProgramRepository struct {
	db database.Executor
}

func NewClassroomProgramRepository(db database.Executor) classroom.IClassroomProgramRepository {
	return &ClassroomProgramRepository{db: db}
}

func scanClassroomProgram(s database.RowScanner) (*models.ClassroomProgramModel, error) {
	var m models.ClassroomProgramModel
	if err := s.Scan(&m.Id, &m.ClassroomProgramId, &m.ClassroomId, &m.ProgramId,
		&m.Note, &m.Status, &m.CreateId, &m.CreateDt, &m.ModifyId, &m.ModifyDt); err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *ClassroomProgramRepository) ListProgramIdsByClassroomId(ctx context.Context, classroomId string) ([]string, error) {
	args := append(classroomProgramActiveArgs(), classroomId)
	query := `SELECT cp.program_id FROM ` + classroomProgramTable + ` cp WHERE ` +
		classroomProgramActiveWhere + ` AND cp.classroom_id = ? ORDER BY cp.id ASC`

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("classroom_program repo list by classroom: %w", err)
	}
	defer rows.Close()

	var out []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("classroom_program repo scan row: %w", err)
		}
		out = append(out, id)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("classroom_program repo rows iteration: %w", err)
	}
	return out, nil
}

func (r *ClassroomProgramRepository) ListProgramIdsByClassroomIds(ctx context.Context, classroomIds []string) (map[string][]string, error) {
	if len(classroomIds) == 0 {
		return map[string][]string{}, nil
	}
	placeholders := make([]string, len(classroomIds))
	args := classroomProgramActiveArgs()
	for i, id := range classroomIds {
		placeholders[i] = "?"
		args = append(args, id)
	}

	query := `SELECT cp.classroom_id, cp.program_id FROM ` + classroomProgramTable + ` cp WHERE ` +
		classroomProgramActiveWhere +
		` AND cp.classroom_id IN (` + strings.Join(placeholders, ",") + `)` +
		` ORDER BY cp.id ASC`

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("classroom_program repo list by classrooms: %w", err)
	}
	defer rows.Close()

	out := make(map[string][]string, len(classroomIds))
	for rows.Next() {
		var classroomId, programId string
		if err := rows.Scan(&classroomId, &programId); err != nil {
			return nil, fmt.Errorf("classroom_program repo scan row: %w", err)
		}
		out[classroomId] = append(out[classroomId], programId)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("classroom_program repo rows iteration: %w", err)
	}
	return out, nil
}

func (r *ClassroomProgramRepository) findBareById(ctx context.Context, id int64) (*classroom.ClassroomProgram, error) {
	args := append(classroomProgramActiveArgs(), id)
	query := `SELECT ` + classroomProgramColumns + ` FROM ` + classroomProgramTable + ` cp WHERE ` +
		classroomProgramActiveWhere + ` AND cp.id = ?`

	m, err := scanClassroomProgram(r.db.QueryRow(ctx, query, args...))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("classroom_program repo find bare by id: %w", err)
	}
	return ModelToDomainClassroomProgram(m), nil
}

func (r *ClassroomProgramRepository) Create(ctx context.Context, cp *classroom.ClassroomProgram) (*classroom.ClassroomProgram, error) {
	query := `
		INSERT INTO ` + classroomProgramTable + `
			(classroom_program_id, classroom_id, program_id, note, create_id)
		VALUES (?, ?, ?, ?, ?)
	`
	result, err := r.db.Exec(ctx, query,
		cp.ClassroomProgramId(), cp.ClassroomId(), cp.ProgramId(),
		cp.Note(), cp.CreateId())
	if err != nil {
		return nil, fmt.Errorf("classroom_program repo create: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("classroom_program repo last insert id: %w", err)
	}
	return r.findBareById(ctx, id)
}

func (r *ClassroomProgramRepository) DeleteByPair(ctx context.Context, classroomId, programId string) error {
	query := `DELETE FROM ` + classroomProgramTable +
		` WHERE classroom_id = ? AND program_id = ?`
	if _, err := r.db.Exec(ctx, query, classroomId, programId); err != nil {
		return fmt.Errorf("classroom_program repo delete pair: %w", err)
	}
	return nil
}

func (r *ClassroomProgramRepository) DeleteByClassroomId(ctx context.Context, classroomId string) error {
	query := `DELETE FROM ` + classroomProgramTable + ` WHERE classroom_id = ?`
	if _, err := r.db.Exec(ctx, query, classroomId); err != nil {
		return fmt.Errorf("classroom_program repo delete by classroom: %w", err)
	}
	return nil
}

func ModelToDomainClassroomProgram(m *models.ClassroomProgramModel) *classroom.ClassroomProgram {
	cp := classroom.NewClassroomProgram()
	cp.SetId(m.Id)
	cp.SetClassroomProgramId(m.ClassroomProgramId)
	cp.SetClassroomId(m.ClassroomId)
	cp.SetProgramId(m.ProgramId)
	cp.SetNote(m.Note)
	cp.SetStatus(m.Status)
	cp.SetCreateId(m.CreateId)
	cp.SetCreateDt(mtime.MathTime{Time: m.CreateDt})
	cp.SetModifyId(m.ModifyId)
	cp.SetModifyDt(mtime.MathTime{Time: m.ModifyDt})
	return cp
}
