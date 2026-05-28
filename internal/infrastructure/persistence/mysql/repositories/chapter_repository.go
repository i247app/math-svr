package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"math-ai.com/math-ai/internal/domain/chapter"
	mtime "math-ai.com/math-ai/internal/domain/shared/time"
	"math-ai.com/math-ai/internal/infrastructure/database"
	"math-ai.com/math-ai/internal/infrastructure/persistence/mysql/models"
	"math-ai.com/math-ai/internal/shared/enum"
	"math-ai.com/math-ai/internal/shared/pagination"
)

const (
	chapterTable             = "ma_chapters"
	chapterTranslationsTable = "ma_chapter_translations"

	// LEFT JOIN so a chapter without a translation in the requested
	// language still surfaces with its base label / description.
	chapterFromJoin = chapterTable + ` c
	LEFT JOIN ` + chapterTranslationsTable + ` t
	  ON t.chapter_id = c.chapter_id
	  AND t.language = ?
	  AND t.status = ?
	  AND t.deleted_dt IS NULL`

	chapterColumns = `c.id, c.chapter_id, c.program_id, c.grade_id, c.semester_id,
		COALESCE(t.label, c.label) AS label,
		COALESCE(t.description, c.description) AS description,
		c.display_order, c.note,
		c.chapter_status, c.status,
		c.create_id, c.create_dt, c.modify_id, c.modify_dt`

	chapterActiveWhere = `c.status IN (?) AND c.deleted_dt IS NULL
		AND (c.chapter_status IS NULL OR c.chapter_status != ?)`

	chapterBareActiveWhere = `c.status IN (?) AND c.deleted_dt IS NULL
		AND (c.chapter_status IS NULL OR c.chapter_status != ?)`
)

func chapterJoinArgs(lang enum.LanguageType) []any {
	if lang == "" {
		lang = enum.LanguageTypeVietnamese
	}
	return []any{lang, enum.StatusActive}
}

func chapterActiveArgs() []any {
	return []any{enum.StatusActive, enum.StatusInactive}
}

type ChapterRepository struct {
	db database.Executor
}

func NewChapterRepository(db database.Executor) chapter.IRepository {
	return &ChapterRepository{db: db}
}

func scanChapter(s database.RowScanner) (*models.ChapterModel, error) {
	var m models.ChapterModel
	if err := s.Scan(&m.Id, &m.ChapterId, &m.ProgramId, &m.GradeId, &m.SemesterId,
		&m.Label, &m.Description, &m.DisplayOrder, &m.Note,
		&m.ChapterStatus, &m.Status,
		&m.CreateId, &m.CreateDt, &m.ModifyId, &m.ModifyDt); err != nil {
		return nil, err
	}
	return &m, nil
}

// findOneBy is the single-row read helper. `where` is a package-controlled
// SQL fragment; args supply placeholders. chapterActiveWhere is prepended
// so every read excludes soft-deleted and inactive rows.
func (r *ChapterRepository) findOneBy(ctx context.Context, lang enum.LanguageType, where string, args ...any) (*chapter.Chapter, error) {
	fullArgs := append(chapterJoinArgs(lang), chapterActiveArgs()...)
	fullArgs = append(fullArgs, args...)
	query := `SELECT ` + chapterColumns + ` FROM ` + chapterFromJoin +
		` WHERE ` + chapterActiveWhere + ` AND (` + where + `)`

	m, err := scanChapter(r.db.QueryRow(ctx, query, fullArgs...))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("chapter repo find (%s): %w", where, err)
	}
	return ModelToDomainChapter(m), nil
}

// findBareById hydrates a chapter right after INSERT using its surrogate
// id. The translation rows haven't been written yet at that point, so we
// skip the LEFT JOIN and just read the base label/description.
func (r *ChapterRepository) findBareById(ctx context.Context, id int64) (*chapter.Chapter, error) {
	query := `SELECT c.id, c.chapter_id, c.program_id, c.grade_id, c.semester_id,
			c.label, c.description AS description,
			c.display_order, c.note, c.chapter_status, c.status,
			c.create_id, c.create_dt, c.modify_id, c.modify_dt
		FROM ` + chapterTable + ` c
		WHERE ` + chapterBareActiveWhere + ` AND c.id = ?`
	args := append(chapterActiveArgs(), id)

	m, err := scanChapter(r.db.QueryRow(ctx, query, args...))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("chapter repo find bare by id: %w", err)
	}
	return ModelToDomainChapter(m), nil
}

func (r *ChapterRepository) FindByChapterId(ctx context.Context, chapterId string, language enum.LanguageType) (*chapter.Chapter, error) {
	return r.findOneBy(ctx, language, "c.chapter_id = ?", chapterId)
}

func (r *ChapterRepository) ListChapters(ctx context.Context, params *chapter.ListChaptersParams) ([]*chapter.Chapter, *pagination.Pagination, error) {
	filterWhere, filterArgs := buildChapterListFilterClause(params)

	countArgs := append(chapterJoinArgs(params.Language), chapterActiveArgs()...)
	countArgs = append(countArgs, filterArgs...)
	countQuery := `SELECT COUNT(*) FROM ` + chapterFromJoin +
		` WHERE ` + chapterActiveWhere + filterWhere

	var total int64
	if err := r.db.QueryRow(ctx, countQuery, countArgs...).Scan(&total); err != nil {
		return nil, nil, fmt.Errorf("chapter repo count: %w", err)
	}

	listArgs := append(chapterJoinArgs(params.Language), chapterActiveArgs()...)
	listArgs = append(listArgs, filterArgs...)

	query := `SELECT ` + chapterColumns + ` FROM ` + chapterFromJoin +
		` WHERE ` + chapterActiveWhere + filterWhere +
		` ORDER BY c.display_order ASC, c.id ASC`

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
		return nil, nil, fmt.Errorf("chapter repo list: %w", err)
	}
	defer rows.Close()

	var chapters []*chapter.Chapter
	for rows.Next() {
		m, err := scanChapter(rows)
		if err != nil {
			return nil, nil, fmt.Errorf("chapter repo scan row: %w", err)
		}
		chapters = append(chapters, ModelToDomainChapter(m))
	}
	if err := rows.Err(); err != nil {
		return nil, nil, fmt.Errorf("chapter repo rows iteration: %w", err)
	}
	return chapters, pg, nil
}

// buildChapterListFilterClause appends "AND c.<col> = ?" pairs for each
// non-nil filter, keeping placeholder ordering in lockstep with args.
func buildChapterListFilterClause(params *chapter.ListChaptersParams) (string, []any) {
	if params == nil {
		return "", nil
	}
	var (
		clause string
		args   []any
	)
	if params.ProgramID != nil && *params.ProgramID != "" {
		clause += ` AND c.program_id = ?`
		args = append(args, *params.ProgramID)
	}
	if params.GradeID != nil && *params.GradeID != "" {
		clause += ` AND c.grade_id = ?`
		args = append(args, *params.GradeID)
	}
	if params.SemesterID != nil && *params.SemesterID != "" {
		clause += ` AND c.semester_id = ?`
		args = append(args, *params.SemesterID)
	}
	return clause, args
}

func (r *ChapterRepository) Create(ctx context.Context, c *chapter.Chapter) (*chapter.Chapter, error) {
	query := `
		INSERT INTO ` + chapterTable + `
			(chapter_id, program_id, grade_id, semester_id, label, description, display_order, note, chapter_status)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	result, err := r.db.Exec(ctx, query,
		c.ChapterId(), c.ProgramId(), c.GradeId(), c.SemesterId(),
		c.Label(), c.Description(), c.DisplayOrder(),
		c.Note(), c.ChapterStatus())
	if err != nil {
		return nil, fmt.Errorf("chapter repo create: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("chapter repo last insert id: %w", err)
	}
	return r.findBareById(ctx, id)
}

// Update applies a partial update using COALESCE(?, col) for every column.
// Callers pass empty/zero values to keep the existing column unchanged;
// non-zero values overwrite. display_order is int8 — a literal zero is a
// legal value, so it is always written and not coalesced.
func (r *ChapterRepository) Update(ctx context.Context, c *chapter.Chapter) error {
	var label, description, programID, gradeID, semesterID any
	if c.Label() != "" {
		label = c.Label()
	}
	if c.Description() != "" {
		description = c.Description()
	}
	if c.ProgramId() != "" {
		programID = c.ProgramId()
	}
	if c.GradeId() != "" {
		gradeID = c.GradeId()
	}
	if c.SemesterId() != "" {
		semesterID = c.SemesterId()
	}

	query := `
		UPDATE ` + chapterTable + `
		SET program_id    = COALESCE(?, program_id),
			grade_id      = COALESCE(?, grade_id),
			semester_id   = COALESCE(?, semester_id),
			label         = COALESCE(?, label),
			description   = COALESCE(?, description),
			display_order = ?,
			note          = COALESCE(?, note),
			modify_dt     = ?
		WHERE chapter_id = ?
	`
	if _, err := r.db.Exec(ctx, query,
		programID, gradeID, semesterID, label, description,
		c.DisplayOrder(), c.Note(), mtime.Now().Time, c.ChapterId()); err != nil {
		return fmt.Errorf("chapter repo update: %w", err)
	}
	return nil
}

func (r *ChapterRepository) SoftDeleteByChapterId(ctx context.Context, chapterId string) error {
	query := `
		UPDATE ` + chapterTable + `
		SET chapter_status = ?,
			status         = ?,
			deleted_dt     = ?,
			modify_dt      = ?
		WHERE chapter_id = ?
	`
	now := mtime.Now().Time
	if _, err := r.db.Exec(ctx, query,
		enum.StatusInactive, enum.StatusInactive, now, now, chapterId); err != nil {
		return fmt.Errorf("chapter repo soft delete: %w", err)
	}
	return nil
}

func (r *ChapterRepository) ForceDeleteByChapterId(ctx context.Context, chapterId string) error {
	query := `
		DELETE FROM ` + chapterTable + `
		WHERE chapter_id = ?
	`
	if _, err := r.db.Exec(ctx, query, chapterId); err != nil {
		return fmt.Errorf("chapter repo force delete: %w", err)
	}
	return nil
}

func ModelToDomainChapter(m *models.ChapterModel) *chapter.Chapter {
	c := chapter.NewChapter()
	c.SetId(m.Id)
	c.SetChapterId(m.ChapterId)
	c.SetProgramId(m.ProgramId)
	c.SetGradeId(m.GradeId)
	c.SetSemesterId(m.SemesterId)
	c.SetLabel(m.Label)
	c.SetDescription(m.Description)
	c.SetDisplayOrder(m.DisplayOrder)
	c.SetNote(m.Note)
	c.SetChapterStatus(m.ChapterStatus)
	c.SetStatus(m.Status)
	c.SetCreateId(m.CreateId)
	c.SetCreateDt(mtime.MathTime{Time: m.CreateDt})
	c.SetModifyId(m.ModifyId)
	c.SetModifyDt(mtime.MathTime{Time: m.ModifyDt})
	return c
}
