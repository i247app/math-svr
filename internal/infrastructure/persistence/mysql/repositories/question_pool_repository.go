package repositories

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"

	"math-ai.com/math-ai/internal/domain/questionpool"
	"math-ai.com/math-ai/internal/domain/shared/mtime"
	"math-ai.com/math-ai/internal/infrastructure/database"
	"math-ai.com/math-ai/internal/infrastructure/persistence/mysql/models"
	"math-ai.com/math-ai/internal/shared/enum"
	"math-ai.com/math-ai/internal/shared/pagination"
)

const (
	questionPoolTable = "ma_question_pools"

	questionPoolColumns = `q.question_pool_id, q.question_kind, q.question_text, q.question_image_key,
		q.options, q.correct_answer, q.explanation, q.question_grade, q.question_topic,
		q.question_level, q.question_sub_level, q.question_tags, q.question_source, q.verified_count, q.question_hash,
		q.question_status, q.rpt_flg, q.kwords, q.note, q.status,
		q.create_id, q.create_dt, q.modify_id, q.modify_dt`

	questionPoolActiveWhere = `q.status IN (?) AND q.deleted_dt IS NULL`

	// questionInsertRow is one VALUES tuple of CreateMany, in the column
	// order of its INSERT list.
	questionPoolInsertRow = `(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
)

func questionPoolActiveArgs() []any {
	return []any{enum.StatusActive}
}

// optionJSON / correctAnswerJSON are the stored shapes of the two JSON
// columns (migration 036 header).
type optionJSON struct {
	Id       string  `json:"id"`
	Text     string  `json:"text"`
	ImageKey *string `json:"image_key,omitempty"`
}

type correctAnswerJSON struct {
	OptionIds []string `json:"option_ids,omitempty"`
	Value     *bool    `json:"value,omitempty"`
	Values    []string `json:"values,omitempty"`
}

type QuestionPoolRepository struct {
	db database.Executor
}

func NewQuestionPoolRepository(db database.Executor) questionpool.IRepository {
	return &QuestionPoolRepository{db: db}
}

func scanQuestionPool(s database.RowScanner) (*models.QuestionPoolModel, error) {
	var m models.QuestionPoolModel
	if err := s.Scan(&m.QuestionPoolId, &m.QuestionKind, &m.QuestionText, &m.QuestionImageKey,
		&m.Options, &m.CorrectAnswer, &m.Explanation, &m.QuestionGrade, &m.QuestionTopic,
		&m.QuestionLevel, &m.QuestionSubLevel, &m.QuestionTags, &m.QuestionSource, &m.VerifiedCount, &m.QuestionHash,
		&m.QuestionStatus, &m.RptFlg, &m.Kwords, &m.Note, &m.Status,
		&m.CreateId, &m.CreateDt, &m.ModifyId, &m.ModifyDt); err != nil {
		return nil, err
	}
	return &m, nil
}

// questionPoolInsertChunk caps the rows of one INSERT: 200 rows x 22
// columns stays far below MySQL's 65,535 placeholders and keeps one
// statement's packet small.
const questionPoolInsertChunk = 200

// questionPoolHashChunk caps the hashes of one FindLiveHashes IN list.
const questionPoolHashChunk = 500

// CreateMany is a plain multi-row INSERT on purpose: ON DUPLICATE KEY
// UPDATE would fold a duplicate (uk_live_question_hash) or a colliding
// minted id into an unrelated row. Any collision fails the statement, so
// the caller's transaction rolls the whole batch back. Rows go out
// questionPoolInsertChunk per statement, all in the caller's transaction.
func (r *QuestionPoolRepository) CreateMany(ctx context.Context, questions []*questionpool.QuestionPool, createId *int64) error {
	for chunk := range slices.Chunk(questions, questionPoolInsertChunk) {
		if err := r.insertChunk(ctx, chunk, createId); err != nil {
			return err
		}
	}
	return nil
}

func (r *QuestionPoolRepository) insertChunk(ctx context.Context, questions []*questionpool.QuestionPool, createId *int64) error {
	now := mtime.Now().Time
	tuples := make([]string, 0, len(questions))
	args := make([]any, 0, len(questions)*22)
	for _, q := range questions {
		options, answer, tags, err := marshalQuestionPoolJSON(q)
		if err != nil {
			return err
		}
		tuples = append(tuples, questionPoolInsertRow)
		args = append(args,
			q.QuestionPoolId(), string(q.QuestionKind()), q.QuestionText(), q.QuestionImageKey(),
			options, answer, q.Explanation(), q.QuestionGrade(), q.QuestionTopic(),
			q.QuestionLevel(), q.QuestionSubLevel(), tags, string(q.QuestionSource()), q.VerifiedCount(), q.QuestionHash(),
			q.QuestionStatus(), q.RptFlg(), q.Kwords(), q.Note(),
			createId, now, now)
	}
	query := `INSERT INTO ` + questionPoolTable + `
			(question_pool_id, question_kind, question_text, question_image_key,
			 options, correct_answer, explanation, question_grade, question_topic,
			 question_level, question_sub_level, question_tags, question_source, verified_count, question_hash,
			 question_status, rpt_flg, kwords, note,
			 create_id, create_dt, modify_dt)
		VALUES ` + strings.Join(tuples, ", ")
	if _, err := r.db.Exec(ctx, query, args...); err != nil {
		if isDuplicateEntry(err) {
			return fmt.Errorf("question pool repo create many: %w", questionpool.ErrDuplicate)
		}
		return fmt.Errorf("question pool repo create many: %w", err)
	}
	return nil
}

func (r *QuestionPoolRepository) LockById(ctx context.Context, id int64) (*questionpool.QuestionPool, error) {
	query := `SELECT ` + questionPoolColumns + ` FROM ` + questionPoolTable + ` q WHERE (q.question_pool_id = ?) AND ` +
		questionPoolActiveWhere + ` FOR UPDATE`
	m, err := scanQuestionPool(r.db.QueryRow(ctx, query, slices.Concat([]any{id}, questionPoolActiveArgs())...))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("question pool repo lock by id: %w", err)
	}
	return ModelToDomainQuestionPool(m)
}

// UpdateContent leaves question_pool_id, source, status, verified_count and
// create_* alone: an edit is not a review.
func (r *QuestionPoolRepository) UpdateContent(ctx context.Context, q *questionpool.QuestionPool, modifyId *int64) error {
	options, answer, tags, err := marshalQuestionPoolJSON(q)
	if err != nil {
		return err
	}
	query := `UPDATE ` + questionPoolTable + ` q SET q.question_kind = ?, q.question_text = ?, q.question_image_key = ?,
			q.options = ?, q.correct_answer = ?, q.explanation = ?, q.question_grade = ?, q.question_topic = ?,
			q.question_level = ?, q.question_sub_level = ?, q.question_tags = ?, q.question_hash = ?, q.modify_id = ?
		WHERE (q.question_pool_id = ?) AND ` + questionPoolActiveWhere
	args := slices.Concat([]any{string(q.QuestionKind()), q.QuestionText(), q.QuestionImageKey(),
		options, answer, q.Explanation(), q.QuestionGrade(), q.QuestionTopic(),
		q.QuestionLevel(), q.QuestionSubLevel(), tags, q.QuestionHash(), modifyId, q.QuestionPoolId()}, questionPoolActiveArgs())
	if _, err := r.db.Exec(ctx, query, args...); err != nil {
		if isDuplicateEntry(err) {
			return fmt.Errorf("question pool repo update content: %w", questionpool.ErrDuplicate)
		}
		return fmt.Errorf("question pool repo update content: %w", err)
	}
	return nil
}

func (r *QuestionPoolRepository) MarkStatus(ctx context.Context, id int64, status string, modifyId *int64) error {
	query := `UPDATE ` + questionPoolTable + ` q SET q.question_status = ?, q.modify_id = ?
		WHERE (q.question_pool_id = ?) AND ` + questionPoolActiveWhere
	if _, err := r.db.Exec(ctx, query, slices.Concat([]any{status, modifyId, id}, questionPoolActiveArgs())...); err != nil {
		return fmt.Errorf("question pool repo mark status: %w", err)
	}
	return nil
}

func (r *QuestionPoolRepository) IncVerified(ctx context.Context, id int64, modifyId *int64) error {
	query := `UPDATE ` + questionPoolTable + ` q SET q.verified_count = q.verified_count + 1, q.modify_id = ?
		WHERE (q.question_pool_id = ?) AND ` + questionPoolActiveWhere
	if _, err := r.db.Exec(ctx, query, slices.Concat([]any{modifyId, id}, questionPoolActiveArgs())...); err != nil {
		return fmt.Errorf("question pool repo inc verified: %w", err)
	}
	return nil
}

// SoftDeleteById leaves question_status (the review state) as it was.
func (r *QuestionPoolRepository) SoftDeleteById(ctx context.Context, id int64, modifyId *int64) (bool, error) {
	query := `UPDATE ` + questionPoolTable + ` q SET q.status = ?, q.deleted_dt = ?, q.modify_id = ?
		WHERE (q.question_pool_id = ?) AND ` + questionPoolActiveWhere
	res, err := r.db.Exec(ctx, query, slices.Concat([]any{enum.StatusInactive, mtime.Now().Time, modifyId, id}, questionPoolActiveArgs())...)
	if err != nil {
		return false, fmt.Errorf("question pool repo soft delete: %w", err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("question pool repo soft delete: %w", err)
	}
	return affected > 0, nil
}

func (r *QuestionPoolRepository) ForceDeleteById(ctx context.Context, id int64) (bool, error) {
	res, err := r.db.Exec(ctx, `DELETE FROM `+questionPoolTable+` WHERE question_pool_id = ?`, id)
	if err != nil {
		return false, fmt.Errorf("question pool repo force delete: %w", err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("question pool repo force delete: %w", err)
	}
	return affected > 0, nil
}

func (r *QuestionPoolRepository) ListByIds(ctx context.Context, ids []int64) ([]*questionpool.QuestionPool, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	args := make([]any, 0, len(ids)+1)
	for _, id := range ids {
		args = append(args, id)
	}
	query := `SELECT ` + questionPoolColumns + ` FROM ` + questionPoolTable + ` q WHERE (q.question_pool_id IN (?` +
		strings.Repeat(", ?", len(ids)-1) + `)) AND ` + questionPoolActiveWhere
	return r.list(ctx, "list by ids", query, slices.Concat(args, questionPoolActiveArgs()))
}

func (r *QuestionPoolRepository) FindLiveHashes(ctx context.Context, hashes []string) ([]string, error) {
	var out []string
	for chunk := range slices.Chunk(hashes, questionPoolHashChunk) {
		found, err := r.findLiveHashes(ctx, chunk)
		if err != nil {
			return nil, err
		}
		out = append(out, found...)
	}
	return out, nil
}

// findLiveHashes looks up one chunk (at most questionPoolHashChunk hashes).
func (r *QuestionPoolRepository) findLiveHashes(ctx context.Context, hashes []string) ([]string, error) {
	args := make([]any, 0, len(hashes)+1)
	for _, h := range hashes {
		args = append(args, h)
	}
	query := `SELECT q.question_hash FROM ` + questionPoolTable + ` q WHERE (q.question_hash IN (?` +
		strings.Repeat(", ?", len(hashes)-1) + `)) AND ` + questionPoolActiveWhere
	rows, err := r.db.Query(ctx, query, slices.Concat(args, questionPoolActiveArgs())...)
	if err != nil {
		return nil, fmt.Errorf("question pool repo find live hashes: %w", err)
	}
	defer rows.Close()

	var out []string
	for rows.Next() {
		var h string
		if err := rows.Scan(&h); err != nil {
			return nil, fmt.Errorf("question pool repo scan hash: %w", err)
		}
		out = append(out, h)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("question pool repo rows iteration: %w", err)
	}
	return out, nil
}

func (r *QuestionPoolRepository) ListPage(ctx context.Context, filter questionpool.ListQuestionPoolsFilter, page, size int64) ([]*questionpool.QuestionPool, *pagination.Pagination, error) {
	size = pagination.ClampSize(size)
	if page < 1 {
		page = 1
	}
	filterWhere, filterArgs, err := buildQuestionPoolFilterClause(filter)
	if err != nil {
		return nil, nil, err
	}
	where := whereActive(filterWhere, questionPoolActiveWhere)
	whereArgs := slices.Concat(filterArgs, questionPoolActiveArgs())

	var total int64
	if err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM `+questionPoolTable+` q`+where, whereArgs...).Scan(&total); err != nil {
		return nil, nil, fmt.Errorf("question pool repo count: %w", err)
	}

	// LIMIT / OFFSET bind after the active filter's args.
	query := `SELECT ` + questionPoolColumns + ` FROM ` + questionPoolTable + ` q` + where +
		` ORDER BY q.create_dt DESC, q.question_pool_id DESC LIMIT ? OFFSET ?`
	out, err := r.list(ctx, "list page", query, slices.Concat(whereArgs, []any{size, (page - 1) * size}))
	if err != nil {
		return nil, nil, err
	}
	return out, pagination.NewPagination(page, size, total), nil
}

// PickForExam draws random approved candidates for an exam paper.
//
// ponytail: ORDER BY RAND() sorts every candidate of the grade/level; fine
// for a bank of thousands. Past ~100k rows, pick random ids from the
// ix_status_grade_created range instead.
func (r *QuestionPoolRepository) PickForExam(ctx context.Context, filter questionpool.PickFilter, limit int) ([]*questionpool.QuestionPool, error) {
	if limit <= 0 {
		return nil, nil
	}
	clause, args := buildQuestionPoolPickClause(filter)
	where := whereActive(clause, questionPoolActiveWhere)
	query := `SELECT ` + questionPoolColumns + ` FROM ` + questionPoolTable + ` q` + where +
		` ORDER BY RAND() LIMIT ?`
	// LIMIT binds after the active filter's args.
	return r.list(ctx, "pick for exam", query, slices.Concat(args, questionPoolActiveArgs(), []any{limit}))
}

// buildQuestionPoolPickClause keeps placeholder order in lockstep with the
// returned args slice.
func buildQuestionPoolPickClause(filter questionpool.PickFilter) (string, []any) {
	clause := ` AND q.question_status = ? AND q.question_grade = ?`
	args := []any{questionpool.StatusApproved, filter.Grade}
	if filter.Level != nil {
		clause += ` AND q.question_level = ?`
		args = append(args, *filter.Level)
	}
	if n := len(filter.ExcludeIds); n > 0 {
		clause += ` AND q.question_pool_id NOT IN (?` + strings.Repeat(", ?", n-1) + `)`
		for _, id := range filter.ExcludeIds {
			args = append(args, id)
		}
	}
	if n := len(filter.AvoidTexts); n > 0 {
		clause += ` AND q.question_text NOT IN (?` + strings.Repeat(", ?", n-1) + `)`
		for _, t := range filter.AvoidTexts {
			args = append(args, t)
		}
	}
	return clause, args
}

// buildQuestionFilterClause keeps placeholder order in lockstep with the
// returned args slice.
func buildQuestionPoolFilterClause(filter questionpool.ListQuestionPoolsFilter) (string, []any, error) {
	var (
		clause string
		args   []any
	)
	if filter.Grade != nil {
		clause += ` AND q.question_grade = ?`
		args = append(args, *filter.Grade)
	}
	if filter.Level != nil {
		clause += ` AND q.question_level = ?`
		args = append(args, *filter.Level)
	}
	if filter.SubLevel != nil {
		clause += ` AND q.question_sub_level = ?`
		args = append(args, *filter.SubLevel)
	}
	if filter.Topic != nil {
		clause += ` AND q.question_topic = ?`
		args = append(args, *filter.Topic)
	}
	if len(filter.Tags) > 0 {
		raw, err := json.Marshal(filter.Tags)
		if err != nil {
			return "", nil, fmt.Errorf("question pool repo marshal tag filter: %w", err)
		}
		clause += ` AND JSON_OVERLAPS(q.question_tags, CAST(? AS JSON))`
		args = append(args, string(raw))
	}
	if n := len(filter.Kinds); n > 0 {
		clause += ` AND q.question_kind IN (?` + strings.Repeat(", ?", n-1) + `)`
		for _, k := range filter.Kinds {
			args = append(args, string(k))
		}
	}
	if n := len(filter.Statuses); n > 0 {
		clause += ` AND q.question_status IN (?` + strings.Repeat(", ?", n-1) + `)`
		for _, s := range filter.Statuses {
			args = append(args, s)
		}
	}
	if filter.Source != nil {
		clause += ` AND q.question_source = ?`
		args = append(args, string(*filter.Source))
	}
	if filter.IsVerified != nil {
		if *filter.IsVerified {
			clause += ` AND q.verified_count > 0`
		} else {
			clause += ` AND q.verified_count = 0`
		}
	}
	return clause, args, nil
}

func (r *QuestionPoolRepository) list(ctx context.Context, op, query string, args []any) ([]*questionpool.QuestionPool, error) {
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("question pool repo %s: %w", op, err)
	}
	defer rows.Close()

	var out []*questionpool.QuestionPool
	for rows.Next() {
		m, err := scanQuestionPool(rows)
		if err != nil {
			return nil, fmt.Errorf("question pool repo scan row: %w", err)
		}
		q, err := ModelToDomainQuestionPool(m)
		if err != nil {
			return nil, err
		}
		out = append(out, q)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("question pool repo rows iteration: %w", err)
	}
	return out, nil
}

// marshalQuestionJSON renders the three JSON columns. options is NULL for
// a kind with no options. question_tags is never NULL and never omitted:
// the column has no DEFAULT (migration 036, error 3903 once
// ix_question_tags exists), so no tags is '[]'.
func marshalQuestionPoolJSON(q *questionpool.QuestionPool) (*string, string, string, error) {
	var options *string
	if q.Options() != nil {
		out := make([]optionJSON, 0, len(q.Options()))
		for _, o := range q.Options() {
			out = append(out, optionJSON{Id: o.Id, Text: o.Text, ImageKey: o.ImageKey})
		}
		raw, err := json.Marshal(out)
		if err != nil {
			return nil, "", "", fmt.Errorf("question pool repo marshal options: %w", err)
		}
		s := string(raw)
		options = &s
	}
	a := q.CorrectAnswer()
	rawAnswer, err := json.Marshal(correctAnswerJSON{OptionIds: a.OptionIds, Value: a.Value, Values: a.Values})
	if err != nil {
		return nil, "", "", fmt.Errorf("question pool repo marshal correct_answer: %w", err)
	}
	tags := q.QuestionTags()
	if tags == nil {
		tags = []string{}
	}
	rawTags, err := json.Marshal(tags)
	if err != nil {
		return nil, "", "", fmt.Errorf("question pool repo marshal question_tags: %w", err)
	}
	return options, string(rawAnswer), string(rawTags), nil
}

func ModelToDomainQuestionPool(m *models.QuestionPoolModel) (*questionpool.QuestionPool, error) {
	var options []questionpool.Option
	if m.Options != nil {
		var stored []optionJSON
		if err := json.Unmarshal([]byte(*m.Options), &stored); err != nil {
			return nil, fmt.Errorf("question pool repo decode options of %d: %w", m.QuestionPoolId, err)
		}
		options = make([]questionpool.Option, 0, len(stored))
		for _, o := range stored {
			options = append(options, questionpool.Option{Id: o.Id, Text: o.Text, ImageKey: o.ImageKey})
		}
	}
	var answer correctAnswerJSON
	if err := json.Unmarshal([]byte(m.CorrectAnswer), &answer); err != nil {
		return nil, fmt.Errorf("question pool repo decode correct_answer of %d: %w", m.QuestionPoolId, err)
	}
	var tags []string
	if err := json.Unmarshal([]byte(m.QuestionTags), &tags); err != nil {
		return nil, fmt.Errorf("question pool repo decode question_tags of %d: %w", m.QuestionPoolId, err)
	}

	q := questionpool.NewQuestionPool()
	q.SetQuestionPoolId(m.QuestionPoolId)
	q.SetQuestionKind(questionpool.Kind(m.QuestionKind))
	q.SetQuestionText(m.QuestionText)
	q.SetQuestionImageKey(m.QuestionImageKey)
	q.SetOptions(options)
	q.SetCorrectAnswer(questionpool.CorrectAnswer{OptionIds: answer.OptionIds, Value: answer.Value, Values: answer.Values})
	q.SetExplanation(m.Explanation)
	q.SetQuestionGrade(m.QuestionGrade)
	q.SetQuestionTopic(m.QuestionTopic)
	q.SetQuestionLevel(m.QuestionLevel)
	q.SetQuestionSubLevel(m.QuestionSubLevel)
	q.SetQuestionTags(tags)
	q.SetQuestionSource(questionpool.Source(m.QuestionSource))
	q.SetVerifiedCount(m.VerifiedCount)
	q.SetQuestionHash(m.QuestionHash)
	q.SetQuestionStatus(m.QuestionStatus)
	q.SetRptFlg(m.RptFlg)
	q.SetKwords(m.Kwords)
	q.SetNote(m.Note)
	q.SetStatus(m.Status)
	q.SetCreateId(m.CreateId)
	q.SetCreateDt(mtime.MathTime{Time: m.CreateDt})
	q.SetModifyId(m.ModifyId)
	q.SetModifyDt(mtime.MathTime{Time: m.ModifyDt})
	return q, nil
}
