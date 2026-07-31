# FEATURE-SPEC — Quiz Learning Progress (per-profile, quizzes only)

> Parent/student-facing endpoint showing ONE child's learning-progress
> over their standalone **quizzes** (`ma_quizzes`) — a per-quiz score
> line chart plus a small summary that drives the trend banner. Read-only,
> single endpoint, no commands, no UoW.

Aggregate: `quiz` (lives inside `internal/module/quiz/`).
Data source: `ma_quizzes` ONLY. Classroom exercises (`ma_exercises` /
`ma_exercise_submissions`) are deliberately **excluded** — they are a
separate aggregate and are surfaced by feature `002` instead.

UI reference: the "Tiến độ học tập" screen attached to the request — a
line chart "Điểm số theo bài đánh giá" with one point per completed quiz
(Bài 1…Bài 6, oldest→newest), a "Lọc thời gian" date-range dialog, and a
single banner ("Con đang tiến bộ tốt! — Điểm trung bình tăng 1.2 điểm so
với giai đoạn trước").

This is the **quiz** sibling of `002-profile-learning-progress` (which is
the **classroom-exercise** version of the same screen). The two endpoints
share wire conventions and reuse the same trend/score helpers so the
mobile client renders both screens with one shape.

---

## 1. Story

> "As a parent (or the student themself), I want to open my child's quiz
> progress and see:
> 1. A **line chart** with one point per completed quiz, in time order,
>    so I can watch how each round went.
> 2. By default the **10 most recent** completed quizzes; optionally I can
>    pick a **date range** and/or change how many points to show.
> 3. A **banner** that congratulates or nudges based on the trend, with
>    the average-score delta vs the previous period."

Only the child's **own** quizzes are shown; the acting session user must
own the target profile.

---

## 2. Decisions (signed off)

| # | Decision |
|---|---|
| 1 | One read-only endpoint returns the whole screen (chart series + summary). New route `POST /quizzes/analytics/progress`, auth-gated. |
| 2 | Chart x-axis is **per quiz** — one point per completed quiz, ordered by completion time ASC. `sequence` (1..N) is the "Bài N" label (quiz `title` is the AI grade-level string, not a "Bài N" label, so we number positionally). |
| 3 | "Completed quiz" = `quiz_status = 'SUBMITTED'`, `status = 'ACTIVE'`, `deleted_dt IS NULL`, `score_percentage IS NOT NULL`. |
| 4 | **All** quiz purposes counted by default; optional `purpose` request filter (`ASSESSMENT` / `PRACTICE` / `EXAM`) narrows it. |
| 5 | **`limit`** caps the number of chart points; default `10`, clamped `[1, 100]`. The endpoint returns the `limit` **most recent** completed quizzes (then reversed to ASC for display). `002` has no limit — this is the quiz-specific addition. |
| 6 | Completion time source = `ma_quizzes.modify_dt` of the SUBMITTED row (no schema change). Safe today: `UpdateAnswersAndGrading` is the only mutator and soft-deleted rows are filtered out, so `modify_dt` ≈ submission time. Documented fragility (see §8). |
| 7 | Score wire format matches `002`: `score` (10-point, 2 dp via `PctTo10Pt`) **and** `score_pct` (0–100 int). Client rounds for display (mockup shows integers). |
| 8 | "Giai đoạn trước" delta (`average_delta`) = current-window avg − prior-window avg, both on the 10-point scale. Prior window = the `limit` completed quizzes **immediately preceding** the current window (count-based, anchored at the current window's oldest point). `null` when the prior window has no data. This is count-based (not date-based like `002`) because the quiz screen is limit-driven. |
| 9 | Trend banner reuses `enum.ProgressComment` + `Classify(count, avg10, slope)`; `slope` = `LinearSlope` over the per-quiz 10-point series (pts-per-quiz, x = quiz index 0..n-1). |
| 10 | **No pass-rate / "tỷ lệ làm đúng" card** — the quiz mockup does not show it (that card is `002`-only). `PassScorePctThreshold` stays in the classroom package. |
| 11 | Date range: optional `from_dt`/`to_dt` (datetime strings) + `tz` (numeric offset, default `+07:00`). Filter compares `modify_dt` against the parsed bounds, same as `002`. |
| 12 | Ownership enforced: the target `profile_id` must belong to the session user, else `QUIZ_ANALYTICS_PROFILE_NOT_OWNED`. |
| 13 | Endpoint is unpaginated beyond `limit` (a single child's recent quizzes is small). |
| 14 | Repository read uses a **lightweight projection** (`quiz.ProgressPoint`) selecting only the needed columns — the LONGTEXT `questions`/`answers` blobs are NOT loaded. |

---

## 3. Endpoint

### `POST /quizzes/analytics/progress`

**Request**
```json
{
  "profile_id": 7001,            // target child; must belong to session user
  "purpose": null,               // null=all; ASSESSMENT|PRACTICE|EXAM
  "from_dt": "2026-05-01 00:00:00",  // optional
  "to_dt":   "2026-06-24 23:59:59",  // optional
  "limit": 10,                   // optional; default 10; clamp [1,100]
  "tz": "+07:00",                // optional; default +07:00
  "metadata": { ... }
}
```

`user_id` is taken from the session, never the body.

**Response**
```json
{
  "status": "Success", "mstatus": 200,
  "profile_id": 7001,
  "from_dt": "2026-05-01 00:00:00",
  "to_dt":   "2026-06-24 23:59:59",
  "tz": "+07:00",
  "purpose": null,
  "limit": 10,
  "series": [
    {
      "sequence": 1,
      "quiz_id": 900001,
      "purpose": "ASSESSMENT",
      "type_of_quiz": "GENERAL",
      "title": "Grade 1 - Level 1",
      "short_text": "Các số trong phạm vi 20",
      "completed_dt": "2026-05-05 09:41:00.000000",
      "score": 6.0, "score_pct": 60,
      "correct_number": 6, "total_questions": 10
    }
    // … one entry per completed quiz, completed_dt ASC (oldest→newest)
  ],
  "summary": {
    "count":              6,
    "average_score":      7.70, "average_score_pct": 77.0,
    "average_delta":      1.20,                 // null when no prior-period data
    "highest_score":      10.0, "highest_score_pct": 100,
    "highest_quiz_id":    900006,
    "lowest_score":       6.0,
    "trend":              "GOOD_PROGRESS"       // enum.ProgressComment
  }
}
```

Empty result → `series: []`, `summary.count = 0`, all nullable score
fields `null`, `trend: "NO_DATA"`. Still `mstatus: 200` (empty state, not
an error).

### Field → UI mapping

| Screen element | Response field |
|---|---|
| Line chart points (Bài N, ngày, điểm) | `series[].sequence` / `completed_dt` / `score` |
| "6 bài đánh giá" | `summary.count` |
| Banner label ("Con đang tiến bộ tốt!") | `summary.trend` (client maps enum → copy) |
| "Điểm trung bình tăng 1.2 điểm so với giai đoạn trước" | `summary.average_delta` (+ `average_score`) |
| Point tooltip (topic) | `series[].title` / `short_text` |

---

## 4. Trend enum + helpers (reused, promoted to shared)

`enum.ProgressComment` (already in `internal/shared/enum/classroom_progress.go`)
is reused verbatim. First-match-wins rules (identical to `002` §4):

| Symbol | UI label (VN) | Rule |
|---|---|---|
| `NO_DATA` | (no chip) | `count == 0` |
| `INSUFFICIENT` | (no chip) | `count == 1` |
| `NEED_TO_TRY` | Cần cố gắng | `avg10 < 5.0` OR `slope ≤ -0.05` |
| `GOOD_PROGRESS` | Tiến bộ tốt | `avg10 ≥ 7.5` AND `slope ≥ 0` |
| `PROGRESS` | Tiến bộ | everything else with `count ≥ 2` |

`slope` unit = 10-point pts per quiz (x = quiz index), matching `002`'s
per-item convention.

**Promotion (Phase 0):** `PctTo10Pt`, `LinearSlope`, and `Classify`
currently live in `internal/application/query/classroomprogress/trend.go`.
They are generic (percentage→10-point rounding, least-squares slope, enum
classification) and now have two consumers, so they move to a neutral
shared package:

- New package `internal/application/query/progress` (package `progress`)
  holding `PctTo10Pt`, `LinearSlope`, `Classify`.
- `classroomprogress` updates its call sites to `progress.PctTo10Pt` /
  `progress.LinearSlope` / `progress.Classify`. `PassScorePctThreshold`
  stays in `classroomprogress` (quiz has no pass-rate).
- The quiz query imports the same `progress` package.

This is a small mechanical refactor; `002`'s behavior is unchanged
(verified by `go build` + `go vet` + its existing tests).

---

## 5. Permission rule

`profile_id` is the target child. The acting user is the session `uid`.

```
uid  := sessionUID(r)                       // else UNAUTHORIZED
prof := profileRepo.FindByProfileId(profile_id)
if prof == nil:            -> PROFILE_NOT_FOUND
if prof.UserId() != uid:   -> QUIZ_ANALYTICS_PROFILE_NOT_OWNED
```

The quiz `Service` already holds `profileRepo`, so no new dependency. This
mirrors `home`'s `HOME_PROFILE_NOT_OWNED` posture (session user may only
view their own child's data).

---

## 6. Status codes

Extend the QUIZ block (11000–11099; last live code is
`QUIZ_ANSWER_SCHEMA_INVALID = 11016`). Lockstep across `code.go`,
`message_en.go`, `message_vn.go`:

| Code | Symbol | EN | VN |
|---|---|---|---|
| 11017 | `QUIZ_ANALYTICS_MISSING_PROFILE` | "Profile is required." | "Hồ sơ học sinh là bắt buộc." |
| 11018 | `QUIZ_ANALYTICS_PROFILE_NOT_OWNED` | "This profile does not belong to you." | "Hồ sơ này không thuộc về bạn." |
| 11019 | `QUIZ_ANALYTICS_INVALID_DATE_RANGE` | "Invalid date range." | "Khoảng thời gian không hợp lệ." |
| 11020 | `QUIZ_ANALYTICS_INVALID_TZ` | "Invalid time zone." | "Múi giờ không hợp lệ." |
| 11021 | `QUIZ_ANALYTICS_INVALID_PURPOSE` | "Invalid quiz type." | "Loại quiz không hợp lệ." |

Re-used: `UNAUTHORIZED`, `PROFILE_NOT_FOUND`, `FAIL`.

`limit` out of range is **clamped silently** (no error code) — matches the
pagination-clamp convention.

---

## 7. Data source & query strategy

Table: `ma_quizzes` (domain `quiz.Quiz`, repo `quiz.IRepository`).

New repository method returning a **lightweight projection** (no LONGTEXT):

```go
// domain/quiz/repository.go
type ProgressPoint struct {
    QuizId          int64
    Purpose         string
    TypeOfQuiz      *string
    Title           *string
    ShortText       *string
    ScorePercentage int64
    CorrectNumber   *int64
    TotalQuestions  *int64
    CompletedDt     mtime.MathTime   // = modify_dt of the SUBMITTED row
}

type ProgressPointsParams struct {
    ProfileID       int64
    Purpose         *string          // nil => all
    From            *mtime.MathTime  // nil => open lower bound
    To              *mtime.MathTime  // nil => open upper bound
    CompletedBefore *mtime.MathTime  // nil => none; used for the prior window anchor
    Limit           int64            // already clamped by the caller
}

ListProgressPoints(ctx, params ProgressPointsParams) ([]*ProgressPoint, error)
```

SQL (rows ordered newest-first; the query layer reverses to ASC):

```sql
SELECT q.quiz_id, q.purpose, q.type_of_quiz, q.title, q.short_text,
       q.score_percentage, q.correct_number, q.total_questions, q.modify_dt
FROM ma_quizzes q
WHERE q.status = 'ACTIVE' AND q.deleted_dt IS NULL
  AND q.quiz_status = 'SUBMITTED'
  AND q.profile_id = ?
  AND q.score_percentage IS NOT NULL
  [AND q.purpose = ?]
  [AND q.modify_dt >= ?]     -- From
  [AND q.modify_dt <= ?]     -- To
  [AND q.modify_dt <  ?]     -- CompletedBefore (prior window)
ORDER BY q.modify_dt DESC, q.quiz_id DESC
LIMIT ?
```

Query-layer algorithm (`application/query/quiz/get_quiz_progress_query.go`):

1. **Current window:** `ListProgressPoints` with `From/To` (if any), `Limit`.
   Reverse to ASC → `series`; assign `sequence` 1..N.
2. **Prior window:** if the current window is non-empty, call again with
   `CompletedBefore = oldest(current).CompletedDt`, same `Limit`, same
   `Purpose` (no `From/To`). `priorAvg10 = avg10(priorPoints)` or `nil`.
3. **Summary:** count, `average_score`/`average_score_pct`,
   `highest_*` (max by pct; ties → most recent), `lowest_score`,
   `average_delta = avg10(current) − priorAvg10` (nil-guarded), and
   `trend = Classify(count, avg10, LinearSlope(scores10))`.

Recommended index (migration, forward-only, `ask`-gated, applied manually
per known-issues §5):

```sql
-- migration up
ALTER TABLE ma_quizzes
  ADD KEY ix_quiz_profile_status_modify (profile_id, quiz_status, modify_dt);
```

Equality on `profile_id, quiz_status` + range/sort on `modify_dt`;
`status` / `deleted_dt` / `score_percentage` / `purpose` are residual
filters. (There is a commented-out `idx_profile_status` stub in migration
009 — this new index supersedes it for the analytics access path.)

Purpose filtering is applied **in SQL** (single low-cardinality equality),
unlike `002` which filters in Go after a JOIN — the quiz table needs no
hydration JOIN, so pushing it down is simpler here.

---

## 8. Edge cases

| Case | Handling |
|---|---|
| No completed quizzes (or none in range) | `series: []`, `summary.count = 0`, nullable fields null, `trend: NO_DATA`, `mstatus 200` |
| Exactly one point | `trend: INSUFFICIENT`; `average_delta` null unless a prior window exists |
| Same `modify_dt` on multiple rows | Deterministic tie-break `quiz_id DESC` in ORDER BY |
| `score_percentage` NULL (legacy / failed grading) | Excluded by `score_percentage IS NOT NULL` |
| GENERATED / DELETED quizzes | Excluded by `quiz_status = 'SUBMITTED'` + active predicate |
| Anonymous quizzes (`profile_id` NULL) | Naturally excluded (filter is `profile_id = ?`) |
| `limit` > 100 or < 1 | Clamped to `[1,100]`, default 10 |
| `from_dt` > `to_dt` (both set) | `QUIZ_ANALYTICS_INVALID_DATE_RANGE` |
| Span > 2 years | `QUIZ_ANALYTICS_INVALID_DATE_RANGE` (mirror `002` §11) |
| Bad `tz` | `QUIZ_ANALYTICS_INVALID_TZ`; empty tz → default `+07:00` |
| Invalid `purpose` | `QUIZ_ANALYTICS_INVALID_PURPOSE` |
| Profile not owned / not found | `QUIZ_ANALYTICS_PROFILE_NOT_OWNED` / `PROFILE_NOT_FOUND` |
| **`modify_dt` fragility** | Today `modify_dt` ≈ submission time (only mutator is submit; soft-deletes filtered out). If a future non-submit mutator is added to `ma_quizzes`, introduce a dedicated `completed_dt` column and switch the completion-time source — the projection isolates this to one repo method. |

---

## 9. Impact map (files)

| Layer | Change |
|---|---|
| **Shared (Phase 0)** | New `internal/application/query/progress/trend.go` (`PctTo10Pt`, `LinearSlope`, `Classify`); update `classroomprogress` call sites. |
| **DB** | New migration: `ix_quiz_profile_status_modify` index. No column change. |
| **Domain** | `quiz.ProgressPoint` + `quiz.ProgressPointsParams` + `ListProgressPoints` on `quiz.IRepository`. |
| **Repo** | `QuizRepository.ListProgressPoints` (narrow SELECT) + projection scan. |
| **Query** | `application/query/quiz/get_quiz_progress_query.go` (+ summary aggregation). |
| **DTO** | `application/dto/quiz/quiz_progress_dto.go` (`QuizProgressReq`/`Res`, `QuizPoint`, `QuizProgressSummary`, mappers). |
| **Status** | 5 new `QUIZ_ANALYTICS_*` codes (lockstep 3 files). |
| **Module** | `Service.GetQuizProgress` + validator + `QuizHandler.HandleGetQuizProgress`. |
| **Route** | `reg("POST /quizzes/analytics/progress", ..., authMiddleware)`. |
| **Container** | None — quiz `Service` already receives `quizRepo` + `profileRepo`. Wire the new query handler inside `NewService`. |
| **Tests** | Repo test (real MySQL), query aggregation test (fake repo), validator test, `progress` helper tests (move existing coverage if any). |
| **Docs** | This spec + `IMPLEMENTATION-PLAN.md`. |

No `transaction.Repositories` change (pure read, no UoW, no `seq`).

---

## 10. Out of scope (v1)

- No pre-aggregation / materialized stats; query hits `ma_quizzes` directly.
- No background job.
- No CSV/export (the screen's top-right ↓ icon is a client concern).
- No per-question breakdown — `correct_number` / `total_questions` are
  surfaced raw per point only.
- No `completed_dt` schema column (using `modify_dt`; see §8 for the
  upgrade path).
- No cross-child / user-level aggregation (per-profile only).
- No classroom-exercise data (that is feature `002`).
