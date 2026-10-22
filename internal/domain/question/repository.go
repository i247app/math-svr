package question

import "context"

// IRepository owns ma_questions.
type IRepository interface {
	// ListByExamId returns the live bank rows copied from one pool set,
	// ordered by question_number.
	ListByExamId(ctx context.Context, examId int64) ([]*Question, error)
	// CreateMany inserts new rows in one plain INSERT. A collision on any
	// unique key is an error, never absorbed into the colliding row.
	CreateMany(ctx context.Context, questions []*Question, createId *int64) error
	// UpdateContent overwrites the content columns (type, stem, answers,
	// topic, grade, tags, verified_count) of the live row with
	// QuestionId(); question_id, exam_id, question_number and create_* stay.
	UpdateContent(ctx context.Context, q *Question, modifyId *int64) error
	// SetVerifiedCountByExamId copies a pool's verified_count onto every
	// live bank row of that pool, in one statement.
	SetVerifiedCountByExamId(ctx context.Context, examId int64, verifiedCount int, modifyId *int64) error
}
