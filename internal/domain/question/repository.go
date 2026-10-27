package question

import "context"

// IRepository owns ma_questions.
type IRepository interface {
	// CreateMany inserts new rows (ids already minted) in one plain INSERT.
	// A live (exam_id, question_number) already held fails the statement.
	CreateMany(ctx context.Context, questions []*Question, createId *int64) error
	// ListByExamId returns the live rows of one exam, by question_number.
	ListByExamId(ctx context.Context, examId int64) ([]*Question, error)
	// ReplaceByExamId soft-deletes every live row of the exam (stamping
	// modify_id) and inserts questions in their place, created by actorId.
	ReplaceByExamId(ctx context.Context, examId int64, questions []*Question, actorId *int64) error
}
