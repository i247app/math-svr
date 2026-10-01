package repositories

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"math-ai.com/math-ai/internal/domain/exam"
	"math-ai.com/math-ai/internal/domain/shared/mtime"
	"math-ai.com/math-ai/internal/infrastructure/database"
	"math-ai.com/math-ai/internal/infrastructure/persistence/mysql/models"
	"math-ai.com/math-ai/internal/shared/enum"
)

const (
	examSessionLineTable = "ma_exam_session_lines"

	examSessionLineColumns = `d.esess_ln_id, d.elink_id, d.esess_id, d.exam_id, d.req_exam_type,
		d.question_number, d.question_type, d.question_name, d.question_topic, d.question_grade, d.question_level,
		d.right_answer_label, d.right_answer_content,
		d.selected_label, d.selected_content, d.is_correct,
		d.rpt_flg, d.kwords, d.note, d.esess_ln_status, d.status,
		d.create_id, d.create_dt, d.modify_id, d.modify_dt`

	examSessionLineActiveWhere = `d.status IN (?) AND d.deleted_dt IS NULL`

	// examSessionLineInsertColumns and its placeholder tuple are kept next
	// to each other: CreateBatch repeats the tuple once per row, so the
	// two must be edited together or the batch insert misaligns.
	examSessionLineInsertColumns = `(esess_ln_id, elink_id, esess_id, exam_id, req_exam_type,
			question_number, question_type, question_name, question_topic, question_grade, question_level,
			right_answer_label, right_answer_content,
			selected_label, selected_content, is_correct,
			rpt_flg, kwords, note, esess_ln_status, create_id, create_dt, modify_dt)`

	examSessionLineValueTuple = `(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
)

func examSessionLineActiveArgs() []any {
	return []any{enum.StatusActive}
}

type ExamSessionLineRepository struct {
	db database.Executor
}

func NewExamSessionLineRepository(db database.Executor) exam.IExamSessionLineRepository {
	return &ExamSessionLineRepository{db: db}
}

func scanExamSessionLine(s database.RowScanner) (*models.ExamSessionLineModel, error) {
	var m models.ExamSessionLineModel
	if err := s.Scan(&m.EsessLnId, &m.ElinkId, &m.EsessId, &m.ExamId, &m.ReqExamType,
		&m.QuestionNumber, &m.QuestionType, &m.QuestionName, &m.QuestionTopic, &m.QuestionGrade, &m.QuestionLevel,
		&m.RightAnswerLabel, &m.RightAnswerContent,
		&m.SelectedLabel, &m.SelectedContent, &m.IsCorrect,
		&m.RptFlg, &m.Kwords, &m.Note, &m.EsessLnStatus, &m.Status,
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
func (r *ExamSessionLineRepository) CreateBatch(ctx context.Context, details []*exam.ExamSessionLine) error {
	if len(details) == 0 {
		return nil
	}

	now := mtime.Now().Time
	placeholders := make([]string, 0, len(details))
	args := make([]any, 0, len(details)*21)

	for _, d := range details {
		esessLnStatus := d.EsessLnStatus()
		if esessLnStatus == nil {
			s := string(enum.EsessLnStatusActive)
			esessLnStatus = &s
		}
		placeholders = append(placeholders, examSessionLineValueTuple)
		args = append(args,
			d.EsessLnId(), d.ElinkId(), d.EsessId(), d.ExamId(), d.ReqExamType(),
			d.QuestionNumber(), d.QuestionType(), d.QuestionName(), d.QuestionTopic(), d.QuestionGrade(), d.QuestionLevel(),
			d.RightAnswerLabel(), d.RightAnswerContent(),
			d.SelectedLabel(), d.SelectedContent(), d.IsCorrect(),
			d.RptFlg(), d.Kwords(), d.Note(), esessLnStatus, d.CreateId(), now, now)
	}

	query := `INSERT INTO ` + examSessionLineTable + ` ` + examSessionLineInsertColumns +
		` VALUES ` + strings.Join(placeholders, ", ")

	if _, err := r.db.Exec(ctx, query, args...); err != nil {
		return fmt.Errorf("user exam detail repo create batch: %w", err)
	}
	return nil
}

// list binds placeholders in the order they appear in the SQL: the WHERE
// args, then the active filter's, then orderArgs for any placeholder in
// orderLimit (a LIMIT ?). orderArgs must stay separate from args — passed
// together, the LIMIT value lands in the active filter's slot.
func (r *ExamSessionLineRepository) list(ctx context.Context, where string, args []any, orderLimit string, orderArgs ...any) ([]*exam.ExamSessionLine, error) {
	fullArgs := slices.Concat(args, examSessionLineActiveArgs(), orderArgs)
	query := `SELECT ` + examSessionLineColumns + ` FROM ` + examSessionLineTable + ` d WHERE (` +
		where + `) AND ` + examSessionLineActiveWhere + ` ` + orderLimit

	rows, err := r.db.Query(ctx, query, fullArgs...)
	if err != nil {
		return nil, fmt.Errorf("user exam detail repo list (%s): %w", where, err)
	}
	defer rows.Close()

	var details []*exam.ExamSessionLine
	for rows.Next() {
		m, err := scanExamSessionLine(rows)
		if err != nil {
			return nil, fmt.Errorf("user exam detail repo scan row: %w", err)
		}
		details = append(details, ModelToDomainExamSessionLine(m))
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("user exam detail repo rows iteration: %w", err)
	}
	return details, nil
}

// ListByElinkId returns one sitting in question order — the review
// screen's read.
func (r *ExamSessionLineRepository) ListByElinkId(ctx context.Context, elinkId int64) ([]*exam.ExamSessionLine, error) {
	return r.list(ctx, "d.elink_id = ?", []any{elinkId}, "ORDER BY d.question_number ASC")
}

// ListByEsessId returns one row of a journey's log — its ASSESSMENT
// sittings or its PRACTICE sittings, never both. Ordered by sitting id
// first — ids are minted monotonically, so that IS chronological — then
// by question number, so a client can walk it exam by exam.
func (r *ExamSessionLineRepository) ListByEsessId(ctx context.Context, esessId int64, examType string) ([]*exam.ExamSessionLine, error) {
	return r.list(ctx, "d.esess_id = ? AND d.req_exam_type = ?", []any{esessId, examType},
		"ORDER BY d.elink_id ASC, d.question_number ASC")
}

// ListRecentByEsessId returns the child's most recently answered
// questions across every sitting, newest first. It is the input a
// placement or weak-topic rule reads, which is why it is bounded: those
// rules look at a recent window, never the whole lifetime.
func (r *ExamSessionLineRepository) ListRecentByEsessId(ctx context.Context, esessId int64, examType string, limit int64) ([]*exam.ExamSessionLine, error) {
	if limit <= 0 {
		limit = 50
	}
	return r.list(ctx, "d.esess_id = ? AND d.req_exam_type = ?", []any{esessId, examType},
		"ORDER BY d.esess_ln_id DESC LIMIT ?", limit)
}

func ModelToDomainExamSessionLine(m *models.ExamSessionLineModel) *exam.ExamSessionLine {
	d := exam.NewExamSessionLine()
	d.SetEsessLnId(m.EsessLnId)
	d.SetElinkId(m.ElinkId)
	d.SetEsessId(m.EsessId)
	d.SetExamId(m.ExamId)
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
	d.SetRptFlg(m.RptFlg)
	d.SetKwords(m.Kwords)
	d.SetNote(m.Note)
	d.SetEsessLnStatus(m.EsessLnStatus)
	d.SetStatus(m.Status)
	d.SetCreateId(m.CreateId)
	d.SetCreateDt(mtime.MathTime{Time: m.CreateDt})
	d.SetModifyId(m.ModifyId)
	d.SetModifyDt(mtime.MathTime{Time: m.ModifyDt})
	return d
}
