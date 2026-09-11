package exercise

import (
	"context"

	botAdapter "math-ai.com/math-ai/internal/adapter/bot"
	"math-ai.com/math-ai/internal/application/dto/question"
	domainBot "math-ai.com/math-ai/internal/domain/bot"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/infrastructure/logger"
	"math-ai.com/math-ai/internal/shared/enum"
)

// botClient is the classroomexercise module's typed surface over the
// generic bot adapter. It owns prompt construction (delegated to
// domain/bot via BuildExercisePrompt) and parses the response through
// application/dto/question, the package that owns the shared MCQ
// JSON-mode contract.
//
// Temperature 0.2 matches QuizPromptKindGenerate — generation creativity
// stays moderate so the model can vary phrasing across calls but won't
// drift the schema.
type botClient struct {
	adapter *botAdapter.Adapter
}

func newBotClient(adapter *botAdapter.Adapter) *botClient {
	return &botClient{adapter: adapter}
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
	Questions []question.Question
}

func (c *botClient) GenerateExercise(ctx context.Context, in generateExerciseInput) (*generateExerciseOutput, error) {
	log := logger.From(ctx)
	if c.adapter == nil {
		return nil, errs.NewError(ctx, status.BOT_CONFIG_INVALID, nil,
			ErrBotAdapterNotConfigured)
	}

	system, user, err := domainBot.BuildExercisePrompt(
		domainBot.ExercisePromptKindGenerate,
		domainBot.ExercisePromptInput{
			Language:     domainBot.QuizLanguage(normalizeLanguage(in.Language)),
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

	res, err := c.adapter.Chat(ctx, botAdapter.ChatRequest{
		Messages: []botAdapter.Message{
			{Role: botAdapter.RoleSystem, Content: system},
			{Role: botAdapter.RoleUser, Content: user},
		},
		Temperature: 0.2,
		TopP:        0.95,
		JSONMode:    true,
	})
	if err != nil {
		return nil, err
	}

	log.Infof("BOT RESPONSE: %s", res.Content)

	// The exercise schema shares the {short_text, questions} shape; we
	// ignore the AI title because the teacher supplies the exercise title,
	// but we keep the AI short_text as the auto-generated topic description.
	gen, err := question.ParseGeneration(res.Content)
	if err != nil {
		log.Warnf("classroom_exercise.bot.parse_failed err=%v", err)
		return nil, errs.NewError(ctx, status.CLASSROOM_EXERCISE_GENERATION_FAILED, map[string]any{"reason": err.Error()}, err)
	}

	// Clamp render discriminators and report icon-token drift, so the
	// visual contract is enforced identically for classroom exercises.
	// Grading is label-based, so this never changes correctness.
	questions, warnings := question.Normalize(gen.Questions)
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

func (c *botClient) GradeExercise(ctx context.Context, in gradeExerciseInput) (*question.GradingResult, error) {
	log := logger.From(ctx)
	if c.adapter == nil {
		return nil, errs.NewError(ctx, status.BOT_CONFIG_INVALID, nil,
			ErrClassroomExerciseSubmissionBotAdapterNotConfigured)
	}

	system, user, err := domainBot.BuildExercisePrompt(
		domainBot.ExercisePromptKindGrade,
		domainBot.ExercisePromptInput{
			Language:  domainBot.QuizLanguage(normalizeLanguage(in.Language)),
			Questions: in.Questions,
			Answers:   in.Answers,
		})
	if err != nil {
		return nil, errs.NewError(ctx, status.CLASSROOM_EXERCISE_SUBMISSION_GRADING_FAILED, nil, err)
	}

	log.Infof("PROMPT CLASSROOM EXERCISE SUBMISSION GRADE: system=%s user=%s", system, user)

	res, err := c.adapter.Chat(ctx, botAdapter.ChatRequest{
		Messages: []botAdapter.Message{
			{Role: botAdapter.RoleSystem, Content: system},
			{Role: botAdapter.RoleUser, Content: user},
		},
		Temperature: 0.1,
		TopP:        0.95,
		JSONMode:    true,
	})
	if err != nil {
		return nil, err
	}

	log.Infof("BOT RESPONSE: %s", res.Content)

	grading, err := question.ParseGrading(res.Content)
	if err != nil {
		log.Warnf("classroom_exercise_submission.bot.parse_failed err=%v", err)
		return nil, errs.NewError(ctx, status.CLASSROOM_EXERCISE_SUBMISSION_GRADING_FAILED,
			map[string]any{"reason": err.Error()}, err)
	}
	return grading, nil
}
