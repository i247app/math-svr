package quiz

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
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
	"math-ai.com/math-ai/internal/shared/enum"
)

// Service is the quiz module's public façade. It composes the command /
// query handlers behind validators, owns the bot client, and handles
// curriculum-name lookups so the bot prompts always see resolved labels
// rather than raw UUIDs.
type Service struct {
	getQuizByIdQuery     *query.GetQuizByQuizIdQueryHandler
	listQuizzesQuery     *query.ListQuizzesByProfileIdQueryHandler
	createQuizCmd        *command.CreateQuizCommandHandler
	submitQuizAnswersCmd *command.SubmitQuizAnswersCommandHandler
	softDeleteQuizCmd    *command.SoftDeleteQuizCommandHandler
	quizRepo             domain.IRepository
	profileRepo          profileDomain.IRepository
	programRepo          programDomain.IRepository
	gradeRepo            gradeDomain.IRepository
	semesterRepo         semesterDomain.IRepository
	bot                  *botClient
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
		getQuizByIdQuery:     query.NewGetQuizByQuizIdQueryHandler(quizRepo),
		listQuizzesQuery:     query.NewListQuizzesByProfileIdQueryHandler(quizRepo),
		createQuizCmd:        command.NewCreateQuizCommandHandler(uow),
		submitQuizAnswersCmd: command.NewSubmitQuizAnswersCommandHandler(uow),
		softDeleteQuizCmd:    command.NewSoftDeleteQuizCommandHandler(uow),
		quizRepo:             quizRepo,
		profileRepo:          profileRepo,
		programRepo:          programRepo,
		gradeRepo:            gradeRepo,
		semesterRepo:         semesterRepo,
		bot:                  newBotClient(bot),
	}
}

// GenerateQuiz drives the create flow: validate, load profile context,
// resolve curriculum labels, optionally load the previous quiz for
// reinforce mode, call the bot, persist the result. The bot call sits
// OUTSIDE the UoW (slow I/O must not hold a tx).
func (s *Service) GenerateQuiz(ctx context.Context, req *dto.GenerateQuizReq) (*dto.GenerateQuizRes, error) {
	log := logger.From(ctx)

	quizType, err := ValidateGenerateQuiz(ctx, req)
	if err != nil {
		return nil, err
	}

	profile, err := s.profileRepo.FindByProfileId(ctx, req.ProfileID)
	if err != nil {
		return nil, errs.NewError(ctx, status.FAIL, nil, err)
	}
	if profile == nil {
		return nil, errs.NewError(ctx, status.PROFILE_NOT_FOUND, nil,
			errors.New("profile not found"))
	}

	// cc, err := s.resolveCurriculumContext(ctx, req, profile)
	// if err != nil {
	// 	return nil, err
	// }

	genIn := generateQuizInput{
		Language: req.Language,
		QuizType: quizType,
		// GradeLabel:    cc.GradeLabel,
		// SemesterLabel: cc.SemesterLabel,
		// ProgramLabel:  cc.ProgramLabel,
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
		if prev.ProfileId() != profile.ProfileId() {
			return nil, errs.NewError(ctx, status.QUIZ_NOT_OWNED, nil,
				errors.New("previous quiz does not belong to this profile"))
		}
		genIn.PreviousQuestions = derefString(prev.Questions())
		genIn.PreviousAnswers = derefString(prev.Answers())
		genIn.PreviousAIReview = *prev.AIReview()
	}

	questions, err := s.bot.GenerateQuiz(ctx, genIn)
	if err != nil {
		return nil, err
	}

	questionsJSON, err := json.Marshal(questions)
	if err != nil {
		return nil, errs.NewError(ctx, status.QUIZ_GENERATION_FAILED, nil,
			fmt.Errorf("quiz: marshal questions: %w", err))
	}

	created, err := s.createQuizCmd.Handle(ctx, command.CreateQuizCommand{
		UserID:         profile.UserId(),
		ProfileID:      profile.ProfileId(),
		QuizType:       quizType,
		QuestionsJSON:  string(questionsJSON),
		PreviousQuizID: req.PreviousQuizID,
	})
	if err != nil {
		return nil, err
	}

	log.Info("quiz.generated",
		"quiz_id", created.QuizId(),
		"profile_id", created.ProfileId(),
		"type", created.QuizType(),
		"reinforce", req.PreviousQuizID != nil,
	)

	// Live quizzes do NOT expose right_answer — the student would see
	// the key in the same payload. After SUBMITTED, review endpoints
	// flip this flag on.
	return &dto.GenerateQuizRes{Quiz: dto.DomainToResponse(created, false)}, nil
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

	gradeIn := gradeQuizInput{
		Language:    req.Language,
		QuizType:    enum.QuizType(existing.QuizType()),
		Questions:   *existing.Questions(),
		Answers:     string(answersJSON),
		IsReinforce: existing.PreviousQuizId() != nil,
	}

	if gradeIn.IsReinforce {
		currentLabel, err := s.resolveCurrentGradeLabel(ctx, existing.ProfileId(), req.Language)
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
		gradingUpdate.AIDetectGrade = grading.AIDetectGrade
	}
	if grading.TotalQuestions > 0 {
		v := grading.TotalQuestions
		gradingUpdate.TotalQuestions = &v
	}
	if grading.CorrectNumber >= 0 {
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

	log.Info("quiz.submitted",
		"quiz_id", updated.QuizId(),
		"profile_id", updated.ProfileId(),
		"score", grading.ScorePercentage,
	)

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

func (s *Service) ListQuizzesByProfileId(ctx context.Context, req *dto.ListQuizzesByProfileIdReq) (*dto.ListQuizzesByProfileIdRes, error) {
	if err := ValidateListQuizzes(ctx, req); err != nil {
		return nil, err
	}

	quizzes, pg, err := s.listQuizzesQuery.Handle(ctx, query.ListQuizzesByProfileIdQuery{
		ProfileID: req.ProfileID,
		Page:      int64(req.Page),
		Limit:     int64(req.Size),
	})
	if err != nil {
		return nil, err
	}

	return &dto.ListQuizzesByProfileIdRes{
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

// resolveCurriculumLabels fetches the program / grade / semester names
// in the requested language, in three batched IN-queries. Falls back to
// Vietnamese when language is empty (matches the project default).
func (s *Service) resolveCurriculumLabels(ctx context.Context,
	programID, gradeID, semesterID *uuid.UUID, lang enum.LanguageType) (gradeLabel, semesterLabel, programLabel string, err error) {

	if lang == "" {
		lang = enum.LanguageTypeVietnamese
	}

	if programID == nil {
		programLabel = "Cánh diều"
	} else {
		programs, err := s.programRepo.ListProgramsByIds(ctx, []uuid.UUID{*programID}, lang)
		if err != nil {
			return "", "", "", errs.NewError(ctx, status.FAIL, nil, err)
		}
		if len(programs) > 0 {
			programLabel = programs[0].Label()
		}
	}

	if gradeID == nil {
		gradeLabel = "Lớp 1"
	} else {
		grades, err := s.gradeRepo.ListGradesByIds(ctx, []uuid.UUID{*gradeID}, lang)
		if err != nil {
			return "", "", "", errs.NewError(ctx, status.FAIL, nil, err)
		}
		if len(grades) > 0 {
			gradeLabel = grades[0].Label()
		}

		if semesterID == nil {
			semesterLabel = "Học kỳ 1"
		} else {
			semesters, err := s.semesterRepo.ListSemestersByIds(ctx, []uuid.UUID{*semesterID}, lang)
			if err != nil {
				return "", "", "", errs.NewError(ctx, status.FAIL, nil, err)
			}
			if len(semesters) > 0 {
				semesterLabel = semesters[0].Name()
			}
		}
	}
	return gradeLabel, semesterLabel, programLabel, nil
}

// resolveCurrentGradeLabel fetches just the grade label for the profile —
// used by reinforce grading where only the current grade is needed.
func (s *Service) resolveCurrentGradeLabel(ctx context.Context, profileID uuid.UUID, lang enum.LanguageType) (string, error) {
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
	grades, err := s.gradeRepo.ListGradesByIds(ctx, []uuid.UUID{*profile.GradeId()}, lang)
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
