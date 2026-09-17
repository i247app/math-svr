package exam

import (
	"context"

	dto "math-ai.com/math-ai/internal/application/dto/exam"
	query "math-ai.com/math-ai/internal/application/query/exam"
	domainBot "math-ai.com/math-ai/internal/domain/bot"
	examDomain "math-ai.com/math-ai/internal/domain/exam"
	profileDomain "math-ai.com/math-ai/internal/domain/profile"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/infrastructure/logger"
	"math-ai.com/math-ai/internal/shared/enum"
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

// findReusableExam returns a stored question set only once the tag's pool
// is deep enough to pick from, and never one this child has sat before.
// A miss is an ordinary outcome, not an error — the caller generates
// instead, and that generation deepens the pool for everyone.
func (s *Service) findReusableExam(ctx context.Context, tag string, profileID int64) (*examDomain.AiExam, error) {
	log := logger.From(ctx)

	variants, err := s.aiExamRepo.CountByExtras(ctx, tag)
	if err != nil {
		return nil, errs.NewError(ctx, status.FAIL, nil, err)
	}
	if variants < MinCacheVariants {
		log.Infof("exam.cache.warming tag=%s variants=%d/%d", tag, variants, MinCacheVariants)
		return nil, nil
	}

	cached, err := s.aiExamRepo.FindReusableByExtras(ctx, tag, profileID)
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
func (s *Service) loadOwnedProfile(ctx context.Context, userID *int64, profileID int64) (*profileDomain.Profile, error) {
	if userID == nil || *userID == 0 {
		return nil, errs.NewError(ctx, status.UNAUTHORIZED, nil, ErrUidNotFoundFromSession)
	}
	p, err := s.profileRepo.FindByProfileId(ctx, profileID)
	if err != nil {
		return nil, errs.NewError(ctx, status.FAIL, nil, err)
	}
	if p == nil {
		return nil, errs.NewError(ctx, status.EXAM_PROFILE_NOT_FOUND, nil, ErrProfileNotFound)
	}
	if p.UserId() != *userID {
		return nil, errs.NewError(ctx, status.EXAM_PROFILE_NOT_OWNED, nil, ErrProfileNotOwned)
	}
	return p, nil
}

// resolvePlacement decides which band the next paper is written at, and
// reports the open journey (if any) so the sitting can be pinned to it.
//
// Grade, in order of preference: what the client stated on this request,
// then the journey's current grade (what the client stated last time),
// then the class the profile says the child attends, and finally
// kindergarten. Nothing here is measured by the server any more — the
// journey's grade is the client's statement, recorded at hand-out — so
// there is no longer a "carry the last measurement forward" step between
// journeys: a new journey starts where the client, or the profile, says.
//
// Level is not resolved. It is recorded when stated and read by no rule.
func (s *Service) resolvePlacement(ctx context.Context, req *dto.GenerateExamReq, examType enum.ExamType, profile *profileDomain.Profile) (grade int, openJourney *int64, err error) {
	// The open journey is looked up regardless of a stated grade: even a
	// pinned sitting belongs to the journey that is open.
	active, err := s.statsRepo.FindActiveByUserProfileType(ctx, profile.UserId(), profile.ProfileId(), string(examType))
	if err != nil {
		return 0, nil, errs.NewError(ctx, status.FAIL, nil, err)
	}
	if active != nil {
		id := active.UserExamId()
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
// writes at a level, and kindergarten cannot be pushed past 4 (the child
// cannot read yet); asking for more degrades to the ceiling, with a log
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
func (s *Service) getAttempt(ctx context.Context, userAiExamID int64, profile *profileDomain.Profile) (*dto.GetExamRes, error) {
	detail, err := s.getAttemptQuery.Handle(ctx, query.GetExamAttemptQuery{
		UserAiExamID: userAiExamID,
		UserID:       profile.UserId(),
		ProfileID:    profile.ProfileId(),
	})
	if err != nil {
		return nil, err
	}

	submitted := detail.Attempt.UserAiExamStatus() != nil &&
		*detail.Attempt.UserAiExamStatus() == string(enum.UserAiExamStatusSubmitted)

	return &dto.GetExamRes{
		Exam: dto.AttemptToResponse(detail.Attempt, detail.AiExam, submitted),
		Details: dto.DetailsToResponse(detail.Details, dto.ShufflesOf(detail.Attempt),
			map[int64]*examDomain.AiExam{detail.AiExam.AiExamId(): detail.AiExam}),
	}, nil
}

// getJourney is the whole-journey review: running totals, every sitting
// that fed them (as cards, no question blobs), and every answered
// question across those sittings. Every sitting here is SUBMITTED by
// construction — that is the moment a sitting joins a journey — so the
// answer key is always safe to expose.
func (s *Service) getJourney(ctx context.Context, userExamID int64, examType string, profile *profileDomain.Profile) (*dto.GetExamRes, error) {
	detail, err := s.getJourneyQuery.Handle(ctx, query.GetExamJourneyQuery{
		UserExamID: userExamID,
		ExamType:   examType,
		UserID:     profile.UserId(),
		ProfileID:  profile.ProfileId(),
	})
	if err != nil {
		return nil, err
	}

	return &dto.GetExamRes{
		Stats:           dto.StatsToSingleResponse(detail.Journey),
		Exams:           dto.AttemptListToResponse(detail.Attempts, detail.AiExams),
		Details:         dto.DetailsToResponse(detail.Details, dto.ShufflesOf(detail.Attempts...), detail.AiExams),
		PracticePreview: dto.PracticePreviewFrom(detail.PracticeBase, detail.PracticeBrief),
	}, nil
}
