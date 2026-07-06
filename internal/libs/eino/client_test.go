package eino

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	einomodel "github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
	"google.golang.org/genai"
)

// fakeChatModel is a hand-rolled einomodel.BaseChatModel double.
type fakeChatModel struct {
	generateFn func(ctx context.Context, in []*schema.Message, opts ...einomodel.Option) (*schema.Message, error)
	streamFn   func(ctx context.Context, in []*schema.Message, opts ...einomodel.Option) (*schema.StreamReader[*schema.Message], error)
	calls      int
}

func (f *fakeChatModel) Generate(ctx context.Context, in []*schema.Message, opts ...einomodel.Option) (*schema.Message, error) {
	f.calls++
	return f.generateFn(ctx, in, opts...)
}

func (f *fakeChatModel) Stream(ctx context.Context, in []*schema.Message, opts ...einomodel.Option) (*schema.StreamReader[*schema.Message], error) {
	f.calls++
	return f.streamFn(ctx, in, opts...)
}

func newTestClient(plain, jsonModel einomodel.BaseChatModel, jsonOpts []einomodel.Option) *Client {
	cfg := Config{
		Backend:     BackendGoogleAI,
		APIKey:      "test",
		Model:       "test-model",
		Temperature: -1,
		TopP:        -1,
		Timeout:     2 * time.Second,
		MaxRetries:  2,
		RetryDelay:  time.Millisecond,
	}
	if jsonModel == nil {
		jsonModel = plain
	}
	return &Client{cfg: cfg, plain: plain, jsonModel: jsonModel, jsonCallOpts: jsonOpts}
}

func okMessage(content string) *schema.Message {
	return &schema.Message{
		Role:    schema.Assistant,
		Content: content,
		ResponseMeta: &schema.ResponseMeta{
			FinishReason: "stop",
			Usage: &schema.TokenUsage{
				PromptTokens:     10,
				CompletionTokens: 5,
			},
		},
	}
}

func TestGenerateHappyPath(t *testing.T) {
	fake := &fakeChatModel{
		generateFn: func(_ context.Context, in []*schema.Message, _ ...einomodel.Option) (*schema.Message, error) {
			if len(in) != 2 {
				t.Fatalf("messages len = %d, want 2", len(in))
			}
			if in[0].Role != schema.System || in[1].Role != schema.User {
				t.Fatalf("roles = %s/%s, want system/user", in[0].Role, in[1].Role)
			}
			return okMessage("hello"), nil
		},
	}
	c := newTestClient(fake, nil, nil)

	out, err := c.Generate(context.Background(), ChatRequest{
		Messages: []Message{
			{Role: RoleSystem, Content: "sys"},
			{Role: RoleUser, Content: "hi"},
		},
		Temperature: -1,
		TopP:        -1,
	})
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	if out.Content != "hello" {
		t.Errorf("Content = %q, want %q", out.Content, "hello")
	}
	if out.Model != "test-model" {
		t.Errorf("Model = %q, want config default", out.Model)
	}
	if out.FinishReason != "stop" {
		t.Errorf("FinishReason = %q, want stop", out.FinishReason)
	}
	// TotalTokens synthesised from prompt+completion.
	if out.Usage.TotalTokens != 15 {
		t.Errorf("TotalTokens = %d, want 15", out.Usage.TotalTokens)
	}
}

func TestGenerateEmptyMessages(t *testing.T) {
	c := newTestClient(&fakeChatModel{}, nil, nil)
	_, err := c.Generate(context.Background(), ChatRequest{Temperature: -1, TopP: -1})
	if !errors.Is(err, ErrInvalidConfig) {
		t.Fatalf("error = %v, want ErrInvalidConfig", err)
	}
}

func TestGenerateNilResponse(t *testing.T) {
	fake := &fakeChatModel{
		generateFn: func(context.Context, []*schema.Message, ...einomodel.Option) (*schema.Message, error) {
			return nil, nil
		},
	}
	c := newTestClient(fake, nil, nil)
	_, err := c.Generate(context.Background(), ChatRequest{
		Messages:    []Message{{Role: RoleUser, Content: "hi"}},
		Temperature: -1, TopP: -1,
	})
	if !errors.Is(err, ErrDecodeResponse) {
		t.Fatalf("error = %v, want ErrDecodeResponse", err)
	}
}

func TestGenerateRetriesTransientThenSucceeds(t *testing.T) {
	fake := &fakeChatModel{}
	fake.generateFn = func(context.Context, []*schema.Message, ...einomodel.Option) (*schema.Message, error) {
		if fake.calls < 2 {
			return nil, errors.New("connection reset by peer")
		}
		return okMessage("recovered"), nil
	}
	c := newTestClient(fake, nil, nil)

	out, err := c.Generate(context.Background(), ChatRequest{
		Messages:    []Message{{Role: RoleUser, Content: "hi"}},
		Temperature: -1, TopP: -1,
	})
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	if fake.calls != 2 {
		t.Errorf("calls = %d, want 2 (one retry)", fake.calls)
	}
	if out.Content != "recovered" {
		t.Errorf("Content = %q, want recovered", out.Content)
	}
}

func TestGenerateDoesNotRetryAuthError(t *testing.T) {
	fake := &fakeChatModel{}
	fake.generateFn = func(context.Context, []*schema.Message, ...einomodel.Option) (*schema.Message, error) {
		return nil, fmt.Errorf("call failed: %w", &genai.APIError{Code: 401, Message: "bad key"})
	}
	c := newTestClient(fake, nil, nil)

	_, err := c.Generate(context.Background(), ChatRequest{
		Messages:    []Message{{Role: RoleUser, Content: "hi"}},
		Temperature: -1, TopP: -1,
	})
	if err == nil {
		t.Fatal("Generate() = nil error, want auth failure")
	}
	if fake.calls != 1 {
		t.Errorf("calls = %d, want 1 (no retry on 401)", fake.calls)
	}
	if !IsAuthError(mustLift(t, c, err)) {
		t.Errorf("expected lifted 401 APIError, got %v", err)
	}
}

// mustLift re-runs vendor lifting so the assertion reads the typed shape
// the adapter layer would see after translateError.
func mustLift(t *testing.T, c *Client, err error) error {
	t.Helper()
	if lifted := liftVendorError(c.cfg.Backend, err); lifted != nil {
		return lifted
	}
	return err
}

func TestGenerateRateLimitedNotRetriedAndTranslated(t *testing.T) {
	fake := &fakeChatModel{}
	fake.generateFn = func(context.Context, []*schema.Message, ...einomodel.Option) (*schema.Message, error) {
		return nil, fmt.Errorf("call failed: %w", &genai.APIError{Code: 429, Message: "quota"})
	}
	c := newTestClient(fake, nil, nil)

	_, err := c.Generate(context.Background(), ChatRequest{
		Messages:    []Message{{Role: RoleUser, Content: "hi"}},
		Temperature: -1, TopP: -1,
	})
	if !errors.Is(err, ErrRateLimited) {
		t.Fatalf("error = %v, want ErrRateLimited", err)
	}
	if fake.calls != 1 {
		t.Errorf("calls = %d, want 1 (429 is not auto-retried)", fake.calls)
	}
}

func TestPickRoutesJSONMode(t *testing.T) {
	plain := &fakeChatModel{}
	jsonModel := &fakeChatModel{}
	jsonOpts := []einomodel.Option{einomodel.WithMaxTokens(123)}
	c := newTestClient(plain, jsonModel, jsonOpts)

	m, opts := c.pick(ChatRequest{JSONMode: true, Temperature: -1, TopP: -1})
	if m != einomodel.BaseChatModel(jsonModel) {
		t.Error("JSONMode=true did not route to jsonModel")
	}
	got := einomodel.GetCommonOptions(&einomodel.Options{}, opts...)
	if got.MaxTokens == nil || *got.MaxTokens != 123 {
		t.Error("jsonCallOpts were not appended for JSONMode=true")
	}

	m, _ = c.pick(ChatRequest{Temperature: -1, TopP: -1})
	if m != einomodel.BaseChatModel(plain) {
		t.Error("JSONMode=false did not route to plain model")
	}
}

func TestBuildCallOptionsPerCallOverridesConfig(t *testing.T) {
	c := newTestClient(&fakeChatModel{}, nil, nil)
	c.cfg.Temperature = 0.9
	c.cfg.TopP = 0.8
	c.cfg.MaxTokens = 100

	opts := c.buildCallOptions(ChatRequest{
		Model:       "override-model",
		Temperature: 0.1,
		TopP:        0.2,
		MaxTokens:   50,
		Stop:        []string{"END"},
	})
	got := einomodel.GetCommonOptions(&einomodel.Options{}, opts...)

	if got.Model == nil || *got.Model != "override-model" {
		t.Error("per-call model override lost")
	}
	if got.Temperature == nil || *got.Temperature != float32(0.1) {
		t.Error("per-call temperature override lost")
	}
	if got.TopP == nil || *got.TopP != float32(0.2) {
		t.Error("per-call topP override lost")
	}
	if got.MaxTokens == nil || *got.MaxTokens != 50 {
		t.Error("per-call maxTokens override lost")
	}
	if len(got.Stop) != 1 || got.Stop[0] != "END" {
		t.Error("stop sequences lost")
	}

	// Negative request values fall back to config defaults.
	opts = c.buildCallOptions(ChatRequest{Temperature: -1, TopP: -1})
	got = einomodel.GetCommonOptions(&einomodel.Options{}, opts...)
	if got.Temperature == nil || *got.Temperature != float32(0.9) {
		t.Error("config temperature default not applied")
	}
	if got.MaxTokens == nil || *got.MaxTokens != 100 {
		t.Error("config maxTokens default not applied")
	}
}

func TestBuildCallOptionsMaxTokensRoutingByBackend(t *testing.T) {
	req := ChatRequest{MaxTokens: 7, Temperature: -1, TopP: -1}

	// Non-openai backend: MaxTokens flows through the common option,
	// which the vendor component maps onto its standard max-tokens field.
	gcli := &Client{cfg: Config{Backend: BackendGoogleAI, Model: "m", Temperature: -1, TopP: -1}}
	gCommon := einomodel.GetCommonOptions(&einomodel.Options{}, gcli.buildCallOptions(req)...)
	if gCommon.MaxTokens == nil || *gCommon.MaxTokens != 7 {
		t.Fatalf("googleai: common MaxTokens = %v, want 7", gCommon.MaxTokens)
	}

	// OpenAI backend: the deprecated common `max_tokens` MUST NOT be set;
	// the value is routed to the openai-specific `max_completion_tokens`
	// option instead. That option struct is unexported so it cannot be
	// read back here, but its absence from the common options is the
	// observable proof that routing happened (newer OpenAI models reject
	// `max_tokens`).
	ocli := &Client{cfg: Config{Backend: BackendOpenAI, Model: "m", Temperature: -1, TopP: -1}}
	oOpts := ocli.buildCallOptions(req)
	oCommon := einomodel.GetCommonOptions(&einomodel.Options{}, oOpts...)
	if oCommon.MaxTokens != nil {
		t.Fatalf("openai: common MaxTokens = %d, want nil (routed to max_completion_tokens)", *oCommon.MaxTokens)
	}
	// The model option plus a max-completion option must have been emitted.
	if len(oOpts) < 2 {
		t.Fatalf("openai: options len = %d, want >= 2 (model + max_completion_tokens)", len(oOpts))
	}
}

func TestGenerateStream(t *testing.T) {
	fake := &fakeChatModel{
		streamFn: func(context.Context, []*schema.Message, ...einomodel.Option) (*schema.StreamReader[*schema.Message], error) {
			return schema.StreamReaderFromArray([]*schema.Message{
				{Role: schema.Assistant, Content: "Hel"},
				{Role: schema.Assistant, Content: "lo"},
				{Role: schema.Assistant, Content: "", ResponseMeta: &schema.ResponseMeta{
					FinishReason: "stop",
					Usage:        &schema.TokenUsage{PromptTokens: 3, CompletionTokens: 2, TotalTokens: 5},
				}},
			}), nil
		},
	}
	c := newTestClient(fake, nil, nil)

	var deltas []string
	var doneSeen bool
	var finalUsage *Usage
	out, err := c.GenerateStream(context.Background(), ChatRequest{
		Messages:    []Message{{Role: RoleUser, Content: "hi"}},
		Temperature: -1, TopP: -1,
	}, func(chunk StreamChunk) error {
		if chunk.Done {
			doneSeen = true
			finalUsage = chunk.Usage
			return nil
		}
		deltas = append(deltas, chunk.Delta)
		return nil
	})
	if err != nil {
		t.Fatalf("GenerateStream() error = %v", err)
	}
	if len(deltas) != 2 || deltas[0] != "Hel" || deltas[1] != "lo" {
		t.Errorf("deltas = %v, want [Hel lo]", deltas)
	}
	if !doneSeen {
		t.Error("final Done chunk was not delivered")
	}
	if out.Content != "Hello" {
		t.Errorf("assembled Content = %q, want Hello", out.Content)
	}
	if finalUsage == nil || finalUsage.TotalTokens != 5 {
		t.Errorf("final usage = %+v, want TotalTokens 5", finalUsage)
	}
}

func TestTranslateError(t *testing.T) {
	c := newTestClient(&fakeChatModel{}, nil, nil)

	tests := []struct {
		name string
		in   error
		want error // sentinel expected via errors.Is; nil means passthrough
	}{
		{"nil", nil, nil},
		{"sentinel preserved", ErrContextTooLarge, ErrContextTooLarge},
		{"context window substring", errors.New("input exceeds maximum context length"), ErrContextTooLarge},
		{"rate limit substring", errors.New("429 too many requests"), ErrRateLimited},
		{"resource exhausted substring", errors.New("googleapi: RESOURCE_EXHAUSTED"), ErrRateLimited},
		{"quota substring", errors.New("quota exceeded for project"), ErrRateLimited},
		{"typed genai 429", fmt.Errorf("x: %w", &genai.APIError{Code: 429}), ErrRateLimited},
		{"unknown passthrough", errors.New("some transport hiccup"), nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := c.translateError(tt.in)
			if tt.in == nil {
				if got != nil {
					t.Fatalf("translateError(nil) = %v", got)
				}
				return
			}
			if tt.want != nil {
				if !errors.Is(got, tt.want) {
					t.Fatalf("translateError(%v) = %v, want %v", tt.in, got, tt.want)
				}
				return
			}
			if got == nil {
				t.Fatal("translateError() = nil for non-nil input")
			}
		})
	}

	// Typed genai 401 lifts into *APIError with auth semantics.
	lifted := c.translateError(fmt.Errorf("x: %w", &genai.APIError{Code: 401, Message: "no"}))
	if !IsAuthError(lifted) {
		t.Errorf("translateError(genai 401) = %v, want IsAuthError=true", lifted)
	}
}

func TestToEinoRole(t *testing.T) {
	tests := []struct {
		in   Role
		want schema.RoleType
	}{
		{RoleSystem, schema.System},
		{RoleUser, schema.User},
		{RoleAssistant, schema.Assistant},
		{RoleTool, schema.Tool},
		{Role("unknown"), schema.User},
	}
	for _, tt := range tests {
		if got := toEinoRole(tt.in); got != tt.want {
			t.Errorf("toEinoRole(%s) = %s, want %s", tt.in, got, tt.want)
		}
	}
}

func TestFromTokenUsage(t *testing.T) {
	if got := fromTokenUsage(nil); got != (Usage{}) {
		t.Errorf("fromTokenUsage(nil) = %+v, want zero", got)
	}
	got := fromTokenUsage(&schema.TokenUsage{PromptTokens: 7, CompletionTokens: 3})
	if got.TotalTokens != 10 {
		t.Errorf("TotalTokens = %d, want synthesised 10", got.TotalTokens)
	}
	got = fromTokenUsage(&schema.TokenUsage{PromptTokens: 7, CompletionTokens: 3, TotalTokens: 11})
	if got.TotalTokens != 11 {
		t.Errorf("TotalTokens = %d, want vendor-reported 11", got.TotalTokens)
	}
}
