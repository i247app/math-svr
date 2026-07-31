# Quiz Learning Progress Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add `POST /quizzes/analytics/progress` — a read-only, per-profile learning-progress chart over standalone quizzes (`ma_quizzes`), returning a time-ordered score series + summary/trend, with optional date range and a `limit` (default 10 most recent).

**Architecture:** Pure read (CQRS-lite query side, no UoW, no `seq`). A lightweight repo projection avoids loading LONGTEXT columns. Generic score/trend helpers are promoted from `classroomprogress` into a shared `progress` package and reused. The endpoint aligns its wire shape with feature `002-profile-learning-progress`.

**Tech Stack:** Go 1.24, `gex` HTTP, MySQL 8, project `MathError`/`response.WriteJson` envelope.

> Spec: [`FEATURE-SPEC.md`](./FEATURE-SPEC.md). Bottom-up build order:
> shared helpers → migration → status codes → domain → repo → dto → query
> → validator → module/handler/wiring → route → manual smoke. Each phase
> ends with `make build` + `go vet ./...` (the Stop hook runs vet). Stop
> for review at every **🛑 LAYER BOUNDARY** marker.
>
> Legend: `[ ]` todo · `[~]` in progress · `[x]` done. Paths repo-relative.
>
> **Commit note:** `git commit` is `ask`-gated and `Edit(migrations/**)` is
> `ask`-gated (settings.json). Commit steps will prompt for approval.

---

## File structure

| File | Responsibility |
|---|---|
| `internal/application/query/progress/trend.go` (new) | Shared `PctTo10Pt`, `LinearSlope`, `Classify`. |
| `internal/application/query/progress/trend_test.go` (new) | Table-driven tests for the three helpers. |
| `internal/application/query/classroomprogress/trend.go` (delete) | Contents moved to `progress`. |
| `internal/application/query/classroomprogress/profile_progress_query.go` (modify) | Re-point helper calls to `progress.*`. |
| `migrations/028_ma_quizzes_progress_index.sql` (new) | Index backing the analytics access path. |
| `internal/domain/shared/status/{code,message_en,message_vn}.go` (modify) | 5 `QUIZ_ANALYTICS_*` codes. |
| `internal/domain/quiz/repository.go` (modify) | `ProgressPoint`, `ProgressPointsParams`, `ListProgressPoints`. |
| `internal/infrastructure/persistence/mysql/repositories/quiz_repository.go` (modify) | `ListProgressPoints` impl + projection scan. |
| `internal/application/dto/quiz/quiz_progress_dto.go` (new) | `QuizProgressReq/Res`, `QuizPoint`, `QuizProgressSummary`, mapper. |
| `internal/application/query/quiz/get_quiz_progress_query.go` (new) | Aggregation: current + prior window → series + summary. |
| `internal/application/query/quiz/get_quiz_progress_query_test.go` (new) | Aggregation tests with a fake reader. |
| `internal/module/quiz/errors.go` (modify) | New sentinel errors. |
| `internal/module/quiz/validator.go` (modify) | `ValidateQuizProgress`. |
| `internal/module/quiz/validator_test.go` (new) | Validator tests. |
| `internal/module/quiz/service.go` (modify) | `GetQuizProgress` + wire query handler in `NewService`. |
| `internal/module/quiz/handler.go` (modify) | `HandleGetQuizProgress`. |
| `internal/bootstrap/routes/routes.go` (modify) | Register the route. |

No changes to `transaction.Repositories`, container `type.go`, or `NewService`'s argument list (the query handler is built from the already-injected `quizRepo`).

---

## Phase 0 — Promote shared score/trend helpers

🛑 LAYER BOUNDARY (touches feature `002` code). Behavior must stay identical.

### Task 0: Move `PctTo10Pt` / `LinearSlope` / `Classify` into `progress`

**Files:**
- Create: `internal/application/query/progress/trend.go`
- Create: `internal/application/query/progress/trend_test.go`
- Delete: `internal/application/query/classroomprogress/trend.go`
- Modify: `internal/application/query/classroomprogress/profile_progress_query.go`

- [ ] **Step 1: Create the shared package** — `internal/application/query/progress/trend.go`:

```go
// Package progress holds score/trend math shared by the learning-progress
// query handlers (classroom exercises + standalone quizzes). Pure
// functions, no I/O — the single source of truth for the 10-point
// rounding rule, the least-squares slope, and the trend classification.
package progress

import (
	"math"

	"math-ai.com/math-ai/internal/shared/enum"
)

// PctTo10Pt converts a 0–100 score_percentage to the 10-point UI scale
// with TWO decimal places: pct=87.5 → 8.75, pct=100 → 10.00.
func PctTo10Pt(pct float64) float64 {
	return math.Round(pct*10) / 100
}

// LinearSlope returns the least-squares slope of values onto their
// integer indices. NaN entries are skipped (keeping original index as x).
// Returns 0 with fewer than 2 non-NaN points or a vanishing denominator.
func LinearSlope(values []float64) float64 {
	type pt struct{ x, y float64 }
	pts := make([]pt, 0, len(values))
	for i, v := range values {
		if math.IsNaN(v) {
			continue
		}
		pts = append(pts, pt{x: float64(i), y: v})
	}
	if len(pts) < 2 {
		return 0
	}
	var sx, sy float64
	for _, p := range pts {
		sx += p.x
		sy += p.y
	}
	n := float64(len(pts))
	meanX := sx / n
	meanY := sy / n
	var num, den float64
	for _, p := range pts {
		dx := p.x - meanX
		num += dx * (p.y - meanY)
		den += dx * dx
	}
	if den == 0 {
		return 0
	}
	return num / den
}

// Classify maps (count, avg10pt, slope) to a comment enum in
// first-match-wins order. Threshold magic-numbers live here. See
// docs/features/002-profile-learning-progress/FEATURE-SPEC.md §4.
func Classify(count int, avg10pt, slope float64) enum.ProgressComment {
	switch {
	case count == 0:
		return enum.ProgressCommentNoData
	case count == 1:
		return enum.ProgressCommentInsufficient
	case avg10pt < 5.0 || slope <= -0.05:
		return enum.ProgressCommentNeedToTry
	case avg10pt >= 7.5 && slope >= 0.0:
		return enum.ProgressCommentGoodProgress
	default:
		return enum.ProgressCommentProgress
	}
}
```

- [ ] **Step 2: Write the helper tests** — `internal/application/query/progress/trend_test.go`:

```go
package progress

import (
	"math"
	"testing"

	"math-ai.com/math-ai/internal/shared/enum"
)

func TestPctTo10Pt(t *testing.T) {
	cases := []struct {
		pct  float64
		want float64
	}{
		{0, 0}, {60, 6}, {77, 7.7}, {87.5, 8.75}, {100, 10},
	}
	for _, c := range cases {
		if got := PctTo10Pt(c.pct); math.Abs(got-c.want) > 1e-9 {
			t.Errorf("PctTo10Pt(%v)=%v want %v", c.pct, got, c.want)
		}
	}
}

func TestLinearSlope(t *testing.T) {
	if got := LinearSlope([]float64{1, 2, 3}); math.Abs(got-1) > 1e-9 {
		t.Errorf("rising slope = %v want 1", got)
	}
	if got := LinearSlope([]float64{5}); got != 0 {
		t.Errorf("single point slope = %v want 0", got)
	}
	if got := LinearSlope([]float64{3, 2, 1}); math.Abs(got+1) > 1e-9 {
		t.Errorf("falling slope = %v want -1", got)
	}
}

func TestClassify(t *testing.T) {
	cases := []struct {
		count      int
		avg, slope float64
		want       enum.ProgressComment
	}{
		{0, 0, 0, enum.ProgressCommentNoData},
		{1, 9, 1, enum.ProgressCommentInsufficient},
		{5, 4.9, 1, enum.ProgressCommentNeedToTry},
		{5, 8, -0.1, enum.ProgressCommentNeedToTry},
		{5, 8, 0.2, enum.ProgressCommentGoodProgress},
		{5, 6, 0.1, enum.ProgressCommentProgress},
	}
	for _, c := range cases {
		if got := Classify(c.count, c.avg, c.slope); got != c.want {
			t.Errorf("Classify(%d,%v,%v)=%v want %v", c.count, c.avg, c.slope, got, c.want)
		}
	}
}
```

- [ ] **Step 3: Delete the old helper file**

Run: `git rm internal/application/query/classroomprogress/trend.go`
Expected: file removed.

- [ ] **Step 4: Re-point `classroomprogress` call sites** — in `internal/application/query/classroomprogress/profile_progress_query.go`:

Add to the import block:
```go
	"math-ai.com/math-ai/internal/application/query/progress"
```
Then replace the three bare calls (there are 4 call sites total):
- `avg10OfPoints`: `PctTo10Pt(...)` → `progress.PctTo10Pt(...)`
- `buildPoints`: `PctTo10Pt(float64(pct))` → `progress.PctTo10Pt(float64(pct))`
- `buildSummary`: `PctTo10Pt(avgPct)` → `progress.PctTo10Pt(avgPct)`, `LinearSlope(scores10)` → `progress.LinearSlope(scores10)`, `Classify(int(graded), avg10, slope)` → `progress.Classify(int(graded), avg10, slope)`

- [ ] **Step 5: Build, vet, test**

Run: `make build && go vet ./... && go test ./internal/application/query/progress/... ./internal/application/query/classroomprogress/...`
Expected: build clean, vet clean, tests PASS.

- [ ] **Step 6: Commit**

```bash
git add internal/application/query/progress internal/application/query/classroomprogress
git commit -m "refactor(progress): promote score/trend helpers to shared package"
```

---

## Phase 1 — Migration (index)

🛑 LAYER BOUNDARY (DB). `Edit(migrations/**)` is `ask`-gated. Boot-time `database.Migrate` is disabled (known-issues §5) — apply manually.

### Task 1: Add the analytics index

**Files:**
- Create: `migrations/028_ma_quizzes_progress_index.sql`

- [ ] **Step 1: Write the migration**

```sql
-- migration up
-- Backs POST /quizzes/analytics/progress: per-profile SUBMITTED-quiz range
-- scan ordered by completion time (modify_dt). Supersedes the commented-out
-- idx_profile_status stub in 009 for the analytics access path.
ALTER TABLE ma_quizzes
  ADD KEY ix_quiz_profile_status_modify (profile_id, quiz_status, modify_dt);
```

- [ ] **Step 2: Sanity compile** (no schema parser runs at build)

Run: `make build`
Expected: build clean.

- [ ] **Step 3: Commit**

```bash
git add migrations/028_ma_quizzes_progress_index.sql
git commit -m "feat(db): index ma_quizzes for quiz progress analytics"
```

> The index is applied to each environment manually (see Phase 10).

---

## Phase 2 — Status codes

No layer crossing — pure constants. Lockstep across 3 files.

### Task 2: Add `QUIZ_ANALYTICS_*` codes

**Files:**
- Modify: `internal/domain/shared/status/code.go`
- Modify: `internal/domain/shared/status/message_en.go`
- Modify: `internal/domain/shared/status/message_vn.go`

- [ ] **Step 1: Append codes** in `code.go`, immediately after `QUIZ_ANSWER_SCHEMA_INVALID StatusCode = 11016`:

```go
	QUIZ_ANALYTICS_MISSING_PROFILE   StatusCode = 11017
	QUIZ_ANALYTICS_PROFILE_NOT_OWNED StatusCode = 11018
	QUIZ_ANALYTICS_INVALID_DATE_RANGE StatusCode = 11019
	QUIZ_ANALYTICS_INVALID_TZ        StatusCode = 11020
	QUIZ_ANALYTICS_INVALID_PURPOSE   StatusCode = 11021
```

- [ ] **Step 2: Append EN messages** in `message_en.go`, after the `case QUIZ_ANSWER_SCHEMA_INVALID:` block:

```go
	case QUIZ_ANALYTICS_MISSING_PROFILE:
		return "Profile is required."
	case QUIZ_ANALYTICS_PROFILE_NOT_OWNED:
		return "This profile does not belong to you."
	case QUIZ_ANALYTICS_INVALID_DATE_RANGE:
		return "Invalid date range."
	case QUIZ_ANALYTICS_INVALID_TZ:
		return "Invalid time zone."
	case QUIZ_ANALYTICS_INVALID_PURPOSE:
		return "Invalid quiz type."
```

- [ ] **Step 3: Append VN messages** in `message_vn.go`, after the `case QUIZ_ANSWER_SCHEMA_INVALID:` block:

```go
	case QUIZ_ANALYTICS_MISSING_PROFILE:
		return "Hồ sơ học sinh là bắt buộc."
	case QUIZ_ANALYTICS_PROFILE_NOT_OWNED:
		return "Hồ sơ này không thuộc về bạn."
	case QUIZ_ANALYTICS_INVALID_DATE_RANGE:
		return "Khoảng thời gian không hợp lệ."
	case QUIZ_ANALYTICS_INVALID_TZ:
		return "Múi giờ không hợp lệ."
	case QUIZ_ANALYTICS_INVALID_PURPOSE:
		return "Loại quiz không hợp lệ."
```

- [ ] **Step 4: Build + vet**

Run: `make build && go vet ./...`
Expected: clean.

- [ ] **Step 5: Commit**

```bash
git add internal/domain/shared/status
git commit -m "feat(status): add QUIZ_ANALYTICS_* codes (11017-11021)"
```

---

## Phase 3 — Domain: projection type + repo interface

🛑 LAYER BOUNDARY (domain). No behavior — interface + value types only.

### Task 3: Add `ProgressPoint`, `ProgressPointsParams`, `ListProgressPoints`

**Files:**
- Modify: `internal/domain/quiz/repository.go`

- [ ] **Step 1: Add the types + method.** Add `"math-ai.com/math-ai/internal/domain/shared/mtime"` to the import block, then append:

```go
// ProgressPoint is a lightweight read projection of a completed quiz for
// the learning-progress chart. It deliberately omits the LONGTEXT
// questions/answers blobs — the analytics query never needs them.
// ScorePercentage is non-null by construction (the query filters
// score_percentage IS NOT NULL). CompletedDt is the SUBMITTED row's
// modify_dt (see FEATURE-SPEC §2 decision 6).
type ProgressPoint struct {
	QuizId          int64
	Purpose         string
	TypeOfQuiz      *string
	Title           *string
	ShortText       *string
	ScorePercentage int64
	CorrectNumber   *int64
	TotalQuestions  *int64
	CompletedDt     mtime.MathTime
}

// ProgressPointsParams drives ListProgressPoints. From/To bound the
// completion time (nil = open). CompletedBefore fetches the prior window
// (rows strictly older than the anchor); nil for the current window.
// Purpose nil = all purposes. Limit is pre-clamped by the caller.
type ProgressPointsParams struct {
	ProfileID       int64
	Purpose         *string
	From            *mtime.MathTime
	To              *mtime.MathTime
	CompletedBefore *mtime.MathTime
	Limit           int64
}
```

Add this method to the `IRepository` interface:
```go
	ListProgressPoints(ctx context.Context, params ProgressPointsParams) ([]*ProgressPoint, error)
```

- [ ] **Step 2: Build (expect a compile FAILURE — repo does not implement the method yet)**

Run: `go build ./internal/domain/... ./internal/infrastructure/persistence/mysql/...`
Expected: FAIL — `*QuizRepository` does not implement `quiz.IRepository (missing ListProgressPoints)`. This confirms the interface is wired; Phase 4 satisfies it.

---

## Phase 4 — Repository implementation

🛑 LAYER BOUNDARY (infra). Only place SQL is written.

### Task 4: Implement `ListProgressPoints`

**Files:**
- Modify: `internal/infrastructure/persistence/mysql/repositories/quiz_repository.go`

- [ ] **Step 1: Add the projection columns + scan + method.** Append to the file (the package already imports `context`, `fmt`, `mtime`, `quiz`, `database`):

```go
// progressPointColumns is the narrow projection for the analytics chart —
// no LONGTEXT questions/answers. Order == scanProgressPoint's Scan order.
const progressPointColumns = `q.quiz_id, q.purpose, q.type_of_quiz, q.title, q.short_text,
	q.score_percentage, q.correct_number, q.total_questions, q.modify_dt`

func scanProgressPoint(s database.RowScanner) (*quiz.ProgressPoint, error) {
	var (
		p          quiz.ProgressPoint
		typeOfQuiz *string
		title      *string
		shortText  *string
		correct    *int64
		total      *int64
		modifyDt   time.Time
	)
	if err := s.Scan(&p.QuizId, &p.Purpose, &typeOfQuiz, &title, &shortText,
		&p.ScorePercentage, &correct, &total, &modifyDt); err != nil {
		return nil, err
	}
	p.TypeOfQuiz = typeOfQuiz
	p.Title = title
	p.ShortText = shortText
	p.CorrectNumber = correct
	p.TotalQuestions = total
	p.CompletedDt = mtime.NewTime(modifyDt)
	return &p, nil
}

// ListProgressPoints returns completed (SUBMITTED, graded) quizzes for one
// profile, newest first, capped at params.Limit. The caller reverses to
// chronological order. Only the columns the chart needs are selected.
func (r *QuizRepository) ListProgressPoints(ctx context.Context, params quiz.ProgressPointsParams) ([]*quiz.ProgressPoint, error) {
	where := quizActiveWhere +
		` AND q.quiz_status = ? AND q.profile_id = ? AND q.score_percentage IS NOT NULL`
	args := append(quizActiveArgs(), string(enum.QuizStatusTypeSubmitted), params.ProfileID)

	if params.Purpose != nil && *params.Purpose != "" {
		where += ` AND q.purpose = ?`
		args = append(args, *params.Purpose)
	}
	if params.From != nil && !params.From.IsZero() {
		where += ` AND q.modify_dt >= ?`
		args = append(args, params.From.Time)
	}
	if params.To != nil && !params.To.IsZero() {
		where += ` AND q.modify_dt <= ?`
		args = append(args, params.To.Time)
	}
	if params.CompletedBefore != nil && !params.CompletedBefore.IsZero() {
		where += ` AND q.modify_dt < ?`
		args = append(args, params.CompletedBefore.Time)
	}

	limit := params.Limit
	if limit <= 0 {
		limit = 10
	}
	args = append(args, limit)

	query := `SELECT ` + progressPointColumns + ` FROM ` + quizTable + ` q WHERE ` +
		where + ` ORDER BY q.modify_dt DESC, q.quiz_id DESC LIMIT ?`

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("quiz repo list progress points: %w", err)
	}
	defer rows.Close()

	var points []*quiz.ProgressPoint
	for rows.Next() {
		p, err := scanProgressPoint(rows)
		if err != nil {
			return nil, fmt.Errorf("quiz repo scan progress point: %w", err)
		}
		points = append(points, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("quiz repo progress points iteration: %w", err)
	}
	return points, nil
}
```

- [ ] **Step 2: Add the `time` import** if not present. Check the import block of `quiz_repository.go`; it currently imports `database/sql`, `errors`, `fmt`, `context`, and project packages but NOT `time`. Add:
```go
	"time"
```

- [ ] **Step 3: Build + vet**

Run: `make build && go vet ./...`
Expected: clean (the Phase 3 interface is now satisfied).

- [ ] **Step 4: Commit**

```bash
git add internal/domain/quiz/repository.go internal/infrastructure/persistence/mysql/repositories/quiz_repository.go
git commit -m "feat(quiz): add ListProgressPoints projection query"
```

---

## Phase 5 — DTO

### Task 5: Progress request/response shapes

**Files:**
- Create: `internal/application/dto/quiz/quiz_progress_dto.go`

- [ ] **Step 1: Write the DTO file**

```go
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
```

- [ ] **Step 2: Build + vet**

Run: `make build && go vet ./...`
Expected: clean.

- [ ] **Step 3: Commit**

```bash
git add internal/application/dto/quiz/quiz_progress_dto.go
git commit -m "feat(quiz): add quiz progress DTO shapes"
```

---

## Phase 6 — Query handler (aggregation, TDD)

### Task 6: `GetQuizProgressQueryHandler`

**Files:**
- Create: `internal/application/query/quiz/get_quiz_progress_query.go`
- Test: `internal/application/query/quiz/get_quiz_progress_query_test.go`

- [ ] **Step 1: Write the handler + aggregation** — `get_quiz_progress_query.go`:

```go
package query

import (
	"context"

	dto "math-ai.com/math-ai/internal/application/dto/quiz"
	"math-ai.com/math-ai/internal/application/query/progress"
	"math-ai.com/math-ai/internal/domain/quiz"
	"math-ai.com/math-ai/internal/domain/shared/mtime"
	"math-ai.com/math-ai/internal/shared/enum"
)

// quizProgressReader is the narrow slice of quiz.IRepository this query
// needs. Accepting an interface (not the concrete repo) keeps the handler
// unit-testable with a hand-rolled fake. quiz.IRepository satisfies it.
type quizProgressReader interface {
	ListProgressPoints(ctx context.Context, params quiz.ProgressPointsParams) ([]*quiz.ProgressPoint, error)
}

// GetQuizProgressQuery carries the resolved inputs. From/To are optional
// (empty string = open bound). Purpose nil = all. Limit is pre-clamped.
type GetQuizProgressQuery struct {
	ProfileID int64
	Purpose   *string
	From      string
	To        string
	Limit     int64
}

// GetQuizProgressResult is the pure-data output; the module wraps it with
// the request echoes into dto.QuizProgressRes.
type GetQuizProgressResult struct {
	Series  []dto.QuizPoint
	Summary dto.QuizProgressSummary
}

type GetQuizProgressQueryHandler struct {
	reader quizProgressReader
}

func NewGetQuizProgressQueryHandler(reader quizProgressReader) *GetQuizProgressQueryHandler {
	return &GetQuizProgressQueryHandler{reader: reader}
}

func (h *GetQuizProgressQueryHandler) Handle(ctx context.Context, q GetQuizProgressQuery) (*GetQuizProgressResult, error) {
	params := quiz.ProgressPointsParams{
		ProfileID: q.ProfileID,
		Purpose:   q.Purpose,
		Limit:     q.Limit,
	}
	if q.From != "" {
		from, err := mtime.ParseFromString(q.From)
		if err != nil {
			return nil, err
		}
		params.From = &from
	}
	if q.To != "" {
		to, err := mtime.ParseFromString(q.To)
		if err != nil {
			return nil, err
		}
		params.To = &to
	}

	// Current window (newest-first from the repo).
	desc, err := h.reader.ListProgressPoints(ctx, params)
	if err != nil {
		return nil, err
	}
	series := toSeriesAsc(desc)

	// Prior window: the same-size set immediately before the current
	// window's oldest point. Only meaningful when the current window has
	// data.
	var priorAvg10 *float64
	if len(series) > 0 {
		anchor := desc[len(desc)-1].CompletedDt // oldest of current
		priorParams := quiz.ProgressPointsParams{
			ProfileID:       q.ProfileID,
			Purpose:         q.Purpose,
			CompletedBefore: &anchor,
			Limit:           q.Limit,
		}
		prior, err := h.reader.ListProgressPoints(ctx, priorParams)
		if err != nil {
			return nil, err
		}
		if avg, ok := avg10OfPoints(prior); ok {
			priorAvg10 = &avg
		}
	}

	return &GetQuizProgressResult{
		Series:  series,
		Summary: buildSummary(series, priorAvg10),
	}, nil
}

// toSeriesAsc reverses the newest-first projection into chronological
// order and assigns the 1..N sequence label.
func toSeriesAsc(desc []*quiz.ProgressPoint) []dto.QuizPoint {
	n := len(desc)
	out := make([]dto.QuizPoint, 0, n)
	for i := n - 1; i >= 0; i-- {
		p := desc[i]
		out = append(out, dto.QuizPoint{
			Sequence:       int64(n - i),
			QuizID:         p.QuizId,
			Purpose:        p.Purpose,
			TypeOfQuiz:     p.TypeOfQuiz,
			Title:          p.Title,
			ShortText:      p.ShortText,
			CompletedDt:    p.CompletedDt.String(),
			Score:          progress.PctTo10Pt(float64(p.ScorePercentage)),
			ScorePct:       p.ScorePercentage,
			CorrectNumber:  p.CorrectNumber,
			TotalQuestions: p.TotalQuestions,
		})
	}
	return out
}

// avg10OfPoints returns the 10-point average of a projection slice and
// whether it had any points.
func avg10OfPoints(points []*quiz.ProgressPoint) (float64, bool) {
	if len(points) == 0 {
		return 0, false
	}
	var sum int64
	for _, p := range points {
		sum += p.ScorePercentage
	}
	return progress.PctTo10Pt(float64(sum) / float64(len(points))), true
}

// buildSummary aggregates count, averages, highest/lowest, delta, trend.
func buildSummary(series []dto.QuizPoint, priorAvg10 *float64) dto.QuizProgressSummary {
	count := int64(len(series))
	summary := dto.QuizProgressSummary{Count: count}
	if count == 0 {
		summary.Trend = string(enum.ProgressCommentNoData)
		return summary
	}

	var sumPct int64
	scores10 := make([]float64, 0, len(series))
	hi := series[0]
	lo := series[0]
	for _, p := range series {
		sumPct += p.ScorePct
		scores10 = append(scores10, p.Score)
		if p.ScorePct >= hi.ScorePct { // ties → latest (series is ASC)
			hi = p
		}
		if p.ScorePct < lo.ScorePct {
			lo = p
		}
	}

	avgPct := float64(sumPct) / float64(count)
	avg10 := progress.PctTo10Pt(avgPct)
	slope := progress.LinearSlope(scores10)

	summary.AverageScore = &avg10
	summary.AverageScorePct = &avgPct
	summary.Trend = string(progress.Classify(int(count), avg10, slope))

	hiScore := hi.Score
	hiPct := hi.ScorePct
	hiID := hi.QuizID
	summary.HighestScore = &hiScore
	summary.HighestScorePct = &hiPct
	summary.HighestQuizID = &hiID

	loScore := lo.Score
	summary.LowestScore = &loScore

	if priorAvg10 != nil {
		delta := avg10 - *priorAvg10
		summary.AverageDelta = &delta
	}
	return summary
}
```

- [ ] **Step 2: Write the aggregation test** — `get_quiz_progress_query_test.go`:

```go
package query

import (
	"context"
	"math"
	"testing"
	"time"

	"math-ai.com/math-ai/internal/domain/quiz"
	"math-ai.com/math-ai/internal/domain/shared/mtime"
	"math-ai.com/math-ai/internal/shared/enum"
)

// fakeReader returns queued result sets in call order: first call = current
// window, second call = prior window.
type fakeReader struct {
	calls   int
	results [][]*quiz.ProgressPoint
}

func (f *fakeReader) ListProgressPoints(_ context.Context, _ quiz.ProgressPointsParams) ([]*quiz.ProgressPoint, error) {
	i := f.calls
	f.calls++
	if i < len(f.results) {
		return f.results[i], nil
	}
	return nil, nil
}

func pt(id, pct int64, day int) *quiz.ProgressPoint {
	return &quiz.ProgressPoint{
		QuizId:          id,
		Purpose:         "PRACTICE",
		ScorePercentage: pct,
		CompletedDt:     mtime.NewTime(time.Date(2026, 5, day, 9, 0, 0, 0, time.UTC)),
	}
}

func TestHandle_EmptyIsNoData(t *testing.T) {
	h := NewGetQuizProgressQueryHandler(&fakeReader{results: [][]*quiz.ProgressPoint{{}}})
	res, err := h.Handle(context.Background(), GetQuizProgressQuery{ProfileID: 1, Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Series) != 0 {
		t.Errorf("series len = %d want 0", len(res.Series))
	}
	if res.Summary.Trend != string(enum.ProgressCommentNoData) {
		t.Errorf("trend = %s want NO_DATA", res.Summary.Trend)
	}
	if res.Summary.AverageScore != nil || res.Summary.AverageDelta != nil {
		t.Error("nullable summary fields should be nil on empty")
	}
}

func TestHandle_SeriesAscAndSummary(t *testing.T) {
	// Repo returns newest-first: day6=100, day5=70, day4=90 ... use 3 points.
	current := []*quiz.ProgressPoint{pt(3, 90, 6), pt(2, 70, 5), pt(1, 60, 4)}
	prior := []*quiz.ProgressPoint{pt(0, 50, 3)} // avg10 = 5.0
	h := NewGetQuizProgressQueryHandler(&fakeReader{results: [][]*quiz.ProgressPoint{current, prior}})

	res, err := h.Handle(context.Background(), GetQuizProgressQuery{ProfileID: 1, Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	// Ascending: id1 (seq1), id2 (seq2), id3 (seq3).
	if len(res.Series) != 3 || res.Series[0].QuizID != 1 || res.Series[0].Sequence != 1 ||
		res.Series[2].QuizID != 3 || res.Series[2].Sequence != 3 {
		t.Fatalf("series order/sequence wrong: %+v", res.Series)
	}
	if res.Series[0].Score != 6.0 || res.Series[2].Score != 9.0 {
		t.Errorf("score mapping wrong: %v %v", res.Series[0].Score, res.Series[2].Score)
	}
	// avg pct = (90+70+60)/3 = 73.33 → 7.33.
	if res.Summary.AverageScore == nil || math.Abs(*res.Summary.AverageScore-7.33) > 0.01 {
		t.Errorf("average_score = %v want ~7.33", res.Summary.AverageScore)
	}
	if res.Summary.HighestQuizID == nil || *res.Summary.HighestQuizID != 3 {
		t.Errorf("highest quiz id wrong: %v", res.Summary.HighestQuizID)
	}
	if res.Summary.LowestScore == nil || *res.Summary.LowestScore != 6.0 {
		t.Errorf("lowest score = %v want 6.0", res.Summary.LowestScore)
	}
	// delta = 7.33 - 5.0 = 2.33.
	if res.Summary.AverageDelta == nil || math.Abs(*res.Summary.AverageDelta-2.33) > 0.01 {
		t.Errorf("average_delta = %v want ~2.33", res.Summary.AverageDelta)
	}
}
```

- [ ] **Step 3: Run tests**

Run: `go test ./internal/application/query/quiz/... -run TestHandle -v`
Expected: PASS (both tests).

- [ ] **Step 4: Build + vet + commit**

Run: `make build && go vet ./...`
```bash
git add internal/application/query/quiz/get_quiz_progress_query.go internal/application/query/quiz/get_quiz_progress_query_test.go
git commit -m "feat(quiz): quiz progress aggregation query"
```

---

## Phase 7 — Validator (TDD)

### Task 7: `ValidateQuizProgress` + sentinel errors

**Files:**
- Modify: `internal/module/quiz/errors.go`
- Modify: `internal/module/quiz/validator.go`
- Test: `internal/module/quiz/validator_test.go`

- [ ] **Step 1: Add sentinel errors** in `errors.go` (inside the `var (...)` block):

```go
	ErrProgressProfileRequired  = errors.New("profile_id is required")
	ErrProgressInvalidDateRange = errors.New("from_dt must be before to_dt and within 2 years")
	ErrProgressInvalidTz        = errors.New("tz must be a numeric offset like +07:00")
	ErrProgressInvalidPurpose   = errors.New("purpose must be one of ASSESSMENT, PRACTICE, EXAM")
```

- [ ] **Step 2: Add the validator** in `validator.go`. Add imports `"time"`, `dto` is already imported, `enum` already imported; add `mtime "math-ai.com/math-ai/internal/domain/shared/mtime"`. Then:

```go
// ProgressLimitDefault / Min / Max bound the number of chart points.
const (
	ProgressLimitDefault = 10
	ProgressLimitMin     = 1
	ProgressLimitMax     = 100
	// progressMaxRangeDays caps the date window to ~2 years (mirrors 002).
	progressMaxRange = 2 * 365 * 24 * time.Hour
)

// ValidateQuizProgress checks the analytics request and normalises Limit
// (clamped in place) + Tz (default applied). ProfileID/ownership are
// enforced in the service (needs a repo lookup); this covers pure input.
func ValidateQuizProgress(ctx context.Context, req *dto.QuizProgressReq) error {
	if req.ProfileID <= 0 {
		return errs.NewError(ctx, status.QUIZ_ANALYTICS_MISSING_PROFILE, nil, ErrProgressProfileRequired)
	}

	if req.Purpose != nil {
		p := enum.QuizPurpose(strings.ToUpper(strings.TrimSpace(*req.Purpose)))
		if !p.IsValid() {
			return errs.NewError(ctx, status.QUIZ_ANALYTICS_INVALID_PURPOSE, nil, ErrProgressInvalidPurpose)
		}
		up := string(p)
		req.Purpose = &up
	}

	if req.Tz == "" {
		req.Tz = enum.DefaultProgressTz
	} else if !enum.IsValidTzOffset(req.Tz) {
		return errs.NewError(ctx, status.QUIZ_ANALYTICS_INVALID_TZ, nil, ErrProgressInvalidTz)
	}

	if req.FromDt != "" || req.ToDt != "" {
		if req.FromDt == "" || req.ToDt == "" {
			return errs.NewError(ctx, status.QUIZ_ANALYTICS_INVALID_DATE_RANGE, nil, ErrProgressInvalidDateRange)
		}
		from, err := mtime.ParseFromString(req.FromDt)
		if err != nil {
			return errs.NewError(ctx, status.QUIZ_ANALYTICS_INVALID_DATE_RANGE, nil, err)
		}
		to, err := mtime.ParseFromString(req.ToDt)
		if err != nil {
			return errs.NewError(ctx, status.QUIZ_ANALYTICS_INVALID_DATE_RANGE, nil, err)
		}
		if !from.Time.Before(to.Time) || to.Time.Sub(from.Time) > progressMaxRange {
			return errs.NewError(ctx, status.QUIZ_ANALYTICS_INVALID_DATE_RANGE, nil, ErrProgressInvalidDateRange)
		}
	}

	if req.Limit == 0 {
		req.Limit = ProgressLimitDefault
	}
	if req.Limit < ProgressLimitMin {
		req.Limit = ProgressLimitMin
	}
	if req.Limit > ProgressLimitMax {
		req.Limit = ProgressLimitMax
	}
	return nil
}
```

- [ ] **Step 3: Write validator tests** — `validator_test.go`:

```go
package quiz

import (
	"context"
	"testing"

	dto "math-ai.com/math-ai/internal/application/dto/quiz"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
)

func codeOf(t *testing.T, err error) status.StatusCode {
	t.Helper()
	mErr, ok := errs.IsbinbaseError(err)
	if !ok {
		t.Fatalf("expected MathError, got %v", err)
	}
	return mErr.GetStatusCode()
}

func TestValidateQuizProgress(t *testing.T) {
	ctx := context.Background()
	practice := "practice"
	bad := "SHOPPING"

	t.Run("missing profile", func(t *testing.T) {
		err := ValidateQuizProgress(ctx, &dto.QuizProgressReq{ProfileID: 0})
		if codeOf(t, err) != status.QUIZ_ANALYTICS_MISSING_PROFILE {
			t.Fatal("wrong code")
		}
	})

	t.Run("defaults applied", func(t *testing.T) {
		req := &dto.QuizProgressReq{ProfileID: 1, Purpose: &practice}
		if err := ValidateQuizProgress(ctx, req); err != nil {
			t.Fatal(err)
		}
		if req.Limit != ProgressLimitDefault {
			t.Errorf("limit = %d want %d", req.Limit, ProgressLimitDefault)
		}
		if req.Tz != "+07:00" {
			t.Errorf("tz = %q want +07:00", req.Tz)
		}
		if req.Purpose == nil || *req.Purpose != "PRACTICE" {
			t.Errorf("purpose not upper-cased: %v", req.Purpose)
		}
	})

	t.Run("limit clamp", func(t *testing.T) {
		req := &dto.QuizProgressReq{ProfileID: 1, Limit: 9999}
		_ = ValidateQuizProgress(ctx, req)
		if req.Limit != ProgressLimitMax {
			t.Errorf("limit = %d want %d", req.Limit, ProgressLimitMax)
		}
	})

	t.Run("bad purpose", func(t *testing.T) {
		err := ValidateQuizProgress(ctx, &dto.QuizProgressReq{ProfileID: 1, Purpose: &bad})
		if codeOf(t, err) != status.QUIZ_ANALYTICS_INVALID_PURPOSE {
			t.Fatal("wrong code")
		}
	})

	t.Run("bad tz", func(t *testing.T) {
		err := ValidateQuizProgress(ctx, &dto.QuizProgressReq{ProfileID: 1, Tz: "Asia/Ho_Chi_Minh"})
		if codeOf(t, err) != status.QUIZ_ANALYTICS_INVALID_TZ {
			t.Fatal("wrong code")
		}
	})

	t.Run("reversed range", func(t *testing.T) {
		err := ValidateQuizProgress(ctx, &dto.QuizProgressReq{
			ProfileID: 1, FromDt: "2026-06-24 00:00:00", ToDt: "2026-05-01 00:00:00",
		})
		if codeOf(t, err) != status.QUIZ_ANALYTICS_INVALID_DATE_RANGE {
			t.Fatal("wrong code")
		}
	})
}
```

- [ ] **Step 4: Run tests**

Run: `go test ./internal/module/quiz/... -run TestValidateQuizProgress -v`
Expected: PASS.

- [ ] **Step 5: Build + vet + commit**

Run: `make build && go vet ./...`
```bash
git add internal/module/quiz/errors.go internal/module/quiz/validator.go internal/module/quiz/validator_test.go
git commit -m "feat(quiz): validate quiz progress request"
```

---

## Phase 8 — Module service + handler + wiring

🛑 LAYER BOUNDARY (presentation). Ties query + validator + ownership together.

### Task 8: `Service.GetQuizProgress` + handler

**Files:**
- Modify: `internal/module/quiz/service.go`
- Modify: `internal/module/quiz/handler.go`

- [ ] **Step 1: Wire the query handler into the Service struct.** In `service.go`, add a field to `Service`:
```go
	getQuizProgressQuery   *query.GetQuizProgressQueryHandler
```
And in `NewService`, in the struct literal, add:
```go
		getQuizProgressQuery:   query.NewGetQuizProgressQueryHandler(quizRepo),
```

- [ ] **Step 2: Add the service method** to `service.go`:

```go
// GetQuizProgress returns the per-profile quiz learning-progress chart.
// Validates input, enforces that the target profile belongs to the
// session user, then delegates aggregation to the query handler.
func (s *Service) GetQuizProgress(ctx context.Context, req *dto.QuizProgressReq) (*dto.QuizProgressRes, error) {
	if err := ValidateQuizProgress(ctx, req); err != nil {
		return nil, err
	}

	profile, err := s.profileRepo.FindByProfileId(ctx, req.ProfileID)
	if err != nil {
		return nil, errs.NewError(ctx, status.FAIL, nil, err)
	}
	if profile == nil {
		return nil, errs.NewError(ctx, status.PROFILE_NOT_FOUND, nil, ErrProfileNotFound)
	}
	if req.UserID == nil || profile.UserId() != *req.UserID {
		return nil, errs.NewError(ctx, status.QUIZ_ANALYTICS_PROFILE_NOT_OWNED, nil, ErrProgressProfileNotOwned)
	}

	result, err := s.getQuizProgressQuery.Handle(ctx, query.GetQuizProgressQuery{
		ProfileID: req.ProfileID,
		Purpose:   req.Purpose,
		From:      req.FromDt,
		To:        req.ToDt,
		Limit:     int64(req.Limit),
	})
	if err != nil {
		return nil, errs.NewError(ctx, status.FAIL, nil, err)
	}

	return &dto.QuizProgressRes{
		ProfileID: req.ProfileID,
		FromDt:    req.FromDt,
		ToDt:      req.ToDt,
		Tz:        req.Tz,
		Purpose:   req.Purpose,
		Limit:     req.Limit,
		Series:    result.Series,
		Summary:   result.Summary,
	}, nil
}
```

- [ ] **Step 3: Add the ownership sentinel** to `errors.go` `var (...)` block:
```go
	ErrProgressProfileNotOwned = errors.New("profile does not belong to this user")
```

- [ ] **Step 4: Add the handler** to `handler.go`:

```go
// POST /quizzes/analytics/progress
func (h *QuizHandler) HandleGetQuizProgress(w http.ResponseWriter, r *http.Request) {
	var req dto.QuizProgressReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, nil, err)
		return
	}

	session, err := h.appResource.GetRequestSession(r)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	uid, ok := session.UID()
	if !ok {
		response.WriteJson(w, nil, errs.NewError(r.Context(), status.UNAUTHORIZED, nil, ErrUidNotFoundFromSession))
		return
	}
	req.UserID = &uid

	res, err := h.quizSvc.GetQuizProgress(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, nil, err)
		return
	}
	response.WriteJson(w, res, nil)
}
```

- [ ] **Step 5: Build + vet**

Run: `make build && go vet ./...`
Expected: clean.

- [ ] **Step 6: Commit**

```bash
git add internal/module/quiz/service.go internal/module/quiz/handler.go internal/module/quiz/errors.go
git commit -m "feat(quiz): quiz progress service + handler with ownership check"
```

---

## Phase 9 — Route

### Task 9: Register the endpoint

**Files:**
- Modify: `internal/bootstrap/routes/routes.go`

- [ ] **Step 1: Add the route** in the quiz routes block (after the `soft-delete` line, ~`routes.go:204`):

```go
		reg("POST /quizzes/analytics/progress", quizHandler.HandleGetQuizProgress, authMiddleware)
```

- [ ] **Step 2: Build + vet**

Run: `make build && go vet ./...`
Expected: clean.

- [ ] **Step 3: Full test sweep**

Run: `go test ./...`
Expected: PASS (or unchanged pre-existing skips; no new failures).

- [ ] **Step 4: Commit**

```bash
git add internal/bootstrap/routes/routes.go
git commit -m "feat(quiz): register POST /quizzes/analytics/progress"
```

---

## Phase 10 — Manual smoke (needs live MySQL)

🛑 LAYER BOUNDARY (runtime). Boot-time migrate is disabled — apply the index by hand.

- [ ] **Step 1: Apply the index** on each target DB:

```bash
mysql -h "$DB_HOST" -u "$DB_USER" -p "$DB_NAME" < migrations/028_ma_quizzes_progress_index.sql
```
Expected: `Query OK`. If `Duplicate key name`, an equivalent index already exists — safe to ignore.

- [ ] **Step 2: Run the server**

Run: `make run`
Expected: boots clean.

- [ ] **Step 3: Smoke the endpoint** (replace the Bearer token + profile_id with a session that owns quizzes):

```bash
curl -s -X POST http://localhost:8080/quizzes/analytics/progress \
  -H 'Content-Type: application/json' \
  -d '{"profile_id": 7001, "limit": 10, "metadata": {"authorization": "Bearer <jwt>", "language": "vi"}}'
```
Expected: `mstatus: 200`, `series` ordered oldest→newest with `sequence` 1..N, `score` on the 10-point scale, `summary.trend` set. A profile you do not own → `mstatus: 11018`. Empty history → `series: []`, `trend: "NO_DATA"`.

- [ ] **Step 4: Verify the query plan** (optional, confirms the index is used):

```sql
EXPLAIN SELECT q.quiz_id FROM ma_quizzes q
WHERE q.status='ACTIVE' AND q.deleted_dt IS NULL
  AND q.quiz_status='SUBMITTED' AND q.profile_id=7001
  AND q.score_percentage IS NOT NULL
ORDER BY q.modify_dt DESC, q.quiz_id DESC LIMIT 10;
```
Expected: `key: ix_quiz_profile_status_modify`.

---

## Self-review notes

- **Spec coverage:** contract (§3) → Tasks 5,8,9; business rules (§2) → Tasks 4,6,8; validation (§5,§8) → Task 7; query strategy + index (§7) → Tasks 1,4,6; helper reuse (§4) → Task 0; status codes (§6) → Task 2; edge cases (§8) → Tasks 6,7 tests. All covered.
- **Type consistency:** `ListProgressPoints`, `ProgressPoint`, `ProgressPointsParams`, `GetQuizProgressQuery`, `GetQuizProgressResult`, `QuizProgressReq/Res`, `QuizPoint`, `QuizProgressSummary`, `ValidateQuizProgress` names are used identically across Tasks 3–9.
- **Reuse:** `enum.DefaultProgressTz`, `enum.IsValidTzOffset`, `enum.ProgressComment*` (feature 002) and `mtime.ParseFromString` are reused, not re-created.
- **`errs.IsbinbaseError`** is the (legacy-named) MathError detector — see known-issues §1.
