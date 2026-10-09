package exam

import (
	"context"
	"slices"
	"strings"
	"time"

	dto "math-ai.com/math-ai/internal/application/dto/exam"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/mtime"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/shared/enum"
)

// DefaultNumQuestions is the size of every exam. It is fixed, not a
// default: the per-grade system prompt in ma_exam_prompts is written for
// exactly 10 questions (probes at Q3 and Q6 included) and fills nothing in
// at runtime, so a request asking for another count is served 10. The
// num_questions request field is kept so older clients still decode, and
// ignored.
const DefaultNumQuestions = 10

// ValidatedGenerate carries the normalised values the service needs, so
// nothing downstream has to re-parse the request's strings.
type ValidatedGenerate struct {
	ExamType enum.ExamType
}

func ValidateGenerateExam(ctx context.Context, req *dto.GenerateExamReq) (ValidatedGenerate, error) {
	if req.ProfileID <= 0 {
		return ValidatedGenerate{}, errs.NewError(ctx, status.EXAM_MISSING_PROFILE_ID, nil, ErrProfileIDRequired)
	}

	raw := strings.ToUpper(strings.TrimSpace(req.ExamType))
	if raw == "" {
		return ValidatedGenerate{}, errs.NewError(ctx, status.EXAM_MISSING_EXAM_TYPE, nil, ErrExamTypeRequired)
	}
	examType := enum.ExamType(raw)
	if !examType.IsValid() {
		return ValidatedGenerate{}, errs.NewError(ctx, status.EXAM_INVALID_EXAM_TYPE, nil, ErrExamTypeInvalid)
	}
	req.ExamType = raw

	// A PRACTICE round lives inside a journey and is aimed at that
	// journey's latest sitting, so the journey must be named.
	if examType == enum.ExamTypePractice && (req.EsessID == nil || *req.EsessID <= 0) {
		return ValidatedGenerate{}, errs.NewError(ctx, status.EXAM_MISSING_JOURNEY_ID, nil, ErrPracticeJourneyRequired)
	}

	// Grade is optional; when pinned it must name a band the product
	// actually serves. Nothing may be placed at the probe ceiling.
	if req.Grade != nil && (*req.Grade < enum.ExamGradeMin || *req.Grade > enum.ExamGradeMax) {
		return ValidatedGenerate{}, errs.NewError(ctx, status.EXAM_INVALID_GRADE, nil, ErrGradeOutOfRange)
	}
	// Level is optional and recorded as stated, so the only check is the
	// scale itself.
	if req.Level != nil && (*req.Level < enum.ExamLevelMin || *req.Level > enum.ExamLevelMax) {
		return ValidatedGenerate{}, errs.NewError(ctx, status.EXAM_INVALID_LEVEL, nil, ErrLevelOutOfRange)
	}

	req.NumQuestions = DefaultNumQuestions

	return ValidatedGenerate{ExamType: examType}, nil
}

// ValidateGenerateExamPool checks a pool generation. Same rules as a
// hand-out's placement fields, except that grade is required (there is
// no child to fall back on) and PRACTICE is refused.
func ValidateGenerateExamPool(ctx context.Context, req *dto.GenerateExamPoolReq) (enum.ExamType, error) {
	raw := strings.ToUpper(strings.TrimSpace(req.ExamType))
	if raw == "" {
		return "", errs.NewError(ctx, status.EXAM_MISSING_EXAM_TYPE, nil, ErrExamTypeRequired)
	}
	examType := enum.ExamType(raw)
	if !examType.IsValid() {
		return "", errs.NewError(ctx, status.EXAM_INVALID_EXAM_TYPE, nil, ErrExamTypeInvalid)
	}
	if examType == enum.ExamTypePractice {
		return "", errs.NewError(ctx, status.EXAM_INVALID_EXAM_TYPE, nil, ErrPracticeNotPooled)
	}
	if req.Grade == nil {
		return "", errs.NewError(ctx, status.EXAM_MISSING_GRADE, nil, ErrGradeRequired)
	}
	if *req.Grade < enum.ExamGradeMin || *req.Grade > enum.ExamGradeMax {
		return "", errs.NewError(ctx, status.EXAM_INVALID_GRADE, nil, ErrGradeOutOfRange)
	}
	if req.Level != nil && (*req.Level < enum.ExamLevelMin || *req.Level > enum.ExamLevelMax) {
		return "", errs.NewError(ctx, status.EXAM_INVALID_LEVEL, nil, ErrLevelOutOfRange)
	}
	req.NumQuestions = DefaultNumQuestions
	return examType, nil
}

func ValidateListExamPools(ctx context.Context, req *dto.ListExamPoolsReq) error {
	var examTypes []string
	for _, raw := range req.ExamTypes {
		examType, err := normalizeExamType(ctx, &raw)
		if err != nil {
			return err
		}
		if examType != nil && !slices.Contains(examTypes, *examType) {
			examTypes = append(examTypes, *examType)
		}
	}
	req.ExamTypes = examTypes
	if req.Grade != nil && (*req.Grade < enum.ExamGradeMin || *req.Grade > enum.ExamGradeMax) {
		return errs.NewError(ctx, status.EXAM_INVALID_GRADE, nil, ErrGradeOutOfRange)
	}
	return nil
}

func ValidateMarkExamPoolVerify(ctx context.Context, req *dto.MarkExamPoolVerifyReq) error {
	if req.ExamID <= 0 {
		return errs.NewError(ctx, status.EXAM_MISSING_EXAM_ID, nil, ErrExamIDRequired)
	}
	if req.IsVerify == nil {
		return errs.NewError(ctx, status.EXAM_MISSING_IS_VERIFY, nil, ErrIsVerifyRequired)
	}
	return nil
}

// ValidateVerifyExamPool checks what needs no stored row; the shape
// against the stored set is checked in the command.
func ValidateVerifyExamPool(ctx context.Context, req *dto.VerifyExamPoolReq) error {
	if req.ExamID <= 0 {
		return errs.NewError(ctx, status.EXAM_MISSING_EXAM_ID, nil, ErrExamIDRequired)
	}
	if len(req.Questions) == 0 {
		return errs.NewError(ctx, status.EXAM_POOL_INVALID_QUESTIONS, nil, ErrQuestionsRequired)
	}
	return nil
}

func ValidateGetExamPool(ctx context.Context, req *dto.GetExamPoolReq) error {
	if req.ExamID <= 0 {
		return errs.NewError(ctx, status.EXAM_MISSING_EXAM_ID, nil, ErrExamIDRequired)
	}
	return nil
}

func ValidateSubmitExam(ctx context.Context, req *dto.SubmitExamReq) error {
	if req.ProfileID <= 0 {
		return errs.NewError(ctx, status.EXAM_MISSING_PROFILE_ID, nil, ErrProfileIDRequired)
	}
	if req.ElinkID <= 0 {
		return errs.NewError(ctx, status.EXAM_MISSING_ATTEMPT_ID, nil, ErrAttemptIDRequired)
	}
	if len(req.Answers) == 0 {
		return errs.NewError(ctx, status.EXAM_MISSING_ANSWERS, nil, ErrAnswersRequired)
	}

	// Duplicates are caught here rather than in the scorer so the client
	// gets a field-level error instead of a grading failure.
	seen := make(map[int]struct{}, len(req.Answers))
	for i, a := range req.Answers {
		if a.QuestionNumber <= 0 {
			return errs.NewError(ctx, status.EXAM_INVALID_ANSWERS,
				map[string]any{"index": i}, ErrQuestionNumberInvalid)
		}
		if strings.TrimSpace(a.Label) == "" {
			return errs.NewError(ctx, status.EXAM_INVALID_ANSWERS,
				map[string]any{"index": i}, ErrAnswerLabelRequired)
		}
		if _, dup := seen[a.QuestionNumber]; dup {
			return errs.NewError(ctx, status.EXAM_INVALID_ANSWERS,
				map[string]any{"question_number": a.QuestionNumber}, ErrDuplicateAnswer)
		}
		seen[a.QuestionNumber] = struct{}{}
	}
	return nil
}

// ValidateGetExam accepts elink_id (a sitting) or esess_id (a
// journey); when both are sent the journey wins. Neither is a
// missing-attempt error — the older of the two shapes, kept so existing
// clients see the code they already handle.
//
// A journey read also settles req_exam_type: it names which row of the
// journey to read and is normalised in place. Left empty, the journey's
// owning row is read — ASSESSMENT or GRADE, whichever it was opened as —
// which is what clients that predate the PRACTICE row always got. It is
// deliberately NOT defaulted to ASSESSMENT: that made every GRADE journey
// read as "not found".
func ValidateGetExam(ctx context.Context, req *dto.GetExamReq) error {
	if req.ProfileID <= 0 {
		return errs.NewError(ctx, status.EXAM_MISSING_PROFILE_ID, nil, ErrProfileIDRequired)
	}
	hasAttempt := req.ElinkID > 0
	hasJourney := req.EsessID > 0
	if !hasAttempt && !hasJourney {
		return errs.NewError(ctx, status.EXAM_MISSING_ATTEMPT_ID, nil, ErrDetailIDRequired)
	}

	if hasJourney {
		examType, err := normalizeExamType(ctx, req.ExamType)
		if err != nil {
			return err
		}
		req.ExamType = examType
	}
	return nil
}

// normalizeExamType upper-cases and validates an optional exam type. An
// absent or blank value comes back nil so the caller can apply its own
// default.
func normalizeExamType(ctx context.Context, raw *string) (*string, error) {
	if raw == nil {
		return nil, nil
	}
	normalized := strings.ToUpper(strings.TrimSpace(*raw))
	if normalized == "" {
		return nil, nil
	}
	if !enum.ExamType(normalized).IsValid() {
		return nil, errs.NewError(ctx, status.EXAM_INVALID_EXAM_TYPE, nil, ErrExamTypeInvalid)
	}
	return &normalized, nil
}

// requireJourneyForPractice is the rule shared by the history reads: a
// PRACTICE round only exists inside a journey, so listing "the practice
// rounds" is only a question about one.
func requireJourneyForPractice(ctx context.Context, examType *string, esessID *int64) error {
	if examType != nil && *examType == string(enum.ExamTypePractice) && (esessID == nil || *esessID <= 0) {
		return errs.NewError(ctx, status.EXAM_MISSING_JOURNEY_ID, nil, ErrPracticeJourneyRequired)
	}
	return nil
}

func ValidateListExams(ctx context.Context, req *dto.ListExamsReq) error {
	if req.ProfileID <= 0 {
		return errs.NewError(ctx, status.EXAM_MISSING_PROFILE_ID, nil, ErrProfileIDRequired)
	}
	examType, err := normalizeExamType(ctx, req.ExamType)
	if err != nil {
		return err
	}
	req.ExamType = examType
	if err := requireJourneyForPractice(ctx, req.ExamType, req.EsessID); err != nil {
		return err
	}
	if req.Status != nil {
		normalized := strings.ToUpper(strings.TrimSpace(*req.Status))
		if normalized == "" {
			req.Status = nil
		} else {
			req.Status = &normalized
		}
	}
	return nil
}

func ValidateListExamSessions(ctx context.Context, req *dto.ListExamSessionsReq) error {
	if req.ProfileID <= 0 {
		return errs.NewError(ctx, status.EXAM_MISSING_PROFILE_ID, nil, ErrProfileIDRequired)
	}
	if req.JourneyExamStatus != nil {
		normalized := strings.ToUpper(strings.TrimSpace(*req.JourneyExamStatus))
		if normalized == "" {
			req.JourneyExamStatus = nil
		} else {
			if !enum.EsessStatusType(normalized).IsValid() {
				return errs.NewError(ctx, status.EXAM_INVALID_JOURNEY_STATUS, nil, ErrJourneyStatusInvalid)
			}
			req.JourneyExamStatus = &normalized
		}
	}
	// Normalised, blanks dropped, duplicates folded; nothing left means
	// every type.
	var examTypes []string
	for _, raw := range req.ExamTypes {
		examType, err := normalizeExamType(ctx, &raw)
		if err != nil {
			return err
		}
		if examType == nil || slices.Contains(examTypes, *examType) {
			continue
		}
		// PRACTICE is not a journey: its totals are read inside the
		// journey they belong to. Answering with bare practice rows would
		// hand the client two entries per id.
		if *examType == string(enum.ExamTypePractice) {
			return errs.NewError(ctx, status.EXAM_INVALID_EXAM_TYPE, nil, ErrStatsPracticeNotAJourney)
		}
		examTypes = append(examTypes, *examType)
	}
	req.ExamTypes = examTypes
	return req.Request.Validate(ctx)
}

// Progress chart bounds. The window is capped at roughly two years
// because the chart is a learning trend, not an archive: a wider range
// costs a bigger scan and shows a line nobody reads.
const (
	ProgressLimitDefault = 10
	ProgressLimitMin     = 1
	ProgressLimitMax     = 100
	progressMaxRange     = 2 * 365 * 24 * time.Hour
)

// ValidateExamProgress checks the sittings chart request and normalises
// Limit (clamped in place), Tz (default applied) and ExamType
// (upper-cased). Ownership needs a repository read and so stays in the
// service.
func ValidateExamProgress(ctx context.Context, req *dto.ExamProgressReq) error {
	if req.ProfileID <= 0 {
		return errs.NewError(ctx, status.EXAM_MISSING_PROFILE_ID, nil, ErrProfileIDRequired)
	}

	examType, err := normalizeExamType(ctx, req.ExamType)
	if err != nil {
		return err
	}
	req.ExamType = examType
	if err := requireJourneyForPractice(ctx, req.ExamType, req.EsessID); err != nil {
		return err
	}
	return validateProgressWindow(ctx, &req.ProgressWindow)
}

// ValidateJourneyProgress checks the journeys chart request the same way.
// PRACTICE is refused for the reason /exams/sessions/list refuses it: a PRACTICE row is
// not a journey, it is a part of one.
func ValidateJourneyProgress(ctx context.Context, req *dto.JourneyProgressReq) error {
	if req.ProfileID <= 0 {
		return errs.NewError(ctx, status.EXAM_MISSING_PROFILE_ID, nil, ErrProfileIDRequired)
	}

	examType, err := normalizeExamType(ctx, req.ExamType)
	if err != nil {
		return err
	}
	if examType != nil && *examType == string(enum.ExamTypePractice) {
		return errs.NewError(ctx, status.EXAM_INVALID_EXAM_TYPE, nil, ErrStatsPracticeNotAJourney)
	}
	req.ExamType = examType
	return validateProgressWindow(ctx, &req.ProgressWindow)
}

// ValidateJourneyReview checks the review request: a profile and a journey.
func ValidateJourneyReview(ctx context.Context, req *dto.JourneyReviewReq) error {
	if req.ProfileID <= 0 {
		return errs.NewError(ctx, status.EXAM_MISSING_PROFILE_ID, nil, ErrProfileIDRequired)
	}
	if req.EsessID <= 0 {
		return errs.NewError(ctx, status.EXAM_MISSING_JOURNEY_ID, nil, ErrJourneyIDRequired)
	}
	return nil
}

// ValidateLatestJourney checks the latest-journey request. Both filters
// are optional; an empty exam_type means any type.
func ValidateLatestJourney(ctx context.Context, req *dto.LatestJourneyReq) error {
	if req.ProfileID <= 0 {
		return errs.NewError(ctx, status.EXAM_MISSING_PROFILE_ID, nil, ErrProfileIDRequired)
	}
	examType, err := normalizeExamType(ctx, req.ExamType)
	if err != nil {
		return err
	}
	// A PRACTICE row lives inside its journey (under "practice"); it is
	// never a journey of its own, so it cannot be the latest one.
	if examType != nil && *examType == string(enum.ExamTypePractice) {
		return errs.NewError(ctx, status.EXAM_INVALID_EXAM_TYPE, nil, ErrStatsPracticeNotAJourney)
	}
	req.ExamType = examType
	if req.Grade != nil && (*req.Grade < enum.ExamGradeMin || *req.Grade > enum.ExamGradeMax) {
		return errs.NewError(ctx, status.EXAM_INVALID_GRADE, nil, ErrGradeOutOfRange)
	}
	return nil
}

// ValidateGradeLevels checks the level-ladder request: a profile and a
// grade the product teaches.
func ValidateGradeLevels(ctx context.Context, req *dto.GradeLevelsReq) error {
	return validateProfileAndGrade(ctx, req.ProfileID, req.Grade)
}

// ValidateGradeMap checks /exams/grade/map the same way.
func ValidateGradeMap(ctx context.Context, req *dto.GradeMapReq) error {
	return validateProfileAndGrade(ctx, req.ProfileID, req.Grade)
}

// validateProfileAndGrade is the rule every per-grade read shares: a
// profile, and a required grade the product teaches.
func validateProfileAndGrade(ctx context.Context, profileID int64, grade *int) error {
	if profileID <= 0 {
		return errs.NewError(ctx, status.EXAM_MISSING_PROFILE_ID, nil, ErrProfileIDRequired)
	}
	if grade == nil {
		return errs.NewError(ctx, status.EXAM_MISSING_GRADE, nil, ErrGradeRequired)
	}
	if *grade < enum.ExamGradeMin || *grade > enum.ExamGradeMax {
		return errs.NewError(ctx, status.EXAM_INVALID_GRADE, nil, ErrGradeOutOfRange)
	}
	return nil
}

// validateProgressWindow normalises the part both charts share: Tz gets
// its default, the date range must be complete and bounded, Limit is
// clamped in place.
func validateProgressWindow(ctx context.Context, w *dto.ProgressWindow) error {
	// A numeric offset, never an IANA name: the value reaches CONVERT_TZ
	// verbatim and production MySQL is not guaranteed to carry the
	// timezone tables that would make "Asia/Ho_Chi_Minh" resolve.
	if w.Tz == "" {
		w.Tz = enum.DefaultProgressTz
	} else if !enum.IsValidTzOffset(w.Tz) {
		return errs.NewError(ctx, status.EXAM_ANALYTICS_INVALID_TZ, nil, ErrInvalidTz)
	}

	if w.FromDt != "" || w.ToDt != "" {
		if w.FromDt == "" || w.ToDt == "" {
			return errs.NewError(ctx, status.EXAM_ANALYTICS_INVALID_RANGE, nil, ErrInvalidDateRange)
		}
		from, err := mtime.ParseFromString(w.FromDt)
		if err != nil {
			return errs.NewError(ctx, status.EXAM_ANALYTICS_INVALID_RANGE, nil, err)
		}
		to, err := mtime.ParseFromString(w.ToDt)
		if err != nil {
			return errs.NewError(ctx, status.EXAM_ANALYTICS_INVALID_RANGE, nil, err)
		}
		if !from.Time.Before(to.Time) || to.Time.Sub(from.Time) > progressMaxRange {
			return errs.NewError(ctx, status.EXAM_ANALYTICS_INVALID_RANGE, nil, ErrInvalidDateRange)
		}
	}

	if w.Limit == 0 {
		w.Limit = ProgressLimitDefault
	}
	if w.Limit < ProgressLimitMin {
		w.Limit = ProgressLimitMin
	}
	if w.Limit > ProgressLimitMax {
		w.Limit = ProgressLimitMax
	}
	return nil
}

// ValidatedMarkJourney carries the normalised status so the service does
// not re-parse the request's string.
type ValidatedMarkJourney struct {
	Status enum.EsessStatusType
}

// ValidateMarkExamJourney checks an end-of-journey request. Only COMPLETE
// and CANCEL are accepted here: ACTIVE would be a reopen, which does not
// exist, and DELETED is a soft-delete with its own path.
func ValidateMarkExamJourney(ctx context.Context, req *dto.MarkExamJourneyReq) (ValidatedMarkJourney, error) {
	if req.ProfileID <= 0 {
		return ValidatedMarkJourney{}, errs.NewError(ctx, status.EXAM_MISSING_PROFILE_ID, nil, ErrProfileIDRequired)
	}
	if req.EsessID <= 0 {
		return ValidatedMarkJourney{}, errs.NewError(ctx, status.EXAM_MISSING_JOURNEY_ID, nil, ErrJourneyIDRequired)
	}
	st := enum.EsessStatusType(strings.ToUpper(strings.TrimSpace(req.Status)))
	if !st.IsMarkable() {
		return ValidatedMarkJourney{}, errs.NewError(ctx, status.EXAM_INVALID_JOURNEY_STATUS, nil, ErrJourneyStatusInvalid)
	}
	req.Status = string(st)
	return ValidatedMarkJourney{Status: st}, nil
}
