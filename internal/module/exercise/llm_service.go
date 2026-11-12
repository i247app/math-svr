package exercise

import (
	"context"

	"math-ai.com/math-ai/internal/application/dto"
	"math-ai.com/math-ai/internal/domain/prompt/input"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/infrastructure/logger"
	llmProvider "math-ai.com/math-ai/internal/provider/llm"
	"math-ai.com/math-ai/internal/shared/enum"
)

// llmClient is the classroomexercise module's typed surface over the
// generic LLM provider. It owns prompt construction (delegated to
// domain/prompt/input via BuildExercisePrompt) and parses the response through
// application/dto (aiquestion_*.go), which owns the shared MCQ
// JSON-mode contract.
//
// Temperature 0.2 matches QuizPromptKindGenerate — generation creativity
// stays moderate so the model can vary phrasing across calls but won't
// drift the schema.
type llmClient struct {
	provider *llmProvider.Registry
}

func newLLMClient(provider *llmProvider.Registry) *llmClient {
	return &llmClient{provider: provider}
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
	Questions []dto.Question
}

func (c *llmClient) GenerateExercise(ctx context.Context, in generateExerciseInput) (*generateExerciseOutput, error) {
	log := logger.From(ctx)
	if c.provider == nil {
		return nil, errs.NewError(ctx, status.LLM_CONFIG_INVALID, nil,
			ErrLLMProviderNotConfigured)
	}

	system, user, err := input.BuildExercisePrompt(
		input.ExercisePromptKindGenerate,
		input.ExercisePromptInput{
			Language:     input.QuizLanguage(normalizeLanguage(in.Language)),
			Grade:        in.GradeLabel,
			Program:      in.ProgramLabel,
			Description:  in.Description,
			ChapterName:  in.ChapterName,
			LessonName:   in.LessonName,
			NumQuestions: in.NumQuestions,
		})
	if err != nil {
		return nil, errs.NewError(ctx, status.CLASSROOM_EXERCISE_GENERATION_FAILED, nil, err)
	}

	log.Infof("PROMPT CLASSROOM EXERCISE: system=%s user=%s", system, user)

	res, err := c.provider.Chat(ctx, llmProvider.ChatRequest{
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
	gen, err := dto.ParseGeneration(res.Content)
	if err != nil {
		log.Warnf("classroom_exercise.llm.parse_failed err=%v", err)
		return nil, errs.NewError(ctx, status.CLASSROOM_EXERCISE_GENERATION_FAILED, map[string]any{"reason": err.Error()}, err)
	}

	// Clamp render discriminators and report icon-token drift, so the
	// visual contract is enforced identically for classroom exercises.
	// Grading is label-based, so this never changes correctness.
	questions, warnings := dto.Normalize(gen.Questions)
	for _, w := range warnings {
		log.Warnf("classroom_exercise.normalize.%s value=%q q=%d", w.Kind, w.Value, w.QuestionNumber)
	}

	return &generateExerciseOutput{ShortText: gen.ShortText, Questions: questions}, nil
}

type gradeExerciseInput struct {
	Language  enum.LanguageType
	Questions string
	Answers   string
}

func (c *llmClient) GradeExercise(ctx context.Context, in gradeExerciseInput) (*dto.GradingResult, error) {
	log := logger.From(ctx)
	if c.provider == nil {
		return nil, errs.NewError(ctx, status.LLM_CONFIG_INVALID, nil,
			ErrClassroomExerciseSubmissionLLMProviderNotConfigured)
	}

	system, user, err := input.BuildExercisePrompt(
		input.ExercisePromptKindGrade,
		input.ExercisePromptInput{
			Language:  input.QuizLanguage(normalizeLanguage(in.Language)),
			Questions: in.Questions,
			Answers:   in.Answers,
		})
	if err != nil {
		return nil, errs.NewError(ctx, status.CLASSROOM_EXERCISE_SUBMISSION_GRADING_FAILED, nil, err)
	}

	log.Infof("PROMPT CLASSROOM EXERCISE SUBMISSION GRADE: system=%s user=%s", system, user)

	res, err := c.provider.Chat(ctx, llmProvider.ChatRequest{
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
		return nil, errs.NewError(ctx, status.CLASSROOM_EXERCISE_SUBMISSION_GRADING_FAILED,
			map[string]any{"reason": err.Error()}, err)
	}
	return grading, nil
}
