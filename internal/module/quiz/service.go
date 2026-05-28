package quiz

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	botAdapter "math-ai.com/math-ai/internal/adapter/bot"
	command "math-ai.com/math-ai/internal/application/command/quiz"
	dto "math-ai.com/math-ai/internal/application/dto/quiz"
	query "math-ai.com/math-ai/internal/application/query/quiz"
	"math-ai.com/math-ai/internal/application/transaction"
	chapterDomain "math-ai.com/math-ai/internal/domain/chapter"
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
	getQuizByIdQuery     *query.GetQuizByQuizIdQueryHandler
	listQuizzesQuery     *query.ListQuizzesQueryHandler
	createQuizCmd        *command.CreateQuizCommandHandler
	submitQuizAnswersCmd *command.SubmitQuizAnswersCommandHandler
	softDeleteQuizCmd    *command.SoftDeleteQuizCommandHandler
	quizRepo             domain.IRepository
	profileRepo          profileDomain.IRepository
	programRepo          programDomain.IRepository
	gradeRepo            gradeDomain.IRepository
	semesterRepo         semesterDomain.IRepository
	chapterRepo          chapterDomain.IRepository
	bot                  *botClient
}

// NewService wires the quiz module. botAdapter may be nil — in a deploy
// where the bot adapter is disabled, generate/submit endpoints return
// MathError(BOT_CONFIG_INVALID) so consumers see a uniform error shape.
// chapterRepo enables chapter-aware prompts: when the profile pins a
// (program, grade, semester) triple, the resolver hydrates the matching
// chapter labels and feeds them into the bot prompt. The arg may be nil
// in dev/test wiring — chapter injection then silently no-ops.
func NewService(
	quizRepo domain.IRepository,
	uow transaction.UnitOfWork,
	bot *botAdapter.Adapter,
	profileRepo profileDomain.IRepository,
	programRepo programDomain.IRepository,
	gradeRepo gradeDomain.IRepository,
	semesterRepo semesterDomain.IRepository,
	chapterRepo chapterDomain.IRepository,
) *Service {
	return &Service{
		getQuizByIdQuery:     query.NewGetQuizByQuizIdQueryHandler(quizRepo),
		listQuizzesQuery:     query.NewListQuizzesQueryHandler(quizRepo),
		createQuizCmd:        command.NewCreateQuizCommandHandler(uow),
		submitQuizAnswersCmd: command.NewSubmitQuizAnswersCommandHandler(uow),
		softDeleteQuizCmd:    command.NewSoftDeleteQuizCommandHandler(uow),
		quizRepo:             quizRepo,
		profileRepo:          profileRepo,
		programRepo:          programRepo,
		gradeRepo:            gradeRepo,
		semesterRepo:         semesterRepo,
		chapterRepo:          chapterRepo,
		bot:                  newBotClient(bot),
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
			return nil, errs.NewError(ctx, status.PROFILE_NOT_FOUND, nil,
				errors.New("profile not found"))
		}
		profile = p
	}

	cc, err := s.resolveCurriculumContext(ctx, req, profile)
	if err != nil {
		return nil, err
	}

	genIn := generateQuizInput{
		Language:            metadata.GetClientLanguage(ctx).ToEnumLanguage(),
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
				errors.New("previous quiz not found"))
		}
		if prev.AIReview() == nil || prev.Answers() == nil {
			return nil, errs.NewError(ctx, status.QUIZ_PREVIOUS_NOT_GRADED, nil,
				errors.New("previous quiz must be submitted before generating a reinforce round"))
		}
		// Ownership check only applies when both sides have a profile.
		// An anonymous reinforce round (or a reinforce off an anonymous
		// previous quiz) is allowed — they share no owner to mismatch.
		if profile != nil && prev.ProfileId() != nil && *prev.ProfileId() != profile.ProfileId() {
			return nil, errs.NewError(ctx, status.QUIZ_NOT_OWNED, nil,
				errors.New("previous quiz does not belong to this profile"))
		}
		genIn.PreviousQuestions = derefString(prev.Questions())
		genIn.PreviousAnswers = derefString(prev.Answers())
		genIn.PreviousAIReview = *prev.AIReview()
	}

	generated, err := s.bot.GenerateQuiz(ctx, genIn)
	if err != nil {
		return nil, err
	}

	questionsJSON, err := json.Marshal(generated.Questions)
	if err != nil {
		return nil, errs.NewError(ctx, status.QUIZ_GENERATION_FAILED, nil,
			fmt.Errorf("quiz: marshal questions: %w", err))
	}

	log.Infof("questionsJSON: %s", string(questionsJSON))

	// Owner fields are NULL for anonymous quizzes (no profile supplied).
	var ownerUserID, ownerProfileID *string
	if req.UserID != nil && *req.UserID != "" {
		ownerUserID = req.UserID
	}

	if profile != nil {
		uid := profile.UserId()
		pid := profile.ProfileId()
		ownerUserID = &uid
		ownerProfileID = &pid
	}
	created, err := s.createQuizCmd.Handle(ctx, command.CreateQuizCommand{
		UserID:         ownerUserID,
		ProfileID:      ownerProfileID,
		Purpose:        validated.Purpose,
		TypeOfQuiz:     validated.TypeOfQuiz,
		Title:          sanitizeQuizTitle(generated.Title),
		QuestionsJSON:  string(questionsJSON),
		PreviousQuizID: req.PreviousQuizID,
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
	)

	// Live quizzes do NOT expose right_answer — the student would see
	// the key in the same payload. After SUBMITTED, review endpoints
	// flip this flag on.
	return &dto.GenerateQuizRes{Quiz: dto.DomainToResponse(created, true)}, nil
}

// SubmitQuizAnswers grades the answers against the quiz's right-answers,
// persists them, and returns the graded quiz. Bot call sits outside the
// UoW; the UoW only does the row update.
func (s *Service) SubmitQuizAnswers(ctx context.Context, req *dto.SubmitQuizAnswersReq) (*dto.SubmitQuizAnswersRes, error) {
	log := logger.From(ctx)

	if err := ValidateSubmitAnswers(ctx, req); err != nil {
		return nil, err
	}

	existing, err := s.quizRepo.FindByQuizId(ctx, req.QuizID)
	if err != nil {
		return nil, errs.NewError(ctx, status.FAIL, nil, err)
	}
	if existing == nil {
		return nil, errs.NewError(ctx, status.QUIZ_NOT_FOUND, nil,
			errors.New("quiz not found"))
	}
	if existing.QuizStatus() != nil && *existing.QuizStatus() == string(enum.QuizStatusTypeSubmitted) {
		return nil, errs.NewError(ctx, status.QUIZ_ALREADY_SUBMITTED, nil,
			errors.New("quiz has already been submitted"))
	}
	if existing.Questions() == nil || *existing.Questions() == "" {
		return nil, errs.NewError(ctx, status.QUIZ_GRADING_FAILED, nil,
			errors.New("quiz has no questions to grade"))
	}

	answersJSON, err := json.Marshal(req.Answers)
	if err != nil {
		return nil, errs.NewError(ctx, status.QUIZ_INVALID_ANSWERS, nil,
			fmt.Errorf("quiz: marshal answers: %w", err))
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
		Language:   req.Language,
		Purpose:    enum.QuizPurpose(existing.Purpose()),
		TypeOfQuiz: typeOfQuiz,
		Questions:  *existing.Questions(),
		Answers:    string(answersJSON),
	}

	if typeOfQuiz == enum.QuizTypeOfQuizReinforcement && existing.ProfileId() != nil {
		// Anonymous reinforce rounds have no profile to look up; the
		// prompt's "current grade: unknown" branch handles that case.
		currentLabel, err := s.resolveCurrentGradeLabel(ctx, *existing.ProfileId(), req.Language)
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
		AIReview: grading.AIReview,
	}
	if grading.AIDetectGrade != nil && *grading.AIDetectGrade != "" {
		log.Infof("quiz.submitted.ai_detect_grade: %s", *grading.AIDetectGrade)
		gradingUpdate.AIDetectGrade = grading.AIDetectGrade
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

	// Submitted quizzes are review-mode — surface right_answer so the
	// client can render correct/incorrect indicators.
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
			errors.New("quiz not found"))
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
			errors.New("quiz not found"))
	}

	if err := s.softDeleteQuizCmd.Handle(ctx, command.SoftDeleteQuizCommand{QuizID: req.QuizID}); err != nil {
		return nil, err
	}
	return &dto.DeleteQuizRes{}, nil
}

// curriculumContext is the resolved (request-or-profile-or-empty) view
// of the academic fields the bot prompt may render. Any field can be
// blank — the prompt template renders only the lines whose value is
// non-empty so a partial or absent context still produces a valid quiz.
//
// ChapterDescriptions carries the curriculum chapters the quiz should
// prioritise. It is filled in priority order:
//  1. req.ChapterDescriptions — pinned by the client (works for
//     anonymous quizzes that have no profile at all).
//  2. The profile's (program, grade, semester) triple via the chapter
//     repo — derived only when all three IDs are pinned because chapter
//     rows are keyed on the full triple.
//  3. Empty — the prompt degrades gracefully to grade/semester/program
//     guidance only.
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

	lang := metadata.GetClientLanguage(ctx).ToEnumLanguage()

	if cc.ProgramLabel == "" && profile.ProgramId() != nil {
		programs, err := s.programRepo.ListProgramsByIds(ctx, []string{*profile.ProgramId()}, lang)
		if err != nil {
			return cc, errs.NewError(ctx, status.FAIL, nil, err)
		}
		if len(programs) > 0 {
			cc.ProgramLabel = programs[0].Label()
		}
	}
	if cc.GradeLabel == "" && profile.GradeId() != nil {
		grades, err := s.gradeRepo.ListGradesByIds(ctx, []string{*profile.GradeId()}, lang)
		if err != nil {
			return cc, errs.NewError(ctx, status.FAIL, nil, err)
		}
		if len(grades) > 0 {
			cc.GradeLabel = grades[0].Label()
		}
	}
	if cc.SemesterLabel == "" && profile.SemesterId() != nil {
		semesters, err := s.semesterRepo.ListSemestersByIds(ctx, []string{*profile.SemesterId()}, lang)
		if err != nil {
			return cc, errs.NewError(ctx, status.FAIL, nil, err)
		}
		if len(semesters) > 0 {
			cc.SemesterLabel = semesters[0].Name()
		}
	}

	if len(cc.ChapterDescriptions) == 0 {
		cc.ChapterDescriptions = s.resolveProfileChapterDescriptions(ctx, profile, lang)
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
// when no entry survives so the resolver can fall back to the
// profile-derived list with a simple len() == 0 check.
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

// resolveProfileChapterLabels fetches chapters for the profile's
// (program, grade, semester) triple and returns their localized labels
// in display_order. Returns nil for every failure mode (missing
// profile, partial curriculum, repo error, no rows) — chapter
// injection is a *prompt enhancement*, never a hard requirement, so
// errors are logged at warn level and the quiz still generates with
// grade/semester/program-only context.
func (s *Service) resolveProfileChapterDescriptions(ctx context.Context,
	profile *profileDomain.Profile, lang enum.LanguageType) []string {

	if s.chapterRepo == nil || profile == nil {
		return nil
	}
	programID := profile.ProgramId()
	gradeID := profile.GradeId()
	semesterID := profile.SemesterId()
	if programID == nil || gradeID == nil || semesterID == nil {
		// Chapter rows are keyed on all three IDs; with any missing we'd
		// either over-fetch (cross-curriculum) or under-fetch (empty).
		// Skip silently — the prompt still has grade/semester/program.
		return nil
	}

	chapters, _, err := s.chapterRepo.ListChapters(ctx, &chapterDomain.ListChaptersParams{
		ProgramID:  programID,
		GradeID:    gradeID,
		SemesterID: semesterID,
		Language:   lang,
		TakeAll:    true,
	})
	if err != nil {
		logger.From(ctx).Warnf("quiz.resolve_chapters_failed program_id=%s grade_id=%s semester_id=%s err=%v",
			*programID, *gradeID, *semesterID, err)
		return nil
	}
	if len(chapters) == 0 {
		return nil
	}

	limit := min(len(chapters), maxPromptChapters)
	descriptions := make([]string, 0, limit)
	for _, c := range chapters[:limit] {
		description := strings.TrimSpace(c.Description())
		if description == "" {
			continue
		}
		descriptions = append(descriptions, description)
	}
	return descriptions
}

// resolveCurrentGradeLabel fetches just the grade label for the profile —
// used by reinforce grading where only the current grade matters. Returns
// "" when the profile has no grade configured; the grade prompt handles
// the "unknown" case gracefully.
func (s *Service) resolveCurrentGradeLabel(ctx context.Context, profileID string, lang enum.LanguageType) (string, error) {
	profile, err := s.profileRepo.FindByProfileId(ctx, profileID)
	if err != nil {
		return "", errs.NewError(ctx, status.FAIL, nil, err)
	}
	if profile == nil || profile.GradeId() == nil {
		return "", nil
	}
	if lang == "" {
		lang = enum.LanguageTypeVietnamese
	}
	grades, err := s.gradeRepo.ListGradesByIds(ctx, []string{*profile.GradeId()}, lang)
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
