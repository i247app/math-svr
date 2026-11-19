package exam

import "context"

// IExamSessionLineRepository owns the per-question log.
//
// CreateBatch takes the whole sitting at once: one INSERT with N value
// tuples rather than N round-trips, since every row is written in the same
// transaction anyway.
type IExamSessionLineRepository interface {
	CreateBatch(ctx context.Context, details []*ExamSessionLine) error
	ListByElinkId(ctx context.Context, elinkId int64) ([]*ExamSessionLine, error)
	// ListByElinkIds returns the answered questions of several sittings,
	// newest sitting first (elink_id DESC), then question order.
	ListByElinkIds(ctx context.Context, elinkIds []int64) ([]*ExamSessionLine, error)
	// ListByEsessId returns EVERY answered question of one row of a
	// journey — its ASSESSMENT sittings or its PRACTICE sittings — in
	// sitting order then question order: the journey review screen.
	ListByEsessId(ctx context.Context, esessId int64, examType string) ([]*ExamSessionLine, error)
	ListRecentByEsessId(ctx context.Context, esessId int64, examType string, limit int64) ([]*ExamSessionLine, error)
}
