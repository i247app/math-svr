package repositories

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"math-ai.com/math-ai/internal/domain/question"
	"math-ai.com/math-ai/internal/domain/shared/mtime"
	"math-ai.com/math-ai/internal/infrastructure/database"
	"math-ai.com/math-ai/internal/infrastructure/persistence/mysql/models"
	"math-ai.com/math-ai/internal/shared/enum"
)

const (
	questionTable = "ma_questions"

	questionColumns = `q.question_id, q.exam_id, q.question_number, q.question_type, q.question_name,
		q.answers, q.question_topic, q.question_grade, q.question_tags, q.verified_count,
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

// answerJSON is the stored shape of one element of ma_questions.answers.
type answerJSON struct {
	Content   string `json:"content"`
	IsCorrect bool   `json:"is_correct"`
}

type QuestionRepository struct {
	db database.Executor
}

func NewQuestionRepository(db database.Executor) question.IRepository {
	return &QuestionRepository{db: db}
}

func scanQuestion(s database.RowScanner) (*models.QuestionModel, error) {
	var m models.QuestionModel
	if err := s.Scan(&m.QuestionId, &m.ExamId, &m.QuestionNumber, &m.QuestionType, &m.QuestionName,
		&m.Answers, &m.QuestionTopic, &m.QuestionGrade, &m.QuestionTags, &m.VerifiedCount,
		&m.QuestionStatus, &m.RptFlg, &m.Kwords, &m.Note, &m.Status,
		&m.CreateId, &m.CreateDt, &m.ModifyId, &m.ModifyDt); err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *QuestionRepository) ListByExamId(ctx context.Context, examId int64) ([]*question.Question, error) {
	query := `SELECT ` + questionColumns + ` FROM ` + questionTable + ` q WHERE (q.exam_id = ?) AND ` +
		questionActiveWhere + ` ORDER BY q.question_number`
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

// CreateMany is a plain multi-row INSERT on purpose: an INSERT … ON
// DUPLICATE KEY UPDATE would redirect a row whose minted question_id
// collides (ma_seqs behind the data) into an unrelated question of another
// exam. The command reads the existing rows first and only inserts numbers
// the bank does not hold, so a collision here is reported, not absorbed.
func (r *QuestionRepository) CreateMany(ctx context.Context, questions []*question.Question, createId *int64) error {
	if len(questions) == 0 {
		return nil
	}
	now := mtime.Now().Time
	tuples := make([]string, 0, len(questions))
	args := make([]any, 0, len(questions)*17)
	for _, q := range questions {
		answers, tags, err := marshalQuestionJSON(q)
		if err != nil {
			return err
		}
		tuples = append(tuples, questionInsertRow)
		args = append(args,
			q.QuestionId(), q.ExamId(), q.QuestionNumber(), q.QuestionType(), q.QuestionName(),
			answers, q.QuestionTopic(), q.QuestionGrade(), tags, q.VerifiedCount(),
			q.QuestionStatus(), q.RptFlg(), q.Kwords(), q.Note(),
			createId, now, now)
	}
	query := `INSERT INTO ` + questionTable + `
			(question_id, exam_id, question_number, question_type, question_name,
			 answers, question_topic, question_grade, question_tags, verified_count,
			 question_status, rpt_flg, kwords, note,
			 create_id, create_dt, modify_dt)
		VALUES ` + strings.Join(tuples, ", ")
	if _, err := r.db.Exec(ctx, query, args...); err != nil {
		return fmt.Errorf("question repo create many: %w", err)
	}
	return nil
}

func (r *QuestionRepository) UpdateContent(ctx context.Context, q *question.Question, modifyId *int64) error {
	answers, tags, err := marshalQuestionJSON(q)
	if err != nil {
		return err
	}
	query := `UPDATE ` + questionTable + ` q SET q.question_type = ?, q.question_name = ?, q.answers = ?,
			q.question_topic = ?, q.question_grade = ?, q.question_tags = ?, q.verified_count = ?, q.modify_id = ?
		WHERE (q.question_id = ?) AND ` + questionActiveWhere
	args := slices.Concat([]any{q.QuestionType(), q.QuestionName(), answers,
		q.QuestionTopic(), q.QuestionGrade(), tags, q.VerifiedCount(), modifyId,
		q.QuestionId()}, questionActiveArgs())
	if _, err := r.db.Exec(ctx, query, args...); err != nil {
		return fmt.Errorf("question repo update content: %w", err)
	}
	return nil
}

func (r *QuestionRepository) SetVerifiedCountByExamId(ctx context.Context, examId int64, verifiedCount int, modifyId *int64) error {
	query := `UPDATE ` + questionTable + ` q SET q.verified_count = ?, q.modify_id = ?` +
		` WHERE (q.exam_id = ?) AND ` + questionActiveWhere
	if _, err := r.db.Exec(ctx, query, slices.Concat([]any{verifiedCount, modifyId, examId}, questionActiveArgs())...); err != nil {
		return fmt.Errorf("question repo set verified count: %w", err)
	}
	return nil
}

// marshalQuestionJSON renders the two JSON columns. question_tags is never NULL
// and never omitted: the column has no DEFAULT (migration 036, error 3903
// once ix_question_tags exists), so no tags is '[]'.
func marshalQuestionJSON(q *question.Question) (string, string, error) {
	answers := make([]answerJSON, 0, len(q.Answers()))
	for _, a := range q.Answers() {
		answers = append(answers, answerJSON{Content: a.Content, IsCorrect: a.IsCorrect})
	}
	rawAnswers, err := json.Marshal(answers)
	if err != nil {
		return "", "", fmt.Errorf("question repo marshal answers: %w", err)
	}
	tags := q.QuestionTags()
	if tags == nil {
		tags = []string{}
	}
	rawTags, err := json.Marshal(tags)
	if err != nil {
		return "", "", fmt.Errorf("question repo marshal question_tags: %w", err)
	}
	return string(rawAnswers), string(rawTags), nil
}

func ModelToDomainQuestion(m *models.QuestionModel) (*question.Question, error) {
	var answers []answerJSON
	if err := json.Unmarshal([]byte(m.Answers), &answers); err != nil {
		return nil, fmt.Errorf("question repo decode answers of %d: %w", m.QuestionId, err)
	}
	var tags []string
	if err := json.Unmarshal([]byte(m.QuestionTags), &tags); err != nil {
		return nil, fmt.Errorf("question repo decode question_tags of %d: %w", m.QuestionId, err)
	}
	domainAnswers := make([]question.Answer, 0, len(answers))
	for _, a := range answers {
		domainAnswers = append(domainAnswers, question.Answer{Content: a.Content, IsCorrect: a.IsCorrect})
	}

	q := question.NewQuestion()
	q.SetQuestionId(m.QuestionId)
	q.SetExamId(m.ExamId)
	q.SetQuestionNumber(m.QuestionNumber)
	q.SetQuestionType(m.QuestionType)
	q.SetQuestionName(m.QuestionName)
	q.SetAnswers(domainAnswers)
	q.SetQuestionTopic(m.QuestionTopic)
	q.SetQuestionGrade(m.QuestionGrade)
	q.SetQuestionTags(tags)
	q.SetVerifiedCount(m.VerifiedCount)
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
