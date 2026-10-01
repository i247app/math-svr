package exam

import (
	"context"
	"fmt"
	"strings"
	"time"

	dto "math-ai.com/math-ai/internal/application/dto/exam"
	query "math-ai.com/math-ai/internal/application/query/exam"
	domainBot "math-ai.com/math-ai/internal/domain/bot"
	examDomain "math-ai.com/math-ai/internal/domain/exam"
	profileDomain "math-ai.com/math-ai/internal/domain/profile"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/mtime"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/infrastructure/logger"
	"math-ai.com/math-ai/internal/shared/enum"
	"math-ai.com/math-ai/internal/shared/utils"
)

// MinCacheVariants is how many distinct question sets must exist under a
// cache tag before the tag starts serving from cache.
//
// Without it the cache would defeat itself: "found one, reuse it" means
// the FIRST set generated for a tag is the only one ever generated, so a
// child practising twice at the same placement meets the identical paper.
// Below the threshold the server keeps paying for generations, which is
// how the pool fills; above it, reads are free and varied.
//
// Three is a starting point, not a measurement. Raise it if children
// report repeats; lower it if the generation bill says so.
const MinCacheVariants = 100

// RecentExamsForAvoid is how many of the child's latest sittings at the
// requested grade have their stems listed in the prompt as "do not
// repeat". One is the case that hurts — the child asks again straight
// after finishing and meets the same sums. Raise it if repeats across
// older papers turn out to bother children too; every paper costs about
// ten short lines of prompt.
const RecentExamsForAvoid = 1

// recentStems collects the question stems of the child's latest sittings
// at grade, newest paper first, deduplicated. It reads the stored
// (canonical) question set, so the order within a paper is the stored
// one — the model only needs the stems, not the order they were served.
func (s *Service) recentStems(ctx context.Context, profileID int64, grade int) ([]string, error) {
	attempts, err := s.attemptRepo.ListRecentByProfileGrade(ctx, profileID, grade, RecentExamsForAvoid)
	if err != nil {
		return nil, errs.NewError(ctx, status.FAIL, nil, err)
	}
	if len(attempts) == 0 {
		return nil, nil
	}
	ids := make([]int64, 0, len(attempts))
	for _, a := range attempts {
		ids = append(ids, a.ExamId())
	}
	sets, err := s.examPoolRepo.ListByExamIds(ctx, ids)
	if err != nil {
		return nil, errs.NewError(ctx, status.FAIL, nil, err)
	}
	byID := make(map[int64]*examDomain.ExamPool, len(sets))
	for _, e := range sets {
		byID[e.ExamId()] = e
	}

	seen := make(map[string]struct{})
	var stems []string
	for _, a := range attempts {
		e := byID[a.ExamId()]
		if e == nil {
			continue
		}
		questions, err := decodeStoredQuestions(ctx, e.AiQuestionsJson())
		if err != nil {
			return nil, err
		}
		for _, q := range questions {
			stem := strings.TrimSpace(q.QuestionName)
			if stem == "" {
				continue
			}
			if _, dup := seen[stem]; dup {
				continue
			}
			seen[stem] = struct{}{}
			stems = append(stems, stem)
		}
	}
	return stems, nil
}

// findReusableExam returns a stored question set only once the tag's pool
// is deep enough to pick from, and never one this child has sat before.
// A miss is an ordinary outcome, not an error — the caller generates
// instead, and that generation deepens the pool for everyone.
func (s *Service) findReusableExam(ctx context.Context, tag string, profileID int64) (*examDomain.ExamPool, error) {
	log := logger.From(ctx)

	variants, err := s.examPoolRepo.CountByExtras(ctx, tag)
	if err != nil {
		return nil, errs.NewError(ctx, status.FAIL, nil, err)
	}
	if variants < MinCacheVariants {
		log.Infof("exam.cache.warming tag=%s variants=%d/%d", tag, variants, MinCacheVariants)
		return nil, nil
	}

	cached, err := s.examPoolRepo.FindReusableByExtras(ctx, tag, profileID)
	if err != nil {
		return nil, errs.NewError(ctx, status.FAIL, nil, err)
	}
	if cached == nil {
		log.Infof("exam.cache.exhausted tag=%s profile=%d variants=%d", tag, profileID, variants)
	}
	return cached, nil
}

// loadOwnedProfile resolves the acting child and proves it belongs to the
// session's user. Every entry point goes through it: a profile id is a
// plain integer, so without this check one parent could read — or submit
// against — another family's exams.
func (s *Service) loadOwnedProfile(ctx context.Context, uid *int64, profileID int64) (*profileDomain.Profile, error) {
	if uid == nil || *uid == 0 {
		return nil, errs.NewError(ctx, status.UNAUTHORIZED, nil, ErrUidNotFoundFromSession)
	}
	p, err := s.profileRepo.FindByProfileId(ctx, profileID)
	if err != nil {
		return nil, errs.NewError(ctx, status.FAIL, nil, err)
	}
	if p == nil {
		return nil, errs.NewError(ctx, status.EXAM_PROFILE_NOT_FOUND, nil, ErrProfileNotFound)
	}
	if p.Uid() != *uid {
		return nil, errs.NewError(ctx, status.EXAM_PROFILE_NOT_OWNED, nil, ErrProfileNotOwned)
	}
	return p, nil
}

// resolvePlacement decides which band the next paper is written at, and
// reports the open journey (if any) so the sitting can be pinned to it.
//
// Grade, in order of preference: what the client stated on this request,
// then the open journey's current grade (what the client stated last
// time), then the class the profile says the child attends, and finally
// kindergarten.
//
// Which journey is "the open one" depends on the type (exam.JourneyKey).
// An ASSESSMENT child has one, whatever the grade. A GRADE child has one
// per grade: a stated grade picks the journey of that grade — none open
// means the command opens a new one — and never folds into a journey of
// another grade. With no grade stated, the GRADE journey worked last is
// continued at its own grade. Nothing here is measured by the server any more — the
// journey's grade is the client's statement, recorded at hand-out — so
// there is no longer a "carry the last measurement forward" step between
// journeys: a new journey starts where the client, or the profile, says.
//
// Level is not resolved. It is recorded when stated and read by no rule.
func (s *Service) resolvePlacement(ctx context.Context, req *dto.GenerateExamReq, examType enum.ExamType, profile *profileDomain.Profile) (grade int, openJourney *int64, err error) {
	// The open journey is looked up regardless of a stated grade: even a
	// pinned sitting belongs to the journey that is open in its slot.
	key := examDomain.JourneyKeyOf(profile.Uid(), profile.ProfileId(), string(examType), req.Grade)
	active, err := s.statsRepo.FindActiveJourney(ctx, key)
	if err != nil {
		return 0, nil, errs.NewError(ctx, status.FAIL, nil, err)
	}
	if active != nil {
		id := active.EsessId()
		openJourney = &id
	}

	if req.Grade != nil {
		return *req.Grade, openJourney, nil
	}
	if active != nil && active.CurrentGrade() != nil {
		return *active.CurrentGrade(), openJourney, nil
	}

	grade, err = s.gradeFromProfile(ctx, profile)
	return grade, openJourney, err
}

// resolveLevel is the one place a stated level is clamped, so the number
// that reaches the cache tag, the stored rows, the journey and the
// prompt is the same number. A GRADE review is the only round that
// writes at a level, and a band may cap it (domainBot.ClampLevelToGrade —
// kindergarten, whose child cannot read yet, is the case the cap exists
// for); asking for more degrades to the ceiling, with a log
// line, rather than being refused — the request was reasonable, the
// band just has nowhere higher to go. Other types keep the stated value
// as a record only.
func resolveLevel(ctx context.Context, examType enum.ExamType, grade int, stated *int) *int {
	if stated == nil || examType != enum.ExamTypeGrade {
		return stated
	}
	clamped := domainBot.ClampLevelToGrade(grade, *stated)
	if clamped != *stated {
		logger.From(ctx).Warnf("exam.level.clamped grade=%d stated=%d applied=%d", grade, *stated, clamped)
	}
	return &clamped
}

// gradeFromProfile reads the class the child attends. The profile stores
// a grade id, and the band number lives in that row's label ("Lớp 1",
// "Mẫu giáo"), so it has to be resolved rather than cast. An unreadable
// or absent label falls back to kindergarten — the safest wrong answer,
// since an exam that is too easy is recoverable and one that is far too
// hard is not.
func (s *Service) gradeFromProfile(ctx context.Context, profile *profileDomain.Profile) (int, error) {
	if profile.GradeId() == nil {
		return enum.ExamGradeMin, nil
	}
	grades, err := s.gradeRepo.ListGradesByIds(ctx, []int64{*profile.GradeId()})
	if err != nil {
		return 0, errs.NewError(ctx, status.FAIL, nil, err)
	}
	if len(grades) == 0 {
		return enum.ExamGradeMin, nil
	}
	if number, ok := domainBot.ResolveGradeNumber(grades[0].Label()); ok {
		return number, nil
	}
	logger.From(ctx).Warnf("exam.grade_label_unresolved label=%q profile=%d",
		grades[0].Label(), profile.ProfileId())
	return enum.ExamGradeMin, nil
}

// getAttempt is the single-sitting review: the exam as it was served,
// plus what the child answered on it.
func (s *Service) getAttempt(ctx context.Context, elinkID int64, profile *profileDomain.Profile) (*dto.GetExamRes, error) {
	detail, err := s.getAttemptQuery.Handle(ctx, query.GetExamAttemptQuery{
		ElinkID:   elinkID,
		UID:       profile.Uid(),
		ProfileID: profile.ProfileId(),
	})
	if err != nil {
		return nil, err
	}

	submitted := detail.Attempt.ElinkStatus() != nil &&
		*detail.Attempt.ElinkStatus() == string(enum.ElinkStatusSubmitted)

	return &dto.GetExamRes{
		Exam: dto.AttemptToResponse(detail.Attempt, detail.ExamPool, submitted),
		Details: dto.DetailsToResponse(detail.Details, dto.ShufflesOf(detail.Attempt),
			map[int64]*examDomain.ExamPool{detail.ExamPool.ExamId(): detail.ExamPool}),
	}, nil
}

// getJourney is the whole-journey review: running totals, every sitting
// that fed them (as cards, no question blobs), and every answered
// question across those sittings. Every sitting here is SUBMITTED by
// construction — that is the moment a sitting joins a journey — so the
// answer key is always safe to expose.
func (s *Service) getJourney(ctx context.Context, esessID int64, examType string, profile *profileDomain.Profile) (*dto.GetExamRes, error) {
	detail, err := s.getJourneyQuery.Handle(ctx, query.GetExamJourneyQuery{
		EsessID:   esessID,
		ExamType:  examType,
		UID:       profile.Uid(),
		ProfileID: profile.ProfileId(),
	})
	if err != nil {
		return nil, err
	}

	return &dto.GetExamRes{
		Stats:           dto.StatsToSingleResponse(detail.Journey),
		Exams:           dto.AttemptListToResponse(detail.Attempts, detail.ExamPools),
		Details:         dto.DetailsToResponse(detail.Details, dto.ShufflesOf(detail.Attempts...), detail.ExamPools),
		PracticePreview: dto.PracticePreviewFrom(detail.PracticeBase, detail.PracticeBrief),
	}, nil
}

// Guest ceilings. Everything a guest is handed is a paid model call made
// for someone who has not registered, so the trial is bounded in two
// directions: what they may ask for, and how much of it per day.
const (
	// GuestDailyExamLimit is how many rounds one guest child may be
	// handed in a rolling 24 hours. Raise it here; nothing else reads a
	// number.
	GuestDailyExamLimit = 30
	// guestExamWindow is what "per day" means. A rolling window rather
	// than midnight-to-midnight, so a guest cannot take the day's quota
	// twice by starting just before midnight — and so the rule needs no
	// timezone to be fair.
	guestExamWindow = 24 * time.Hour
)

// guardGuest applies both ceilings. It reads the profile's own
// identity_code, already loaded for the ownership check, so a registered
// child costs nothing here.
func (s *Service) guardGuest(ctx context.Context, p *profileDomain.Profile, examType enum.ExamType) error {
	if !enum.IdentityCodeType(utils.DerefString(p.IdentityCode())).IsGuest() {
		return nil
	}

	// A guest is trying the product, not using it. PRACTICE is built from
	// a child's own mistakes across a journey and GRADE is a review at a
	// level — both belong to a child whose progress is being followed.
	if examType != enum.ExamTypeAssessment {
		return errs.NewError(ctx, status.EXAM_GUEST_TYPE_NOT_ALLOWED, nil, ErrGuestAssessmentOnly)
	}

	since := mtime.MathTime{Time: time.Now().UTC().Add(-guestExamWindow)}
	handed, err := s.attemptRepo.CountHandedOutSince(ctx, p.ProfileId(), since)
	if err != nil {
		return errs.NewError(ctx, status.FAIL, nil, err)
	}
	if handed >= GuestDailyExamLimit {
		logger.From(ctx).Info("exam.guest.limit_reached", "profile_id", p.ProfileId(), "handed", handed)
		return errs.NewError(ctx, status.EXAM_GUEST_DAILY_LIMIT,
			map[string]any{"limit": GuestDailyExamLimit}, ErrGuestDailyLimit)
	}
	return nil
}

// resumeOpenSitting returns the child's IN_PROGRESS sitting of examType in
// the journey, served exactly as it was handed out (same elink_id, same
// ordering, no answer key), or (nil, nil) when there is none. It is a
// plain read that saves the model call in the ordinary case; the
// generate command repeats the check under a lock for the concurrent one.
func (s *Service) resumeOpenSitting(ctx context.Context, profile *profileDomain.Profile, esessID int64, examType enum.ExamType) (*dto.GenerateExamRes, error) {
	open, err := s.attemptRepo.FindInProgressInJourney(ctx, profile.ProfileId(), esessID, string(examType))
	if err != nil {
		return nil, errs.NewError(ctx, status.FAIL, nil, err)
	}
	if open == nil {
		return nil, nil
	}
	pool, err := s.examPoolRepo.FindByExamId(ctx, open.ExamId())
	if err != nil {
		return nil, errs.NewError(ctx, status.FAIL, nil, err)
	}
	if pool == nil {
		return nil, errs.NewError(ctx, status.EXAM_NOT_FOUND, nil,
			fmt.Errorf("exam: open sitting %d points at exam_pool %d, which is gone", open.ElinkId(), open.ExamId()))
	}

	logger.From(ctx).Infof("exam.resumed attempt=%d journey=%d type=%s", open.ElinkId(), esessID, examType)
	return &dto.GenerateExamRes{
		Exam:    dto.AttemptToResponse(open, pool, true),
		Resumed: true,
	}, nil
}

// journeyReviewInput turns the stored answer lines (newest first, as
// ListRecentByEsessId returns them) into the review prompt's input,
// oldest first. Only question content and correctness cross over — no
// profile field ever reaches the prompt.
func journeyReviewInput(journey *examDomain.ExamSession, lines []*examDomain.ExamSessionLine) domainBot.JourneyReviewInput {
	answers := make([]domainBot.ReviewAnswer, 0, len(lines))
	for i := len(lines) - 1; i >= 0; i-- {
		l := lines[i]
		child := l.SelectedLabel()
		if c := utils.DerefString(l.SelectedContent()); c != "" {
			child += " (" + c + ")"
		}
		right := utils.DerefString(l.RightAnswerLabel())
		if c := utils.DerefString(l.RightAnswerContent()); c != "" {
			right += " (" + c + ")"
		}
		answers = append(answers, domainBot.ReviewAnswer{
			Question:    utils.DerefString(l.QuestionName()),
			Topic:       utils.DerefString(l.QuestionTopic()),
			Grade:       l.QuestionGrade(),
			RightAnswer: right,
			ChildAnswer: child,
			Correct:     l.IsCorrect(),
		})
	}
	return domainBot.JourneyReviewInput{
		ExamType:        journey.ReqExamType(),
		Grade:           journey.CurrentGrade(),
		Level:           journey.CurrentLevel(),
		ScorePercentage: journey.ResScorePercentage(),
		Answers:         answers,
	}
}

// requireCompleteForReview refuses an AI review of a journey that is not
// COMPLETE (still ACTIVE, or CANCELLED).
func requireCompleteForReview(ctx context.Context, journey *examDomain.ExamSession) error {
	if st := utils.DerefString(journey.EsessStatus()); st != string(enum.EsessStatusComplete) {
		return errs.NewError(ctx, status.EXAM_REVIEW_JOURNEY_NOT_COMPLETE, nil,
			fmt.Errorf("exam: journey %d is %s; only a COMPLETE journey is reviewed", journey.EsessId(), st))
	}
	return nil
}
