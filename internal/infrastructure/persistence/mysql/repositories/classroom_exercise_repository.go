package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	domain "math-ai.com/math-ai/internal/domain/classroomexercise"
	mtime "math-ai.com/math-ai/internal/domain/shared/time"
	"math-ai.com/math-ai/internal/infrastructure/database"
	"math-ai.com/math-ai/internal/infrastructure/persistence/mysql/models"
	"math-ai.com/math-ai/internal/shared/enum"
	"math-ai.com/math-ai/internal/shared/pagination"
)

const (
	classroomExerciseTable = "ma_classroom_exercises"

	classroomExerciseColumns = `e.id, e.classroom_exercise_id, e.classroom_id,
		e.creator_profile_id, e.visibility, e.program_id,
		e.title, e.chapter_name, e.lesson_name, e.total_questions,
		e.questions, e.answers, e.start_date, e.end_date,
		e.note, e.exercise_status, e.status,
		e.create_id, e.create_dt, e.modify_id, e.modify_dt`

	// classroomExerciseActiveWhere excludes system-inactive and
	// business-DELETED rows but keeps ARCHIVED visible — archived
	// exercises are still surfaced to the UI as read-only history.
	classroomExerciseActiveWhere = `e.status = ? AND e.deleted_dt IS NULL
		AND (e.exercise_status IS NULL OR e.exercise_status != ?)`
)

func classroomExerciseActiveArgs() []any {
	return []any{enum.StatusActive, enum.ClassroomExerciseStatusTypeDeleted}
}

type ClassroomExerciseRepository struct {
	db database.Executor
}

func NewClassroomExerciseRepository(db database.Executor) domain.IRepository {
	return &ClassroomExerciseRepository{db: db}
}

func scanClassroomExercise(s database.RowScanner) (*models.ClassroomExerciseModel, error) {
	var m models.ClassroomExerciseModel
	if err := s.Scan(&m.Id, &m.ClassroomExerciseId, &m.ClassroomId,
		&m.CreatorProfileId, &m.Visibility, &m.ProgramId,
		&m.Title, &m.ChapterName, &m.LessonName, &m.TotalQuestions,
		&m.Questions, &m.Answers, &m.StartDate, &m.EndDate,
		&m.Note, &m.ExerciseStatus, &m.Status,
		&m.CreateId, &m.CreateDt, &m.ModifyId, &m.ModifyDt); err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *ClassroomExerciseRepository) findOneBy(ctx context.Context, where string, args ...any) (*domain.Exercise, error) {
	fullArgs := append(classroomExerciseActiveArgs(), args...)
	query := `SELECT ` + classroomExerciseColumns + ` FROM ` + classroomExerciseTable + ` e WHERE ` +
		classroomExerciseActiveWhere + ` AND (` + where + `)`

	m, err := scanClassroomExercise(r.db.QueryRow(ctx, query, fullArgs...))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("classroom exercise repo find (%s): %w", where, err)
	}
	return modelToDomainClassroomExercise(m), nil
}

func (r *ClassroomExerciseRepository) findBareById(ctx context.Context, id int64) (*domain.Exercise, error) {
	args := append(classroomExerciseActiveArgs(), id)
	query := `SELECT ` + classroomExerciseColumns + ` FROM ` + classroomExerciseTable + ` e WHERE ` +
		classroomExerciseActiveWhere + ` AND e.id = ?`

	m, err := scanClassroomExercise(r.db.QueryRow(ctx, query, args...))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("classroom exercise repo find bare by id: %w", err)
	}
	return modelToDomainClassroomExercise(m), nil
}

func (r *ClassroomExerciseRepository) FindByClassroomExerciseId(ctx context.Context, id int64) (*domain.Exercise, error) {
	return r.findOneBy(ctx, "e.classroom_exercise_id = ?", id)
}

func (r *ClassroomExerciseRepository) ListExercises(ctx context.Context, params domain.ListExercisesParams) ([]*domain.Exercise, *pagination.Pagination, error) {
	if params.Limit <= 0 {
		params.Limit = pagination.DefaultPageSize
	}
	if params.Page < 1 {
		params.Page = 1
	}

	filterClause, filterArgs := buildClassroomExerciseFilter(params)
	orderBy := buildClassroomExerciseOrderBy(params)

	countArgs := append(classroomExerciseActiveArgs(), filterArgs...)
	countQuery := `SELECT COUNT(*) FROM ` + classroomExerciseTable + ` e WHERE ` +
		classroomExerciseActiveWhere + filterClause

	var total int64
	if err := r.db.QueryRow(ctx, countQuery, countArgs...).Scan(&total); err != nil {
		return nil, nil, fmt.Errorf("classroom exercise repo count: %w", err)
	}

	pg := pagination.NewPagination(params.Page, params.Limit, total)

	listArgs := append(classroomExerciseActiveArgs(), filterArgs...)
	listArgs = append(listArgs, pg.Size, pg.Skip)
	query := `SELECT ` + classroomExerciseColumns + ` FROM ` + classroomExerciseTable + ` e WHERE ` +
		classroomExerciseActiveWhere + filterClause +
		` ORDER BY ` + orderBy + ` LIMIT ? OFFSET ?`

	rows, err := r.db.Query(ctx, query, listArgs...)
	if err != nil {
		return nil, nil, fmt.Errorf("classroom exercise repo list: %w", err)
	}
	defer rows.Close()

	var out []*domain.Exercise
	for rows.Next() {
		m, err := scanClassroomExercise(rows)
		if err != nil {
			return nil, nil, fmt.Errorf("classroom exercise repo scan row: %w", err)
		}
		out = append(out, modelToDomainClassroomExercise(m))
	}
	if err := rows.Err(); err != nil {
		return nil, nil, fmt.Errorf("classroom exercise repo rows iteration: %w", err)
	}
	return out, pg, nil
}

// buildClassroomExerciseFilter composes the parameterised WHERE
// fragments used by both the count and the page query in ListExercises.
// All user-supplied values arrive as bound parameters; only the
// fragment string itself is composed from constants. The visibility
// access gate (PUBLIC OR creator=caller) is always applied — the
// optional user-facing visibility filter ANDs on top, so a student
// asking for PRIVATE rows still gets only the rows they own.
func buildClassroomExerciseFilter(params domain.ListExercisesParams) (string, []any) {
	var (
		clause strings.Builder
		args   []any
	)
	if params.ClassroomID != 0 {
		clause.WriteString(` AND e.classroom_id = ?`)
		args = append(args, params.ClassroomID)
	}
	if params.Status != nil && *params.Status != "" {
		clause.WriteString(` AND e.exercise_status = ?`)
		args = append(args, *params.Status)
	}
	// Visibility access gate.
	if params.CallerProfileID != 0 {
		clause.WriteString(` AND (e.visibility = ? OR e.creator_profile_id = ?)`)
		args = append(args,
			enum.ClassroomExerciseVisibilityPublic, params.CallerProfileID)
	} else {
		clause.WriteString(` AND e.visibility = ?`)
		args = append(args, enum.ClassroomExerciseVisibilityPublic)
	}
	// Optional user-facing filters.
	if params.Visibility != nil && *params.Visibility != "" {
		clause.WriteString(` AND e.visibility = ?`)
		args = append(args, *params.Visibility)
	}
	if params.CreatorProfileID != nil && *params.CreatorProfileID != 0 {
		clause.WriteString(` AND e.creator_profile_id = ?`)
		args = append(args, *params.CreatorProfileID)
	}
	if params.ProgramID != nil && *params.ProgramID != 0 {
		clause.WriteString(` AND e.program_id = ?`)
		args = append(args, *params.ProgramID)
	}
	if params.ChapterName != nil && *params.ChapterName != "" {
		clause.WriteString(` AND e.chapter_name = ?`)
		args = append(args, *params.ChapterName)
	}
	if params.LessonName != nil && *params.LessonName != "" {
		clause.WriteString(` AND e.lesson_name = ?`)
		args = append(args, *params.LessonName)
	}
	if params.Search != nil && *params.Search != "" {
		pattern := "%" + escapeLikePattern(*params.Search) + "%"
		clause.WriteString(` AND (e.title LIKE ? ESCAPE '\\' OR e.chapter_name LIKE ? ESCAPE '\\' OR e.lesson_name LIKE ? ESCAPE '\\')`)
		args = append(args, pattern, pattern, pattern)
	}
	return clause.String(), args
}

// buildClassroomExerciseOrderBy maps the validated sort_by token to a
// real column. Unknown tokens fall back to created_dt DESC so the
// function is total — the validator is the source of truth for what's
// "allowed" in the first place.
func buildClassroomExerciseOrderBy(params domain.ListExercisesParams) string {
	column := "e.create_dt"
	if params.SortBy != nil {
		switch *params.SortBy {
		case "created":
			column = "e.create_dt"
		case "modified":
			column = "e.modify_dt"
		case "title":
			column = "e.title"
		case "start_date":
			column = "e.start_date"
		}
	}
	direction := "DESC"
	if params.SortOrder != nil && *params.SortOrder == "asc" {
		direction = "ASC"
	}
	// Stable secondary sort on e.id so equal primary keys paginate
	// deterministically.
	return column + " " + direction + ", e.id " + direction
}

// escapeLikePattern escapes the two LIKE wildcards plus the backslash
// so user-supplied search text matches literally. The query uses
// ESCAPE '\\' to make backslash the escape char.
func escapeLikePattern(s string) string {
	replacer := strings.NewReplacer(
		`\`, `\\`,
		`%`, `\%`,
		`_`, `\_`,
	)
	return replacer.Replace(s)
}

func (r *ClassroomExerciseRepository) Create(ctx context.Context, e *domain.Exercise) (*domain.Exercise, error) {
	var startArg, endArg any
	if e.StartDate().IsValid() {
		startArg = e.StartDate().Time
	}
	if e.EndDate().IsValid() {
		endArg = e.EndDate().Time
	}

	visibility := e.Visibility()
	if visibility == "" {
		visibility = string(enum.ClassroomExerciseVisibilityPublic)
	}

	query := `
		INSERT INTO ` + classroomExerciseTable + `
			(classroom_exercise_id, classroom_id, creator_profile_id, visibility,
			 program_id,
			 title, chapter_name, lesson_name, total_questions,
			 questions, answers, start_date, end_date,
			 note, exercise_status, create_id)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	result, err := r.db.Exec(ctx, query,
		e.ClassroomExerciseId(), e.ClassroomId(), e.CreatorProfileId(), visibility,
		e.ProgramId(),
		e.Title(), e.ChapterName(), e.LessonName(), e.TotalQuestions(),
		e.Questions(), e.Answers(), startArg, endArg,
		e.Note(), e.ExerciseStatus(), e.CreateId())
	if err != nil {
		return nil, fmt.Errorf("classroom exercise repo create: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("classroom exercise repo last insert id: %w", err)
	}
	return r.findBareById(ctx, id)
}

// Update is the COALESCE patch used by the metadata-edit path. The
// AI-generated questions/answers columns are intentionally absent — a
// regeneration would be a separate command that overwrites them
// directly, not a patch.
func (r *ClassroomExerciseRepository) Update(ctx context.Context, classroomExerciseId int64, patch domain.UpdatePatch) error {
	var startArg, endArg any
	if patch.StartDate != nil {
		if patch.StartDate.IsValid() {
			startArg = patch.StartDate.Time
		} else {
			// explicit clear — pass typed NULL via sql.NullTime so the
			// COALESCE doesn't see a Go nil and short-circuit.
			startArg = sql.NullTime{}
		}
	}
	if patch.EndDate != nil {
		if patch.EndDate.IsValid() {
			endArg = patch.EndDate.Time
		} else {
			endArg = sql.NullTime{}
		}
	}

	query := `
		UPDATE ` + classroomExerciseTable + `
		SET title           = COALESCE(?, title),
			chapter_name    = COALESCE(?, chapter_name),
			lesson_name     = COALESCE(?, lesson_name),
			start_date      = COALESCE(?, start_date),
			end_date        = COALESCE(?, end_date),
			note            = COALESCE(?, note),
			exercise_status = COALESCE(?, exercise_status),
			visibility      = COALESCE(?, visibility),
			modify_id       = COALESCE(?, modify_id),
			modify_dt       = ?
		WHERE classroom_exercise_id = ?
	`
	if _, err := r.db.Exec(ctx, query,
		patch.Title, patch.ChapterName, patch.LessonName,
		startArg, endArg,
		patch.Note, patch.ExerciseStatus, patch.Visibility, patch.ModifyID,
		mtime.Now().Time, classroomExerciseId); err != nil {
		return fmt.Errorf("classroom exercise repo update: %w", err)
	}
	return nil
}

func (r *ClassroomExerciseRepository) SoftDelete(ctx context.Context, classroomExerciseId int64, actorID *int64) error {
	query := `
		UPDATE ` + classroomExerciseTable + `
		SET exercise_status = ?,
			status          = ?,
			deleted_dt      = ?,
			modify_id       = COALESCE(?, modify_id),
			modify_dt       = ?
		WHERE classroom_exercise_id = ?
	`
	now := mtime.Now().Time
	if _, err := r.db.Exec(ctx, query,
		enum.ClassroomExerciseStatusTypeDeleted, enum.StatusInactive,
		now, actorID, now, classroomExerciseId); err != nil {
		return fmt.Errorf("classroom exercise repo soft delete: %w", err)
	}
	return nil
}

func modelToDomainClassroomExercise(m *models.ClassroomExerciseModel) *domain.Exercise {
	e := domain.NewExercise()
	e.SetId(m.Id)
	e.SetClassroomExerciseId(m.ClassroomExerciseId)
	e.SetClassroomId(m.ClassroomId)
	if m.CreatorProfileId != nil {
		e.SetCreatorProfileId(*m.CreatorProfileId)
	}
	e.SetVisibility(m.Visibility)
	e.SetProgramId(m.ProgramId)
	e.SetTitle(m.Title)
	e.SetChapterName(m.ChapterName)
	e.SetLessonName(m.LessonName)
	e.SetTotalQuestions(m.TotalQuestions)
	e.SetQuestions(m.Questions)
	e.SetAnswers(m.Answers)
	if m.StartDate != nil {
		e.SetStartDate(mtime.MathTime{Time: *m.StartDate})
	}
	if m.EndDate != nil {
		e.SetEndDate(mtime.MathTime{Time: *m.EndDate})
	}
	e.SetNote(m.Note)
	e.SetExerciseStatus(m.ExerciseStatus)
	e.SetStatus(m.Status)
	e.SetCreateId(m.CreateId)
	e.SetCreateDt(mtime.MathTime{Time: m.CreateDt})
	e.SetModifyId(m.ModifyId)
	e.SetModifyDt(mtime.MathTime{Time: m.ModifyDt})
	return e
}
