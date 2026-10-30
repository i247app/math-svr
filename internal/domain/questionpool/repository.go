package questionpool

import (
	"context"
	"errors"

	"math-ai.com/math-ai/internal/shared/pagination"
)

// ErrDuplicate is returned by CreateMany when a row's question_hash is
// already held by a live question (uk_live_question_hash).
var ErrDuplicate = errors.New("question_pool: duplicate content")

// ListQuestionsFilter narrows the bank listing. Every field is optional;
// set fields are ANDed. Tags matches a row holding at least one of them.
// IsVerified true = verified_count > 0, false = 0, nil = both.
type ListQuestionPoolsFilter struct {
	Grade      *int
	Level      *int
	SubLevel   *int
	Topic      *string
	Tags       []string
	Kinds      []Kind
	Statuses   []string
	Source     *Source
	IsVerified *bool
}

// PickFilter says which questions an exam may draw (PickForExam). Only
// approved, live questions are ever candidates.
type PickFilter struct {
	Grade int
	// Level narrows to one question_level; nil = any level, rated or not.
	Level *int
	// ExcludeIds are questions already chosen for this paper.
	ExcludeIds []int64
	// AvoidTexts are stems the child met recently; a matching question is
	// not a candidate.
	AvoidTexts []string
}

// IRepository owns ma_question_pools.
type IRepository interface {
	// CreateMany inserts new rows in one plain INSERT. A live row with the
	// same question_hash fails the whole statement with ErrDuplicate.
	CreateMany(ctx context.Context, questions []*QuestionPool, createId *int64) error
	// LockById reads the live row and X-locks it until the transaction
	// ends, so two edits of one question cannot overwrite each other.
	// (nil, nil) when there is none.
	LockById(ctx context.Context, id int64) (*QuestionPool, error)
	// UpdateContent overwrites every editable column (content, grade,
	// topic, level, tags, explanation, hash) of the live row and stamps
	// modify_id. A hash another live row holds fails with ErrDuplicate.
	UpdateContent(ctx context.Context, q *QuestionPool, modifyId *int64) error
	// MarkStatus sets question_status of the live row and stamps modify_id;
	// content and hash are left alone.
	MarkStatus(ctx context.Context, id int64, status string, modifyId *int64) error
	// IncVerified adds 1 to verified_count of the live row in one UPDATE
	// (concurrent calls each count) and stamps modify_id.
	IncVerified(ctx context.Context, id int64, modifyId *int64) error
	// SoftDeleteById hides the live row (status INACTIVE, deleted_dt now)
	// and stamps modify_id; its hash is free again. false = no live row.
	SoftDeleteById(ctx context.Context, id int64, modifyId *int64) (bool, error)
	// ForceDeleteById removes the row, live or soft-deleted. false = no row.
	ForceDeleteById(ctx context.Context, id int64) (bool, error)
	// ListByIds returns the live rows among ids, in no particular order.
	ListByIds(ctx context.Context, ids []int64) ([]*QuestionPool, error)
	// FindLiveHashes returns which of hashes a live row already holds, so a
	// refusal can name the duplicate questions.
	FindLiveHashes(ctx context.Context, hashes []string) ([]string, error)
	// ListPage pages the live rows matching filter, newest first
	// (create_dt DESC, question_pool_id DESC).
	ListPage(ctx context.Context, filter ListQuestionPoolsFilter, page, size int64) ([]*QuestionPool, *pagination.Pagination, error)
	// PickForExam returns up to limit candidates of filter in random order.
	PickForExam(ctx context.Context, filter PickFilter, limit int) ([]*QuestionPool, error)
}
