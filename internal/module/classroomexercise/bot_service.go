package classroomexercise

import (
	"context"
	"errors"

	botAdapter "math-ai.com/math-ai/internal/adapter/bot"
	quizDto "math-ai.com/math-ai/internal/application/dto/quiz"
	domainBot "math-ai.com/math-ai/internal/domain/bot"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/infrastructure/logger"
	quizParser "math-ai.com/math-ai/internal/module/quiz"
	"math-ai.com/math-ai/internal/shared/enum"
)

// botClient is the classroomexercise module's typed surface over the
// generic bot adapter. It owns prompt construction (delegated to
// domain/bot via BuildExercisePrompt) and reuses the quiz module's
// parseGeneration so the JSON-mode contract stays identical to quizzes.
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
	ChapterName  string
	LessonName   string
	NumQuestions int
}

type generateExerciseOutput struct {
	Questions []quizDto.QuizQuestion
}

func (c *botClient) GenerateExercise(ctx context.Context, in generateExerciseInput) (*generateExerciseOutput, error) {
	log := logger.From(ctx)
	if c.adapter == nil {
		return nil, errs.NewError(ctx, status.BOT_CONFIG_INVALID, nil,
			errors.New("classroom exercise: bot adapter is not configured"))
	}

	system, user, err := domainBot.BuildExercisePrompt(
		domainBot.ExercisePromptKindGenerate,
		domainBot.ExercisePromptInput{
			Language:     domainBot.QuizLanguage(normalizeLanguage(in.Language)),
			Grade:        in.GradeLabel,
			Program:      in.ProgramLabel,
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

	// Reuse the quiz parser. The exercise schema is the same
	// {title, questions} shape as a quiz; we ignore the title because
	// the teacher already supplied one.
	_, questions, err := quizParser.ParseGeneration(res.Content)
	if err != nil {
		log.Warnf("classroom_exercise.bot.parse_failed err=%v", err)
		return nil, errs.NewError(ctx, status.CLASSROOM_EXERCISE_GENERATION_FAILED,
			map[string]any{"reason": err.Error()}, err)
	}
	return &generateExerciseOutput{Questions: questions}, nil
}
