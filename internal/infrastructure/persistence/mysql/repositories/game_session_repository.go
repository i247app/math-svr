package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"slices"
	"time"

	"math-ai.com/math-ai/internal/domain/game"
	"math-ai.com/math-ai/internal/infrastructure/database"
	"math-ai.com/math-ai/internal/infrastructure/persistence/mysql/models"
	"math-ai.com/math-ai/internal/shared/enum"
	"math-ai.com/math-ai/internal/shared/pagination"
)

const (
	gameSessionTable = "ma_game_sessions"

	gameSessionActiveWhere = `g.status IN (?) AND g.deleted_dt IS NULL`

	// gameSessionListWhere is the WHERE of a child's played rounds, shared by
	// the page, keyset and existence reads so all three see the same rows.
	gameSessionListWhere = `g.uid = ? AND g.profile_id = ? AND g.gsess_status = ?`
	gameSessionListOrder = `ORDER BY g.gsess_id DESC`
)

var (
	gameSessionColumns = `g.gsess_id, g.uid, g.profile_id, g.req_grade, g.req_level, g.req_num_ques, g.questions_json, g.answers_json, g.res_total_questions, g.res_correct_number, g.res_skipped_number, g.res_score_percentage, g.started_dt, g.submitted_dt, g.gsess_status, ` + models.AuditColumns("g")

	// gameSessionListColumns has the same columns in the same order with the
	// two JSON blobs blanked (a round holds up to 50 questions), so
	// scanGameSession reads both.
	gameSessionListColumns = `g.gsess_id, g.uid, g.profile_id, g.req_grade, g.req_level, g.req_num_ques, '' AS questions_json, NULL AS answers_json, g.res_total_questions, g.res_correct_number, g.res_skipped_number, g.res_score_percentage, g.started_dt, g.submitted_dt, g.gsess_status, ` + models.AuditColumns("g")
)

func gameSessionActiveArgs() []any {
	return []any{enum.StatusActive}
}

type GameSessionRepository struct {
	db database.Executor
}

func NewGameSessionRepository(db database.Executor) game.IRepository {
	return &GameSessionRepository{db: db}
}

func scanGameSession(s database.RowScanner) (*models.GameSessionModel, error) {
	var m models.GameSessionModel
	if err := s.Scan(append([]any{&m.GsessId, &m.UID, &m.ProfileId, &m.ReqGrade, &m.ReqLevel, &m.ReqNumQues, &m.QuestionsJson, &m.AnswersJson, &m.ResTotalQuestions, &m.ResCorrectNumber, &m.ResSkippedNumber, &m.ResScorePercentage, &m.StartedDt, &m.SubmittedDt, &m.GsessStatus}, m.Audit.Dest()...)...); err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *GameSessionRepository) findOneBy(ctx context.Context, where string, args ...any) (*game.GameSession, error) {
	query := `SELECT ` + gameSessionColumns + ` FROM ` + gameSessionTable + ` g WHERE (` + where + `) AND ` + gameSessionActiveWhere
	m, err := scanGameSession(r.db.QueryRow(ctx, query, slices.Concat(args, gameSessionActiveArgs())...))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("game session repo find (%s): %w", where, err)
	}
	return ModelToDomainGameSession(m), nil
}

func (r *GameSessionRepository) FindByGsessId(ctx context.Context, gsessId int64) (*game.GameSession, error) {
	return r.findOneBy(ctx, "g.gsess_id = ?", gsessId)
}

func (r *GameSessionRepository) Create(ctx context.Context, s *game.GameSession) (*game.GameSession, error) {
	query := `
		INSERT INTO ` + gameSessionTable + `
			(gsess_id, uid, profile_id, req_grade, req_level, req_num_ques, questions_json,
			 started_dt, rpt_flg, kwords, note, gsess_status, create_id, create_dt, modify_dt)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	now := time.Now().UTC()
	startedDt := s.StartedDt()
	if startedDt.IsZero() {
		startedDt = now
	}
	if _, err := r.db.Exec(ctx, query,
		s.GsessId(), s.UID(), s.ProfileId(), s.ReqGrade(), s.ReqLevel(), s.ReqNumQues(), s.QuestionsJson(),
		startedDt, s.RptFlg(), s.Kwords(), s.Note(), s.GsessStatus(), s.CreateId(), now, now); err != nil {
		return nil, fmt.Errorf("game session repo create: %w", err)
	}
	return r.FindByGsessId(ctx, s.GsessId())
}

// MarkComplete carries the expected state in its WHERE clause: of two
// submits racing on one round, the loser matches zero rows.
func (r *GameSessionRepository) MarkComplete(ctx context.Context, gsessId int64, res game.Result) error {
	query := `
		UPDATE ` + gameSessionTable + ` g
		SET g.answers_json = ?, g.res_total_questions = ?, g.res_correct_number = ?,
			g.res_skipped_number = ?, g.res_score_percentage = ?, g.submitted_dt = ?,
			g.gsess_status = ?, g.modify_id = COALESCE(?, g.modify_id), g.modify_dt = ?
		WHERE (g.gsess_id = ? AND g.gsess_status = ?) AND ` + gameSessionActiveWhere
	now := time.Now().UTC()
	submittedDt := res.SubmittedDt
	if submittedDt.IsZero() {
		submittedDt = now
	}
	args := slices.Concat([]any{
		res.AnswersJson, res.TotalQuestions, res.CorrectNumber,
		res.SkippedNumber, res.ScorePercentage, submittedDt,
		enum.GsessStatusComplete.String(), res.ModifyId, now,
		gsessId, enum.GsessStatusInProgress.String(),
	}, gameSessionActiveArgs())

	result, err := r.db.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("game session repo mark complete: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("game session repo mark complete rows affected: %w", err)
	}
	if affected == 0 {
		return game.ErrSessionNotInProgress
	}
	return nil
}

func gameSessionListArgs(uid, profileId int64) []any {
	return []any{uid, profileId, enum.GsessStatusComplete.String()}
}

func (r *GameSessionRepository) ListCompletePage(ctx context.Context, uid, profileId int64, page, size int64) ([]*game.GameSession, *pagination.Pagination, error) {
	args := gameSessionListArgs(uid, profileId)

	var total int64
	countQuery := `SELECT COUNT(*) FROM ` + gameSessionTable + ` g WHERE (` + gameSessionListWhere + `) AND ` + gameSessionActiveWhere
	if err := r.db.QueryRow(ctx, countQuery, slices.Concat(args, gameSessionActiveArgs())...).Scan(&total); err != nil {
		return nil, nil, fmt.Errorf("game session repo count: %w", err)
	}
	pg := pagination.NewPagination(page, size, total)

	rows, err := r.list(ctx, "list page", gameSessionListWhere, args, gameSessionListOrder+` LIMIT ? OFFSET ?`, pg.Size, pg.Skip)
	if err != nil {
		return nil, nil, err
	}
	return rows, pg, nil
}

// ListCompleteKeyset reads a previous page upward (so LIMIT keeps the
// nearest rows) and reverses it: every caller sees display order.
func (r *GameSessionRepository) ListCompleteKeyset(ctx context.Context, uid, profileId int64, params game.KeysetParams) ([]*game.GameSession, error) {
	where, args, order := gameSessionListWhere, gameSessionListArgs(uid, profileId), gameSessionListOrder
	switch {
	case params.Older != nil:
		where += ` AND g.gsess_id < ?`
		args = append(args, *params.Older)
	case params.Newer != nil:
		where += ` AND g.gsess_id > ?`
		args = append(args, *params.Newer)
		order = `ORDER BY g.gsess_id ASC`
	}
	rows, err := r.list(ctx, "list keyset", where, args, order+` LIMIT ?`, params.Limit)
	if err != nil {
		return nil, err
	}
	if params.Older == nil && params.Newer != nil {
		slices.Reverse(rows)
	}
	return rows, nil
}

func (r *GameSessionRepository) ExistsCompleteAtOrNewer(ctx context.Context, uid, profileId, gsessId int64) (bool, error) {
	return r.exists(ctx, uid, profileId, `g.gsess_id >= ?`, gsessId)
}

func (r *GameSessionRepository) ExistsCompleteAtOrOlder(ctx context.Context, uid, profileId, gsessId int64) (bool, error) {
	return r.exists(ctx, uid, profileId, `g.gsess_id <= ?`, gsessId)
}

func (r *GameSessionRepository) exists(ctx context.Context, uid, profileId int64, keyCond string, gsessId int64) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM ` + gameSessionTable + ` g WHERE (` + gameSessionListWhere + ` AND ` + keyCond + `) AND ` + gameSessionActiveWhere + `)`
	args := slices.Concat(gameSessionListArgs(uid, profileId), []any{gsessId}, gameSessionActiveArgs())
	var exists bool
	if err := r.db.QueryRow(ctx, query, args...).Scan(&exists); err != nil {
		return false, fmt.Errorf("game session repo exists: %w", err)
	}
	return exists, nil
}

// list runs the blob-free SELECT; tailArgs bind the placeholders of tail
// (LIMIT / OFFSET), which follow the active filter's.
func (r *GameSessionRepository) list(ctx context.Context, op, where string, args []any, tail string, tailArgs ...any) ([]*game.GameSession, error) {
	query := `SELECT ` + gameSessionListColumns + ` FROM ` + gameSessionTable + ` g WHERE (` + where + `) AND ` +
		gameSessionActiveWhere + ` ` + tail
	rows, err := r.db.Query(ctx, query, slices.Concat(args, gameSessionActiveArgs(), tailArgs)...)
	if err != nil {
		return nil, fmt.Errorf("game session repo %s: %w", op, err)
	}
	defer rows.Close()

	var out []*game.GameSession
	for rows.Next() {
		m, err := scanGameSession(rows)
		if err != nil {
			return nil, fmt.Errorf("game session repo scan row: %w", err)
		}
		out = append(out, ModelToDomainGameSession(m))
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("game session repo rows iteration: %w", err)
	}
	return out, nil
}

func ModelToDomainGameSession(m *models.GameSessionModel) *game.GameSession {
	s := game.NewGameSession()
	s.SetGsessId(m.GsessId)
	s.SetUID(m.UID)
	s.SetProfileId(m.ProfileId)
	s.SetReqGrade(m.ReqGrade)
	s.SetReqLevel(m.ReqLevel)
	s.SetReqNumQues(m.ReqNumQues)
	s.SetQuestionsJson(m.QuestionsJson)
	s.SetAnswersJson(m.AnswersJson)
	s.SetResTotalQuestions(m.ResTotalQuestions)
	s.SetResCorrectNumber(m.ResCorrectNumber)
	s.SetResSkippedNumber(m.ResSkippedNumber)
	s.SetResScorePercentage(m.ResScorePercentage)
	s.SetStartedDt(m.StartedDt)
	s.SetSubmittedDt(timeOrZero(m.SubmittedDt))
	s.SetGsessStatus(m.GsessStatus)
	m.Audit.ApplyTo(&s.Audit)
	return s
}
