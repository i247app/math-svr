package classroomprogress

// import (
// 	"context"
// 	"fmt"
// 	"sort"
// 	"time"

// 	dto "math-ai.com/math-ai/internal/application/dto/classroomprogress"
// 	"math-ai.com/math-ai/internal/domain/classroom"
// 	exercise "math-ai.com/math-ai/internal/domain/exercise"
// 	profileDomain "math-ai.com/math-ai/internal/domain/profile"
// 	mtime "math-ai.com/math-ai/internal/domain/shared/time"
// 	"math-ai.com/math-ai/internal/shared/enum"
// 	"math-ai.com/math-ai/internal/shared/pagination"
// )

// // StudentsProgressQuery is the read input for the per-student endpoint.
// type StudentsProgressQuery struct {
// 	ClassroomID     int64
// 	Bucket          string
// 	From            mtime.MathTime
// 	To              mtime.MathTime
// 	TzOffset        string
// 	FilterProfileID *int64
// 	Page            int64
// 	Size            int64
// }

// // StudentsProgressQueryHandler owns the per-student aggregation +
// // prior-period delta pass. It composes four repos:
// //
// //   - memberRepo     : the STUDENT roster (denominator + the
// //     always-include-everyone list — decision #18)
// //   - profileRepo    : name + avatar_key hydration
// //   - submissionRepo : the score series (current + prior)
// //   - exerciseRepo   : exercise title lookup for trend_series_by_exercise
// //     (2026-06-10 refinement) — one batch call per request
// //
// // Avatar URL hydration intentionally does NOT live here — it's I/O
// // against the storage adapter, which is a module-layer concern.
// type StudentsProgressQueryHandler struct {
// 	memberRepo     classroom.IMemberRepository
// 	profileRepo    profileDomain.IRepository
// 	submissionRepo exercise.ISubmissionRepository
// 	exerciseRepo   exercise.IRepository
// }

// func NewStudentsProgressQueryHandler(
// 	memberRepo classroom.IMemberRepository,
// 	profileRepo profileDomain.IRepository,
// 	submissionRepo exercise.ISubmissionRepository,
// 	exerciseRepo exercise.IRepository,
// ) *StudentsProgressQueryHandler {
// 	return &StudentsProgressQueryHandler{
// 		memberRepo:     memberRepo,
// 		profileRepo:    profileRepo,
// 		submissionRepo: submissionRepo,
// 		exerciseRepo:   exerciseRepo,
// 	}
// }

// // studentResult is the per-student intermediate carried between the
// // computation pass and the summary/sort/paginate steps. The named type
// // (rather than an anonymous struct) lets buildSummary accept the slice
// // without an identity mismatch — []NamedType is not assignable to a
// // []struct{...} parameter even when the shape is identical.
// type studentResult struct {
// 	row            dto.StudentProgressRow
// 	currentComment enum.ProgressComment
// 	priorComment   enum.ProgressComment
// }

// // commentSortRank orders the per-student list so the teacher sees
// // problems first, then unstarted students, then the people doing fine.
// // Stable rank → tie-break by name in Handle.
// func commentSortRank(c enum.ProgressComment) int {
// 	switch c {
// 	case enum.ProgressCommentNeedToTry:
// 		return 0
// 	case enum.ProgressCommentNoData:
// 		return 1
// 	case enum.ProgressCommentInsufficient:
// 		return 2
// 	case enum.ProgressCommentProgress:
// 		return 3
// 	case enum.ProgressCommentGoodProgress:
// 		return 4
// 	default:
// 		return 5
// 	}
// }

// // Handle runs the full pipeline:
// //
// //  1. Load STUDENT members of the classroom (denominator).
// //  2. Resolve profile rows for name + avatar_key.
// //  3. Fetch current-period submissions; group by profile_id.
// //  4. Fetch prior-period submissions; group by profile_id.
// //  5. For every STUDENT, compute aggregates, trend_series, slope,
// //     current comment, and prior comment.
// //  6. Compute class summary + deltas.
// //  7. Apply filter_profile_id to the student slice.
// //  8. Sort by (commentRank, name).
// //  9. Paginate in-memory.
// func (h *StudentsProgressQueryHandler) Handle(ctx context.Context, q StudentsProgressQuery) (*dto.StudentsProgressRes, error) {
// 	loc, err := ParseTzOffset(q.TzOffset)
// 	if err != nil {
// 		return nil, fmt.Errorf("students-progress: %w", err)
// 	}

// 	// 1. STUDENT-role active members of the classroom.
// 	role := string(enum.ClassroomMemberRoleTypeStudent)
// 	memberStatus := string(enum.ClassroomMemberStatusTypeActive)
// 	classroomID := q.ClassroomID
// 	members, _, err := h.memberRepo.ListMembers(ctx, &classroom.ListMembersParams{
// 		ClassroomId: &classroomID,
// 		Role:        &role,
// 		Status:      &memberStatus,
// 		TakeAll:     true,
// 	})
// 	if err != nil {
// 		return nil, err
// 	}

// 	// 2. Profile hydration in one round trip.
// 	profileIDs := make([]int64, 0, len(members))
// 	for _, m := range members {
// 		profileIDs = append(profileIDs, m.ProfileId())
// 	}
// 	profiles, err := h.profileRepo.ListByProfileIds(ctx, profileIDs)
// 	if err != nil {
// 		return nil, err
// 	}
// 	profileByID := make(map[int64]*profileDomain.Profile, len(profiles))
// 	for _, p := range profiles {
// 		profileByID[p.ProfileId()] = p
// 	}

// 	// 3+4. Current and prior submission rows. Prior window is the same
// 	// length immediately before From — decision #4.
// 	currentRows, err := h.submissionRepo.ListForProgress(ctx, exercise.ProgressRangeParams{
// 		ClassroomID:     q.ClassroomID,
// 		From:            q.From,
// 		To:              q.To,
// 		FilterProfileID: nil, // load all so the summary stays class-level
// 	})
// 	if err != nil {
// 		return nil, err
// 	}
// 	priorFrom, priorTo := priorWindow(q.From.Time, q.To.Time)
// 	priorRows, err := h.submissionRepo.ListForProgress(ctx, exercise.ProgressRangeParams{
// 		ClassroomID: q.ClassroomID,
// 		From:        mtime.MathTime{Time: priorFrom},
// 		To:          mtime.MathTime{Time: priorTo},
// 	})
// 	if err != nil {
// 		return nil, err
// 	}
// 	currentByProfile := groupRowsByProfile(currentRows)
// 	priorByProfile := groupRowsByProfile(priorRows)

// 	// 4b. Exercise title hydration. Walk every student's last-5
// 	// submissions, collect the set of classroom_exercise_ids that will
// 	// land in trend_series_by_exercise, batch-fetch their titles.
// 	// Missing ids fall back to "#<id>" in buildStudentRow.
// 	exerciseTitles, err := h.loadExerciseTitlesForLast5(ctx, currentByProfile, members)
// 	if err != nil {
// 		return nil, err
// 	}

// 	// 5. Per-student computation. Build the full list first, then
// 	// apply filter + sort + paginate.
// 	results := make([]studentResult, 0, len(members))
// 	for _, m := range members {
// 		pid := m.ProfileId()
// 		p := profileByID[pid]

// 		curRows := currentByProfile[pid]
// 		row := buildStudentRow(p, pid, curRows, loc, q.To.Time, exerciseTitles)

// 		// Current comment
// 		curWindow := BuildSlopeWindow(curRows, loc, q.To.Time)
// 		curSlope := LinearSlope(curWindow)
// 		var avg10 float64
// 		if row.AvgScore != nil {
// 			avg10 = *row.AvgScore
// 		}
// 		curComment := Classify(int(row.SubmissionCount), avg10, curSlope)
// 		row.Comment = curComment.String()
// 		if int(row.SubmissionCount) >= 2 {
// 			s := curSlope
// 			row.Slope = &s
// 		}

// 		// Prior comment (slope window relative to priorTo)
// 		priorRowsForP := priorByProfile[pid]
// 		priorWindowVec := BuildSlopeWindow(priorRowsForP, loc, priorTo)
// 		priorSlope := LinearSlope(priorWindowVec)
// 		priorAvg10 := avgFromRows10pt(priorRowsForP)
// 		priorComment := Classify(len(priorRowsForP), priorAvg10, priorSlope)

// 		results = append(results, studentResult{
// 			row:            row,
// 			currentComment: curComment,
// 			priorComment:   priorComment,
// 		})
// 	}

// 	// 6. Class summary.
// 	summary := buildSummary(results)

// 	// 7. Apply filter_profile_id.
// 	if q.FilterProfileID != nil {
// 		filtered := results[:0]
// 		for _, r := range results {
// 			if r.row.ProfileID == *q.FilterProfileID {
// 				filtered = append(filtered, r)
// 			}
// 		}
// 		results = filtered
// 	}

// 	// 8. Sort.
// 	sort.SliceStable(results, func(i, j int) bool {
// 		ri := commentSortRank(results[i].currentComment)
// 		rj := commentSortRank(results[j].currentComment)
// 		if ri != rj {
// 			return ri < rj
// 		}
// 		return results[i].row.ProfileName < results[j].row.ProfileName
// 	})

// 	// 9. Paginate. Pagination.NewPagination clamps size to [1, 1000]
// 	// and defaults page=1 when total=0, so we mirror that in the slice
// 	// arithmetic.
// 	total := int64(len(results))
// 	pg := pagination.NewPagination(q.Page, q.Size, total)
// 	skip := min(max(pg.Skip, 0), total)
// 	end := min(skip+pg.Size, total)
// 	pageResults := results[skip:end]
// 	students := make([]dto.StudentProgressRow, 0, len(pageResults))
// 	for _, r := range pageResults {
// 		students = append(students, r.row)
// 	}

// 	return &dto.StudentsProgressRes{
// 		Bucket:        q.Bucket,
// 		FromDt:        q.From.String(),
// 		ToDt:          q.To.String(),
// 		Tz:            q.TzOffset,
// 		CommentSource: string(enum.TrendSourceLast5Days),
// 		Summary:       summary,
// 		Students:      students,
// 		Pagination:    pg,
// 	}, nil
// }

// // loadExerciseTitlesForLast5 walks every student's last-5 submissions,
// // collects the distinct classroom_exercise_id set, and batch-fetches
// // the title via exercise.IRepository. One round trip regardless of
// // classroom size. Returns an empty map (not nil) when the classroom
// // has no submissions in range, so callers can dereference without a
// // nil check.
// func (h *StudentsProgressQueryHandler) loadExerciseTitlesForLast5(
// 	ctx context.Context,
// 	currentByProfile map[int64][]*exercise.ProgressRow,
// 	members []*classroom.Member,
// ) (map[int64]string, error) {
// 	idSet := make(map[int64]struct{}, len(members)*5)
// 	for _, m := range members {
// 		curRows := currentByProfile[m.ProfileId()]
// 		for _, r := range LastNExerciseRows(curRows, 5) {
// 			idSet[r.ClassroomExerciseID] = struct{}{}
// 		}
// 	}
// 	if len(idSet) == 0 {
// 		return map[int64]string{}, nil
// 	}
// 	ids := make([]int64, 0, len(idSet))
// 	for id := range idSet {
// 		ids = append(ids, id)
// 	}
// 	exs, err := h.exerciseRepo.ListByClassroomExerciseIds(ctx, ids)
// 	if err != nil {
// 		return nil, err
// 	}
// 	out := make(map[int64]string, len(exs))
// 	for _, ex := range exs {
// 		out[ex.ClassroomExerciseId()] = ex.Title()
// 	}
// 	return out, nil
// }

// // --- private helpers ---

// // priorWindow returns the immediately-prior same-length window for
// // delta accounting. The split is 1µs-inclusive (priorTo = from - 1µs)
// // so the prior range and the current range don't overlap a single
// // instant.
// func priorWindow(from, to time.Time) (priorFrom, priorTo time.Time) {
// 	length := to.Sub(from)
// 	priorTo = from.Add(-time.Microsecond)
// 	priorFrom = priorTo.Add(-length)
// 	return
// }

// func groupRowsByProfile(rows []*exercise.ProgressRow) map[int64][]*exercise.ProgressRow {
// 	out := make(map[int64][]*exercise.ProgressRow, 16)
// 	for _, r := range rows {
// 		out[r.ProfileID] = append(out[r.ProfileID], r)
// 	}
// 	return out
// }

// // buildStudentRow handles the DTO scalar fields (counts, avg/min/max,
// // first/last submitted, both sparkline series). Comment + Slope are
// // filled by the caller because they need the slope-window vector.
// //
// // trend_series_by_day is always daily density (the 2026-06-10
// // refinement removed bucket-aware sparklines — chart `bucket` no
// // longer affects the per-student row).
// //
// // trend_series_by_exercise is the last 5 submissions, each as one
// // point. BucketLabel = exercise title from exerciseTitles, falling
// // back to "#<id>" so the chart still renders if a title row went
// // missing.
// func buildStudentRow(
// 	p *profileDomain.Profile,
// 	profileID int64,
// 	rows []*exercise.ProgressRow,
// 	loc *time.Location,
// 	_ time.Time,
// 	exerciseTitles map[int64]string,
// ) dto.StudentProgressRow {
// 	row := dto.StudentProgressRow{
// 		ProfileID:             profileID,
// 		TrendSeriesByDay:      []dto.TrendPoint{},
// 		TrendSeriesByExercise: []dto.ExerciseTrendPoint{},
// 	}
// 	if p != nil {
// 		row.ProfileName = p.Name()
// 		row.AvatarKey = p.AvatarKey()
// 	}
// 	row.SubmissionCount = int64(len(rows))

// 	if len(rows) > 0 {
// 		var sum, minS, maxS int64
// 		minS = rows[0].ScorePercentage
// 		maxS = rows[0].ScorePercentage
// 		for i, r := range rows {
// 			sum += r.ScorePercentage
// 			if r.ScorePercentage < minS {
// 				minS = r.ScorePercentage
// 			}
// 			if r.ScorePercentage > maxS {
// 				maxS = r.ScorePercentage
// 			}
// 			if i == 0 {
// 				row.FirstSubmittedDt = r.SubmittedDt.String()
// 			}
// 			row.LastSubmittedDt = r.SubmittedDt.String()
// 		}
// 		avgPct := float64(sum) / float64(len(rows))
// 		avg10 := PctTo10Pt(avgPct)
// 		max10 := PctTo10Pt(float64(maxS))
// 		min10 := PctTo10Pt(float64(minS))
// 		maxPct := float64(maxS)
// 		minPct := float64(minS)

// 		row.AvgScore = &avg10
// 		row.HighestScore = &max10
// 		row.LowestScore = &min10
// 		row.AvgScorePct = &avgPct
// 		row.HighestScorePct = &maxPct
// 		row.LowestScorePct = &minPct

// 		// trend_series_by_day — always daily density, last 5 buckets.
// 		daySeries := BuildTrendSeries(rows, string(enum.ProgressBucketDay), loc, 5)
// 		dayPoints := make([]dto.TrendPoint, 0, len(daySeries))
// 		for _, s := range daySeries {
// 			avg := PctTo10Pt(s.AvgPct)
// 			pct := s.AvgPct
// 			dayPoints = append(dayPoints, dto.TrendPoint{
// 				BucketLabel: s.Label,
// 				AvgScore:    &avg,
// 				AvgScorePct: &pct,
// 			})
// 		}
// 		row.TrendSeriesByDay = dayPoints

// 		// trend_series_by_exercise — last 5 submissions, one point each.
// 		last5 := LastNExerciseRows(rows, 5)
// 		exPoints := make([]dto.ExerciseTrendPoint, 0, len(last5))
// 		for _, r := range last5 {
// 			pct := float64(r.ScorePercentage)
// 			avg := PctTo10Pt(pct)
// 			label, ok := exerciseTitles[r.ClassroomExerciseID]
// 			if !ok || label == "" {
// 				label = fmt.Sprintf("#%d", r.ClassroomExerciseID)
// 			}
// 			exPoints = append(exPoints, dto.ExerciseTrendPoint{
// 				BucketLabel: label,
// 				ExerciseID:  r.ClassroomExerciseID,
// 				AvgScore:    &avg,
// 				AvgScorePct: &pct,
// 			})
// 		}
// 		row.TrendSeriesByExercise = exPoints
// 	}

// 	return row
// }

// func avgFromRows10pt(rows []*exercise.ProgressRow) float64 {
// 	if len(rows) == 0 {
// 		return 0
// 	}
// 	var sum int64
// 	for _, r := range rows {
// 		sum += r.ScorePercentage
// 	}
// 	return PctTo10Pt(float64(sum) / float64(len(rows)))
// }

// // buildSummary collapses the per-student results into the three-card
// // summary block. Improving = PROGRESS + GOOD_PROGRESS (decision #5);
// // need-support = NEED_TO_TRY only (decision #6). Deltas are vs prior
// // period (decision #4), nil when prior had zero matches.
// func buildSummary(results []studentResult) dto.ProgressSummary {
// 	var (
// 		total          = int64(len(results))
// 		participating  int64
// 		improvingCur   int64
// 		improvingPrior int64
// 		supportCur     int64
// 		supportPrior   int64
// 	)
// 	for _, r := range results {
// 		if r.row.SubmissionCount > 0 {
// 			participating++
// 		}
// 		if isImproving(r.currentComment) {
// 			improvingCur++
// 		}
// 		if isImproving(r.priorComment) {
// 			improvingPrior++
// 		}
// 		if r.currentComment == enum.ProgressCommentNeedToTry {
// 			supportCur++
// 		}
// 		if r.priorComment == enum.ProgressCommentNeedToTry {
// 			supportPrior++
// 		}
// 	}
// 	var rate float64
// 	if total > 0 {
// 		rate = float64(participating) / float64(total)
// 	}
// 	return dto.ProgressSummary{
// 		TotalStudents:       total,
// 		ParticipatingCount:  participating,
// 		ParticipationRate:   rate,
// 		ImprovingCount:      improvingCur,
// 		ImprovingDeltaPct:   deltaPct(improvingCur, improvingPrior),
// 		NeedSupportCount:    supportCur,
// 		NeedSupportDeltaPct: deltaPct(supportCur, supportPrior),
// 	}
// }

// func isImproving(c enum.ProgressComment) bool {
// 	return c == enum.ProgressCommentProgress || c == enum.ProgressCommentGoodProgress
// }

// // deltaPct returns (cur - prior) / prior. Nil when prior is 0 — the
// // "from zero" delta is undefined, and the UI's ↑/↓ arrows would be
// // misleading.
// func deltaPct(cur, prior int64) *float64 {
// 	if prior == 0 {
// 		return nil
// 	}
// 	d := float64(cur-prior) / float64(prior)
// 	return &d
// }
