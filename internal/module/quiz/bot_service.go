package quiz

import (
	"context"
	"errors"
	"strings"

	botAdapter "math-ai.com/math-ai/internal/adapter/bot"
	quizDto "math-ai.com/math-ai/internal/application/dto/quiz"
	domainBot "math-ai.com/math-ai/internal/domain/bot"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/infrastructure/logger"
	"math-ai.com/math-ai/internal/shared/enum"
)

// botClient is the quiz module's typed surface over the generic bot
// adapter. It owns the prompt construction (delegated to domain/bot)
// and the JSON-parsing contract that turns the LLM's text into typed
// QuizQuestion / QuizGradingResult values for the rest of the module.
//
// Behavioural notes:
//   - Generation uses Temperature=0.2 for moderate creativity but
//     consistent shape. Grading uses 0.1 — the model's task is closer to
//     "match the answer keys" and we want determinism.
//   - JSONMode=true is passed through to the adapter; backends that
//     ignore the hint still produce JSON because the system prompt
//     mandates a strict schema.
type botClient struct {
	adapter *botAdapter.Adapter
}

func newBotClient(adapter *botAdapter.Adapter) *botClient {
	return &botClient{adapter: adapter}
}

// generateQuizInput carries every field GenerateQuiz may consume. The
// reinforce branch is selected when PreviousQuestions is non-empty, so
// callers never have to choose a purpose explicitly.
type generateQuizInput struct {
	Language          enum.LanguageType
	QuizType          enum.QuizType
	GradeLabel        string
	SemesterLabel     string
	ProgramLabel      string
	NumQuestions      int
	PreviousQuestions string
	PreviousAnswers   string
	PreviousAIReview  string
}

func (c *botClient) GenerateQuiz(ctx context.Context, in generateQuizInput) ([]quizDto.QuizQuestion, error) {
	log := logger.From(ctx)
	if c.adapter == nil {
		return nil, errs.NewError(ctx, status.BOT_CONFIG_INVALID, nil,
			errors.New("quiz: bot adapter is not configured"))
	}

	purpose := domainBot.QuizPurposeGenerate
	if strings.TrimSpace(in.PreviousQuestions) != "" {
		purpose = domainBot.QuizPurposeReinforce
	}

	system, user, err := domainBot.BuildQuizPrompt(purpose, domainBot.QuizPromptInput{
		Language:          domainBot.QuizLanguage(normalizeLanguage(in.Language)),
		Type:              domainBot.QuizType(in.QuizType),
		Grade:             in.GradeLabel,
		Semester:          in.SemesterLabel,
		Program:           in.ProgramLabel,
		NumQuestions:      in.NumQuestions,
		PreviousQuestions: in.PreviousQuestions,
		PreviousAnswers:   in.PreviousAnswers,
		PreviousAIReview:  in.PreviousAIReview,
	})
	if err != nil {
		return nil, errs.NewError(ctx, status.QUIZ_GENERATION_FAILED, nil, err)
	}

	log.Infof("PROMPT GENERATE QUIZ: system=%s user=%s", system, user)

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

	questions, err := parseGeneratedQuestions(res.Content)
	if err != nil {
		logger.From(ctx).Warnf("quiz.bot.generate_parse_failed err=%v", err)
		return nil, errs.NewError(ctx, status.QUIZ_GENERATION_FAILED,
			map[string]any{"reason": err.Error()}, err)
	}
	return questions, nil
}

type gradeQuizInput struct {
	Language     enum.LanguageType
	QuizType     enum.QuizType
	Questions    string
	Answers      string
	IsReinforce  bool
	CurrentGrade string
}

func (c *botClient) GradeQuiz(ctx context.Context, in gradeQuizInput) (*quizDto.QuizGradingResult, error) {
	log := logger.From(ctx)
	if c.adapter == nil {
		return nil, errs.NewError(ctx, status.BOT_CONFIG_INVALID, nil,
			errors.New("quiz: bot adapter is not configured"))
	}

	purpose := domainBot.QuizPurposeGrade
	if in.IsReinforce {
		purpose = domainBot.QuizPurposeGradeReinforce
	}

	system, user, err := domainBot.BuildQuizPrompt(purpose, domainBot.QuizPromptInput{
		Language:     domainBot.QuizLanguage(normalizeLanguage(in.Language)),
		Type:         domainBot.QuizType(in.QuizType),
		Questions:    in.Questions,
		Answers:      in.Answers,
		CurrentGrade: in.CurrentGrade,
	})
	if err != nil {
		return nil, errs.NewError(ctx, status.QUIZ_GRADING_FAILED, nil, err)
	}

	log.Infof("PROMPT GRADE QUIZ: system=%s user=%s", system, user)

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

	grading, err := parseGradedQuiz(res.Content)
	if err != nil {
		logger.From(ctx).Warnf("quiz.bot.grade_parse_failed err=%v", err)
		return nil, errs.NewError(ctx, status.QUIZ_GRADING_FAILED,
			map[string]any{"reason": err.Error()}, err)
	}
	return grading, nil
}
