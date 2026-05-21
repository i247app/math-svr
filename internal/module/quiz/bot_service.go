package quiz

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
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
		PreviousQuestions: in.PreviousQuestions,
		PreviousAnswers:   in.PreviousAnswers,
		PreviousAIReview:  in.PreviousAIReview,
	})
	if err != nil {
		return nil, errs.NewError(ctx, status.QUIZ_GENERATION_FAILED, nil, err)
	}

	log.Infof("PROMPT: system=%s user=%s", system, user)

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

	grading, err := parseGradedQuiz(res.Content)
	if err != nil {
		logger.From(ctx).Warnf("quiz.bot.grade_parse_failed err=%v", err)
		return nil, errs.NewError(ctx, status.QUIZ_GRADING_FAILED,
			map[string]any{"reason": err.Error()}, err)
	}
	return grading, nil
}

func normalizeLanguage(lang enum.LanguageType) string {
	s := strings.ToLower(strings.TrimSpace(string(lang)))
	if s == "" {
		return string(enum.LanguageTypeEnglish)
	}
	return s
}

func parseGeneratedQuestions(content string) ([]quizDto.QuizQuestion, error) {
	payload := extractJSONPayload(content)

	var out []quizDto.QuizQuestion
	if err := json.Unmarshal([]byte(payload), &out); err == nil {
		if len(out) == 0 {
			return nil, errors.New("quiz: model returned zero questions")
		}
		return out, nil
	}

	// LLM truncated mid-array (max_tokens hit, safety cut, network
	// reset, …). Salvage the longest prefix of complete objects rather
	// than fail the whole request.
	repaired, ok := salvageTruncatedJSONArray(payload)
	if !ok {
		return nil, fmt.Errorf("quiz: parse generated questions: payload not recoverable")
	}
	if err := json.Unmarshal([]byte(repaired), &out); err != nil {
		return nil, fmt.Errorf("quiz: parse generated questions (after salvage): %w", err)
	}
	if len(out) == 0 {
		return nil, errors.New("quiz: model returned zero questions")
	}
	return out, nil
}

// salvageTruncatedJSONArray recovers the longest prefix of a JSON array
// of objects that ends at a complete top-level element. It returns
// (repaired, true) when at least one full object was found; the repaired
// string is guaranteed to be a syntactically-valid JSON array.
//
// The state machine tracks string/escape state so braces inside strings
// don't affect depth, and remembers the index of the last "}" that
// closed back to depth 0 (just inside the leading "["). On truncation
// that index is the cut point — everything after gets dropped and the
// array is re-closed with "]". Payloads that don't start with "["
// (some backends drop the wrapper) are wrapped before salvage is
// attempted.
func salvageTruncatedJSONArray(s string) (string, bool) {
	s = strings.TrimSpace(s)
	if s == "" {
		return "", false
	}
	if s[0] != '[' {
		if s[0] != '{' {
			return "", false
		}
		s = "[" + s
	}

	inStr := false
	esc := false
	depth := 0
	lastGood := -1

	for i := 1; i < len(s); i++ {
		c := s[i]
		if inStr {
			if esc {
				esc = false
				continue
			}
			switch c {
			case '\\':
				esc = true
			case '"':
				inStr = false
			}
			continue
		}
		switch c {
		case '"':
			inStr = true
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				lastGood = i
			}
		case ']':
			if depth == 0 {
				return s[:i+1], true
			}
		}
	}

	if lastGood < 0 {
		return "", false
	}
	return s[:lastGood+1] + "]", true
}

func parseGradedQuiz(content string) (*quizDto.QuizGradingResult, error) {
	payload := extractJSONPayload(content)
	var out quizDto.QuizGradingResult
	if err := json.Unmarshal([]byte(payload), &out); err != nil {
		return nil, fmt.Errorf("quiz: parse graded quiz: %w", err)
	}
	return &out, nil
}

// codeFenceRe matches ```lang\n...\n``` wrappers some backends emit even
// when JSON mode is requested. Strip them so json.Unmarshal sees the
// payload directly.
var codeFenceRe = regexp.MustCompile("(?ms)^```[a-zA-Z0-9_-]*\\n?(.*?)\\n?```$")

func extractJSONPayload(s string) string {
	s = strings.TrimSpace(s)
	if m := codeFenceRe.FindStringSubmatch(s); len(m) == 2 {
		return strings.TrimSpace(m[1])
	}
	return s
}
