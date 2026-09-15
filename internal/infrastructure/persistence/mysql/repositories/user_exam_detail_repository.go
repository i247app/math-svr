package repositories

import (
	"context"
	"fmt"
	"strings"

	"math-ai.com/math-ai/internal/domain/exam"
	"math-ai.com/math-ai/internal/domain/shared/mtime"
	"math-ai.com/math-ai/internal/infrastructure/database"
	"math-ai.com/math-ai/internal/infrastructure/persistence/mysql/models"
	"math-ai.com/math-ai/internal/shared/enum"
)

const (
	userExamDetailTable = "ma_user_exam_details"

	userExamDetailColumns = `d.id, d.user_exam_detail_id, d.user_ai_exam_id, d.user_exam_id, d.ai_exam_id, d.req_exam_type,
		d.question_number, d.question_type, d.question_name, d.question_topic, d.question_grade, d.question_level,
		d.right_answer_label, d.right_answer_content,
		d.selected_label, d.selected_content, d.is_correct,
		d.note, d.detail_status, d.status,
		d.create_id, d.create_dt, d.modify_id, d.modify_dt`

	userExamDetailActiveWhere = `d.status = ? AND (d.detail_status IS NULL OR d.detail_status != ?) AND d.deleted_dt IS NULL`

	// userExamDetailInsertColumns and its placeholder tuple are kept next
	// to each other: CreateBatch repeats the tuple once per row, so the
	// two must be edited together or the batch insert misaligns.
	userExamDetailInsertColumns = `(user_exam_detail_id, user_ai_exam_id, user_exam_id, ai_exam_id, req_exam_type,
			question_number, question_type, question_name, question_topic, question_grade, question_level,
			right_answer_label, right_answer_content,
			selected_label, selected_content, is_correct,
			note, detail_status, create_id, create_dt, modify_dt)`

	userExamDetailValueTuple = `(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
)

func userExamDetailActiveArgs() []any {
	return []any{enum.StatusActive, string(enum.UserExamDetailStatusDeleted)}
}

type UserExamDetailRepository struct {
	db database.Executor
}

func NewUserExamDetailRepository(db database.Executor) exam.IUserExamDetailRepository {
	return &UserExamDetailRepository{db: db}
}

func scanUserExamDetail(s database.RowScanner) (*models.UserExamDetailModel, error) {
	var m models.UserExamDetailModel
	if err := s.Scan(&m.Id, &m.UserExamDetailId, &m.UserAiExamId, &m.UserExamId, &m.AiExamId, &m.ReqExamType,
		&m.QuestionNumber, &m.QuestionType, &m.QuestionName, &m.QuestionTopic, &m.QuestionGrade, &m.QuestionLevel,
		&m.RightAnswerLabel, &m.RightAnswerContent,
		&m.SelectedLabel, &m.SelectedContent, &m.IsCorrect,
		&m.Note, &m.DetailStatus, &m.Status,
		&m.CreateId, &m.CreateDt, &m.ModifyId, &m.ModifyDt); err != nil {
		return nil, err
	}
	return &m, nil
}

// CreateBatch writes a whole sitting in one statement. Every row belongs to
// the same transaction anyway, so N round-trips would buy nothing; the
// batch also means a partial write cannot leave an attempt half-logged.
//
// An empty slice is a no-op rather than an error: a submission with no
// answers is caught by the validator, and a repository is the wrong place
// to re-litigate that.
func (r *UserExamDetailRepository) CreateBatch(ctx context.Context, details []*exam.UserExamDetail) error {
	if len(details) == 0 {
		return nil
	}

	now := mtime.Now().Time
	placeholders := make([]string, 0, len(details))
	args := make([]any, 0, len(details)*21)

	for _, d := range details {
		detailStatus := d.DetailStatus()
		if detailStatus == nil {
			s := string(enum.UserExamDetailStatusActive)
			detailStatus = &s
		}
		placeholders = append(placeholders, userExamDetailValueTuple)
		args = append(args,
			d.UserExamDetailId(), d.UserAiExamId(), d.UserExamId(), d.AiExamId(), d.ReqExamType(),
			d.QuestionNumber(), d.QuestionType(), d.QuestionName(), d.QuestionTopic(), d.QuestionGrade(), d.QuestionLevel(),
			d.RightAnswerLabel(), d.RightAnswerContent(),
			d.SelectedLabel(), d.SelectedContent(), d.IsCorrect(),
			d.Note(), detailStatus, d.CreateId(), now, now)
	}

	query := `INSERT INTO ` + userExamDetailTable + ` ` + userExamDetailInsertColumns +
		` VALUES ` + strings.Join(placeholders, ", ")

	if _, err := r.db.Exec(ctx, query, args...); err != nil {
		return fmt.Errorf("user exam detail repo create batch: %w", err)
	}
	return nil
}

func (r *UserExamDetailRepository) list(ctx context.Context, where string, args []any, orderLimit string) ([]*exam.UserExamDetail, error) {
	fullArgs := append(userExamDetailActiveArgs(), args...)
	query := `SELECT ` + userExamDetailColumns + ` FROM ` + userExamDetailTable + ` d WHERE ` +
		userExamDetailActiveWhere + ` AND (` + where + `) ` + orderLimit

	rows, err := r.db.Query(ctx, query, fullArgs...)
	if err != nil {
		return nil, fmt.Errorf("user exam detail repo list (%s): %w", where, err)
	}
	defer rows.Close()

	var details []*exam.UserExamDetail
	for rows.Next() {
		m, err := scanUserExamDetail(rows)
		if err != nil {
			return nil, fmt.Errorf("user exam detail repo scan row: %w", err)
		}
		details = append(details, ModelToDomainUserExamDetail(m))
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("user exam detail repo rows iteration: %w", err)
	}
	return details, nil
}

// ListByUserAiExamId returns one sitting in question order — the review
// screen's read.
func (r *UserExamDetailRepository) ListByUserAiExamId(ctx context.Context, userAiExamId int64) ([]*exam.UserExamDetail, error) {
	return r.list(ctx, "d.user_ai_exam_id = ?", []any{userAiExamId}, "ORDER BY d.question_number ASC")
}

// ListByUserExamId returns one row of a journey's log — its ASSESSMENT
// sittings or its PRACTICE sittings, never both. Ordered by sitting id
// first — ids are minted monotonically, so that IS chronological — then
// by question number, so a client can walk it exam by exam.
func (r *UserExamDetailRepository) ListByUserExamId(ctx context.Context, userExamId int64, examType string) ([]*exam.UserExamDetail, error) {
	return r.list(ctx, "d.user_exam_id = ? AND d.req_exam_type = ?", []any{userExamId, examType},
		"ORDER BY d.user_ai_exam_id ASC, d.question_number ASC")
}

// ListRecentByUserExamId returns the child's most recently answered
// questions across every sitting, newest first. It is the input a
// placement or weak-topic rule reads, which is why it is bounded: those
// rules look at a recent window, never the whole lifetime.
func (r *UserExamDetailRepository) ListRecentByUserExamId(ctx context.Context, userExamId int64, examType string, limit int64) ([]*exam.UserExamDetail, error) {
	if limit <= 0 {
		limit = 50
	}
	return r.list(ctx, "d.user_exam_id = ? AND d.req_exam_type = ?", []any{userExamId, examType, limit},
		"ORDER BY d.id DESC LIMIT ?")
}

func ModelToDomainUserExamDetail(m *models.UserExamDetailModel) *exam.UserExamDetail {
	d := exam.NewUserExamDetail()
	d.SetId(m.Id)
	d.SetUserExamDetailId(m.UserExamDetailId)
	d.SetUserAiExamId(m.UserAiExamId)
	d.SetUserExamId(m.UserExamId)
	d.SetAiExamId(m.AiExamId)
	d.SetReqExamType(m.ReqExamType)
	d.SetQuestionNumber(m.QuestionNumber)
	d.SetQuestionType(m.QuestionType)
	d.SetQuestionName(m.QuestionName)
	d.SetQuestionTopic(m.QuestionTopic)
	d.SetQuestionGrade(m.QuestionGrade)
	d.SetQuestionLevel(m.QuestionLevel)
	d.SetRightAnswerLabel(m.RightAnswerLabel)
	d.SetRightAnswerContent(m.RightAnswerContent)
	d.SetSelectedLabel(m.SelectedLabel)
	d.SetSelectedContent(m.SelectedContent)
	d.SetIsCorrect(m.IsCorrect)
	d.SetNote(m.Note)
	d.SetDetailStatus(m.DetailStatus)
	d.SetStatus(m.Status)
	d.SetCreateId(m.CreateId)
	d.SetCreateDt(mtime.MathTime{Time: m.CreateDt})
	d.SetModifyId(m.ModifyId)
	d.SetModifyDt(mtime.MathTime{Time: m.ModifyDt})
	return d
}
