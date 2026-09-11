package exam

import (
	"context"

	dto "math-ai.com/math-ai/internal/application/dto/exam"
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
const MinCacheVariants = 10

// findReusableExam returns a stored question set only once the tag's pool
// is deep enough to pick from. A miss is an ordinary outcome, not an
// error — the caller generates instead.
func (s *Service) findReusableExam(ctx context.Context, tag string) (*examDomain.AiExam, error) {
	log := logger.From(ctx)

	variants, err := s.aiExamRepo.CountByExtras(ctx, tag)
	if err != nil {
		return nil, errs.NewError(ctx, status.FAIL, nil, err)
	}
	if variants < MinCacheVariants {
		log.Infof("exam.cache.warming tag=%s variants=%d/%d", tag, variants, MinCacheVariants)
		return nil, nil
	}

	cached, err := s.aiExamRepo.FindReusableByExtras(ctx, tag)
	if err != nil {
		return nil, errs.NewError(ctx, status.FAIL, nil, err)
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

// resolvePlacement decides how hard the next exam should be.
//
// Grade, in order of preference: what the client pinned, then what the
// child has been MEASURED at for this exam type, then the class their
// profile says they attend, and finally kindergarten. The measurement
// outranks the profile on purpose — the two are different facts, and a
// child in Grade 1 may well be answering Grade 3 material.
//
// Grade is the only axis. There is no level to resolve: the teaching team
// has not defined one, so req_level is never set and stays NULL on every
// row this produces. When a rule exists it slots in here, beside grade,
// and the same number must then reach the cache tag, the stored row and
// the prompt — resolve it once, in this function, not in each consumer.
func (s *Service) resolvePlacement(ctx context.Context, req *dto.GenerateExamReq, examType enum.ExamType, profile *profileDomain.Profile) (int, error) {
	stats, err := s.statsRepo.FindByUserProfileType(ctx, profile.UserId(), profile.ProfileId(), string(examType))
	if err != nil {
		return 0, errs.NewError(ctx, status.FAIL, nil, err)
	}
	return s.resolveGrade(ctx, req, stats, profile)
}

func (s *Service) resolveGrade(ctx context.Context, req *dto.GenerateExamReq, stats *examDomain.UserExam, profile *profileDomain.Profile) (int, error) {
	switch {
	case req.Grade != nil:
		return *req.Grade, nil
	case stats != nil && stats.ResGrade() != nil:
		return *stats.ResGrade(), nil
	}
	return s.gradeFromProfile(ctx, profile)
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
