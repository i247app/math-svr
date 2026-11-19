package exam

import (
	"context"

	"math-ai.com/math-ai/internal/shared/pagination"
)

// JourneyStats is one journey as a dashboard reads it: the row that owns
// the lifecycle, the PRACTICE row that shares its id once the child has
// practised in it (nil until then, and always nil for a journey type
// that has no practice), and every sitting of the journey still
// IN_PROGRESS — of either type, oldest first — so the child can be
// offered to finish, or pick between, what they walked away from. A
// child may be handed a new exam while one is open, so this is a list.
type JourneyStats struct {
	Journey    *ExamSession
	Practice   *ExamSession
	InProgress []*ExamLink
}

// ListExamPoolsFilter narrows the pool listing. ExamTypes empty means
// every type; Grade nil means every grade; IsVerified nil means verified
// or not (true = verified_count > 0, false = verified_count = 0).
type ListExamPoolsFilter struct {
	ExamTypes  []string
	Grade      *int
	IsVerified *bool
}

// IExamPoolRepository owns the shared, user-less question sets.
//
// FindReusableByExtras is the cache read. It returns (nil, nil) on a miss
// so the caller can fall through to a real generation without treating a
// miss as an error. A set the given profile has already sat — any
// ma_exam_links row of theirs pointing at it, submitted or not — is
// never a candidate: a child must not meet a paper twice, and the child
// who triggered a generation holds an attempt on it from that moment.
type IExamPoolRepository interface {
	FindByExamId(ctx context.Context, examId int64) (*ExamPool, error)
	FindReusableByExtras(ctx context.Context, extras string, excludeProfileId int64) (*ExamPool, error)
	// CountByExtras reports how many question sets already sit under one
	// cache tag. The caller uses it to decide whether the pool is deep
	// enough to serve from — see the variant threshold in module/exam.
	CountByExtras(ctx context.Context, extras string) (int64, error)
	// ListByExamIds hydrates many question sets at once. A history list
	// renders one title per attempt, and fetching them one by one would
	// turn a twenty-row screen into twenty-one queries.
	ListByExamIds(ctx context.Context, examIds []int64) ([]*ExamPool, error)
	// ListPage returns one page of question sets, newest first.
	ListPage(ctx context.Context, filter ListExamPoolsFilter, page, size int64) ([]*ExamPool, *pagination.Pagination, error)
	// MarkVerified sets verified_count to 1 when it is 0 (verified = true;
	// an already verified set keeps its count) or back to 0 (false).
	MarkVerified(ctx context.Context, examId int64, verified bool, modifyId *int64) error
	// ReplaceQuestions overwrites questions_json and adds 1 to
	// verified_count in one statement.
	ReplaceQuestions(ctx context.Context, examId int64, questionsJSON string, modifyId *int64) error
	Create(ctx context.Context, e *ExamPool) (*ExamPool, error)
}
