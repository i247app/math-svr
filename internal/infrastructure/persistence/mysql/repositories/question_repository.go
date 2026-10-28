package repositories

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"math-ai.com/math-ai/internal/domain/question"
	"math-ai.com/math-ai/internal/domain/questionpool"
	"math-ai.com/math-ai/internal/domain/shared/mtime"
	"math-ai.com/math-ai/internal/infrastructure/database"
	"math-ai.com/math-ai/internal/infrastructure/persistence/mysql/models"
	"math-ai.com/math-ai/internal/shared/enum"
)

const (
	questionTable = "ma_questions"

	questionColumns = `q.question_id, q.exam_id, q.question_number, q.render_type, q.question_kind, q.question_text,
		q.options, q.correct_answer, q.question_topic, q.question_grade,
		q.question_status, q.rpt_flg, q.kwords, q.note, q.status,
		q.create_id, q.create_dt, q.modify_id, q.modify_dt`

	questionActiveWhere = `q.status IN (?) AND q.deleted_dt IS NULL`

	// questionInsertRow is one VALUES tuple of CreateMany, in the column
	// order of its INSERT list.
	questionInsertRow = `(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
)

func questionActiveArgs() []any {
	return []any{enum.StatusActive}
}

// questionOptionJSON is the stored shape of options (migration 039 header) — the
// same keys as the pool's questions JSON. correct_answer is stored as
// correctAnswerJSON, the bank's shape (question_pool_repository.go).
type questionOptionJSON struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}

type QuestionRepository struct {
	db database.Executor
}

func NewQuestionRepository(db database.Executor) question.IRepository {
	return &QuestionRepository{db: db}
}

func scanQuestion(s database.RowScanner) (*models.QuestionModel, error) {
	var m models.QuestionModel
	if err := s.Scan(&m.QuestionId, &m.ExamId, &m.QuestionNumber, &m.RenderType, &m.QuestionKind, &m.QuestionText,
		&m.Options, &m.CorrectAnswer, &m.QuestionTopic, &m.QuestionGrade,
		&m.QuestionStatus, &m.RptFlg, &m.Kwords, &m.Note, &m.Status,
		&m.CreateId, &m.CreateDt, &m.ModifyId, &m.ModifyDt); err != nil {
		return nil, err
	}
	return &m, nil
}

// CreateMany is a plain multi-row INSERT: a colliding minted id or a live
// (exam_id, question_number) fails the statement, so the caller's
// transaction rolls the whole batch back.
func (r *QuestionRepository) CreateMany(ctx context.Context, questions []*question.Question, createId *int64) error {
	if len(questions) == 0 {
		return nil
	}
	now := mtime.Now().Time
	tuples := make([]string, 0, len(questions))
	args := make([]any, 0, len(questions)*17)
	for _, q := range questions {
		options, err := marshalOptions(q.Options())
		if err != nil {
			return err
		}
		a := q.CorrectAnswer()
		answer, err := json.Marshal(correctAnswerJSON{OptionIds: a.OptionIds, Value: a.Value, Values: a.Values})
		if err != nil {
			return fmt.Errorf("question repo marshal correct_answer: %w", err)
		}
		tuples = append(tuples, questionInsertRow)
		args = append(args,
			q.QuestionId(), q.ExamId(), q.QuestionNumber(), q.RenderType(), q.QuestionKind(), q.QuestionText(),
			options, string(answer), q.QuestionTopic(), q.QuestionGrade(),
			q.QuestionStatus(), q.RptFlg(), q.Kwords(), q.Note(),
			createId, now, now)
	}
	query := `INSERT INTO ` + questionTable + `
			(question_id, exam_id, question_number, render_type, question_kind, question_text,
			 options, correct_answer, question_topic, question_grade,
			 question_status, rpt_flg, kwords, note,
			 create_id, create_dt, modify_dt)
		VALUES ` + strings.Join(tuples, ", ")
	if _, err := r.db.Exec(ctx, query, args...); err != nil {
		return fmt.Errorf("question repo create many: %w", err)
	}
	return nil
}

func (r *QuestionRepository) ListByExamId(ctx context.Context, examId int64) ([]*question.Question, error) {
	query := `SELECT ` + questionColumns + ` FROM ` + questionTable + ` q WHERE (q.exam_id = ?) AND ` +
		questionActiveWhere + ` ORDER BY q.question_number ASC`
	rows, err := r.db.Query(ctx, query, slices.Concat([]any{examId}, questionActiveArgs())...)
	if err != nil {
		return nil, fmt.Errorf("question repo list by exam id: %w", err)
	}
	defer rows.Close()

	var out []*question.Question
	for rows.Next() {
		m, err := scanQuestion(rows)
		if err != nil {
			return nil, fmt.Errorf("question repo scan row: %w", err)
		}
		q, err := ModelToDomainQuestion(m)
		if err != nil {
			return nil, err
		}
		out = append(out, q)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("question repo rows iteration: %w", err)
	}
	return out, nil
}

// ReplaceByExamId soft-deletes before it inserts: uk_live_exam_question
// only counts live rows, so the old version stays behind as history and
// the new rows take the same question_numbers. An exam with no rows yet (a
// pool stored before ma_questions was written) just gets its first ones.
func (r *QuestionRepository) ReplaceByExamId(ctx context.Context, examId int64, questions []*question.Question, actorId *int64) error {
	query := `UPDATE ` + questionTable + ` q SET q.status = ?, q.deleted_dt = ?, q.modify_id = ?
		WHERE (q.exam_id = ?) AND ` + questionActiveWhere
	args := slices.Concat([]any{enum.StatusInactive, mtime.Now().Time, actorId, examId}, questionActiveArgs())
	if _, err := r.db.Exec(ctx, query, args...); err != nil {
		return fmt.Errorf("question repo replace by exam id: %w", err)
	}
	return r.CreateMany(ctx, questions, actorId)
}

// marshalOptions never renders null: options is NOT NULL.
func marshalOptions(in []question.Option) (string, error) {
	out := make([]questionOptionJSON, 0, len(in))
	for _, o := range in {
		out = append(out, questionOptionJSON{ID: o.ID, Text: o.Text})
	}
	raw, err := json.Marshal(out)
	if err != nil {
		return "", fmt.Errorf("question repo marshal options: %w", err)
	}
	return string(raw), nil
}

func ModelToDomainQuestion(m *models.QuestionModel) (*question.Question, error) {
	var stored []questionOptionJSON
	if err := json.Unmarshal([]byte(m.Options), &stored); err != nil {
		return nil, fmt.Errorf("question repo decode options of %d: %w", m.QuestionId, err)
	}
	options := make([]question.Option, 0, len(stored))
	for _, o := range stored {
		options = append(options, question.Option{ID: o.ID, Text: o.Text})
	}
	var answer correctAnswerJSON
	if err := json.Unmarshal([]byte(m.CorrectAnswer), &answer); err != nil {
		return nil, fmt.Errorf("question repo decode correct_answer of %d: %w", m.QuestionId, err)
	}

	q := question.NewQuestion()
	q.SetQuestionId(m.QuestionId)
	q.SetExamId(m.ExamId)
	q.SetQuestionNumber(m.QuestionNumber)
	q.SetRenderType(m.RenderType)
	q.SetQuestionKind(m.QuestionKind)
	q.SetQuestionText(m.QuestionText)
	q.SetOptions(options)
	q.SetCorrectAnswer(questionpool.CorrectAnswer{OptionIds: answer.OptionIds, Value: answer.Value, Values: answer.Values})
	q.SetQuestionTopic(m.QuestionTopic)
	q.SetQuestionGrade(m.QuestionGrade)
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
