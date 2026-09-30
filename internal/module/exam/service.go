package exam

import (
	"context"
	"fmt"

	botAdapter "math-ai.com/math-ai/internal/adapter/bot"
	command "math-ai.com/math-ai/internal/application/command/exam"
	"math-ai.com/math-ai/internal/application/command/shared/practice"
	dto "math-ai.com/math-ai/internal/application/dto/exam"
	"math-ai.com/math-ai/internal/application/dto/question"
	query "math-ai.com/math-ai/internal/application/query/exam"
	"math-ai.com/math-ai/internal/application/transaction"
	examDomain "math-ai.com/math-ai/internal/domain/exam"
	gradeDomain "math-ai.com/math-ai/internal/domain/grade"
	profileDomain "math-ai.com/math-ai/internal/domain/profile"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	userDomain "math-ai.com/math-ai/internal/domain/user"
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
	markJourneyCmd  *command.MarkExamSessionCommandHandler
	getAttemptQuery *query.GetExamAttemptQueryHandler
	getJourneyQuery *query.GetExamJourneyQueryHandler
	listQuery       *query.ListExamAttemptsQueryHandler
	statsQuery      *query.GetExamStatsQueryHandler
	progressQuery   *query.GetExamProgressQueryHandler
	journeyProgress *query.GetJourneyProgressQueryHandler
	gradeLevels     *query.GetGradeLevelsQueryHandler

	examPoolRepo examDomain.IExamPoolRepository
	attemptRepo  examDomain.IExamLinkRepository
	statsRepo    examDomain.IExamSessionRepository
	detailRepo   examDomain.IExamSessionLineRepository
	profileRepo  profileDomain.IRepository
	gradeRepo    gradeDomain.IRepository

	bot   *botClient
	guest *guestService

	// shuffleEnabled mirrors EXAM_SHUFFLE_ENABLED. It gates only the
	// DRAWING of a new ordering; reading one back is never gated, or the
	// switch would re-order papers that were already handed out.
	shuffleEnabled bool
}

// NewService wires the module. bot may be nil — a deploy without LLM
// credentials still boots, and generation then fails with a uniform
// BOT_CONFIG_INVALID instead of a panic.
func NewService(
	examPoolRepo examDomain.IExamPoolRepository,
	attemptRepo examDomain.IExamLinkRepository,
	statsRepo examDomain.IExamSessionRepository,
	detailRepo examDomain.IExamSessionLineRepository,
	uow transaction.UnitOfWork,
	bot *botAdapter.Adapter,
	profileRepo profileDomain.IRepository,
	gradeRepo gradeDomain.IRepository,
	userRepo userDomain.IRepository,
	shuffleEnabled bool,
) *Service {
	return &Service{
		generateCmd:     command.NewGenerateExamCommandHandler(uow),
		submitCmd:       command.NewSubmitExamCommandHandler(uow),
		markJourneyCmd:  command.NewMarkExamSessionCommandHandler(uow),
		getAttemptQuery: query.NewGetExamAttemptQueryHandler(attemptRepo, examPoolRepo, detailRepo),
		getJourneyQuery: query.NewGetExamJourneyQueryHandler(statsRepo, attemptRepo, examPoolRepo, detailRepo),
		listQuery:       query.NewListExamAttemptsQueryHandler(attemptRepo, examPoolRepo),
		statsQuery:      query.NewGetExamStatsQueryHandler(statsRepo, attemptRepo, examPoolRepo),
		progressQuery:   query.NewGetExamProgressQueryHandler(attemptRepo),
		journeyProgress: query.NewGetJourneyProgressQueryHandler(statsRepo),
		gradeLevels:     query.NewGetGradeLevelsQueryHandler(statsRepo),
		examPoolRepo:    examPoolRepo,
		attemptRepo:     attemptRepo,
		statsRepo:       statsRepo,
		detailRepo:      detailRepo,
		profileRepo:     profileRepo,
		gradeRepo:       gradeRepo,
		bot:             newBotClient(bot),
		guest:           newGuestService(uow, userRepo, profileRepo),
		shuffleEnabled:  shuffleEnabled,
	}
}

// GenerateExam hands one exam to one child. Two very different rounds
// share the entry point:
//
//   - ASSESSMENT (and GRADE): place the child, look for a reusable
//     question set, and only call the model when there isn't one.
//   - PRACTICE: aim a fresh round at the child's latest sitting in the
//     named journey. Always a model call — the round is built from one
//     child's mistakes and can never be shared.
//
// Either way the bot call sits OUTSIDE any transaction — an LLM round trip
// must never hold one open.
func (s *Service) GenerateExam(ctx context.Context, req *dto.GenerateExamReq) (*dto.GenerateExamRes, error) {
	log := logger.From(ctx)

	validated, err := ValidateGenerateExam(ctx, req)
	if err != nil {
		return nil, err
	}
	profile, err := s.loadOwnedProfile(ctx, req.UID, req.ProfileID)
	if err != nil {
		return nil, err
	}
	if err := s.guardGuest(ctx, profile, validated.ExamType); err != nil {
		return nil, err
	}

	if validated.ExamType == enum.ExamTypePractice {
		return s.generatePractice(ctx, req, profile)
	}

	grade, openJourney, err := s.resolvePlacement(ctx, req, validated.ExamType, profile)
	if err != nil {
		return nil, err
	}
	level := resolveLevel(ctx, validated.ExamType, grade, req.Level)

	// promptLevel is the intensity the paper is written at. Only a GRADE
	// review reads it; for anything else the level is recorded and no
	// more, so it must not fork the cache pool either.
	var promptLevel *int
	if validated.ExamType == enum.ExamTypeGrade {
		promptLevel = level
	}

	tag := BuildCacheTag(validated.ExamType, grade, promptLevel, req.NumQuestions, req.Semester, req.Program)

	cmd := command.GenerateExamCommand{
		UID:         profile.Uid(),
		ProfileID:   profile.ProfileId(),
		ExamType:    validated.ExamType,
		Grade:       grade,
		Level:       level,
		StatedGrade: req.Grade,
		StatedLevel: level,
		EsessID:     openJourney,
	}

	cached, err := s.findReusableExam(ctx, tag, profile.ProfileId())
	if err != nil {
		return nil, err
	}
	// canonical is the question set as stored — the thing the child's own
	// ordering is drawn against below. On a cache hit it is decoded from
	// the shared row; on a miss it is what the model just produced.
	var canonical []question.Question

	if cached != nil {
		log.Infof("exam.cache.hit tag=%s exam_id=%d", tag, cached.ExamId())
		id := cached.ExamId()
		cmd.ReuseExamID = &id
		canonical, err = decodeStoredQuestions(ctx, cached.AiQuestionsJson())
		if err != nil {
			return nil, err
		}
	} else {
		log.Infof("exam.cache.miss tag=%s", tag)
		avoid, err := s.recentStems(ctx, profile.ProfileId(), grade)
		if err != nil {
			return nil, err
		}
		generated, err := s.bot.GenerateExam(ctx, generateExamInput{
			ExamType:     validated.ExamType,
			Grade:        grade,
			NumQuestions: req.NumQuestions,
			Semester:     req.Semester,
			Program:      req.Program,
			Level:        promptLevel,
			Avoid:        avoid,
		})
		if err != nil {
			return nil, err
		}
		content, err := newContentFrom(ctx, req, generated, &tag, promptLevel)
		if err != nil {
			return nil, err
		}
		cmd.NewContent = content
		canonical = generated.Questions
	}

	return s.handOut(ctx, cmd, canonical, profile)
}

// generatePractice draws a PRACTICE round on a finished journey.
//
// The journey's owning row — ASSESSMENT or GRADE, whichever the journey
// was opened as — is the gate: it must exist, belong to this child, and be
// COMPLETE — practice is what comes after a run has been
// finished and measured; an open journey is still being measured, and a
// cancelled one was abandoned. The round is then aimed at the journey's
// latest SUBMITTED sitting, of any type: its grade is the round's grade,
// and its answer log is what the model is told to respond to.
//
// Nothing here touches the cache. A practice set is built from one
// child's mistakes; storing it under a cache tag would serve those
// mistakes to the next child. The row is written with no req_extras so
// it can never be picked up as a variant.
func (s *Service) generatePractice(ctx context.Context, req *dto.GenerateExamReq, profile *profileDomain.Profile) (*dto.GenerateExamRes, error) {
	log := logger.From(ctx)
	journeyID := *req.EsessID

	// By id alone: PRACTICE drills either kind of journey, so naming a type
	// here would refuse every journey of the other kind as "not found".
	journey, err := s.statsRepo.FindByEsessId(ctx, journeyID)
	if err != nil {
		return nil, errs.NewError(ctx, status.FAIL, nil, err)
	}
	if journey == nil {
		return nil, errs.NewError(ctx, status.EXAM_JOURNEY_NOT_FOUND, nil,
			fmt.Errorf("exam: journey %d not found", journeyID))
	}
	if journey.Uid() != profile.Uid() || journey.ProfileId() != profile.ProfileId() {
		return nil, errs.NewError(ctx, status.EXAM_JOURNEY_NOT_OWNED, nil,
			fmt.Errorf("exam: journey %d belongs to another profile", journeyID))
	}
	if st := journey.EsessStatus(); st == nil || *st != string(enum.EsessStatusComplete) {
		return nil, errs.NewError(ctx, status.EXAM_JOURNEY_NOT_COMPLETE, nil,
			fmt.Errorf("exam: journey %d is %s; practice needs it COMPLETE", journeyID, utils.DerefString(st)))
	}

	base, err := s.attemptRepo.FindLatestSubmittedByEsessId(ctx, journeyID)
	if err != nil {
		return nil, errs.NewError(ctx, status.FAIL, nil, err)
	}
	if base == nil {
		return nil, errs.NewError(ctx, status.EXAM_PRACTICE_NO_BASE, nil,
			fmt.Errorf("exam: journey %d has no submitted sitting to practise from", journeyID))
	}
	details, err := s.detailRepo.ListByElinkId(ctx, base.ElinkId())
	if err != nil {
		return nil, errs.NewError(ctx, status.FAIL, nil, err)
	}
	brief := practice.BuildBrief(details)

	// The round follows the sitting it is drawn from, not a pinned grade:
	// a drill at a different grade than the miss would not be a drill.
	grade := base.ReqGrade()
	if req.Grade != nil && *req.Grade != grade {
		log.Warnf("exam.practice.grade_ignored pinned=%d base=%d journey=%d", *req.Grade, grade, journeyID)
	}

	avoid, err := s.recentStems(ctx, profile.ProfileId(), grade)
	if err != nil {
		return nil, err
	}
	generated, err := s.bot.GenerateExam(ctx, generateExamInput{
		ExamType:     enum.ExamTypePractice,
		Grade:        grade,
		NumQuestions: req.NumQuestions,
		Semester:     req.Semester,
		Program:      req.Program,
		Practice:     &brief,
		Avoid:        avoid,
	})
	if err != nil {
		return nil, err
	}
	content, err := newContentFrom(ctx, req, generated, nil, nil)
	if err != nil {
		return nil, err
	}

	log.Infof("exam.practice.drawn journey=%d base=%d mode=%s wrong=%d weak=%v",
		journeyID, base.ElinkId(), brief.Mode, len(brief.Wrong), brief.WeakTopics)

	return s.handOut(ctx, command.GenerateExamCommand{
		UID:        profile.Uid(),
		ProfileID:  profile.ProfileId(),
		ExamType:   enum.ExamTypePractice,
		Grade:      grade,
		Level:      req.Level,
		EsessID:    &journeyID,
		NewContent: content,
	}, generated.Questions, profile)
}

// handOut is the tail every round shares: draw this sitting's ordering,
// persist the attempt, and return the served view without the key.
//
// Every sitting gets its own ordering — the freshly generated one too, so
// the child who paid for the generation is treated no differently from
// the next child who is served the same set from cache. This is what
// stops two children (or one child twice) meeting an identical paper,
// and it is what EXAM_SHUFFLE_ENABLED turns off.
func (s *Service) handOut(ctx context.Context, cmd command.GenerateExamCommand, canonical []question.Question, profile *profileDomain.Profile) (*dto.GenerateExamRes, error) {
	shuffleJSON, err := s.drawShuffle(ctx, canonical)
	if err != nil {
		return nil, err
	}
	cmd.ShuffleJSON = shuffleJSON

	created, err := s.generateCmd.Handle(ctx, cmd)
	if err != nil {
		return nil, err
	}

	logger.From(ctx).Infof("exam.generated attempt=%d exam_pool=%d profile=%d type=%s grade=%d",
		created.Attempt.ElinkId(), created.ExamPool.ExamId(),
		profile.ProfileId(), cmd.ExamType, cmd.Grade)

	// A live exam never ships the answer key.
	return &dto.GenerateExamRes{
		Exam: dto.AttemptToResponse(created.Attempt, created.ExamPool, true),
	}, nil
}

// SubmitExam grades the sitting and folds it into the child's record.
func (s *Service) SubmitExam(ctx context.Context, req *dto.SubmitExamReq) (*dto.SubmitExamRes, error) {
	if err := ValidateSubmitExam(ctx, req); err != nil {
		return nil, err
	}
	profile, err := s.loadOwnedProfile(ctx, req.UID, req.ProfileID)
	if err != nil {
		return nil, err
	}

	result, err := s.submitCmd.Handle(ctx, command.SubmitExamCommand{
		ElinkID:   req.ElinkID,
		UID:       profile.Uid(),
		ProfileID: profile.ProfileId(),
		Answers:   req.Answers,
		Language:  metadata.GetClientLanguage(ctx).ToEnumLanguage(),
	})
	if err != nil {
		return nil, err
	}

	examPool, err := s.examPoolRepo.FindByExamId(ctx, result.Attempt.ExamId())
	if err != nil {
		return nil, errs.NewError(ctx, status.FAIL, nil, err)
	}

	// Submitted: the review screen needs the key to mark each question.
	return &dto.SubmitExamRes{
		Exam:  dto.AttemptToResponse(result.Attempt, examPool, true),
		Stats: dto.StatsToSingleResponse(result.Stats),
	}, nil
}

// GetExam answers for one sitting or one journey, depending on which id
// the request carries; the validator has already guaranteed exactly one.
func (s *Service) GetExam(ctx context.Context, req *dto.GetExamReq) (*dto.GetExamRes, error) {
	if err := ValidateGetExam(ctx, req); err != nil {
		return nil, err
	}
	profile, err := s.loadOwnedProfile(ctx, req.UID, req.ProfileID)
	if err != nil {
		return nil, err
	}

	if req.EsessID > 0 {
		return s.getJourney(ctx, req.EsessID, utils.DerefString(req.ExamType), profile)
	}
	return s.getAttempt(ctx, req.ElinkID, profile)
}

func (s *Service) ListExams(ctx context.Context, req *dto.ListExamsReq) (*dto.ListExamsRes, error) {
	if err := ValidateListExams(ctx, req); err != nil {
		return nil, err
	}
	profile, err := s.loadOwnedProfile(ctx, req.UID, req.ProfileID)
	if err != nil {
		return nil, err
	}

	result, err := s.listQuery.Handle(ctx, query.ListExamAttemptsQuery{
		ProfileID: profile.ProfileId(),
		ExamType:  req.ExamType,
		EsessID:   req.EsessID,
		Status:    req.Status,
		Page:      int64(req.Page),
		Limit:     int64(req.Size),
	})
	if err != nil {
		return nil, err
	}

	return &dto.ListExamsRes{
		Exams:      dto.AttemptListToResponse(result.Attempts, result.ExamPools),
		Pagination: result.Pagination,
	}, nil
}

func (s *Service) GetExamStats(ctx context.Context, req *dto.GetExamStatsReq) (*dto.GetExamStatsRes, error) {
	if err := ValidateGetExamStats(ctx, req); err != nil {
		return nil, err
	}
	profile, err := s.loadOwnedProfile(ctx, req.UID, req.ProfileID)
	if err != nil {
		return nil, err
	}

	rows, err := s.statsQuery.Handle(ctx, query.GetExamStatsQuery{
		UID:       profile.Uid(),
		ProfileID: profile.ProfileId(),
		ExamType:  req.ExamType,
		Status:    req.JourneyExamStatus,
	})
	if err != nil {
		return nil, err
	}
	return &dto.GetExamStatsRes{Stats: dto.JourneyStatsToResponse(rows.Journeys, rows.ExamPools)}, nil
}

// MarkExamJourney ends a journey as COMPLETE or CANCEL. From then on the
// next submission of that exam type opens a fresh journey; the ended one
// stays readable as history, with every attempt and detail row it
// accumulated still pointing at it.
func (s *Service) MarkExamJourney(ctx context.Context, req *dto.MarkExamJourneyReq) (*dto.MarkExamJourneyRes, error) {
	validated, err := ValidateMarkExamJourney(ctx, req)
	if err != nil {
		return nil, err
	}
	profile, err := s.loadOwnedProfile(ctx, req.UID, req.ProfileID)
	if err != nil {
		return nil, err
	}

	journey, err := s.markJourneyCmd.Handle(ctx, command.MarkExamSessionCommand{
		EsessID:   req.EsessID,
		UID:       profile.Uid(),
		ProfileID: profile.ProfileId(),
		Status:    validated.Status,
	})
	if err != nil {
		return nil, err
	}
	return &dto.MarkExamJourneyRes{Stats: dto.StatsToSingleResponse(journey)}, nil
}

// GetExamProgress returns the learning-progress chart for one child.
func (s *Service) GetExamProgress(ctx context.Context, req *dto.ExamProgressReq) (*dto.ExamProgressRes, error) {
	if err := ValidateExamProgress(ctx, req); err != nil {
		return nil, err
	}
	profile, err := s.loadOwnedProfile(ctx, req.UID, req.ProfileID)
	if err != nil {
		return nil, err
	}

	result, err := s.progressQuery.Handle(ctx, query.GetExamProgressQuery{
		ProfileID: profile.ProfileId(),
		ExamType:  req.ExamType,
		EsessID:   req.EsessID,
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

// GetJourneyProgress returns the learning-progress chart over a child's
// journeys — one point per ma_exam_sessions row with a score — bounded by
// the same window as GetExamProgress.
func (s *Service) GetJourneyProgress(ctx context.Context, req *dto.JourneyProgressReq) (*dto.JourneyProgressRes, error) {
	if err := ValidateJourneyProgress(ctx, req); err != nil {
		return nil, err
	}
	profile, err := s.loadOwnedProfile(ctx, req.UID, req.ProfileID)
	if err != nil {
		return nil, err
	}

	result, err := s.journeyProgress.Handle(ctx, query.GetJourneyProgressQuery{
		UID:       profile.Uid(),
		ProfileID: profile.ProfileId(),
		ExamType:  req.ExamType,
		From:      req.FromDt,
		To:        req.ToDt,
		Limit:     int64(req.Limit),
	})
	if err != nil {
		return nil, errs.NewError(ctx, status.FAIL, nil, err)
	}

	return &dto.JourneyProgressRes{
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

// GetGradeLevels tells the client where the child stands on one grade's
// GRADE level ladder: the level of the journey worked last and the highest
// level ever reached, so it can offer the right lock next.
func (s *Service) GetGradeLevels(ctx context.Context, req *dto.GradeLevelsReq) (*dto.GradeLevelsRes, error) {
	if err := ValidateGradeLevels(ctx, req); err != nil {
		return nil, err
	}
	profile, err := s.loadOwnedProfile(ctx, req.UID, req.ProfileID)
	if err != nil {
		return nil, err
	}

	result, err := s.gradeLevels.Handle(ctx, query.GetGradeLevelsQuery{
		UID:       profile.Uid(),
		ProfileID: profile.ProfileId(),
		Grade:     *req.Grade,
	})
	if err != nil {
		return nil, errs.NewError(ctx, status.FAIL, nil, err)
	}

	return &dto.GradeLevelsRes{
		LatestLevel: result.LatestLevel,
		MaxLevel:    result.MaxLevel,
	}, nil
}
