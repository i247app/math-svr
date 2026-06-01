package classroomexercise

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	botAdapter "math-ai.com/math-ai/internal/adapter/bot"
	command "math-ai.com/math-ai/internal/application/command/classroomexercise"
	dto "math-ai.com/math-ai/internal/application/dto/classroomexercise"
	query "math-ai.com/math-ai/internal/application/query/classroomexercise"
	"math-ai.com/math-ai/internal/application/transaction"
	classroomDomain "math-ai.com/math-ai/internal/domain/classroom"
	domain "math-ai.com/math-ai/internal/domain/classroomexercise"
	gradeDomain "math-ai.com/math-ai/internal/domain/grade"
	profileDomain "math-ai.com/math-ai/internal/domain/profile"
	programDomain "math-ai.com/math-ai/internal/domain/program"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/infrastructure/logger"
	"math-ai.com/math-ai/internal/infrastructure/metadata"
	"math-ai.com/math-ai/internal/shared/enum"
)

// Service is the classroomexercise module's public façade. The
// dependency surface mirrors the classroom module: classroom + member
// repos for permission checks, classroom_program for the program-pick
// validation, profile/program/grade repos to resolve curriculum labels
// for the bot prompt, and the bot adapter itself.
type Service struct {
	getExerciseQuery  *query.GetClassroomExerciseByIdQueryHandler
	listExercisesQuery *query.ListClassroomExercisesQueryHandler
	createExerciseCmd *command.CreateClassroomExerciseCommandHandler
	updateExerciseCmd *command.UpdateClassroomExerciseCommandHandler
	softDeleteCmd     *command.SoftDeleteClassroomExerciseCommandHandler

	exerciseRepo         domain.IRepository
	classroomRepo        classroomDomain.IRepository
	classroomMemberRepo  classroomDomain.IMemberRepository
	classroomProgramRepo classroomDomain.IClassroomProgramRepository
	profileRepo          profileDomain.IRepository
	programRepo          programDomain.IRepository
	gradeRepo            gradeDomain.IRepository
	bot                  *botClient
}

func NewService(
	exerciseRepo domain.IRepository,
	uow transaction.UnitOfWork,
	bot *botAdapter.Adapter,
	classroomRepo classroomDomain.IRepository,
	classroomMemberRepo classroomDomain.IMemberRepository,
	classroomProgramRepo classroomDomain.IClassroomProgramRepository,
	profileRepo profileDomain.IRepository,
	programRepo programDomain.IRepository,
	gradeRepo gradeDomain.IRepository,
) *Service {
	return &Service{
		getExerciseQuery:     query.NewGetClassroomExerciseByIdQueryHandler(exerciseRepo),
		listExercisesQuery:   query.NewListClassroomExercisesQueryHandler(exerciseRepo),
		createExerciseCmd:    command.NewCreateClassroomExerciseCommandHandler(uow),
		updateExerciseCmd:    command.NewUpdateClassroomExerciseCommandHandler(uow),
		softDeleteCmd:        command.NewSoftDeleteClassroomExerciseCommandHandler(uow),
		exerciseRepo:         exerciseRepo,
		classroomRepo:        classroomRepo,
		classroomMemberRepo:  classroomMemberRepo,
		classroomProgramRepo: classroomProgramRepo,
		profileRepo:          profileRepo,
		programRepo:          programRepo,
		gradeRepo:            gradeRepo,
		bot:                  newBotClient(bot),
	}
}

// CreateExercise validates → checks the caller is a classroom manager →
// resolves curriculum labels for the bot → calls the bot OUTSIDE the
// UoW → persists the result inside a UoW. ProgramID, when supplied, is
// verified against the classroom's active program junction.
func (s *Service) CreateExercise(ctx context.Context, req *dto.CreateExerciseReq, sessionUserID int64) (*dto.CreateExerciseRes, error) {
	log := logger.From(ctx)

	if err := ValidateCreateExercise(ctx, req); err != nil {
		return nil, err
	}
	caller, err := s.resolveCaller(ctx, req.ProfileID, sessionUserID)
	if err != nil {
		return nil, err
	}
	if _, err := s.requireManager(ctx, req.ClassroomID, caller.ProfileId()); err != nil {
		return nil, err
	}

	classroom, err := s.classroomRepo.FindByClassroomId(ctx, req.ClassroomID)
	if err != nil {
		return nil, errs.NewError(ctx, status.FAIL, nil, err)
	}
	if classroom == nil {
		return nil, errs.NewError(ctx, status.CLASSROOM_NOT_FOUND, nil,
			errors.New("classroom not found"))
	}

	programID, err := s.resolveExerciseProgram(ctx, req.ClassroomID, req.ProgramID)
	if err != nil {
		return nil, err
	}

	lang := metadata.GetClientLanguage(ctx).ToEnumLanguage()
	if req.Language != "" {
		lang = req.Language
	}

	gradeLabel, programLabel := s.resolveCurriculumLabels(ctx, classroom.GradeId(), programID, lang)

	numQuestions := req.NumQuestions
	if numQuestions <= 0 {
		numQuestions = DefaultNumQuestions
	}

	generated, err := s.bot.GenerateExercise(ctx, generateExerciseInput{
		Language:     lang,
		GradeLabel:   gradeLabel,
		ProgramLabel: programLabel,
		ChapterName:  req.ChapterName,
		LessonName:   req.LessonName,
		NumQuestions: numQuestions,
	})
	if err != nil {
		return nil, err
	}

	questionsJSON, err := json.Marshal(generated.Questions)
	if err != nil {
		return nil, errs.NewError(ctx, status.CLASSROOM_EXERCISE_GENERATION_FAILED, nil,
			fmt.Errorf("classroom exercise: marshal questions: %w", err))
	}
	answersJSON := buildAnswerKey(generated.Questions)

	actor := caller.ProfileId()
	questionsStr := string(questionsJSON)
	answersStr := string(answersJSON)
	visibility := string(enum.ClassroomExerciseVisibilityPublic)
	if req.Visibility != nil && *req.Visibility != "" {
		visibility = *req.Visibility
	}
	saved, err := s.createExerciseCmd.Handle(ctx, command.CreateClassroomExerciseCommand{
		ActorID:          &actor,
		ClassroomID:      req.ClassroomID,
		CreatorProfileID: caller.ProfileId(),
		Visibility:       visibility,
		ProgramID:        programID,
		Title:            req.Title,
		ChapterName:      req.ChapterName,
		LessonName:       req.LessonName,
		TotalQuestions:   len(generated.Questions),
		QuestionsJSON:    &questionsStr,
		AnswersJSON:      &answersStr,
		StartDate:        req.StartDate,
		EndDate:          req.EndDate,
		Note:             req.Note,
	})
	if err != nil {
		return nil, err
	}

	log.Info("classroom_exercise.generated",
		"classroom_exercise_id", saved.ClassroomExerciseId(),
		"classroom_id", saved.ClassroomId(),
		"total_questions", saved.TotalQuestions(),
	)

	// Manager-only path — include right-answer key in the response so
	// the teacher can preview the correct labels.
	return &dto.CreateExerciseRes{Exercise: dto.DomainToResponse(saved, true)}, nil
}

func (s *Service) UpdateExercise(ctx context.Context, req *dto.UpdateExerciseReq, sessionUserID int64) (*dto.UpdateExerciseRes, error) {
	if err := ValidateUpdateExercise(ctx, req); err != nil {
		return nil, err
	}
	caller, err := s.resolveCaller(ctx, req.ProfileID, sessionUserID)
	if err != nil {
		return nil, err
	}

	existing, err := s.exerciseRepo.FindByClassroomExerciseId(ctx, req.ClassroomExerciseID)
	if err != nil {
		return nil, errs.NewError(ctx, status.FAIL, nil, err)
	}
	if existing == nil {
		return nil, errs.NewError(ctx, status.CLASSROOM_EXERCISE_NOT_FOUND, nil,
			errors.New("classroom exercise not found"))
	}
	if _, err := s.requireManager(ctx, existing.ClassroomId(), caller.ProfileId()); err != nil {
		return nil, err
	}
	if err := requirePrivateAccess(ctx, existing, caller.ProfileId()); err != nil {
		return nil, err
	}

	actor := caller.ProfileId()
	updated, err := s.updateExerciseCmd.Handle(ctx, command.UpdateClassroomExerciseCommand{
		ActorID:             &actor,
		ClassroomExerciseID: req.ClassroomExerciseID,
		Title:               req.Title,
		ChapterName:         req.ChapterName,
		LessonName:          req.LessonName,
		StartDate:           req.StartDate,
		EndDate:             req.EndDate,
		Note:                req.Note,
		ExerciseStatus:      req.ExerciseStatus,
		Visibility:          req.Visibility,
	})
	if err != nil {
		return nil, err
	}
	return &dto.UpdateExerciseRes{Exercise: dto.DomainToResponse(updated, true)}, nil
}

func (s *Service) GetExercise(ctx context.Context, req *dto.GetExerciseReq, sessionUserID int64) (*dto.GetExerciseRes, error) {
	if err := ValidateGetExercise(ctx, req); err != nil {
		return nil, err
	}

	exercise, err := s.getExerciseQuery.Handle(ctx, query.GetClassroomExerciseByIdQuery{
		ClassroomExerciseID: req.ClassroomExerciseID,
	})
	if err != nil {
		return nil, errs.NewError(ctx, status.FAIL, nil, err)
	}
	if exercise == nil {
		return nil, errs.NewError(ctx, status.CLASSROOM_EXERCISE_NOT_FOUND, nil,
			errors.New("classroom exercise not found"))
	}

	caller, err := s.resolveCaller(ctx, req.ProfileID, sessionUserID)
	if err != nil {
		return nil, err
	}
	callerMember, err := s.requireMember(ctx, exercise.ClassroomId(), caller.ProfileId())
	if err != nil {
		return nil, err
	}
	if err := requirePrivateAccess(ctx, exercise, caller.ProfileId()); err != nil {
		return nil, err
	}

	return &dto.GetExerciseRes{
		Exercise: dto.DomainToResponse(exercise, isManagerRole(callerMember)),
	}, nil
}

func (s *Service) ListExercises(ctx context.Context, req *dto.ListExercisesReq, sessionUserID int64) (*dto.ListExercisesRes, error) {
	if err := ValidateListExercises(ctx, req); err != nil {
		return nil, err
	}
	caller, err := s.resolveCaller(ctx, req.ProfileID, sessionUserID)
	if err != nil {
		return nil, err
	}
	callerMember, err := s.requireMember(ctx, req.ClassroomID, caller.ProfileId())
	if err != nil {
		return nil, err
	}

	exercises, pg, err := s.listExercisesQuery.Handle(ctx, query.ListClassroomExercisesQuery{
		ClassroomID:      req.ClassroomID,
		CallerProfileID:  caller.ProfileId(),
		Status:           req.Status,
		Visibility:       req.Visibility,
		CreatorProfileID: req.CreatorProfileID,
		ProgramID:        req.ProgramID,
		ChapterName:      req.ChapterName,
		LessonName:       req.LessonName,
		Search:           req.Search,
		SortBy:           req.SortBy,
		SortOrder:        req.SortOrder,
		Page:             int64(req.Page),
		Limit:            int64(req.Size),
	})
	if err != nil {
		return nil, errs.NewError(ctx, status.FAIL, nil, err)
	}
	return &dto.ListExercisesRes{
		Exercises:  dto.DomainListToResponse(exercises, isManagerRole(callerMember)),
		Pagination: pg,
	}, nil
}

func (s *Service) SoftDeleteExercise(ctx context.Context, req *dto.DeleteExerciseReq, sessionUserID int64) (*dto.DeleteExerciseRes, error) {
	if err := ValidateDeleteExercise(ctx, req); err != nil {
		return nil, err
	}
	caller, err := s.resolveCaller(ctx, req.ProfileID, sessionUserID)
	if err != nil {
		return nil, err
	}

	existing, err := s.exerciseRepo.FindByClassroomExerciseId(ctx, req.ClassroomExerciseID)
	if err != nil {
		return nil, errs.NewError(ctx, status.FAIL, nil, err)
	}
	if existing == nil {
		return nil, errs.NewError(ctx, status.CLASSROOM_EXERCISE_NOT_FOUND, nil,
			errors.New("classroom exercise not found"))
	}
	if _, err := s.requireManager(ctx, existing.ClassroomId(), caller.ProfileId()); err != nil {
		return nil, err
	}
	if err := requirePrivateAccess(ctx, existing, caller.ProfileId()); err != nil {
		return nil, err
	}

	actor := caller.ProfileId()
	if err := s.softDeleteCmd.Handle(ctx, command.SoftDeleteClassroomExerciseCommand{
		ActorID:             &actor,
		ClassroomExerciseID: req.ClassroomExerciseID,
	}); err != nil {
		return nil, err
	}
	return &dto.DeleteExerciseRes{}, nil
}

// resolveCaller resolves ProfileID with a fallback: when the request
// omits it, the session's authenticated user must have exactly one
// profile (otherwise the caller has to disambiguate). Mirrors how the
// classroom module derives the acting profile.
func (s *Service) resolveCaller(ctx context.Context, profileID *int64, sessionUserID int64) (*profileDomain.Profile, error) {
	if profileID != nil && *profileID != 0 {
		return s.resolveActingProfile(ctx, *profileID, sessionUserID)
	}
	// No ProfileID supplied — require the session and look up by user.
	if sessionUserID == 0 {
		return nil, errs.NewError(ctx, status.CLASSROOM_EXERCISE_PERMISSION_DENIED, nil,
			errors.New("profile_id is required"))
	}
	profiles, err := s.profileRepo.ListByUserId(ctx, sessionUserID)
	if err != nil {
		return nil, errs.NewError(ctx, status.FAIL, nil, err)
	}
	if len(profiles) == 1 {
		return profiles[0], nil
	}
	return nil, errs.NewError(ctx, status.CLASSROOM_EXERCISE_PERMISSION_DENIED, nil,
		errors.New("profile_id is required when the user owns multiple profiles"))
}

// resolveExerciseProgram verifies the supplied program belongs to the
// classroom. When the teacher omits ProgramID we leave it nil — the
// exercise will be persisted without program context and the prompt
// renders a curriculum-agnostic brief.
func (s *Service) resolveExerciseProgram(ctx context.Context, classroomID int64, programID *int64) (*int64, error) {
	if programID == nil || *programID == 0 {
		return nil, nil
	}
	ids, err := s.classroomProgramRepo.ListProgramIdsByClassroomId(ctx, classroomID)
	if err != nil {
		return nil, errs.NewError(ctx, status.FAIL, nil, err)
	}
	for _, id := range ids {
		if id == *programID {
			return programID, nil
		}
	}
	return nil, errs.NewError(ctx, status.CLASSROOM_EXERCISE_PROGRAM_NOT_IN_CLASSROOM, nil,
		errors.New("program is not associated with this classroom"))
}

// resolveCurriculumLabels best-effort hydrates grade + program labels
// for the bot prompt. Errors here are logged but do not fail the
// request — the prompt renders only the lines whose label is non-empty.
func (s *Service) resolveCurriculumLabels(ctx context.Context, gradeID, programID *int64, lang enum.LanguageType) (string, string) {
	log := logger.From(ctx)
	var gradeLabel, programLabel string

	if gradeID != nil && *gradeID != 0 {
		grades, err := s.gradeRepo.ListGradesByIds(ctx, []int64{*gradeID}, lang)
		if err != nil {
			log.Warnf("classroom_exercise.resolve_grade_failed grade_id=%d err=%v", *gradeID, err)
		} else if len(grades) > 0 {
			gradeLabel = grades[0].Label()
		}
	}
	if programID != nil && *programID != 0 {
		programs, err := s.programRepo.ListProgramsByIds(ctx, []int64{*programID}, lang)
		if err != nil {
			log.Warnf("classroom_exercise.resolve_program_failed program_id=%d err=%v", *programID, err)
		} else if len(programs) > 0 {
			programLabel = programs[0].Label()
		}
	}
	return gradeLabel, programLabel
}
