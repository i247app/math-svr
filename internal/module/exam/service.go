package exam

import (
	"context"

	botAdapter "math-ai.com/math-ai/internal/adapter/bot"
	command "math-ai.com/math-ai/internal/application/command/exam"
	dto "math-ai.com/math-ai/internal/application/dto/exam"
	query "math-ai.com/math-ai/internal/application/query/exam"
	"math-ai.com/math-ai/internal/application/transaction"
	domainBot "math-ai.com/math-ai/internal/domain/bot"
	examDomain "math-ai.com/math-ai/internal/domain/exam"
	gradeDomain "math-ai.com/math-ai/internal/domain/grade"
	profileDomain "math-ai.com/math-ai/internal/domain/profile"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/infrastructure/logger"
	"math-ai.com/math-ai/internal/infrastructure/metadata"
	"math-ai.com/math-ai/internal/shared/enum"
	"math-ai.com/math-ai/internal/shared/utils"
)

// Service is the exam module's public face. It owns the decisions that
// are neither pure persistence nor pure prompt-building: where a child is
// placed before an exam is generated, and whether a stored question set
// can be served instead of paying for a new one.
type Service struct {
	generateCmd     *command.GenerateExamCommandHandler
	submitCmd       *command.SubmitExamCommandHandler
	getAttemptQuery *query.GetExamAttemptQueryHandler
	listQuery       *query.ListExamAttemptsQueryHandler
	statsQuery      *query.GetExamStatsQueryHandler
	progressQuery   *query.GetExamProgressQueryHandler

	aiExamRepo  examDomain.IAiExamRepository
	statsRepo   examDomain.IUserExamRepository
	profileRepo profileDomain.IRepository
	gradeRepo   gradeDomain.IRepository

	bot *botClient
}

// NewService wires the module. bot may be nil — a deploy without LLM
// credentials still boots, and generation then fails with a uniform
// BOT_CONFIG_INVALID instead of a panic.
func NewService(
	aiExamRepo examDomain.IAiExamRepository,
	attemptRepo examDomain.IUserAiExamRepository,
	statsRepo examDomain.IUserExamRepository,
	detailRepo examDomain.IUserExamDetailRepository,
	uow transaction.UnitOfWork,
	bot *botAdapter.Adapter,
	profileRepo profileDomain.IRepository,
	gradeRepo gradeDomain.IRepository,
) *Service {
	return &Service{
		generateCmd:     command.NewGenerateExamCommandHandler(uow),
		submitCmd:       command.NewSubmitExamCommandHandler(uow),
		getAttemptQuery: query.NewGetExamAttemptQueryHandler(attemptRepo, aiExamRepo, detailRepo),
		listQuery:       query.NewListExamAttemptsQueryHandler(attemptRepo, aiExamRepo),
		statsQuery:      query.NewGetExamStatsQueryHandler(statsRepo),
		progressQuery:   query.NewGetExamProgressQueryHandler(attemptRepo),
		aiExamRepo:      aiExamRepo,
		statsRepo:       statsRepo,
		profileRepo:     profileRepo,
		gradeRepo:       gradeRepo,
		bot:             newBotClient(bot),
	}
}

// GenerateExam places the child, looks for a reusable question set, and
// only calls the model when there isn't one. The bot call sits OUTSIDE
// any transaction — an LLM round trip must never hold one open.
func (s *Service) GenerateExam(ctx context.Context, req *dto.GenerateExamReq) (*dto.GenerateExamRes, error) {
	log := logger.From(ctx)

	validated, err := ValidateGenerateExam(ctx, req)
	if err != nil {
		return nil, err
	}
	profile, err := s.loadOwnedProfile(ctx, req.UserID, req.ProfileID)
	if err != nil {
		return nil, err
	}

	grade, level, err := s.resolvePlacement(ctx, req, validated.ExamType, profile)
	if err != nil {
		return nil, err
	}

	tag := BuildCacheTag(validated.ExamType, grade, level, req.NumQuestions, req.Semester, req.Program)

	cmd := command.GenerateExamCommand{
		UserID:    profile.UserId(),
		ProfileID: profile.ProfileId(),
		ExamType:  validated.ExamType,
		Grade:     grade,
		Level:     level,
	}

	cached, err := s.findReusableExam(ctx, tag)
	if err != nil {
		return nil, err
	}
	if cached != nil {
		log.Infof("exam.cache.hit tag=%s ai_exam_id=%d", tag, cached.AiExamId())
		id := cached.AiExamId()
		cmd.ReuseAiExamID = &id
	} else {
		log.Infof("exam.cache.miss tag=%s", tag)
		generated, err := s.bot.GenerateExam(ctx, generateExamInput{
			ExamType:     validated.ExamType,
			Grade:        grade,
			Level:        level,
			NumQuestions: req.NumQuestions,
			Semester:     req.Semester,
			Program:      req.Program,
		})
		if err != nil {
			return nil, err
		}

		questionsJSON, err := marshalQuestions(ctx, generated.Questions)
		if err != nil {
			return nil, err
		}

		cmd.NewContent = &command.NewAiExamContent{
			NumQues:       req.NumQuestions,
			Semester:      utils.ToStringPtr(req.Semester),
			Program:       utils.ToStringPtr(req.Program),
			Extras:        &tag,
			Title:         sanitizeExamText(generated.Title),
			ShortText:     sanitizeExamText(generated.ShortText),
			QuestionsJSON: questionsJSON,
		}
	}

	created, err := s.generateCmd.Handle(ctx, cmd)
	if err != nil {
		return nil, err
	}

	log.Infof("exam.generated attempt=%d ai_exam=%d profile=%d type=%s grade=%d",
		created.Attempt.UserAiExamId(), created.AiExam.AiExamId(),
		profile.ProfileId(), validated.ExamType, grade)

	// A live exam never ships the answer key.
	return &dto.GenerateExamRes{
		Exam: dto.AttemptToResponse(created.Attempt, created.AiExam, true),
	}, nil
}

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
const MinCacheVariants = 3

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

// SubmitExam grades the sitting and folds it into the child's record.
func (s *Service) SubmitExam(ctx context.Context, req *dto.SubmitExamReq) (*dto.SubmitExamRes, error) {
	if err := ValidateSubmitExam(ctx, req); err != nil {
		return nil, err
	}
	profile, err := s.loadOwnedProfile(ctx, req.UserID, req.ProfileID)
	if err != nil {
		return nil, err
	}

	result, err := s.submitCmd.Handle(ctx, command.SubmitExamCommand{
		UserAiExamID: req.UserAiExamID,
		UserID:       profile.UserId(),
		ProfileID:    profile.ProfileId(),
		Answers:      req.Answers,
		Language:     metadata.GetClientLanguage(ctx).ToEnumLanguage(),
	})
	if err != nil {
		return nil, err
	}

	aiExam, err := s.aiExamRepo.FindByAiExamId(ctx, result.Attempt.AiExamId())
	if err != nil {
		return nil, errs.NewError(ctx, status.FAIL, nil, err)
	}

	// Submitted: the review screen needs the key to mark each question.
	return &dto.SubmitExamRes{
		Exam:  dto.AttemptToResponse(result.Attempt, aiExam, true),
		Stats: dto.StatsToSingleResponse(result.Stats),
	}, nil
}

func (s *Service) GetExam(ctx context.Context, req *dto.GetExamReq) (*dto.GetExamRes, error) {
	if err := ValidateGetExam(ctx, req); err != nil {
		return nil, err
	}
	profile, err := s.loadOwnedProfile(ctx, req.UserID, req.ProfileID)
	if err != nil {
		return nil, err
	}

	detail, err := s.getAttemptQuery.Handle(ctx, query.GetExamAttemptQuery{
		UserAiExamID: req.UserAiExamID,
		UserID:       profile.UserId(),
		ProfileID:    profile.ProfileId(),
	})
	if err != nil {
		return nil, err
	}

	submitted := detail.Attempt.UserAiExamStatus() != nil &&
		*detail.Attempt.UserAiExamStatus() == string(enum.UserAiExamStatusSubmitted)

	return &dto.GetExamRes{
		Exam:    dto.AttemptToResponse(detail.Attempt, detail.AiExam, submitted),
		Details: dto.DetailsToResponse(detail.Details),
	}, nil
}

func (s *Service) ListExams(ctx context.Context, req *dto.ListExamsReq) (*dto.ListExamsRes, error) {
	if err := ValidateListExams(ctx, req); err != nil {
		return nil, err
	}
	profile, err := s.loadOwnedProfile(ctx, req.UserID, req.ProfileID)
	if err != nil {
		return nil, err
	}

	result, err := s.listQuery.Handle(ctx, query.ListExamAttemptsQuery{
		ProfileID: profile.ProfileId(),
		ExamType:  req.ExamType,
		Status:    req.Status,
		Page:      int64(req.Page),
		Limit:     int64(req.Size),
	})
	if err != nil {
		return nil, err
	}

	return &dto.ListExamsRes{
		Exams:      dto.AttemptListToResponse(result.Attempts, result.AiExams),
		Pagination: result.Pagination,
	}, nil
}

func (s *Service) GetExamStats(ctx context.Context, req *dto.GetExamStatsReq) (*dto.GetExamStatsRes, error) {
	if err := ValidateGetExamStats(ctx, req); err != nil {
		return nil, err
	}
	profile, err := s.loadOwnedProfile(ctx, req.UserID, req.ProfileID)
	if err != nil {
		return nil, err
	}

	rows, err := s.statsQuery.Handle(ctx, query.GetExamStatsQuery{
		UserID:    profile.UserId(),
		ProfileID: profile.ProfileId(),
		ExamType:  req.ExamType,
	})
	if err != nil {
		return nil, err
	}
	return &dto.GetExamStatsRes{Stats: dto.StatsToResponse(rows)}, nil
}

// GetExamProgress returns the learning-progress chart for one child.
func (s *Service) GetExamProgress(ctx context.Context, req *dto.ExamProgressReq) (*dto.ExamProgressRes, error) {
	if err := ValidateExamProgress(ctx, req); err != nil {
		return nil, err
	}
	profile, err := s.loadOwnedProfile(ctx, req.UserID, req.ProfileID)
	if err != nil {
		return nil, err
	}

	result, err := s.progressQuery.Handle(ctx, query.GetExamProgressQuery{
		ProfileID: profile.ProfileId(),
		ExamType:  req.ExamType,
		From:      req.FromDt,
		To:        req.ToDt,
		Limit:     int64(req.Limit),
	})
	if err != nil {
		return nil, errs.NewError(ctx, status.FAIL, nil, err)
	}

	return &dto.ExamProgressRes{
		ProfileID: req.ProfileID,
		FromDt:    req.FromDt,
		ToDt:      req.ToDt,
		Tz:        req.Tz,
		ExamType:  req.ExamType,
		Limit:     req.Limit,
		Series:    result.Series,
		Summary:   result.Summary,
	}, nil
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
// Level follows the same order: what the client pinned, then what the
// child was last measured at, then the bottom of the scale.
//
// The chosen level is then clamped against the grade's ceiling HERE rather
// than deeper down, so the number that reaches the cache tag, the stored
// row and the prompt is one and the same. Clamp it only in the prompt and
// a kindergarten request for level 9 gets stored as 9, tagged as 9, and
// cached apart from the identical level-4 exam it really produced.
func (s *Service) resolvePlacement(ctx context.Context, req *dto.GenerateExamReq, examType enum.ExamType, profile *profileDomain.Profile) (int, *int, error) {
	stats, err := s.statsRepo.FindByUserProfileType(ctx, profile.UserId(), profile.ProfileId(), string(examType))
	if err != nil {
		return 0, nil, errs.NewError(ctx, status.FAIL, nil, err)
	}

	grade, err := s.resolveGrade(ctx, req, stats, profile)
	if err != nil {
		return 0, nil, err
	}

	level := enum.ExamLevelMin
	switch {
	case req.Level != nil:
		level = *req.Level
	case stats != nil && stats.ResLevel() != nil:
		level = *stats.ResLevel()
	}
	level = domainBot.ClampLevelToGrade(grade, level)

	return grade, &level, nil
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
