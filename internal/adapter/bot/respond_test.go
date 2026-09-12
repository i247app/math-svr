package bot

import (
	"context"
	"errors"
	"testing"

	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
)

// statelessProvider implements BotProvider only — the shape of the five
// providers that cannot hold vendor-side state.
type statelessProvider struct{ calls int }

func (p *statelessProvider) Name() BotProviderName { return "stateless" }
func (p *statelessProvider) Chat(context.Context, ChatRequest) (*ChatResult, error) {
	p.calls++
	return &ChatResult{Provider: "stateless"}, nil
}
func (p *statelessProvider) Stream(context.Context, ChatRequest, func(StreamChunk) error) (*ChatResult, error) {
	p.calls++
	return nil, nil
}
func (p *statelessProvider) Embed(context.Context, EmbedRequest) (*EmbedResult, error) {
	p.calls++
	return nil, nil
}

// statefulProvider additionally implements ResponderProvider and records
// what it was handed.
type statefulProvider struct {
	statelessProvider
	got  RespondRequest
	resp *RespondResult
	err  error
}

func (p *statefulProvider) Name() BotProviderName { return "stateful" }
func (p *statefulProvider) Respond(_ context.Context, req RespondRequest) (*RespondResult, error) {
	p.calls++
	p.got = req
	if p.err != nil {
		return nil, p.err
	}
	return p.resp, nil
}

func newRespondAdapter(t *testing.T) (*Adapter, *statelessProvider, *statefulProvider) {
	t.Helper()
	a := NewAdapter()
	stateless := &statelessProvider{}
	stateful := &statefulProvider{resp: &RespondResult{Provider: "stateful", Model: "m", ResponseID: "resp_1", Content: "{}"}}
	a.Register(stateless)
	a.Register(stateful)
	return a, stateless, stateful
}

func TestSupportsRespond(t *testing.T) {
	a, _, _ := newRespondAdapter(t)

	if a.SupportsRespond("stateless") {
		t.Error("a BotProvider-only provider must not report Respond support")
	}
	if !a.SupportsRespond("stateful") {
		t.Error("a ResponderProvider must report Respond support")
	}
	if a.SupportsRespond("missing") {
		t.Error("an unregistered name must not report support")
	}
	// Empty name resolves to the default, which is the first registration.
	if a.SupportsRespond("") {
		t.Error("default (stateless) must not report support")
	}
}

func TestRespondViaUnsupportedProviderMakesNoCall(t *testing.T) {
	a, stateless, _ := newRespondAdapter(t)

	_, err := a.RespondVia(context.Background(), "stateless", RespondRequest{
		Input: []Message{{Role: RoleUser, Content: "hi"}},
		Store: true,
	})
	mErr, ok := errs.IsMathError(err)
	if !ok || mErr.GetStatusCode() != status.BOT_UNSUPPORTED_OP {
		t.Fatalf("error = %v, want BOT_UNSUPPORTED_OP", err)
	}
	if stateless.calls != 0 {
		t.Errorf("provider calls = %d, want 0 — capability is decided before dispatch", stateless.calls)
	}
}

func TestRespondViaPassesThrough(t *testing.T) {
	a, _, stateful := newRespondAdapter(t)

	res, err := a.RespondVia(context.Background(), "stateful", RespondRequest{
		Instructions:       "tutor",
		Input:              []Message{{Role: RoleUser, Content: "next"}},
		PreviousResponseID: "resp_0",
		Store:              true,
		JSONMode:           true,
	})
	if err != nil {
		t.Fatalf("RespondVia() error = %v", err)
	}
	if res.ResponseID != "resp_1" {
		t.Errorf("ResponseID = %q, want resp_1", res.ResponseID)
	}
	if stateful.got.PreviousResponseID != "resp_0" || stateful.got.Instructions != "tutor" || !stateful.got.Store {
		t.Errorf("provider received %+v, want the request unchanged", stateful.got)
	}
}

// The chaining precondition is enforced at the adapter boundary so a
// caller bug never reaches the vendor as a paid 400.
func TestRespondValidateRejectsChainWithoutStore(t *testing.T) {
	a, _, stateful := newRespondAdapter(t)

	_, err := a.RespondVia(context.Background(), "stateful", RespondRequest{
		Input:              []Message{{Role: RoleUser, Content: "hi"}},
		PreviousResponseID: "resp_0",
		Store:              false,
	})
	mErr, ok := errs.IsMathError(err)
	if !ok || mErr.GetStatusCode() != status.BOT_INVALID_PROMPT {
		t.Fatalf("error = %v, want BOT_INVALID_PROMPT", err)
	}
	if stateful.calls != 0 {
		t.Errorf("provider calls = %d, want 0", stateful.calls)
	}
}

func TestRespondRequestValidate(t *testing.T) {
	long := make([]byte, maxMessageBytes+1)
	for i := range long {
		long[i] = 'a'
	}
	tests := []struct {
		name    string
		req     RespondRequest
		wantErr bool
	}{
		{"ok fresh", RespondRequest{Input: []Message{{Role: RoleUser, Content: "x"}}}, false},
		{"ok chained", RespondRequest{Input: []Message{{Role: RoleUser, Content: "x"}}, PreviousResponseID: "r", Store: true}, false},
		{"empty input", RespondRequest{}, true},
		{"blank content", RespondRequest{Input: []Message{{Role: RoleUser, Content: "  "}}}, true},
		{"bad role", RespondRequest{Input: []Message{{Role: "robot", Content: "x"}}}, true},
		{"oversized", RespondRequest{Input: []Message{{Role: RoleUser, Content: string(long)}}}, true},
		{"chain without store", RespondRequest{Input: []Message{{Role: RoleUser, Content: "x"}}, PreviousResponseID: "r"}, true},
		{"negative max tokens", RespondRequest{Input: []Message{{Role: RoleUser, Content: "x"}}, MaxTokens: -1}, true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.req.Validate()
			if (err != nil) != tc.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}

func TestRespondViaProviderErrorIsSurfaced(t *testing.T) {
	a, _, stateful := newRespondAdapter(t)
	stateful.err = errs.NewError(context.Background(), status.BOT_OP_FAILED, nil, errors.New("upstream"))

	_, err := a.RespondVia(context.Background(), "stateful", RespondRequest{
		Input: []Message{{Role: RoleUser, Content: "hi"}}, Store: true,
	})
	mErr, ok := errs.IsMathError(err)
	if !ok || mErr.GetStatusCode() != status.BOT_OP_FAILED {
		t.Fatalf("error = %v, want the provider's MathError unchanged", err)
	}
}
