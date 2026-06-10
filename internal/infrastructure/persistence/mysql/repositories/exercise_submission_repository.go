package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	domain "math-ai.com/math-ai/internal/domain/exercise"
	mtime "math-ai.com/math-ai/internal/domain/shared/time"
	"math-ai.com/math-ai/internal/infrastructure/database"
	"math-ai.com/math-ai/internal/infrastructure/persistence/mysql/models"
	"math-ai.com/math-ai/internal/shared/enum"
	"math-ai.com/math-ai/internal/shared/pagination"
)

const (
	exerciseSubmissionTable = "ma_exercise_submissions"

	exerciseSubmissionColumns = `s.id, s.classroom_exercise_submission_id,
		s.classroom_exercise_id, s.classroom_id, s.profile_id,
		s.answers, s.ai_review,
		s.total_questions, s.correct_number, s.score_percentage,
		s.submitted_dt, s.graded_dt,
		s.note, s.submission_status, s.status,
		s.create_id, s.create_dt, s.modify_id, s.modify_dt`

	// Active-where mirrors the classroom-exercise convention: hide rows
	// the system has flagged INACTIVE plus the business-DELETED tombstone.
	exerciseSubmissionActiveWhere = `s.status = ? AND s.deleted_dt IS NULL
		AND (s.submission_status IS NULL OR s.submission_status != ?)`
)

func exerciseSubmissionActiveArgs() []any {
	return []any{enum.StatusActive, enum.ClassroomExerciseSubmissionStatusDeleted}
}

type ExerciseSubmissionRepository struct {
	db database.Executor
}

func NewExerciseSubmissionRepository(db database.Executor) domain.ISubmissionRepository {
	return &ExerciseSubmissionRepository{db: db}
}

func scanExerciseSubmission(s database.RowScanner) (*models.ExerciseSubmissionModel, error) {
	var m models.ExerciseSubmissionModel
	if err := s.Scan(&m.Id, &m.ClassroomExerciseSubmissionId,
		&m.ClassroomExerciseId, &m.ClassroomId, &m.ProfileId,
		&m.Answers, &m.AIReview,
		&m.TotalQuestions, &m.CorrectNumber, &m.ScorePercentage,
		&m.SubmittedDt, &m.GradedDt,
		&m.Note, &m.SubmissionStatus, &m.Status,
		&m.CreateId, &m.CreateDt, &m.ModifyId, &m.ModifyDt); err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *ExerciseSubmissionRepository) findOneBy(ctx context.Context, where string, args ...any) (*domain.Submission, error) {
	fullArgs := append(exerciseSubmissionActiveArgs(), args...)
	query := `SELECT ` + exerciseSubmissionColumns + ` FROM ` + exerciseSubmissionTable + ` s WHERE ` +
		exerciseSubmissionActiveWhere + ` AND (` + where + `)`

	m, err := scanExerciseSubmission(r.db.QueryRow(ctx, query, fullArgs...))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("classroom exercise submission repo find (%s): %w", where, err)
	}
	return modelToDomainClassroomExerciseSubmission(m), nil
}

func (r *ExerciseSubmissionRepository) findBareById(ctx context.Context, id int64) (*domain.Submission, error) {
	args := append(exerciseSubmissionActiveArgs(), id)
	query := `SELECT ` + exerciseSubmissionColumns + ` FROM ` + exerciseSubmissionTable + ` s WHERE ` +
		exerciseSubmissionActiveWhere + ` AND s.id = ?`

	m, err := scanExerciseSubmission(r.db.QueryRow(ctx, query, args...))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("classroom exercise submission repo find bare by id: %w", err)
	}
	return modelToDomainClassroomExerciseSubmission(m), nil
}

func (r *ExerciseSubmissionRepository) FindBySubmissionId(ctx context.Context, submissionId int64) (*domain.Submission, error) {
	return r.findOneBy(ctx, "s.classroom_exercise_submission_id = ?", submissionId)
}

func (r *ExerciseSubmissionRepository) FindByExerciseAndProfile(ctx context.Context, classroomExerciseId, profileId int64) (*domain.Submission, error) {
	return r.findOneBy(ctx, "s.classroom_exercise_id = ? AND s.profile_id = ?", classroomExerciseId, profileId)
}

// ListByExerciseAndProfileIds returns the active submission rows for
// (classroom_exercise_id, profile_id IN ids). One IN-query; rows that
// don't exist are simply absent from the slice — callers index them
// by profile_id when hydrating.
func (r *ExerciseSubmissionRepository) ListByExerciseAndProfileIds(ctx context.Context, classroomExerciseId int64, profileIds []int64) ([]*domain.Submission, error) {
	if classroomExerciseId == 0 || len(profileIds) == 0 {
		return nil, nil
	}
	placeholders := strings.Repeat("?,", len(profileIds))
	placeholders = placeholders[:len(placeholders)-1]
	query := `SELECT ` + exerciseSubmissionColumns + ` FROM ` + exerciseSubmissionTable + ` s WHERE ` +
		exerciseSubmissionActiveWhere +
		` AND s.classroom_exercise_id = ? AND s.profile_id IN (` + placeholders + `)`
	args := append(exerciseSubmissionActiveArgs(), classroomExerciseId)
	for _, id := range profileIds {
		args = append(args, id)
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("classroom exercise submission repo list by exercise and profile ids: %w", err)
	}
	defer rows.Close()

	out := make([]*domain.Submission, 0, len(profileIds))
	for rows.Next() {
		m, err := scanExerciseSubmission(rows)
		if err != nil {
			return nil, fmt.Errorf("classroom exercise submission repo list by exercise and profile ids scan: %w", err)
		}
		out = append(out, modelToDomainClassroomExerciseSubmission(m))
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("classroom exercise submission repo list by exercise and profile ids iteration: %w", err)
	}
	return out, nil
}

// ListSubmittedExerciseIdsByProfile returns the subset of exerciseIds
// the profile has already submitted (any non-DELETED row counts as
// SUBMITTED for the student-facing list hint). One IN-query, projecting
// only the exercise id so the index ix_exercise_profile_submitted
// covers the lookup without touching the LONGTEXT columns.
func (r *ExerciseSubmissionRepository) ListSubmittedExerciseIdsByProfile(ctx context.Context, profileId int64, exerciseIds []int64) (map[int64]struct{}, error) {
	if profileId == 0 || len(exerciseIds) == 0 {
		return map[int64]struct{}{}, nil
	}
	placeholders := strings.Repeat("?,", len(exerciseIds))
	placeholders = placeholders[:len(placeholders)-1]
	query := `SELECT s.classroom_exercise_id FROM ` + exerciseSubmissionTable + ` s WHERE ` +
		exerciseSubmissionActiveWhere +
		` AND s.profile_id = ? AND s.classroom_exercise_id IN (` + placeholders + `)`
	args := append(exerciseSubmissionActiveArgs(), profileId)
	for _, id := range exerciseIds {
		args = append(args, id)
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("classroom exercise submission repo list submitted exercise ids: %w", err)
	}
	defer rows.Close()

	out := make(map[int64]struct{}, len(exerciseIds))
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("classroom exercise submission repo list submitted scan: %w", err)
		}
		out[id] = struct{}{}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("classroom exercise submission repo list submitted iteration: %w", err)
	}
	return out, nil
}

func (r *ExerciseSubmissionRepository) ListSubmissions(ctx context.Context, params domain.ListSubmissionsParams) ([]*domain.Submission, *pagination.Pagination, error) {
	if params.Limit <= 0 {
		params.Limit = pagination.DefaultPageSize
	}
	if params.Page < 1 {
		params.Page = 1
	}

	filterClause, filterArgs := buildClassroomExerciseSubmissionFilter(params)
	orderBy := buildClassroomExerciseSubmissionOrderBy(params)

	countArgs := append(exerciseSubmissionActiveArgs(), filterArgs...)
	countQuery := `SELECT COUNT(*) FROM ` + exerciseSubmissionTable + ` s WHERE ` +
		exerciseSubmissionActiveWhere + filterClause

	var total int64
	if err := r.db.QueryRow(ctx, countQuery, countArgs...).Scan(&total); err != nil {
		return nil, nil, fmt.Errorf("classroom exercise submission repo count: %w", err)
	}

	pg := pagination.NewPagination(params.Page, params.Limit, total)

	listArgs := append(exerciseSubmissionActiveArgs(), filterArgs...)
	listArgs = append(listArgs, pg.Size, pg.Skip)
	query := `SELECT ` + exerciseSubmissionColumns + ` FROM ` + exerciseSubmissionTable + ` s WHERE ` +
		exerciseSubmissionActiveWhere + filterClause +
		` ORDER BY ` + orderBy + ` LIMIT ? OFFSET ?`

	rows, err := r.db.Query(ctx, query, listArgs...)
	if err != nil {
		return nil, nil, fmt.Errorf("classroom exercise submission repo list: %w", err)
	}
	defer rows.Close()

	var out []*domain.Submission
	for rows.Next() {
		m, err := scanExerciseSubmission(rows)
		if err != nil {
			return nil, nil, fmt.Errorf("classroom exercise submission repo scan row: %w", err)
		}
		out = append(out, modelToDomainClassroomExerciseSubmission(m))
	}
	if err := rows.Err(); err != nil {
		return nil, nil, fmt.Errorf("classroom exercise submission repo rows iteration: %w", err)
	}
	return out, pg, nil
}

func buildClassroomExerciseSubmissionFilter(params domain.ListSubmissionsParams) (string, []any) {
	var (
		clause strings.Builder
		args   []any
	)
	if params.ClassroomExerciseID != 0 {
		clause.WriteString(` AND s.classroom_exercise_id = ?`)
		args = append(args, params.ClassroomExerciseID)
	}
	if params.ClassroomID != 0 {
		clause.WriteString(` AND s.classroom_id = ?`)
		args = append(args, params.ClassroomID)
	}
	if params.ProfileID != 0 {
		clause.WriteString(` AND s.profile_id = ?`)
		args = append(args, params.ProfileID)
	}
	if params.Status != nil && *params.Status != "" {
		clause.WriteString(` AND s.submission_status = ?`)
		args = append(args, *params.Status)
	}
	return clause.String(), args
}

func buildClassroomExerciseSubmissionOrderBy(params domain.ListSubmissionsParams) string {
	column := "s.submitted_dt"
	if params.SortBy != nil {
		switch *params.SortBy {
		case "submitted":
			column = "s.submitted_dt"
		case "graded":
			column = "s.graded_dt"
		case "score":
			column = "s.score_percentage"
		case "created":
			column = "s.create_dt"
		}
	}
	direction := "DESC"
	if params.SortOrder != nil && *params.SortOrder == "asc" {
		direction = "ASC"
	}
	return column + " " + direction + ", s.id " + direction
}

func (r *ExerciseSubmissionRepository) Create(ctx context.Context, sub *domain.Submission) (*domain.Submission, error) {
	var submittedArg, gradedArg any
	if sub.SubmittedDt().IsValid() {
		submittedArg = sub.SubmittedDt().Time
	}
	if sub.GradedDt().IsValid() {
		gradedArg = sub.GradedDt().Time
	}

	query := `
		INSERT INTO ` + exerciseSubmissionTable + `
			(classroom_exercise_submission_id, classroom_exercise_id, classroom_id, profile_id,
			 answers, ai_review,
			 total_questions, correct_number, score_percentage,
			 submitted_dt, graded_dt,
			 note, submission_status, create_id, create_dt, modify_dt)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	result, err := r.db.Exec(ctx, query,
		sub.ClassroomExerciseSubmissionId(), sub.ClassroomExerciseId(), sub.ClassroomId(), sub.ProfileId(),
		sub.Answers(), sub.AIReview(),
		sub.TotalQuestions(), sub.CorrectNumber(), sub.ScorePercentage(),
		submittedArg, gradedArg,
		sub.Note(), sub.SubmissionStatus(), sub.CreateId(),
		mtime.Now().Time, mtime.Now().Time)
	if err != nil {
		return nil, fmt.Errorf("classroom exercise submission repo create: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("classroom exercise submission repo last insert id: %w", err)
	}
	return r.findBareById(ctx, id)
}

// UpdateGrading applies the bot's grading result via COALESCE so a
// caller can refine partial scores without nulling out prior progress.
// GradedDt is written explicitly (non-COALESCE) when valid because the
// row is typically transitioning from un-graded to graded in one shot.
func (r *ExerciseSubmissionRepository) UpdateGrading(ctx context.Context, submissionId int64, patch domain.GradingPatch) error {
	var gradedArg any
	if patch.GradedDt.IsValid() {
		gradedArg = patch.GradedDt.Time
	}

	query := `
		UPDATE ` + exerciseSubmissionTable + `
		SET ai_review         = COALESCE(?, ai_review),
			total_questions   = COALESCE(?, total_questions),
			correct_number    = COALESCE(?, correct_number),
			score_percentage  = COALESCE(?, score_percentage),
			graded_dt         = COALESCE(?, graded_dt),
			submission_status = COALESCE(?, submission_status),
			modify_id         = COALESCE(?, modify_id),
			modify_dt         = ?
		WHERE classroom_exercise_submission_id = ?
	`
	if _, err := r.db.Exec(ctx, query,
		patch.AIReview, patch.TotalQuestions, patch.CorrectNumber, patch.ScorePercentage,
		gradedArg, patch.SubmissionStatus, patch.ModifyID,
		mtime.Now().Time, submissionId); err != nil {
		return fmt.Errorf("classroom exercise submission repo update grading: %w", err)
	}
	return nil
}

func (r *ExerciseSubmissionRepository) SoftDelete(ctx context.Context, submissionId int64, actorID *int64) error {
	query := `
		UPDATE ` + exerciseSubmissionTable + `
		SET submission_status = ?,
			status            = ?,
			deleted_dt        = ?,
			modify_id         = COALESCE(?, modify_id),
			modify_dt         = ?
		WHERE classroom_exercise_submission_id = ?
	`
	now := mtime.Now().Time
	if _, err := r.db.Exec(ctx, query,
		enum.ClassroomExerciseSubmissionStatusDeleted, enum.StatusInactive,
		now, actorID, now, submissionId); err != nil {
		return fmt.Errorf("classroom exercise submission repo soft delete: %w", err)
	}
	return nil
}

// bucketLabelExpr returns the SQL expression that formats a row's
// submitted_dt into the chart's bucket label, plus the number of tz
// placeholder binds that expression consumes. The bucket value is
// validator-normalised upstream (enum.ProgressBucket); an unrecognised
// value falls through to DAY so the query stays well-formed.
//
// DAY  → 1 tz placeholder (one CONVERT_TZ call)
// WEEK → 2 tz placeholders (two CONVERT_TZ calls, for the WEEKDAY
//                           shift to Monday and the date formatting)
// MONTH → 1 tz placeholder
func bucketLabelExpr(bucket string) (string, int) {
	switch bucket {
	case string(enum.ProgressBucketWeek):
		return `DATE_FORMAT(DATE_SUB(CONVERT_TZ(s.submitted_dt,'+00:00',?),` +
			` INTERVAL WEEKDAY(CONVERT_TZ(s.submitted_dt,'+00:00',?)) DAY),'%Y-%m-%d')`, 2
	case string(enum.ProgressBucketMonth):
		return `DATE_FORMAT(CONVERT_TZ(s.submitted_dt,'+00:00',?),'%Y-%m')`, 1
	default: // DAY (and any unrecognised value)
		return `DATE_FORMAT(CONVERT_TZ(s.submitted_dt,'+00:00',?),'%Y-%m-%d')`, 1
	}
}

// ListBucketedScores aggregates submission scores into chart-shaped
// rows. The bucket-label SQL expression lives in bucketLabelExpr; this
// method composes it into a single GROUP BY + ORDER BY query against
// ix_classroom_submitted (covered by migration 021).
//
// Args ordering matches the SQL placeholder order left-to-right:
//
//	1. tz arg(s) for the SELECT expression (1 for DAY/MONTH, 2 for WEEK)
//	2. active-where args (status, deleted-status)
//	3. classroom_id, from, to
//	4. optional filter profile_id
//
// The score_percentage IS NOT NULL filter is added explicitly because
// the active-where only covers row-lifecycle (status / deleted_dt /
// submission_status) — un-graded SUBMITTED rows still pass active-where
// and must be excluded here so the AVG/MIN/MAX numbers reflect graded
// scores only.
func (r *ExerciseSubmissionRepository) ListBucketedScores(ctx context.Context, params domain.BucketedScoresParams) ([]*domain.BucketedScoreRow, error) {
	if params.ClassroomID == 0 {
		return nil, nil
	}

	labelExpr, tzPlaceholders := bucketLabelExpr(params.Bucket)

	var b strings.Builder
	b.WriteString(`SELECT `)
	b.WriteString(labelExpr)
	b.WriteString(` AS bucket_label,
		COUNT(*) AS n,
		AVG(s.score_percentage) AS avg_pct,
		MIN(s.score_percentage) AS min_pct,
		MAX(s.score_percentage) AS max_pct
		FROM `)
	b.WriteString(exerciseSubmissionTable)
	b.WriteString(` s WHERE `)
	b.WriteString(exerciseSubmissionActiveWhere)
	b.WriteString(` AND s.classroom_id = ?
		AND s.submitted_dt IS NOT NULL
		AND s.score_percentage IS NOT NULL
		AND s.submitted_dt BETWEEN ? AND ?`)

	args := make([]any, 0, tzPlaceholders+len(exerciseSubmissionActiveArgs())+4)
	for range tzPlaceholders {
		args = append(args, params.TzOffset)
	}
	args = append(args, exerciseSubmissionActiveArgs()...)
	args = append(args, params.ClassroomID, params.From.Time, params.To.Time)

	if params.FilterProfileID != nil {
		b.WriteString(` AND s.profile_id = ?`)
		args = append(args, *params.FilterProfileID)
	}

	b.WriteString(` GROUP BY bucket_label ORDER BY bucket_label ASC`)

	rows, err := r.db.Query(ctx, b.String(), args...)
	if err != nil {
		return nil, fmt.Errorf("classroom exercise submission repo list bucketed scores: %w", err)
	}
	defer rows.Close()

	out := make([]*domain.BucketedScoreRow, 0)
	for rows.Next() {
		var row domain.BucketedScoreRow
		if err := rows.Scan(&row.BucketLabel, &row.SubmissionCount, &row.AvgPct, &row.MinPct, &row.MaxPct); err != nil {
			return nil, fmt.Errorf("classroom exercise submission repo list bucketed scores scan: %w", err)
		}
		out = append(out, &row)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("classroom exercise submission repo list bucketed scores iteration: %w", err)
	}
	return out, nil
}

// ListForProgress streams the minimum tuple needed by the per-student
// trend math: (profile_id, score_percentage, submitted_dt). Rows come
// back ordered by (profile_id ASC, submitted_dt ASC) so the
// application layer can fold them into per-student slices in one pass.
// Covered by ix_classroom_profile_submitted (migration 021).
//
// score_percentage / submitted_dt IS NOT NULL filters are explicit —
// see the rationale on ListBucketedScores above.
func (r *ExerciseSubmissionRepository) ListForProgress(ctx context.Context, params domain.ProgressRangeParams) ([]*domain.ProgressRow, error) {
	if params.ClassroomID == 0 {
		return nil, nil
	}

	var b strings.Builder
	b.WriteString(`SELECT s.profile_id, s.classroom_exercise_id, s.score_percentage, s.submitted_dt
		FROM `)
	b.WriteString(exerciseSubmissionTable)
	b.WriteString(` s WHERE `)
	b.WriteString(exerciseSubmissionActiveWhere)
	b.WriteString(` AND s.classroom_id = ?
		AND s.submitted_dt IS NOT NULL
		AND s.score_percentage IS NOT NULL
		AND s.submitted_dt BETWEEN ? AND ?`)

	args := make([]any, 0, len(exerciseSubmissionActiveArgs())+4)
	args = append(args, exerciseSubmissionActiveArgs()...)
	args = append(args, params.ClassroomID, params.From.Time, params.To.Time)

	if params.FilterProfileID != nil {
		b.WriteString(` AND s.profile_id = ?`)
		args = append(args, *params.FilterProfileID)
	}

	b.WriteString(` ORDER BY s.profile_id ASC, s.submitted_dt ASC`)

	rows, err := r.db.Query(ctx, b.String(), args...)
	if err != nil {
		return nil, fmt.Errorf("classroom exercise submission repo list for progress: %w", err)
	}
	defer rows.Close()

	out := make([]*domain.ProgressRow, 0)
	for rows.Next() {
		var (
			row         domain.ProgressRow
			submittedDt time.Time
		)
		if err := rows.Scan(&row.ProfileID, &row.ClassroomExerciseID, &row.ScorePercentage, &submittedDt); err != nil {
			return nil, fmt.Errorf("classroom exercise submission repo list for progress scan: %w", err)
		}
		row.SubmittedDt = mtime.MathTime{Time: submittedDt}
		out = append(out, &row)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("classroom exercise submission repo list for progress iteration: %w", err)
	}
	return out, nil
}

func modelToDomainClassroomExerciseSubmission(m *models.ExerciseSubmissionModel) *domain.Submission {
	s := domain.NewSubmission()
	s.SetId(m.Id)
	s.SetClassroomExerciseSubmissionId(m.ClassroomExerciseSubmissionId)
	s.SetClassroomExerciseId(m.ClassroomExerciseId)
	s.SetClassroomId(m.ClassroomId)
	s.SetProfileId(m.ProfileId)
	s.SetAnswers(m.Answers)
	s.SetAIReview(m.AIReview)
	s.SetTotalQuestions(m.TotalQuestions)
	s.SetCorrectNumber(m.CorrectNumber)
	s.SetScorePercentage(m.ScorePercentage)
	if m.SubmittedDt != nil {
		s.SetSubmittedDt(mtime.MathTime{Time: *m.SubmittedDt})
	}
	if m.GradedDt != nil {
		s.SetGradedDt(mtime.MathTime{Time: *m.GradedDt})
	}
	s.SetNote(m.Note)
	s.SetSubmissionStatus(m.SubmissionStatus)
	s.SetStatus(m.Status)
	s.SetCreateId(m.CreateId)
	s.SetCreateDt(mtime.MathTime{Time: m.CreateDt})
	s.SetModifyId(m.ModifyId)
	s.SetModifyDt(mtime.MathTime{Time: m.ModifyDt})
	return s
}
