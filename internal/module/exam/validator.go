package exam

import (
	"context"
	"strings"
	"time"

	dto "math-ai.com/math-ai/internal/application/dto/exam"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/mtime"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/shared/enum"
)

// DefaultNumQuestions is what a generate request gets when it names no
// count. MaxNumQuestions caps prompt size, token cost and latency; a
// larger request is clamped rather than rejected, since the caller's
// intent ("a longer exam") is still servable.
//
// The ASSESSMENT probe positions are fixed at questions 3 and 6, so an
// exam shorter than six loses one or both probes. That is allowed — it
// just measures less.
const (
	DefaultNumQuestions = 10
	MaxNumQuestions     = 20
)

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
	if examType == enum.ExamTypePractice && (req.UserExamID == nil || *req.UserExamID <= 0) {
		return ValidatedGenerate{}, errs.NewError(ctx, status.EXAM_MISSING_JOURNEY_ID, nil, ErrPracticeJourneyRequired)
	}

	// Grade is optional; when pinned it must name a band the product
	// actually serves. Nothing may be placed at the probe ceiling.
	if req.Grade != nil && (*req.Grade < enum.ExamGradeMin || *req.Grade > enum.ExamGradeMax) {
		return ValidatedGenerate{}, errs.NewError(ctx, status.EXAM_INVALID_GRADE, nil, ErrGradeOutOfRange)
	}

	if req.NumQuestions <= 0 {
		req.NumQuestions = DefaultNumQuestions
	} else if req.NumQuestions > MaxNumQuestions {
		req.NumQuestions = MaxNumQuestions
	}

	return ValidatedGenerate{ExamType: examType}, nil
}

func ValidateSubmitExam(ctx context.Context, req *dto.SubmitExamReq) error {
	if req.ProfileID <= 0 {
		return errs.NewError(ctx, status.EXAM_MISSING_PROFILE_ID, nil, ErrProfileIDRequired)
	}
	if req.UserAiExamID <= 0 {
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

// ValidateGetExam accepts user_ai_exam_id (a sitting) or user_exam_id (a
// journey); when both are sent the journey wins. Neither is a
// missing-attempt error — the older of the two shapes, kept so existing
// clients see the code they already handle.
//
// A journey read also settles req_exam_type: it names which row of the
// journey to read, defaults to ASSESSMENT so clients that predate the
// PRACTICE row keep getting what they always did, and is normalised in
// place so the service can use it as-is.
func ValidateGetExam(ctx context.Context, req *dto.GetExamReq) error {
	if req.ProfileID <= 0 {
		return errs.NewError(ctx, status.EXAM_MISSING_PROFILE_ID, nil, ErrProfileIDRequired)
	}
	hasAttempt := req.UserAiExamID > 0
	hasJourney := req.UserExamID > 0
	if !hasAttempt && !hasJourney {
		return errs.NewError(ctx, status.EXAM_MISSING_ATTEMPT_ID, nil, ErrDetailIDRequired)
	}

	if hasJourney {
		examType, err := normalizeExamType(ctx, req.ExamType)
		if err != nil {
			return err
		}
		if examType == nil {
			def := string(enum.ExamTypeAssessment)
			examType = &def
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
func requireJourneyForPractice(ctx context.Context, examType *string, userExamID *int64) error {
	if examType != nil && *examType == string(enum.ExamTypePractice) && (userExamID == nil || *userExamID <= 0) {
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
	if err := requireJourneyForPractice(ctx, req.ExamType, req.UserExamID); err != nil {
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

func ValidateGetExamStats(ctx context.Context, req *dto.GetExamStatsReq) error {
	if req.ProfileID <= 0 {
		return errs.NewError(ctx, status.EXAM_MISSING_PROFILE_ID, nil, ErrProfileIDRequired)
	}
	if req.Status != nil {
		normalized := strings.ToUpper(strings.TrimSpace(*req.Status))
		if normalized == "" {
			req.Status = nil
		} else {
			if !enum.UserExamStatusType(normalized).IsValid() {
				return errs.NewError(ctx, status.EXAM_INVALID_JOURNEY_STATUS, nil, ErrJourneyStatusInvalid)
			}
			req.Status = &normalized
		}
	}
	examType, err := normalizeExamType(ctx, req.ExamType)
	if err != nil {
		return err
	}
	// PRACTICE is not a journey: its totals are read inside the ASSESSMENT
	// journey they belong to. Answering with bare practice rows would
	// hand the client two entries per id.
	if examType != nil && *examType == string(enum.ExamTypePractice) {
		return errs.NewError(ctx, status.EXAM_INVALID_EXAM_TYPE, nil, ErrStatsPracticeNotAJourney)
	}
	req.ExamType = examType
	return nil
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

// ValidateExamProgress checks the analytics request and normalises Limit
// (clamped in place), Tz (default applied) and ExamType (upper-cased).
// Ownership needs a repository read and so stays in the service.
func ValidateExamProgress(ctx context.Context, req *dto.ExamProgressReq) error {
	if req.ProfileID <= 0 {
		return errs.NewError(ctx, status.EXAM_MISSING_PROFILE_ID, nil, ErrProfileIDRequired)
	}

	examType, err := normalizeExamType(ctx, req.ExamType)
	if err != nil {
		return err
	}
	req.ExamType = examType
	if err := requireJourneyForPractice(ctx, req.ExamType, req.UserExamID); err != nil {
		return err
	}

	// A numeric offset, never an IANA name: the value reaches CONVERT_TZ
	// verbatim and production MySQL is not guaranteed to carry the
	// timezone tables that would make "Asia/Ho_Chi_Minh" resolve.
	if req.Tz == "" {
		req.Tz = enum.DefaultProgressTz
	} else if !enum.IsValidTzOffset(req.Tz) {
		return errs.NewError(ctx, status.EXAM_ANALYTICS_INVALID_TZ, nil, ErrInvalidTz)
	}

	if req.FromDt != "" || req.ToDt != "" {
		if req.FromDt == "" || req.ToDt == "" {
			return errs.NewError(ctx, status.EXAM_ANALYTICS_INVALID_RANGE, nil, ErrInvalidDateRange)
		}
		from, err := mtime.ParseFromString(req.FromDt)
		if err != nil {
			return errs.NewError(ctx, status.EXAM_ANALYTICS_INVALID_RANGE, nil, err)
		}
		to, err := mtime.ParseFromString(req.ToDt)
		if err != nil {
			return errs.NewError(ctx, status.EXAM_ANALYTICS_INVALID_RANGE, nil, err)
		}
		if !from.Time.Before(to.Time) || to.Time.Sub(from.Time) > progressMaxRange {
			return errs.NewError(ctx, status.EXAM_ANALYTICS_INVALID_RANGE, nil, ErrInvalidDateRange)
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

// ValidatedMarkJourney carries the normalised status so the service does
// not re-parse the request's string.
type ValidatedMarkJourney struct {
	Status enum.UserExamStatusType
}

// ValidateMarkExamJourney checks an end-of-journey request. Only COMPLETE
// and CANCEL are accepted here: ACTIVE would be a reopen, which does not
// exist, and DELETED is a soft-delete with its own path.
func ValidateMarkExamJourney(ctx context.Context, req *dto.MarkExamJourneyReq) (ValidatedMarkJourney, error) {
	if req.ProfileID <= 0 {
		return ValidatedMarkJourney{}, errs.NewError(ctx, status.EXAM_MISSING_PROFILE_ID, nil, ErrProfileIDRequired)
	}
	if req.UserExamID <= 0 {
		return ValidatedMarkJourney{}, errs.NewError(ctx, status.EXAM_MISSING_JOURNEY_ID, nil, ErrJourneyIDRequired)
	}
	st := enum.UserExamStatusType(strings.ToUpper(strings.TrimSpace(req.Status)))
	if !st.IsMarkable() {
		return ValidatedMarkJourney{}, errs.NewError(ctx, status.EXAM_INVALID_JOURNEY_STATUS, nil, ErrJourneyStatusInvalid)
	}
	req.Status = string(st)
	return ValidatedMarkJourney{Status: st}, nil
}
