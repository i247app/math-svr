package classroomexercisesubmission

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

// botClient is the submission module's typed surface over the bot
// adapter. Grading uses Temperature 0.1 + JSON mode — same shape as
// quiz grading — and reuses the quiz module's salvage-tolerant
// ParseGradedQuiz so the response contract is unified.
type botClient struct {
	adapter *botAdapter.Adapter
}

func newBotClient(adapter *botAdapter.Adapter) *botClient {
	return &botClient{adapter: adapter}
}

type gradeExerciseInput struct {
	Language  enum.LanguageType
	Questions string
	Answers   string
}

func (c *botClient) GradeExercise(ctx context.Context, in gradeExerciseInput) (*quizDto.QuizGradingResult, error) {
	log := logger.From(ctx)
	if c.adapter == nil {
		return nil, errs.NewError(ctx, status.BOT_CONFIG_INVALID, nil,
			errors.New("classroom exercise submission: bot adapter is not configured"))
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

	grading, err := quizParser.ParseGradedQuiz(res.Content)
	if err != nil {
		log.Warnf("classroom_exercise_submission.bot.parse_failed err=%v", err)
		return nil, errs.NewError(ctx, status.CLASSROOM_EXERCISE_SUBMISSION_GRADING_FAILED,
			map[string]any{"reason": err.Error()}, err)
	}
	return grading, nil
}
