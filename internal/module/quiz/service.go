package quiz

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	botAdapter "math-ai.com/math-ai/internal/adapter/bot"
	command "math-ai.com/math-ai/internal/application/command/quiz"
	dto "math-ai.com/math-ai/internal/application/dto/quiz"
	query "math-ai.com/math-ai/internal/application/query/quiz"
	"math-ai.com/math-ai/internal/application/transaction"
	gradeDomain "math-ai.com/math-ai/internal/domain/grade"
	profileDomain "math-ai.com/math-ai/internal/domain/profile"
	programDomain "math-ai.com/math-ai/internal/domain/program"
	domain "math-ai.com/math-ai/internal/domain/quiz"
	semesterDomain "math-ai.com/math-ai/internal/domain/semester"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/infrastructure/logger"
	"math-ai.com/math-ai/internal/infrastructure/metadata"
	"math-ai.com/math-ai/internal/shared/enum"
)

// Service is the quiz module's public façade. It composes the command /
// query handlers behind validators, owns the bot client, and handles
// curriculum-name lookups so the bot prompts always see resolved labels
// rather than raw UUIDs.
type Service struct {
	getQuizByIdQuery       *query.GetQuizByQuizIdQueryHandler
	listQuizzesQuery       *query.ListQuizzesQueryHandler
	getQuizProgressQuery   *query.GetQuizProgressQueryHandler
	createQuizCmd          *command.CreateQuizCommandHandler
	submitQuizAnswersCmd   *command.SubmitQuizAnswersCommandHandler
	submitQuizAnswersV2Cmd *command.SubmitQuizAnswersV2CommandHandler
	softDeleteQuizCmd      *command.SoftDeleteQuizCommandHandler
	quizRepo               domain.IRepository
	profileRepo            profileDomain.IRepository
	programRepo            programDomain.IRepository
	gradeRepo              gradeDomain.IRepository
	semesterRepo           semesterDomain.IRepository
	bot                    *botClient
}

// NewService wires the quiz module. botAdapter may be nil — in a deploy
// where the bot adapter is disabled, generate/submit endpoints return
// MathError(BOT_CONFIG_INVALID) so consumers see a uniform error shape.
func NewService(
	quizRepo domain.IRepository,
	uow transaction.UnitOfWork,
	bot *botAdapter.Adapter,
	profileRepo profileDomain.IRepository,
	programRepo programDomain.IRepository,
	gradeRepo gradeDomain.IRepository,
	semesterRepo semesterDomain.IRepository,
) *Service {
	return &Service{
		getQuizByIdQuery:       query.NewGetQuizByQuizIdQueryHandler(quizRepo),
		listQuizzesQuery:       query.NewListQuizzesQueryHandler(quizRepo),
		getQuizProgressQuery:   query.NewGetQuizProgressQueryHandler(quizRepo),
		createQuizCmd:          command.NewCreateQuizCommandHandler(uow),
		submitQuizAnswersCmd:   command.NewSubmitQuizAnswersCommandHandler(uow),
		submitQuizAnswersV2Cmd: command.NewSubmitQuizAnswersV2CommandHandler(uow),
		softDeleteQuizCmd:      command.NewSoftDeleteQuizCommandHandler(uow),
		quizRepo:               quizRepo,
		profileRepo:            profileRepo,
		programRepo:            programRepo,
		gradeRepo:              gradeRepo,
		semesterRepo:           semesterRepo,
		bot:                    newBotClient(bot),
	}
}

// GenerateQuiz drives the create flow: validate, load profile context,
// resolve curriculum labels, optionally load the previous quiz for
// reinforce mode, call the bot, persist the result. The bot call sits
// OUTSIDE the UoW (slow I/O must not hold a tx).
func (s *Service) GenerateQuiz(ctx context.Context, req *dto.GenerateQuizReq) (*dto.GenerateQuizRes, error) {
	log := logger.From(ctx)

	validated, err := ValidateGenerateQuiz(ctx, req)
	if err != nil {
		return nil, err
	}

	lang := metadata.GetClientLanguage(ctx).ToEnumLanguage()
	log.Infof("Languge from metadata: %s", metadata.GetClientLanguage(ctx))

	// profile_id is optional. When supplied we must find the row (an
	// explicit profile_id pointing nowhere is a client error). When
	// absent, profile stays nil and the quiz is generated anonymously.
	var profile *profileDomain.Profile
	if req.ProfileID != nil {
		p, err := s.profileRepo.FindByProfileId(ctx, *req.ProfileID)
		if err != nil {
			return nil, errs.NewError(ctx, status.FAIL, nil, err)
		}
		if p == nil {
			return nil, errs.NewError(ctx, status.PROFILE_NOT_FOUND, nil, ErrProfileNotFound)
		}
		profile = p
	} else {
		defaultProfile, err := s.profileRepo.FindDefaultProfileByUserId(ctx, *req.UserID)
		if err != nil {
			return nil, err
		}
		if defaultProfile == nil {
			return nil, errs.NewError(ctx, status.PROFILE_NOT_FOUND, nil, ErrDefaultProfileNotFound)
		}
		profile = defaultProfile
	}

	cc, err := s.resolveCurriculumContext(ctx, req, profile)
	if err != nil {
		return nil, err
	}

	genIn := generateQuizInput{
		Language:            lang,
		Purpose:             validated.Purpose,
		TypeOfQuiz:          validated.TypeOfQuiz,
		GradeLabel:          cc.GradeLabel,
		SemesterLabel:       cc.SemesterLabel,
		ProgramLabel:        cc.ProgramLabel,
		ChapterDescriptions: cc.ChapterDescriptions,
		NumQuestions:        req.NumQuestions,
	}

	if req.PreviousQuizID != nil {
		prev, err := s.quizRepo.FindByQuizId(ctx, *req.PreviousQuizID)
		if err != nil {
			return nil, errs.NewError(ctx, status.FAIL, nil, err)
		}
		if prev == nil {
			return nil, errs.NewError(ctx, status.QUIZ_PREVIOUS_NOT_FOUND, nil,
				ErrPreviousQuizNotFound)
		}
		if prev.Review() == nil || prev.Answers() == nil {
			return nil, errs.NewError(ctx, status.QUIZ_PREVIOUS_NOT_GRADED, nil, ErrPreviousQuizMustBeSubmittedBeforeGeneratingReinforceRound)
		}
		// Ownership check only applies when both sides have a profile.
		// An anonymous reinforce round (or a reinforce off an anonymous
		// previous quiz) is allowed — they share no owner to mismatch.
		if profile != nil && prev.ProfileId() != nil && *prev.ProfileId() != profile.ProfileId() {
			return nil, errs.NewError(ctx, status.QUIZ_NOT_OWNED, nil, ErrPreviousQuizDoesNotBelongToThisProfile)
		}
		genIn.PreviousQuestions = derefString(prev.Questions())
		genIn.PreviousAnswers = derefString(prev.Answers())
		genIn.PreviousReview = *prev.Review()
	}

	generated, err := s.bot.GenerateQuiz(ctx, genIn)
	if err != nil {
		return nil, err
	}

	// Clamp render discriminators + log any icon-token drift before the
	// questions are persisted. Grading is label-based, so this never
	// changes correctness — it only keeps the stored JSON renderable.
	generated.Questions = normalizeGeneratedQuestions(ctx, generated.Questions)

	questionsJSON, err := json.Marshal(generated.Questions)
	if err != nil {
		return nil, errs.NewError(ctx, status.QUIZ_GENERATION_FAILED, nil,
			fmt.Errorf("quiz: marshal questions: %w", err))
	}

	log.Infof("questionsJSON: %s", string(questionsJSON))

	// Owner fields are NULL for anonymous quizzes (no profile supplied).
	var ownerUserID, ownerProfileID *int64
	if req.UserID != nil && *req.UserID != 0 {
		ownerUserID = req.UserID
	}

	if profile != nil {
		uid := profile.UserId()
		pid := profile.ProfileId()
		ownerUserID = &uid
		ownerProfileID = &pid
	}
	created, err := s.createQuizCmd.Handle(ctx, command.CreateQuizCommand{
		UserID:          ownerUserID,
		ProfileID:       ownerProfileID,
		Purpose:         validated.Purpose,
		TypeOfQuiz:      validated.TypeOfQuiz,
		Title:           sanitizeQuizText(generated.Title),
		ShortText:       sanitizeQuizText(generated.ShortText),
		AssessmentGrade: sanitizeQuizText(generated.AssessmentGrade),
		QuestionsJSON:   string(questionsJSON),
		PreviousQuizID:  req.PreviousQuizID,
	})
	if err != nil {
		return nil, err
	}

	log.Info("quiz.generated",
		"quiz_id", created.QuizId(),
		"profile_id", created.ProfileId(),
		"purpose", created.Purpose(),
		"type_of_quiz", derefString(created.TypeOfQuiz()),
		"title", derefString(created.Title()),
		"short_text", derefString(created.ShortText()),
	)

	// Live quizzes do NOT expose right_answer — the student would see
	// the key in the same payload. After SUBMITTED, review endpoints
	// flip this flag on.
	return &dto.GenerateQuizRes{Quiz: dto.DomainToResponse(created, true)}, nil
}

// SubmitQuizAnswers grades the answers against the quiz's right-answers,
// persists them, and returns the graded quiz. Bot call sits outside the
// UoW; the UoW only does the row update.
func (s *Service) SubmitQuizAnswersCost(ctx context.Context, req *dto.SubmitQuizAnswersReq) (*dto.SubmitQuizAnswersRes, error) {
	log := logger.From(ctx)

	if err := ValidateSubmitAnswers(ctx, req); err != nil {
		return nil, err
	}

	lang := metadata.GetClientLanguage(ctx).ToEnumLanguage()

	existing, err := s.quizRepo.FindByQuizId(ctx, req.QuizID)
	if err != nil {
		return nil, errs.NewError(ctx, status.FAIL, nil, err)
	}
	if existing == nil {
		return nil, errs.NewError(ctx, status.QUIZ_NOT_FOUND, nil, ErrQuizNotFound)
	}
	if existing.QuizStatus() != nil && *existing.QuizStatus() == string(enum.QuizStatusTypeSubmitted) {
		return nil, errs.NewError(ctx, status.QUIZ_ALREADY_SUBMITTED, nil, ErrQuizHasAlreadyBeenSubmitted)
	}
	if existing.Questions() == nil || *existing.Questions() == "" {
		return nil, errs.NewError(ctx, status.QUIZ_GRADING_FAILED, nil, ErrQuizHasNoQuestionsToGrade)
	}

	answersJSON, err := json.Marshal(req.Answers)
	if err != nil {
		return nil, errs.NewError(ctx, status.QUIZ_INVALID_ANSWERS, nil, fmt.Errorf("quiz: marshal answers: %w", err))
	}

	// Derive the learning intent from the persisted column when present;
	// fall back to inferring it from previous_quiz_id for legacy rows
	// inserted before the column existed.
	typeOfQuiz := enum.QuizTypeOfQuizGeneral
	if existing.TypeOfQuiz() != nil {
		if t := enum.QuizTypeOfQuiz(*existing.TypeOfQuiz()); t.IsValid() {
			typeOfQuiz = t
		}
	} else if existing.PreviousQuizId() != nil {
		typeOfQuiz = enum.QuizTypeOfQuizReinforcement
	}

	gradeIn := gradeQuizInput{
		Language:   lang,
		Purpose:    enum.QuizPurpose(existing.Purpose()),
		TypeOfQuiz: typeOfQuiz,
		Questions:  *existing.Questions(),
		Answers:    string(answersJSON),
	}

	if typeOfQuiz == enum.QuizTypeOfQuizReinforcement && existing.ProfileId() != nil {
		// Anonymous reinforce rounds have no profile to look up; the
		// prompt's "current grade: unknown" branch handles that case.
		currentLabel, err := s.resolveCurrentGradeLabel(ctx, *existing.ProfileId())
		if err != nil {
			return nil, err
		}
		gradeIn.CurrentGrade = currentLabel
	}

	grading, err := s.bot.GradeQuiz(ctx, gradeIn)
	if err != nil {
		return nil, err
	}

	gradingUpdate := domain.GradingUpdate{
		Review: grading.Review,
	}
	if grading.AssessmentGrade != nil && *grading.AssessmentGrade != "" {
		log.Infof("quiz.submitted.assessment_grade: %s", *grading.AssessmentGrade)
		gradingUpdate.AssessmentGrade = grading.AssessmentGrade
	}
	if grading.TotalQuestions > 0 {
		log.Infof("quiz.submitted.total_questions: %d", grading.TotalQuestions)
		v := grading.TotalQuestions
		gradingUpdate.TotalQuestions = &v
	}
	if grading.CorrectNumber >= 0 {
		log.Infof("quiz.submitted.correct_number: %d", grading.CorrectNumber)
		v := grading.CorrectNumber
		gradingUpdate.CorrectNumber = &v
	}
	if grading.ScorePercentage >= 0 {
		v := grading.ScorePercentage
		gradingUpdate.ScorePercentage = &v
	}

	updated, err := s.submitQuizAnswersCmd.Handle(ctx, command.SubmitQuizAnswersCommand{
		QuizID:      req.QuizID,
		AnswersJSON: string(answersJSON),
		Grading:     gradingUpdate,
	})
	if err != nil {
		return nil, err
	}

	res := dto.DomainToResponse(updated, true)

	// Submitted quizzes are review-mode — surface right_answer so the
	// client can render correct/incorrect indicators.
	return &dto.SubmitQuizAnswersRes{Quiz: res}, nil
}

// SubmitQuizAnswers grades the answers deterministically in process —
// no bot call. It exists alongside SubmitQuizAnswers (which still calls
// the bot) so existing mobile clients keep working unchanged; the v2
// endpoint is the cost-saver path the new client is expected to adopt.
//
// Identical validation + read-after-write semantics as v1, minus:
//   - no curriculum lookup (the deterministic review needs nothing
//     beyond per-question topic tags persisted in the row);
//   - no bot client construction;
//   - no assessment_grade signal (see scorer.go design note §3).
func (s *Service) SubmitQuizAnswers(ctx context.Context, req *dto.SubmitQuizAnswersReq) (*dto.SubmitQuizAnswersRes, error) {
	log := logger.From(ctx)
	if err := ValidateSubmitAnswers(ctx, req); err != nil {
		return nil, err
	}

	lang := metadata.GetClientLanguage(ctx).ToEnumLanguage()
	log.Infof("Languge for submit: %s", lang)

	updated, err := s.submitQuizAnswersV2Cmd.Handle(ctx, command.SubmitQuizAnswersV2Command{
		QuizID:   req.QuizID,
		Answers:  req.Answers,
		Language: lang,
	})
	if err != nil {
		return nil, err
	}

	return &dto.SubmitQuizAnswersRes{Quiz: dto.DomainToResponse(updated, true)}, nil
}

func (s *Service) GetQuizByQuizId(ctx context.Context, req *dto.GetQuizByQuizIdReq) (*dto.GetQuizByQuizIdRes, error) {
	if err := ValidateGetQuiz(ctx, req); err != nil {
		return nil, err
	}

	q, err := s.getQuizByIdQuery.Handle(ctx, query.GetQuizByQuizIdQuery{QuizID: req.QuizID})
	if err != nil {
		return nil, err
	}
	if q == nil {
		return nil, errs.NewError(ctx, status.QUIZ_NOT_FOUND, nil,
			ErrQuizNotFound)
	}

	includeAnswers := false
	if s := q.QuizStatus(); s != nil && *s == string(enum.QuizStatusTypeSubmitted) {
		includeAnswers = true
	}
	return &dto.GetQuizByQuizIdRes{Quiz: dto.DomainToResponse(q, includeAnswers)}, nil
}

func (s *Service) ListQuizzes(ctx context.Context, req *dto.ListQuizzesReq) (*dto.ListQuizzesRes, error) {
	if err := ValidateListQuizzes(ctx, req); err != nil {
		return nil, err
	}

	quizzes, pg, err := s.listQuizzesQuery.Handle(ctx, query.ListQuizzesQuery{
		ProfileID: req.ProfileID,
		UserID:    req.UserID,
		Purpose:   req.Purpose,
		Page:      int64(req.Page),
		Limit:     int64(req.Size),
	})
	if err != nil {
		return nil, err
	}

	return &dto.ListQuizzesRes{
		Quizzes:    dto.DomainListToResponse(quizzes),
		Pagination: pg,
	}, nil
}

func (s *Service) SoftDeleteQuiz(ctx context.Context, req *dto.DeleteQuizReq) (*dto.DeleteQuizRes, error) {
	if err := ValidateDeleteQuiz(ctx, req); err != nil {
		return nil, err
	}

	q, err := s.getQuizByIdQuery.Handle(ctx, query.GetQuizByQuizIdQuery{QuizID: req.QuizID})
	if err != nil {
		return nil, err
	}
	if q == nil {
		return nil, errs.NewError(ctx, status.QUIZ_NOT_FOUND, nil,
			ErrQuizNotFound)
	}

	if err := s.softDeleteQuizCmd.Handle(ctx, command.SoftDeleteQuizCommand{QuizID: req.QuizID}); err != nil {
		return nil, err
	}
	return &dto.DeleteQuizRes{}, nil
}

// GetQuizProgress returns the per-profile quiz learning-progress chart.
// Validates input, enforces that the target profile belongs to the
// session user, then delegates aggregation to the query handler.
func (s *Service) GetQuizProgress(ctx context.Context, req *dto.QuizProgressReq) (*dto.QuizProgressRes, error) {
	if err := ValidateQuizProgress(ctx, req); err != nil {
		return nil, err
	}

	profile, err := s.profileRepo.FindByProfileId(ctx, req.ProfileID)
	if err != nil {
		return nil, errs.NewError(ctx, status.FAIL, nil, err)
	}
	if profile == nil {
		return nil, errs.NewError(ctx, status.PROFILE_NOT_FOUND, nil, ErrProfileNotFound)
	}
	if req.UserID == nil || profile.UserId() != *req.UserID {
		return nil, errs.NewError(ctx, status.QUIZ_ANALYTICS_PROFILE_NOT_OWNED, nil, ErrProgressProfileNotOwned)
	}

	result, err := s.getQuizProgressQuery.Handle(ctx, query.GetQuizProgressQuery{
		ProfileID: req.ProfileID,
		Purpose:   req.Purpose,
		From:      req.FromDt,
		To:        req.ToDt,
		Limit:     int64(req.Limit),
	})
	if err != nil {
		return nil, errs.NewError(ctx, status.FAIL, nil, err)
	}

	return &dto.QuizProgressRes{
		ProfileID: req.ProfileID,
		FromDt:    req.FromDt,
		ToDt:      req.ToDt,
		Tz:        req.Tz,
		Purpose:   req.Purpose,
		Limit:     req.Limit,
		Series:    result.Series,
		Summary:   result.Summary,
	}, nil
}

// curriculumContext is the resolved (request-or-profile-or-empty) view
// of the academic fields the bot prompt may render. Any field can be
// blank — the prompt template renders only the lines whose value is
// non-empty so a partial or absent context still produces a valid quiz.
//
// ChapterDescriptions carries the curriculum chapters the quiz should
// prioritise. It is client-supplied only (req.ChapterDescriptions); when
// empty the prompt degrades gracefully to grade/semester/program guidance.
type curriculumContext struct {
	ProgramLabel        string
	GradeLabel          string
	SemesterLabel       string
	ChapterDescriptions []string
}

// resolveCurriculumContext is the single source of truth for the
// request → profile → empty priority chain. Modules and ad-hoc callers
// should use this rather than rolling their own override/fallback
// dance.
//
//   - Labels passed on req take precedence as-is (no DB roundtrip).
//   - Anything still missing is filled from the profile's curriculum
//     IDs via batched IN-queries against the existing repos.
//   - Anything still missing stays empty; the prompt builder adapts.
//
// No hardcoded education-system defaults are introduced here — that's a
// deliberate constraint so the same resolver works for any future
// curriculum or non-VN deployment.
func (s *Service) resolveCurriculumContext(ctx context.Context, req *dto.GenerateQuizReq, profile *profileDomain.Profile) (curriculumContext, error) {
	cc := curriculumContext{
		ProgramLabel:        strings.TrimSpace(req.ProgramLabel),
		GradeLabel:          strings.TrimSpace(req.GradeLabel),
		SemesterLabel:       strings.TrimSpace(req.SemesterLabel),
		ChapterDescriptions: normalizeRequestChapterDescriptions(req.ChapterDescriptions),
	}

	if profile == nil {
		return cc, nil
	}

	if cc.ProgramLabel == "" && profile.ProgramId() != nil {
		programs, err := s.programRepo.ListProgramsByIds(ctx, []int64{*profile.ProgramId()})
		if err != nil {
			return cc, errs.NewError(ctx, status.FAIL, nil, err)
		}
		if len(programs) > 0 {
			cc.ProgramLabel = programs[0].Label()
		}
	}
	if cc.GradeLabel == "" && profile.GradeId() != nil {
		grades, err := s.gradeRepo.ListGradesByIds(ctx, []int64{*profile.GradeId()})
		if err != nil {
			return cc, errs.NewError(ctx, status.FAIL, nil, err)
		}
		if len(grades) > 0 {
			cc.GradeLabel = grades[0].Label()
		}
	}
	if cc.SemesterLabel == "" && profile.SemesterId() != nil {
		semesters, err := s.semesterRepo.ListSemestersByIds(ctx, []int64{*profile.SemesterId()})
		if err != nil {
			return cc, errs.NewError(ctx, status.FAIL, nil, err)
		}
		if len(semesters) > 0 {
			cc.SemesterLabel = semesters[0].Name()
		}
	}
	return cc, nil
}

// maxPromptChapters bounds how many chapter labels are forwarded to the
// LLM. Real curricula ship 5-15 chapters per (program, grade, semester)
// today; the cap is a defensive ceiling so a misconfigured curriculum
// (or an over-eager client payload) can't blow up the prompt size. The
// domain layer adds another defence (per-label length cap), so callers
// don't have to sanitize twice.
const maxPromptChapters = 20

// normalizeRequestChapterDescriptions trims each client-supplied entry,
// drops empties, and caps the slice at maxPromptChapters. Returns nil
// when no entry survives, so an all-blank payload reads the same as no
// chapter preference at all.
func normalizeRequestChapterDescriptions(raw []string) []string {
	if len(raw) == 0 {
		return nil
	}
	out := make([]string, 0, len(raw))
	for _, entry := range raw {
		trimmed := strings.TrimSpace(entry)
		if trimmed == "" {
			continue
		}
		out = append(out, trimmed)
		if len(out) >= maxPromptChapters {
			break
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// resolveCurrentGradeLabel fetches just the grade label for the profile —
// used by reinforce grading where only the current grade matters. Returns
// "" when the profile has no grade configured; the grade prompt handles
// the "unknown" case gracefully.
func (s *Service) resolveCurrentGradeLabel(ctx context.Context, profileID int64) (string, error) {
	profile, err := s.profileRepo.FindByProfileId(ctx, profileID)
	if err != nil {
		return "", errs.NewError(ctx, status.FAIL, nil, err)
	}
	if profile == nil || profile.GradeId() == nil {
		return "", nil
	}
	grades, err := s.gradeRepo.ListGradesByIds(ctx, []int64{*profile.GradeId()})
	if err != nil {
		return "", errs.NewError(ctx, status.FAIL, nil, err)
	}
	if len(grades) == 0 {
		return "", nil
	}
	return grades[0].Label(), nil
}

func derefString(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}
