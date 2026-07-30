package quiz

// QuizProgressReq is the request for POST /quizzes/analytics/progress —
// the per-profile quiz learning-progress chart. UserID is injected from
// the session by the handler, never read from the body. FromDt/ToDt are
// optional datetime strings; Tz is a numeric offset (default +07:00).
// Purpose optionally narrows to one quiz purpose. Limit caps the number
// of chart points (default 10, clamped [1,100]).
type QuizProgressReq struct {
	UserID    *int64  `json:"-"`
	ProfileID int64   `json:"profile_id"`
	Purpose   *string `json:"purpose"`
	FromDt    string  `json:"from_dt"`
	ToDt      string  `json:"to_dt"`
	Tz        string  `json:"tz"`
	Limit     int     `json:"limit"`
}

// QuizPoint is one chart point — a single completed quiz. Sequence is the
// 1..N positional label ("Bài N"). Score is 10-point (2 dp); ScorePct is
// the raw 0–100 value.
type QuizPoint struct {
	Sequence       int64   `json:"sequence"`
	QuizID         int64   `json:"quiz_id"`
	Purpose        string  `json:"purpose"`
	TypeOfQuiz     *string `json:"type_of_quiz"`
	Title          *string `json:"title"`
	ShortText      *string `json:"short_text"`
	CompletedDt    string  `json:"completed_dt"`
	Score          float64 `json:"score"`
	ScorePct       int64   `json:"score_pct"`
	CorrectNumber  *int64  `json:"correct_number"`
	TotalQuestions *int64  `json:"total_questions"`
}

// QuizProgressSummary backs the banner + averages. Nullable score fields
// serialise as null when the window has no data. AverageDelta is the
// 10-point delta vs the prior same-size window (null when it has no data).
type QuizProgressSummary struct {
	Count           int64    `json:"count"`
	AverageScore    *float64 `json:"average_score"`
	AverageScorePct *float64 `json:"average_score_pct"`
	AverageDelta    *float64 `json:"average_delta"`
	HighestScore    *float64 `json:"highest_score"`
	HighestScorePct *int64   `json:"highest_score_pct"`
	HighestQuizID   *int64   `json:"highest_quiz_id"`
	LowestScore     *float64 `json:"lowest_score"`
	Trend           string   `json:"trend"`
}

// QuizProgressRes is the full screen payload: the chart series + summary,
// echoing the resolved request context.
type QuizProgressRes struct {
	ProfileID int64               `json:"profile_id"`
	FromDt    string              `json:"from_dt"`
	ToDt      string              `json:"to_dt"`
	Tz        string              `json:"tz"`
	Purpose   *string             `json:"purpose"`
	Limit     int                 `json:"limit"`
	Series    []QuizPoint         `json:"series"`
	Summary   QuizProgressSummary `json:"summary"`
}
