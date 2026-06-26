package classroomprogress

import (
	mtime "math-ai.com/math-ai/internal/domain/shared/time"
)

// ProfileProgressReq is the request body for
// POST /classrooms/progress/profile — the single-student progress
// detail (002-profile-learning-progress).
//
// ProfileID is the acting (logged-in) profile; TargetProfileID is the
// student being viewed (nil/0 => self). Purpose optionally narrows the
// stats to one exercise intent (HOMEWORK / QUIZ / EXAM).
type ProfileProgressReq struct {
	ProfileID       int64          `json:"profile_id"`
	ClassroomID     int64          `json:"classroom_id"`
	TargetProfileID *int64         `json:"target_profile_id"`
	FromDt          mtime.MathTime `json:"from_dt"`
	ToDt            mtime.MathTime `json:"to_dt"`
	Purpose         *string        `json:"purpose"`
	Tz              string         `json:"tz"`
}

// ProfileProgressRes is the full screen payload: the per-exercise chart
// series plus the four summary cards (and the trend that drives the
// banner).
type ProfileProgressRes struct {
	TargetProfileID int64           `json:"target_profile_id"`
	FromDt          mtime.MathTime  `json:"from_dt"`
	ToDt            mtime.MathTime  `json:"to_dt"`
	Tz              string          `json:"tz"`
	Purpose         *string         `json:"purpose"`
	Series          []ExercisePoint `json:"series"`
	Summary         ProgressSummary `json:"summary"`
}

// ExercisePoint is one chart point — a single graded submission.
// Score is the 10-point value (2 dp); ScorePct is the raw 0–100 form.
type ExercisePoint struct {
	ExerciseID     int64          `json:"exercise_id"`
	Title          string         `json:"title"`
	SubmittedDt    mtime.MathTime `json:"submitted_dt"`
	Score          float64        `json:"score"`
	ScorePct       int64          `json:"score_pct"`
	CorrectNumber  *int64         `json:"correct_number"`
	TotalQuestions *int64         `json:"total_questions"`
	Passed         bool           `json:"passed"`
}

// ProgressSummary backs the four cards. Nullable score fields serialise
// as null when the range has no graded submissions. AverageDelta is the
// 10-point delta vs the immediately-preceding same-length period (null
// when either period has no data).
type ProgressSummary struct {
	TotalExercises       int64          `json:"total_exercises"`
	GradedCount          int64          `json:"graded_count"`
	AverageScore         *float64       `json:"average_score"`
	AverageScorePct      *float64       `json:"average_score_pct"`
	AverageDelta         *float64       `json:"average_delta"`
	HighestScore         *float64       `json:"highest_score"`
	HighestScorePct      *int64         `json:"highest_score_pct"`
	HighestExerciseID    *int64         `json:"highest_exercise_id"`
	HighestExerciseTitle *string        `json:"highest_exercise_title"`
	HighestSubmittedDt   mtime.MathTime `json:"highest_submitted_dt"`
	PassedCount          int64          `json:"passed_count"`
	CorrectRate          *float64       `json:"correct_rate"`
	Trend                string         `json:"trend"`
}
