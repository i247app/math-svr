package exercise

import (
	"context"

	"math-ai.com/math-ai/internal/application/dto"
	routeQuery "math-ai.com/math-ai/internal/application/query/llmroute"
	domain "math-ai.com/math-ai/internal/domain/exercise"
	"math-ai.com/math-ai/internal/domain/llmroute"
	input "math-ai.com/math-ai/internal/domain/prompt/build"
	"math-ai.com/math-ai/internal/infrastructure/logger"
	llmProvider "math-ai.com/math-ai/internal/provider/llm"
	"math-ai.com/math-ai/internal/shared/enum"
	"math-ai.com/math-ai/internal/shared/errs"
)

// llmClient is the classroomexercise module's typed surface over the
// generic LLM provider. It owns prompt construction (delegated to
// domain/prompt/build via BuildExercisePrompt) and parses the response through
// application/dto (question_parse.go), which owns the shared
// JSON-mode contract.
//
// Temperature 0.2 matches QuizPromptKindGenerate — generation creativity
// stays moderate so the model can vary phrasing across calls but won't
// drift the schema.
type llmClient struct {
	provider *llmProvider.Registry
	// routes picks the provider per kind of call (ma_llm_routes); nil =
	// every call goes through the registry's default.
	routes *routeQuery.ResolveLLMRouteQuery
}

// route resolves the provider of one kind of call; zero (= the default)
// without a resolver.
func (c *llmClient) route(ctx context.Context, code llmroute.Code) routeQuery.Route {
	if c.routes == nil {
		return routeQuery.Route{}
	}
	r := c.routes.Handle(ctx, code)
	logger.From(ctx).Infof("exercise.llm.route code=%s provider=%s model=%s source=%s", code, r.Provider, r.Model, r.Source)
	return r
}

// generateExerciseInput is the union of fields the exercise prompt
// consumes. GradeLabel and ProgramLabel are optional — the prompt
// renders only the lines that are populated.
type generateExerciseInput struct {
	Language     enum.LanguageType
	GradeLabel   string
	ProgramLabel string
	// Description is the teacher's free-form guidance — passed through
	// verbatim to the prompt's additional-context block. Empty means
	// the prompt renders no teacher-guidance line.
	Description  string
	ChapterName  string
	LessonName   string
	NumQuestions int
}

type generateExerciseOutput struct {
	// ShortText is the AI-generated short topic description (e.g. "Phép
	// cộng trong phạm vi 10"). The teacher supplies the exercise title
	// separately, so only short_text is consumed from the model here.
	ShortText string
	Questions []dto.ExerciseQuestion
}

func (c *llmClient) GenerateExercise(ctx context.Context, in generateExerciseInput) (*generateExerciseOutput, error) {
	log := logger.From(ctx)
	if c.provider == nil {
		return nil, errs.NewError(ctx, llmProvider.LLMConfigInvalid, nil,
			ErrLLMProviderNotConfigured)
	}

	system, user, err := input.BuildExercisePrompt(
		input.ExercisePromptKindGenerate,
		input.ExercisePromptInput{
			Language:     in.Language,
			Grade:        in.GradeLabel,
			Program:      in.ProgramLabel,
			Description:  in.Description,
			ChapterName:  in.ChapterName,
			LessonName:   in.LessonName,
			NumQuestions: in.NumQuestions,
		})
	if err != nil {
		return nil, errs.NewError(ctx, domain.ClassroomExerciseGenerationFailed, nil, err)
	}

	log.Infof("PROMPT CLASSROOM EXERCISE: system=%s user=%s", system, user)

	route := c.route(ctx, llmroute.CodeExerciseGenerate)
	res, err := c.provider.Chat(ctx, llmProvider.ChatRequest{
		Provider: route.Provider,
		Model:    route.Model,
		Messages: []llmProvider.Message{
			{Role: llmProvider.RoleSystem, Content: system},
			{Role: llmProvider.RoleUser, Content: user},
		},
		Temperature: 0.4,
		TopP:        -1, // temperature is the knob; top_p comes from LLM_*_TOP_P or the model default
		JSONMode:    true,
	})
	if err != nil {
		return nil, err
	}

	log.Infof("LLM RESPONSE: %s", res.Content)

	// The exercise schema shares the {short_text, questions} shape; we
	// ignore the AI title because the teacher supplies the exercise title,
	// but we keep the AI short_text as the auto-generated topic description.
	gen, err := dto.ParseGeneration[dto.ExerciseQuestion](res.Content)
	if err != nil {
		log.Warnf("classroom_exercise.llm.parse_failed err=%v", err)
		return nil, errs.NewError(ctx, domain.ClassroomExerciseGenerationFailed, map[string]any{"reason": err.Error()}, err)
	}
	if err := checkAnswerKeys(gen.Questions); err != nil {
		log.Warnf("classroom_exercise.llm.unusable err=%v", err)
		return nil, errs.NewError(ctx, domain.ClassroomExerciseGenerationFailed, map[string]any{"reason": err.Error()}, err)
	}

	return &generateExerciseOutput{ShortText: gen.ShortText, Questions: gen.Questions}, nil
}

type gradeExerciseInput struct {
	Language  enum.LanguageType
	Questions string
	Answers   string
}

func (c *llmClient) GradeExercise(ctx context.Context, in gradeExerciseInput) (*dto.GradingResult, error) {
	log := logger.From(ctx)
	if c.provider == nil {
		return nil, errs.NewError(ctx, llmProvider.LLMConfigInvalid, nil,
			ErrClassroomExerciseSubmissionLLMProviderNotConfigured)
	}

	system, user, err := input.BuildExercisePrompt(
		input.ExercisePromptKindGrade,
		input.ExercisePromptInput{
			Language:  in.Language,
			Questions: in.Questions,
			Answers:   in.Answers,
		})
	if err != nil {
		return nil, errs.NewError(ctx, domain.ClassroomExerciseSubmissionGradingFailed, nil, err)
	}

	log.Infof("PROMPT CLASSROOM EXERCISE SUBMISSION GRADE: system=%s user=%s", system, user)

	route := c.route(ctx, llmroute.CodeExerciseGrade)
	res, err := c.provider.Chat(ctx, llmProvider.ChatRequest{
		Provider: route.Provider,
		Model:    route.Model,
		Messages: []llmProvider.Message{
			{Role: llmProvider.RoleSystem, Content: system},
			{Role: llmProvider.RoleUser, Content: user},
		},
		Temperature: 0.4,
		TopP:        -1, // temperature is the knob; top_p comes from LLM_*_TOP_P or the model default
		JSONMode:    true,
	})
	if err != nil {
		return nil, err
	}

	log.Infof("LLM RESPONSE: %s", res.Content)

	grading, err := dto.ParseGrading(res.Content)
	if err != nil {
		log.Warnf("classroom_exercise_submission.llm.parse_failed err=%v", err)
		return nil, errs.NewError(ctx, domain.ClassroomExerciseSubmissionGradingFailed,
			map[string]any{"reason": err.Error()}, err)
	}
	return grading, nil
}
