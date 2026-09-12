package exam

import (
	"context"

	botAdapter "math-ai.com/math-ai/internal/adapter/bot"
	command "math-ai.com/math-ai/internal/application/command/exam"
	dto "math-ai.com/math-ai/internal/application/dto/exam"
	"math-ai.com/math-ai/internal/application/dto/question"
	query "math-ai.com/math-ai/internal/application/query/exam"
	"math-ai.com/math-ai/internal/application/transaction"
	examDomain "math-ai.com/math-ai/internal/domain/exam"
	gradeDomain "math-ai.com/math-ai/internal/domain/grade"
	profileDomain "math-ai.com/math-ai/internal/domain/profile"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/infrastructure/logger"
	"math-ai.com/math-ai/internal/infrastructure/metadata"
	"math-ai.com/math-ai/internal/shared/utils"
)

// Service is the exam module's public face. It owns the decisions that
// are neither pure persistence nor pure prompt-building: where a child is
// placed before an exam is generated, and whether a stored question set
// can be served instead of paying for a new one.
type Service struct {
	generateCmd     *command.GenerateExamCommandHandler
	submitCmd       *command.SubmitExamCommandHandler
	markJourneyCmd  *command.MarkUserExamCommandHandler
	getAttemptQuery *query.GetExamAttemptQueryHandler
	getJourneyQuery *query.GetExamJourneyQueryHandler
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
		markJourneyCmd:  command.NewMarkUserExamCommandHandler(uow),
		getAttemptQuery: query.NewGetExamAttemptQueryHandler(attemptRepo, aiExamRepo, detailRepo),
		getJourneyQuery: query.NewGetExamJourneyQueryHandler(statsRepo, attemptRepo, aiExamRepo, detailRepo),
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

	grade, err := s.resolvePlacement(ctx, req, validated.ExamType, profile)
	if err != nil {
		return nil, err
	}

	tag := BuildCacheTag(validated.ExamType, grade, req.NumQuestions, req.Semester, req.Program)

	cmd := command.GenerateExamCommand{
		UserID:    profile.UserId(),
		ProfileID: profile.ProfileId(),
		ExamType:  validated.ExamType,
		Grade:     grade,
	}

	cached, err := s.findReusableExam(ctx, tag)
	if err != nil {
		return nil, err
	}
	// canonical is the question set as stored — the thing the child's own
	// ordering is drawn against below. On a cache hit it is decoded from
	// the shared row; on a miss it is what the model just produced.
	var canonical []question.Question

	if cached != nil {
		log.Infof("exam.cache.hit tag=%s ai_exam_id=%d", tag, cached.AiExamId())
		id := cached.AiExamId()
		cmd.ReuseAiExamID = &id
		canonical, err = decodeStoredQuestions(ctx, cached.AiQuestionsJson())
		if err != nil {
			return nil, err
		}
	} else {
		log.Infof("exam.cache.miss tag=%s", tag)
		generated, err := s.bot.GenerateExam(ctx, generateExamInput{
			ExamType:     validated.ExamType,
			Grade:        grade,
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
		canonical = generated.Questions

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

	// Every sitting gets its own ordering — the freshly generated one too,
	// so the child who paid for the generation is treated no differently
	// from the next child served the same set from cache. This is what
	// stops two children (or one child twice) meeting an identical paper.
	shuffleJSON, err := drawShuffle(ctx, canonical)
	if err != nil {
		return nil, err
	}
	cmd.ShuffleJSON = shuffleJSON

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

// GetExam answers for one sitting or one journey, depending on which id
// the request carries; the validator has already guaranteed exactly one.
func (s *Service) GetExam(ctx context.Context, req *dto.GetExamReq) (*dto.GetExamRes, error) {
	if err := ValidateGetExam(ctx, req); err != nil {
		return nil, err
	}
	profile, err := s.loadOwnedProfile(ctx, req.UserID, req.ProfileID)
	if err != nil {
		return nil, err
	}

	if req.UserExamID > 0 {
		return s.getJourney(ctx, req.UserExamID, profile)
	}
	return s.getAttempt(ctx, req.UserAiExamID, profile)
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
		Status:    req.Status,
	})
	if err != nil {
		return nil, err
	}
	return &dto.GetExamStatsRes{Stats: dto.StatsToResponse(rows)}, nil
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
	profile, err := s.loadOwnedProfile(ctx, req.UserID, req.ProfileID)
	if err != nil {
		return nil, err
	}

	journey, err := s.markJourneyCmd.Handle(ctx, command.MarkUserExamCommand{
		UserExamID: req.UserExamID,
		UserID:     profile.UserId(),
		ProfileID:  profile.ProfileId(),
		Status:     validated.Status,
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
