package quiz

import (
	"context"

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
// reinforce branch is selected explicitly via TypeOfQuiz (the persisted
// learning intent) rather than inferred from PreviousQuestions, so the
// row's stored value is the single source of truth for the prompt shape.
type generateQuizInput struct {
	Language            enum.LanguageType
	Purpose             enum.QuizPurpose
	TypeOfQuiz          enum.QuizTypeOfQuiz
	GradeLabel          string
	SemesterLabel       string
	ProgramLabel        string
	ChapterDescriptions []string
	NumQuestions        int
	PreviousQuestions   string
	PreviousAnswers     string
	PreviousReview      string
}

// generateQuizOutput pairs the parsed quiz title + short_text with its
// questions. Title is the AI-generated grade/level label (e.g. "Grade 1 -
// Level 1") and ShortText is the AI-generated topic description (e.g.
// "Các số trong phạm vi 20"). Both may be empty when the model omits them
// or the payload was truncated past recovery; callers must tolerate ""
// without aborting the quiz.
type generateQuizOutput struct {
	Title     string
	ShortText string
	// AssessmentGrade is the grade level the model calibrated the quiz to
	// (one of "Kindergarten", "Grade 1".."Grade 5"). May be "" when the
	// model omits it or the payload was truncated past recovery.
	AssessmentGrade string
	Questions       []quizDto.QuizQuestion
}

func (c *botClient) GenerateQuiz(ctx context.Context, in generateQuizInput) (*generateQuizOutput, error) {
	log := logger.From(ctx)
	if c.adapter == nil {
		return nil, errs.NewError(ctx, status.BOT_CONFIG_INVALID, nil,
			ErrQuizBotAdapterNotConfigured)
	}

	// Prompt kind is driven by the persisted learning intent. We still
	// require the previous-round payloads when the intent is reinforce,
	// so a caller that asked for REINFORCEMENT without supplying prior
	// context surfaces a prompt validation error (caught below) rather
	// than silently degrading to a generate-style prompt.
	kind := domainBot.QuizPromptKindGenerate
	if in.TypeOfQuiz == enum.QuizTypeOfQuizReinforcement {
		kind = domainBot.QuizPromptKindReinforce
	}

	system, user, err := domainBot.BuildQuizPrompt(kind, domainBot.QuizPromptInput{
		Language:            domainBot.QuizLanguage(normalizeLanguage(in.Language)),
		Purpose:             domainBot.QuizPurpose(in.Purpose),
		TypeOfQuiz:          domainBot.QuizTypeOfQuiz(in.TypeOfQuiz),
		Grade:               in.GradeLabel,
		Semester:            in.SemesterLabel,
		Program:             in.ProgramLabel,
		ChapterDescriptions: in.ChapterDescriptions,
		NumQuestions:        in.NumQuestions,
		PreviousQuestions:   in.PreviousQuestions,
		PreviousAnswers:     in.PreviousAnswers,
		PreviousReview:      in.PreviousReview,
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

	title, shortText, assessmentGrade, questions, err := parseGeneration(res.Content)
	if err != nil {
		logger.From(ctx).Warnf("quiz.bot.generate_parse_failed err=%v", err)
		return nil, errs.NewError(ctx, status.QUIZ_GENERATION_FAILED,
			map[string]any{"reason": err.Error()}, err)
	}
	return &generateQuizOutput{
		Title:           title,
		ShortText:       shortText,
		AssessmentGrade: assessmentGrade,
		Questions:       questions,
	}, nil
}

// gradeQuizInput pairs the persisted row's purpose + learning intent
// with the payloads being graded. The grading prompt branches on
// TypeOfQuiz the same way GenerateQuiz does, so an explicit
// REINFORCEMENT round is graded with the reinforce-shaped prompt even
// when the caller doesn't pass IsReinforce manually.
type gradeQuizInput struct {
	Language     enum.LanguageType
	Purpose      enum.QuizPurpose
	TypeOfQuiz   enum.QuizTypeOfQuiz
	Questions    string
	Answers      string
	CurrentGrade string
}

func (c *botClient) GradeQuiz(ctx context.Context, in gradeQuizInput) (*quizDto.QuizGradingResult, error) {
	log := logger.From(ctx)
	if c.adapter == nil {
		return nil, errs.NewError(ctx, status.BOT_CONFIG_INVALID, nil, ErrQuizBotAdapterNotConfigured)
	}

	kind := domainBot.QuizPromptKindGrade
	if in.TypeOfQuiz == enum.QuizTypeOfQuizReinforcement {
		kind = domainBot.QuizPromptKindGradeReinforce
	}

	system, user, err := domainBot.BuildQuizPrompt(kind, domainBot.QuizPromptInput{
		Language:     domainBot.QuizLanguage(normalizeLanguage(in.Language)),
		Purpose:      domainBot.QuizPurpose(in.Purpose),
		TypeOfQuiz:   domainBot.QuizTypeOfQuiz(in.TypeOfQuiz),
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
		return nil, errs.NewError(ctx, status.QUIZ_GRADING_FAILED, map[string]any{"reason": err.Error()}, err)
	}
	return grading, nil
}
