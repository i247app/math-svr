package exam

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"strings"
	"unicode/utf8"

	"math-ai.com/math-ai/internal/application/dto"
	routeQuery "math-ai.com/math-ai/internal/application/query/llmroute"
	"math-ai.com/math-ai/internal/domain/llmroute"
	llmProvider "math-ai.com/math-ai/internal/provider/llm"

	domain "math-ai.com/math-ai/internal/domain/exam"
	promptDomain "math-ai.com/math-ai/internal/domain/prompt"
	input "math-ai.com/math-ai/internal/domain/prompt/build"
	"math-ai.com/math-ai/internal/infrastructure/logger"
	"math-ai.com/math-ai/internal/shared/enum"
	"math-ai.com/math-ai/internal/shared/errs"
	"math-ai.com/math-ai/internal/shared/utils"
)

// llmClient is the exam module's typed surface over the generic LLM
// provider: prompt in, validated questions out.
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
	provider *llmProvider.Registry
	// routes picks the provider per kind of call (ma_llm_routes); nil =
	// every call goes through the registry's default.
	routes *routeQuery.ResolveLLMRouteQuery
}

func newLLMClient(provider *llmProvider.Registry) *llmClient {
	return &llmClient{provider: provider}
}

// llmRoute is the provider (and optional model) one call goes through.
// The zero value is the registry's default provider with its env model.
type llmRoute = routeQuery.Route

// route resolves the provider of one kind of call; zero without a resolver.
func (c *llmClient) route(ctx context.Context, code llmroute.Code) llmRoute {
	if c.routes == nil {
		return llmRoute{}
	}
	r := c.routes.Handle(ctx, code)
	logger.From(ctx).Infof("exam.llm.route code=%s provider=%s model=%s source=%s", code, r.Provider, r.Model, r.Source)
	return r
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
	// Level is the stated level (clamped for GRADE); nil = not stated. With
	// SubLevel it decides the level pair of every question
	// (input.QuestionLevels); a GRADE prompt also reads it as {{LEVEL}}.
	Level *int
	// SubLevel is the stated sub-level; nil = not stated.
	SubLevel *int
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
	Questions []dto.Question
	// RespID is the journey's new chain pointer; nil breaks the chain.
	RespID *string
}

func (c *llmClient) GenerateExam(ctx context.Context, in generateExamInput) (*generateExamOutput, error) {
	log := logger.From(ctx)
	if c.provider == nil {
		return nil, errs.NewError(ctx, llmProvider.LLMConfigInvalid, nil, ErrLLMProviderNotConfigured)
	}

	// Drawn once: the user message states these pairs and the server stamps
	// the same ones (an ASSESSMENT level is a random pick from its set).
	n := in.NumQuestions
	if n <= 0 {
		n = DefaultNumQuestions
	}
	levels, subLevels := input.QuestionLevels(in.ExamType, n, in.Level, in.SubLevel)

	promptIn := input.ExamPromptInput{
		ExamType:          in.ExamType,
		Grade:             in.Grade,
		NumQuestions:      in.NumQuestions,
		Semester:          in.Semester,
		Program:           in.Program,
		Level:             in.Level,
		SubLevel:          in.SubLevel,
		Practice:          in.Practice,
		QuestionLevels:    levels,
		QuestionSubLevels: subLevels,
		// Read from the stored text before rendering fills it in.
		SystemDefinesLevels: in.Prompt != nil && strings.Contains(in.Prompt.SystemPrompt(), input.PlaceholderLevel),
		Avoid:               in.Avoid,
	}

	if in.Prompt == nil || in.Prompt.SystemPrompt() == "" {
		return nil, errs.NewError(ctx, promptDomain.PromptNotFound, nil,
			fmt.Errorf("exam: no system prompt for %s grade %d", in.ExamType, in.Grade))
	}
	system := input.RenderSystemPrompt(in.Prompt.SystemPrompt(), in.Level, rand.Int64N(1_000_000))
	if input.HasUnfilledLevel(system) {
		log.Warnf("exam.prompt.level_unfilled prompt_id=%d type=%s grade=%d — the prompt names %s but the round has no level",
			in.Prompt.PromptId(), in.ExamType, in.Grade, input.PlaceholderLevel)
	}
	user, err := input.BuildExamUserPrompt(promptIn)
	if err != nil {
		return nil, errs.NewError(ctx, domain.ExamGenerationFailed, nil, err)
	}

	// The full system text is logged on purpose: a rewrite overwrites the
	// row, so this line is the only record of what a past set was made from.
	log.Infof("PROMPT GENERATE EXAM: code=%s grade=%d prompt_id=%d prompt_version=%d system=%s user=%s",
		in.Prompt.PromptCode(), in.Grade, in.Prompt.PromptId(), in.Prompt.PromptVersion(), system, user)

	// Every exam type that reaches the model has a route; another type
	// (none today) would go through the default.
	var route llmRoute
	if code, ok := llmroute.ExamCode(in.ExamType); ok {
		route = c.route(ctx, code)
	}
	content, respID, err := c.complete(ctx, route, system, user, 0.4, in.Chain)
	if err != nil {
		return nil, err
	}

	log.Infof("LLM RESPONSE: %s", content)

	// Parsed with the exam vocabulary (question_text, options[{id, text}],
	// correct_answer, ...) and converted straight into the shared shape, which is what the
	// normalisers and the scorer operate on.
	gen, err := dto.ParseGenerationOf[dto.ExamQuestion](content)
	if err != nil {
		log.Warnf("exam.llm.generate_parse_failed err=%v", err)
		return nil, errs.NewError(ctx, domain.ExamGenerationFailed,
			map[string]any{"reason": err.Error()}, err)
	}
	if len(gen.Questions) == 0 {
		return nil, errs.NewError(ctx, domain.ExamNoQuestions, nil, ErrModelReturnedNothing)
	}

	// The model writes single-choice questions only: whatever question_kind
	// it claims, the key it gave is one option id. Ids are normalised ("A",
	// "a " and "a" are the same option) so the list and the key agree, and
	// the key is moved to the option whose text the model gave as
	// right_answer_text when the two disagree (its working wins). A round
	// written under an older field vocabulary (a prompt not rewritten yet)
	// parses as questions without a correct_answer, or options without ids
	// — refused here, before anything is stored. The explanation and tags
	// are cleaned to the bank's rules (NormalizeExtras), never refused.
	for i := range gen.Questions {
		q := &gen.Questions[i]
		q.QuestionKind = dto.KindSingleChoice
		q.NormalizeIDs()
		q.NormalizeExtras()
		if dropped := q.NormalizeTranslations(); len(dropped) > 0 {
			log.Warnf("exam.llm.translation_dropped q=%d langs=%v", q.QuestionNumber, dropped)
		}
		if q.ReconcileKey() {
			log.Warnf("exam.llm.key_moved_to_answer_text q=%d option=%s", q.QuestionNumber, q.CorrectAnswer.OptionIDs[0])
		}
	}
	if err := dto.CheckSet(gen.Questions); err != nil {
		log.Warnf("exam.llm.generate_unusable err=%v", err)
		return nil, errs.NewError(ctx, domain.ExamGenerationFailed,
			map[string]any{"reason": err.Error()}, err)
	}

	// Two normalisation passes, and they answer different questions.
	//
	// The first clamps render types and reports icon drift: cosmetic, never
	// fatal, since grading is label-based.
	questions, warnings := dto.Normalize(dto.ToSharedQuestions(gen.Questions))
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
	questions, levelMismatches := StampQuestionLevels(questions, levels, subLevels)
	for _, m := range levelMismatches {
		log.Warnf("exam.question_level.mismatch q=%d model=%v/%v applied=%d/%d", m.QuestionNumber,
			utils.DerefInt(m.ModelLevel), utils.DerefInt(m.ModelSubLevel), m.AppliedLevel, m.AppliedSubLevel)
	}

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
func (c *llmClient) complete(ctx context.Context, route llmRoute, system, user string, temperature float64, chain *string) (string, *string, error) {
	log := logger.From(ctx)

	// The chain is only usable on the provider this turn goes through: a
	// route on a stateless provider takes the Chat path and returns a nil
	// id, so the journey's pointer is cleared rather than sent elsewhere.
	if chain == nil || !c.provider.SupportsRespond(route.Provider) {
		res, err := c.provider.Chat(ctx, llmProvider.ChatRequest{
			Provider: route.Provider,
			Model:    route.Model,
			Messages: []llmProvider.Message{
				{Role: llmProvider.RoleSystem, Content: system},
				{Role: llmProvider.RoleUser, Content: user},
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

	req := llmProvider.RespondRequest{
		Provider: route.Provider,
		Model:    route.Model,
		// Instructions are not inherited along the chain: resent every turn.
		Instructions:       system,
		Input:              []llmProvider.Message{{Role: llmProvider.RoleUser, Content: user}},
		PreviousResponseID: *chain,
		Store:              true,
		Temperature:        temperature,
		TopP:               -1,
		JSONMode:           true,
	}
	res, err := c.provider.Respond(ctx, req)
	if err != nil && req.PreviousResponseID != "" && llmProvider.IsChainPointerGone(err) {
		log.Warnf("exam.chain.gone previous_response_id=%s, starting a new chain", req.PreviousResponseID)
		req.PreviousResponseID = ""
		res, err = c.provider.Respond(ctx, req)
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
	if c.provider == nil {
		return nil, errs.NewError(ctx, llmProvider.LLMConfigInvalid, nil, ErrLLMProviderNotConfigured)
	}
	if prompt == nil || prompt.SystemPrompt() == "" {
		return nil, errs.NewError(ctx, promptDomain.PromptNotFound, nil, fmt.Errorf("exam: no system prompt for the journey review"))
	}
	system := prompt.SystemPrompt()

	user, err := input.BuildJourneyReviewUserPrompt(in)
	if err != nil {
		return nil, errs.NewError(ctx, domain.ExamReviewNoAnswers, nil, err)
	}

	// The full system text is logged on purpose: a rewrite overwrites the
	// row, so this line is the only record of what a past review was made from.
	log.Infof("PROMPT REVIEW JOURNEY: prompt_id=%d prompt_version=%d system=%s user=%s",
		prompt.PromptId(), prompt.PromptVersion(), system, user)

	content, respID, err := c.complete(ctx, c.route(ctx, llmroute.CodeExamReview), system, user, 0.4, &chain)
	if err != nil {
		return nil, err
	}

	var out struct {
		Short string `json:"short"`
		Long  string `json:"long"`
	}
	if err := json.Unmarshal([]byte(stripJSONFence(content)), &out); err != nil {
		log.Warnf("exam.review.parse_failed err=%v len=%d", err, len(content))
		return nil, errs.NewError(ctx, domain.ExamReviewFailed, nil, err)
	}
	short, long := strings.TrimSpace(out.Short), strings.TrimSpace(out.Long)
	if short == "" || long == "" {
		return nil, errs.NewError(ctx, domain.ExamReviewFailed, nil, ErrReviewEmpty)
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
