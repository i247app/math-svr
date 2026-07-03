package exercise

import (
	"context"
	"encoding/json"

	botAdapter "math-ai.com/math-ai/internal/adapter/bot"
	"math-ai.com/math-ai/internal/adapter/storage"
	command "math-ai.com/math-ai/internal/application/command/exercise"
	dto "math-ai.com/math-ai/internal/application/dto/exercise"
	query "math-ai.com/math-ai/internal/application/query/exercise"
	"math-ai.com/math-ai/internal/application/transaction"
	classroomDomain "math-ai.com/math-ai/internal/domain/classroom"
	domain "math-ai.com/math-ai/internal/domain/exercise"
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
	getExerciseQuery   *query.GetClassroomExerciseByIdQueryHandler
	listExercisesQuery *query.ListClassroomExercisesQueryHandler
	createExerciseCmd  *command.CreateClassroomExerciseCommandHandler
	updateExerciseCmd  *command.UpdateClassroomExerciseCommandHandler
	softDeleteCmd      *command.SoftDeleteClassroomExerciseCommandHandler

	getSubmissionQuery      *query.GetSubmissionByIdQueryHandler
	listSubmissionsQuery    *query.ListSubmissionsQueryHandler
	submitAnswersCmd        *command.SubmitExerciseAnswersCommandHandler
	submitAnswersV2Cmd      *command.SubmitExerciseAnswersV2CommandHandler
	softDeleteSubmissionCmd *command.SoftDeleteSubmissionCommandHandler

	exerciseRepo         domain.IRepository
	submissionRepo       domain.ISubmissionRepository
	classroomRepo        classroomDomain.IRepository
	classroomMemberRepo  classroomDomain.IMemberRepository
	classroomProgramRepo classroomDomain.IClassroomProgramRepository
	profileRepo          profileDomain.IRepository
	programRepo          programDomain.IRepository
	gradeRepo            gradeDomain.IRepository
	bot                  *botClient
	storageProvider      *storage.Adapter
}

func NewService(
	exerciseRepo domain.IRepository,
	submissionRepo domain.ISubmissionRepository,
	uow transaction.UnitOfWork,
	bot *botAdapter.Adapter,
	classroomRepo classroomDomain.IRepository,
	classroomMemberRepo classroomDomain.IMemberRepository,
	classroomProgramRepo classroomDomain.IClassroomProgramRepository,
	profileRepo profileDomain.IRepository,
	programRepo programDomain.IRepository,
	gradeRepo gradeDomain.IRepository,
	storageProvider *storage.Adapter,
) *Service {
	return &Service{
		getExerciseQuery:   query.NewGetClassroomExerciseByIdQueryHandler(exerciseRepo),
		listExercisesQuery: query.NewListClassroomExercisesQueryHandler(exerciseRepo),
		createExerciseCmd:  command.NewCreateClassroomExerciseCommandHandler(uow),
		updateExerciseCmd:  command.NewUpdateClassroomExerciseCommandHandler(uow),
		softDeleteCmd:      command.NewSoftDeleteClassroomExerciseCommandHandler(uow),

		getSubmissionQuery:      query.NewGetSubmissionByIdQueryHandler(submissionRepo),
		listSubmissionsQuery:    query.NewListSubmissionsQueryHandler(submissionRepo),
		submitAnswersCmd:        command.NewSubmitExerciseAnswersCommandHandler(uow),
		submitAnswersV2Cmd:      command.NewSubmitExerciseAnswersV2CommandHandler(uow),
		softDeleteSubmissionCmd: command.NewSoftDeleteSubmissionCommandHandler(uow),

		exerciseRepo:         exerciseRepo,
		submissionRepo:       submissionRepo,
		classroomRepo:        classroomRepo,
		classroomMemberRepo:  classroomMemberRepo,
		classroomProgramRepo: classroomProgramRepo,
		profileRepo:          profileRepo,
		programRepo:          programRepo,
		gradeRepo:            gradeRepo,
		bot:                  newBotClient(bot),
		storageProvider:      storageProvider,
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
		return nil, errs.NewError(ctx, status.CLASSROOM_NOT_FOUND, nil, ErrClassroomNotFound)
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

	description := ""
	if req.Description != nil {
		description = *req.Description
	}
	generated, err := s.bot.GenerateExercise(ctx, generateExerciseInput{
		Language:     lang,
		GradeLabel:   gradeLabel,
		ProgramLabel: programLabel,
		Description:  description,
		ChapterName:  req.ChapterName,
		LessonName:   req.LessonName,
		NumQuestions: numQuestions,
	})
	if err != nil {
		return nil, err
	}

	questionsJSON, err := json.Marshal(generated.Questions)
	if err != nil {
		return nil, errs.NewError(ctx, status.CLASSROOM_EXERCISE_GENERATION_FAILED, nil, err)
	}
	answersJSON := buildAnswerKey(generated.Questions)

	actor := caller.ProfileId()
	questionsStr := string(questionsJSON)
	answersStr := string(answersJSON)
	visibility := string(enum.ClassroomExerciseVisibilityPublic)
	if req.Visibility != nil && *req.Visibility != "" {
		visibility = *req.Visibility
	}
	purpose := string(enum.ClassroomExercisePurposeHomework)
	if req.Purpose != nil && *req.Purpose != "" {
		purpose = *req.Purpose
	}
	saved, err := s.createExerciseCmd.Handle(ctx, command.CreateClassroomExerciseCommand{
		ActorID:          &actor,
		ClassroomID:      req.ClassroomID,
		CreatorProfileID: caller.ProfileId(),
		Visibility:       visibility,
		Purpose:          purpose,
		ProgramID:        programID,
		Title:            req.Title,
		ShortText:        sanitizeExerciseText(generated.ShortText),
		Description:      req.Description,
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
		return nil, errs.NewError(ctx, status.CLASSROOM_EXERCISE_NOT_FOUND, nil, ErrClassroomExerciseNotFound)
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
		Description:         req.Description,
		ChapterName:         req.ChapterName,
		LessonName:          req.LessonName,
		StartDate:           req.StartDate,
		EndDate:             req.EndDate,
		Note:                req.Note,
		ExerciseStatus:      req.ExerciseStatus,
		Visibility:          req.Visibility,
		Purpose:             req.Purpose,
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
		return nil, errs.NewError(ctx, status.CLASSROOM_EXERCISE_NOT_FOUND, nil, ErrClassroomExerciseNotFound)
	}

	// caller, err := s.resolveCaller(ctx, req.ProfileID, sessionUserID)
	// if err != nil {
	// 	return nil, err
	// }
	// callerMember, err := s.requireMember(ctx, exercise.ClassroomId(), caller.ProfileId())
	// if err != nil {
	// 	return nil, err
	// }
	// if err := requirePrivateAccess(ctx, exercise, caller.ProfileId()); err != nil {
	// 	return nil, err
	// }

	resp := dto.DomainToResponse(exercise, true)
	exercises := []*domain.Exercise{exercise}
	responses := []*dto.ExerciseResponse{resp}
	if err := s.hydrateSubmissionStatus(ctx, req.ProfileID, exercises, responses); err != nil {
		return nil, err
	}
	if err := s.hydrateClassroomAndProgram(ctx, exercises, responses); err != nil {
		return nil, err
	}
	return &dto.GetExerciseRes{
		// Exercise: dto.DomainToResponse(exercise, isManagerRole(callerMember)),
		Exercise: resp,
	}, nil
}

func (s *Service) ListExercises(ctx context.Context, req *dto.ListExercisesReq, sessionUserID int64) (*dto.ListExercisesRes, error) {
	if err := ValidateListExercises(ctx, req); err != nil {
		return nil, err
	}
	// caller, err := s.resolveCaller(ctx, req.ProfileID, sessionUserID)
	// if err != nil {
	// 	return nil, err
	// }

	var callerProfileId int64
	if req.ProfileID != nil {
		callerProfileId = *req.ProfileID
	}

	callerMember, err := s.requireMember(ctx, req.ClassroomID, callerProfileId)
	if err != nil {
		return nil, err
	}

	exercises, pg, err := s.listExercisesQuery.Handle(ctx, query.ListClassroomExercisesQuery{
		ClassroomID:      req.ClassroomID,
		CallerProfileID:  callerProfileId,
		Status:           req.Status,
		Visibility:       req.Visibility,
		CreatorProfileID: req.CreatorProfileID,
		ProgramID:        req.ProgramID,
		ChapterName:      req.ChapterName,
		LessonName:       req.LessonName,
		Search:           req.Search,
		Purpose:          req.Purpose,
		SortBy:           req.SortBy,
		SortOrder:        req.SortOrder,
		Page:             int64(req.Page),
		Limit:            int64(req.Size),
	})
	if err != nil {
		return nil, errs.NewError(ctx, status.FAIL, nil, err)
	}
	responses := dto.DomainListToResponse(exercises, isManagerRole(callerMember))
	// Hydrate from the resolved caller's profile id, not from the raw
	// req.ProfileID. Even when the caller omits it, resolveCaller has
	// inferred the acting profile, so the field reflects the student's
	// own submission state consistently.
	// callerProfileID := caller.ProfileId()
	if err := s.hydrateSubmissionStatus(ctx, &callerProfileId, exercises, responses); err != nil {
		return nil, err
	}
	if err := s.hydrateClassroomAndProgram(ctx, exercises, responses); err != nil {
		return nil, err
	}
	return &dto.ListExercisesRes{
		Exercises:  responses,
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
		return nil, errs.NewError(ctx, status.CLASSROOM_EXERCISE_NOT_FOUND, nil, ErrClassroomExerciseNotFound)
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
