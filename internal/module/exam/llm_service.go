package exam

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"strings"
	"unicode/utf8"

	llmAdapter "math-ai.com/math-ai/internal/adapter/llm"
	dto "math-ai.com/math-ai/internal/application/dto/exam"
	"math-ai.com/math-ai/internal/application/dto/question"

	// input "math-ai.com/math-ai/internal/domain/prompt/input"
	promptDomain "math-ai.com/math-ai/internal/domain/prompt"
	"math-ai.com/math-ai/internal/domain/prompt/input"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	"math-ai.com/math-ai/internal/infrastructure/logger"
	"math-ai.com/math-ai/internal/shared/enum"
	"math-ai.com/math-ai/internal/shared/utils"
)

// llmClient is the exam module's typed surface over the generic LLM
// adapter: prompt in, validated questions out.
//
// It has one method. Grading never reaches the model — the server scores
// every submission — so unlike the quiz module there is no GradeExam here
// and no grading prompt behind it.
//
// Temperature 0.2 keeps phrasing varied between calls without letting the
// schema drift; JSON mode is requested even though the system prompt
// mandates the shape anyway, because backends that honour it fail earlier
// and more cheaply.
type llmClient struct {
	adapter *llmAdapter.Adapter
}

func newLLMClient(adapter *llmAdapter.Adapter) *llmClient {
	return &llmClient{adapter: adapter}
}

type generateExamInput struct {
	// Prompt is the static system prompt (ma_prompts), sent verbatim. The
	// caller loads it once per request so the text sent and the version
	// written into the cache tag are the same row.
	Prompt       *promptDomain.Prompt
	ExamType     enum.ExamType
	Grade        int
	NumQuestions int
	Semester     string
	Program      string
	// Level is the clamped level of a GRADE paper; nil otherwise. It fills
	// the prompt's {{LEVEL}} and is stamped on every question.
	Level *int
	// Practice aims a PRACTICE round; nil for every other type.
	Practice *input.PracticeBrief
	// Avoid is the stems of the child's recent sittings at this grade.
	Avoid []string
	// Chain is the journey's last_resp_id: nil when the set belongs to no
	// journey (admin pool generation), "" for a journey with no live chain.
	Chain *string
}

type generateExamOutput struct {
	Title     string
	ShortText string
	Questions []question.Question
	// RespID is the journey's new chain pointer; nil breaks the chain.
	RespID *string
}

func (c *llmClient) GenerateExam(ctx context.Context, in generateExamInput) (*generateExamOutput, error) {
	log := logger.From(ctx)
	if c.adapter == nil {
		return nil, errs.NewError(ctx, status.LLM_CONFIG_INVALID, nil, ErrLLMAdapterNotConfigured)
	}

	promptIn := input.ExamPromptInput{
		ExamType:     in.ExamType,
		Grade:        in.Grade,
		NumQuestions: in.NumQuestions,
		Semester:     in.Semester,
		Program:      in.Program,
		Level:        in.Level,
		Practice:     in.Practice,
		// Read from the stored text before rendering fills it in.
		SystemDefinesLevels: in.Prompt != nil && strings.Contains(in.Prompt.SystemPrompt(), input.PlaceholderLevel),
		Avoid:               in.Avoid,
	}

	if in.Prompt == nil || in.Prompt.SystemPrompt() == "" {
		return nil, errs.NewError(ctx, status.PROMPT_NOT_FOUND, nil,
			fmt.Errorf("exam: no system prompt for %s grade %d", in.ExamType, in.Grade))
	}
	system := input.RenderSystemPrompt(in.Prompt.SystemPrompt(), in.Level, rand.Int64N(1_000_000))
	if input.HasUnfilledLevel(system) {
		log.Warnf("exam.prompt.level_unfilled prompt_id=%d type=%s grade=%d — the prompt names %s but the round has no level",
			in.Prompt.PromptId(), in.ExamType, in.Grade, input.PlaceholderLevel)
	}
	user, err := input.BuildExamUserPrompt(promptIn)
	if err != nil {
		return nil, errs.NewError(ctx, status.EXAM_GENERATION_FAILED, nil, err)
	}

	// The full system text is logged on purpose: a rewrite overwrites the
	// row, so this line is the only record of what a past set was made from.
	log.Infof("PROMPT GENERATE EXAM: code=%s grade=%d prompt_id=%d prompt_version=%d system=%s user=%s",
		in.Prompt.PromptCode(), in.Grade, in.Prompt.PromptId(), in.Prompt.PromptVersion(), system, user)

	content, respID, err := c.complete(ctx, system, user, 0.4, in.Chain)
	if err != nil {
		return nil, err
	}

	log.Infof("LLM RESPONSE: %s", content)

	// Parsed with the exam vocabulary (right_answer_label, question_topic,
	// ...) and converted straight into the shared shape, which is what the
	// normalisers and the scorer operate on.
	gen, err := question.ParseGenerationOf[dto.ExamQuestion](content)
	if err != nil {
		log.Warnf("exam.llm.generate_parse_failed err=%v", err)
		return nil, errs.NewError(ctx, status.EXAM_GENERATION_FAILED,
			map[string]any{"reason": err.Error()}, err)
	}
	if len(gen.Questions) == 0 {
		return nil, errs.NewError(ctx, status.EXAM_NO_QUESTIONS, nil, ErrModelReturnedNothing)
	}

	// Two normalisation passes, and they answer different questions.
	//
	// The first clamps render types and reports icon drift: cosmetic, never
	// fatal, since grading is label-based.
	questions, warnings := question.Normalize(dto.ToSharedQuestions(gen.Questions))
	for _, w := range warnings {
		log.Warnf("exam.normalize.%s value=%q q=%d", w.Kind, w.Value, w.QuestionNumber)
	}

	// The second stamps question_grade from the request. This one is not
	// cosmetic: question_grade feeds placement, so a value the model
	// invented could move a child up or down a year. The mismatches are
	// logged because a run of them means the prompt has stopped landing.
	questions, mismatches := NormalizeQuestionBands(questions, in.ExamType, in.Grade)
	for _, m := range mismatches {
		log.Warnf("exam.question_grade.mismatch q=%d model=%v applied=%d",
			m.QuestionNumber, utils.DerefInt(m.ModelGrade), m.Applied)
	}
	questions = StampQuestionLevel(questions, in.Level)

	// The title is a pure function of the type and the grade, so the server
	// stamps it (the prompt does not ask for one); the short_text is the
	// model's own summary of the round and is kept.
	return &generateExamOutput{
		Title:     input.ExamTitle(in.ExamType, in.Grade),
		ShortText: gen.ShortText,
		Questions: questions,
		RespID:    respID,
	}, nil
}

// maxChainInputTokens caps what a journey's chain may cost. Every chained
// turn is billed for the whole chain's input again, so once a turn's
// input passes this the pointer is dropped and the next turn starts a new
// chain — continuity then rests on the Avoid list, which is sent anyway.
// ponytail: fixed constant; make it config if it needs tuning per model.
const maxChainInputTokens = 30000

// complete sends one system+user turn and returns the model's text.
//
// chain nil means the turn belongs to no journey: a stateless Chat. A
// non-nil chain is a journey turn: on a provider that keeps server-side
// state (OpenAI Responses API) it is chained on *chain ("" starts a new
// chain), stored, and the new response id is returned for the caller to
// persist; on any other provider it falls back to Chat and returns a nil
// id. A chain pointer the vendor no longer holds (30-day retention) is
// retried once as a new chain.
func (c *llmClient) complete(ctx context.Context, system, user string, temperature float64, chain *string) (string, *string, error) {
	log := logger.From(ctx)

	if chain == nil || !c.adapter.SupportsRespond("") {
		res, err := c.adapter.Chat(ctx, llmAdapter.ChatRequest{
			Messages: []llmAdapter.Message{
				{Role: llmAdapter.RoleSystem, Content: system},
				{Role: llmAdapter.RoleUser, Content: user},
			},
			Temperature: temperature,
			TopP:        -1, // temperature is the knob; top_p comes from LLM_*_TOP_P or the model default
			JSONMode:    true,
		})
		if err != nil {
			return "", nil, err
		}
		return res.Content, nil, nil
	}

	req := llmAdapter.RespondRequest{
		// Instructions are not inherited along the chain: resent every turn.
		Instructions:       system,
		Input:              []llmAdapter.Message{{Role: llmAdapter.RoleUser, Content: user}},
		PreviousResponseID: *chain,
		Store:              true,
		Temperature:        temperature,
		TopP:               -1,
		JSONMode:           true,
	}
	res, err := c.adapter.Respond(ctx, req)
	if err != nil && req.PreviousResponseID != "" && llmAdapter.IsChainPointerGone(err) {
		log.Warnf("exam.chain.gone previous_response_id=%s, starting a new chain", req.PreviousResponseID)
		req.PreviousResponseID = ""
		res, err = c.adapter.Respond(ctx, req)
	}
	if err != nil {
		return "", nil, err
	}
	if res.Usage.PromptTokens > maxChainInputTokens {
		log.Infof("exam.chain.capped response_id=%s input_tokens=%d", res.ResponseID, res.Usage.PromptTokens)
		return res.Content, nil, nil
	}
	return res.Content, &res.ResponseID, nil
}

type journeyReviewOutput struct {
	Short string
	Long  string
	// RespID is the journey's new chain pointer; nil breaks the chain.
	RespID *string
}

// ReviewJourney asks the model to review the answers of one journey and
// returns the review in two lengths. Temperature 0.3: the wording may vary
// between calls, the facts it states may not. The short version is cut to
// ReviewShortMaxChars runes if the model overruns it, so it always fits
// its column; the long one goes to a TEXT column and is kept whole.
//
// prompt is the EXAM_REVIEW row of ma_prompts, sent verbatim as the system
// message. chain is the journey's last_resp_id ("" when none), so the
// review is one more turn of the journey's chain.
func (c *llmClient) ReviewJourney(ctx context.Context, prompt *promptDomain.Prompt, in input.JourneyReviewInput, chain string) (*journeyReviewOutput, error) {
	log := logger.From(ctx)
	if c.adapter == nil {
		return nil, errs.NewError(ctx, status.LLM_CONFIG_INVALID, nil, ErrLLMAdapterNotConfigured)
	}
	if prompt == nil || prompt.SystemPrompt() == "" {
		return nil, errs.NewError(ctx, status.PROMPT_NOT_FOUND, nil, fmt.Errorf("exam: no system prompt for the journey review"))
	}
	system := prompt.SystemPrompt()

	user, err := input.BuildJourneyReviewUserPrompt(in)
	if err != nil {
		return nil, errs.NewError(ctx, status.EXAM_REVIEW_NO_ANSWERS, nil, err)
	}

	// The full system text is logged on purpose: a rewrite overwrites the
	// row, so this line is the only record of what a past review was made from.
	log.Infof("PROMPT REVIEW JOURNEY: prompt_id=%d prompt_version=%d system=%s user=%s",
		prompt.PromptId(), prompt.PromptVersion(), system, user)

	content, respID, err := c.complete(ctx, system, user, 0.4, &chain)
	if err != nil {
		return nil, err
	}

	var out struct {
		Short string `json:"short"`
		Long  string `json:"long"`
	}
	if err := json.Unmarshal([]byte(stripJSONFence(content)), &out); err != nil {
		log.Warnf("exam.review.parse_failed err=%v len=%d", err, len(content))
		return nil, errs.NewError(ctx, status.EXAM_REVIEW_FAILED, nil, err)
	}
	short, long := strings.TrimSpace(out.Short), strings.TrimSpace(out.Long)
	if short == "" || long == "" {
		return nil, errs.NewError(ctx, status.EXAM_REVIEW_FAILED, nil, ErrReviewEmpty)
	}
	if utf8.RuneCountInString(short) > input.ReviewShortMaxChars {
		log.Warnf("exam.review.short_truncated runes=%d", utf8.RuneCountInString(short))
		short = string([]rune(short)[:input.ReviewShortMaxChars])
	}
	return &journeyReviewOutput{Short: short, Long: long, RespID: respID}, nil
}

// stripJSONFence tolerates a model that wraps its JSON in a Markdown code
// fence despite JSON mode.
func stripJSONFence(s string) string {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "```") {
		s = strings.TrimPrefix(s, "```json")
		s = strings.TrimPrefix(s, "```")
		s = strings.TrimSuffix(s, "```")
	}
	return strings.TrimSpace(s)
}
