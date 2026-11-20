package llmroute

import (
	"context"
	"fmt"
	"sync"
	"time"

	"math-ai.com/math-ai/internal/application/dto"
	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
	llmProvider "math-ai.com/math-ai/internal/provider/llm"
	"math-ai.com/math-ai/internal/shared/mtime"
)

// ModelListTTL is how long one provider's model list is served from this
// instance's memory before the vendor is called again. Each instance keeps
// its own cache (3 instances = up to 3 vendor calls per TTL, lists at most
// one TTL apart) — acceptable for an informational list; nothing is synced.
const ModelListTTL = time.Hour

// modelCache holds the last SUCCESSFUL list per provider. Failures are not
// cached, so a vendor that recovers is seen on the next call.
//
// ponytail: concurrent misses for one provider each call the vendor (no
// singleflight); admin-only traffic, add x/sync/singleflight if it matters.
type modelCache struct {
	mu      sync.Mutex
	ttl     time.Duration
	now     func() time.Time
	entries map[llmProvider.LLMProviderName]modelEntry
}

type modelEntry struct {
	models    []llmProvider.ModelInfo
	fetchedAt time.Time
}

func newModelCache(ttl time.Duration) *modelCache {
	return &modelCache{ttl: ttl, now: time.Now, entries: map[llmProvider.LLMProviderName]modelEntry{}}
}

func (c *modelCache) get(name llmProvider.LLMProviderName) (modelEntry, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	e, ok := c.entries[name]
	if !ok || c.now().Sub(e.fetchedAt) >= c.ttl {
		return modelEntry{}, false
	}
	return e, true
}

func (c *modelCache) put(name llmProvider.LLMProviderName, models []llmProvider.ModelInfo) modelEntry {
	e := modelEntry{models: models, fetchedAt: c.now()}
	c.mu.Lock()
	c.entries[name] = e
	c.mu.Unlock()
	return e
}

// ListLLMModels lists the models of one registered provider, or of every
// one when req.provider is empty. Providers are called in parallel, each
// under its own client timeout; one that fails is reported in its own
// error field and the others are still returned.
func (s *Service) ListLLMModels(ctx context.Context, req *dto.ListLLMModelsReq) (*dto.ListLLMModelsRes, error) {
	if _, err := requireAdminGate(ctx); err != nil {
		return nil, err
	}
	names, err := s.modelTargets(ctx, req)
	if err != nil {
		return nil, err
	}

	res := &dto.ListLLMModelsRes{
		Providers: make([]*dto.LLMProviderModels, len(names)),
	}

	var wg sync.WaitGroup
	for i, name := range names {
		wg.Go(func() {
			res.Providers[i] = s.providerModels(ctx, name, req.ChatOnly)
		})
	}
	wg.Wait()
	return res, nil
}

// modelTargets is the requested provider (lower-cased, must be registered)
// or every registered one, sorted.
func (s *Service) modelTargets(ctx context.Context, req *dto.ListLLMModelsReq) ([]llmProvider.LLMProviderName, error) {
	if p := blankToNil(req.Provider, true); p != nil {
		name := llmProvider.LLMProviderName(*p)
		if s.registry == nil || !s.registry.Has(name) {
			return nil, errs.NewError(ctx, status.LLM_ROUTE_INVALID_PROVIDER, nil,
				fmt.Errorf("provider %q is not registered (registered: %s)", name, registeredList(s.registry)))
		}
		return []llmProvider.LLMProviderName{name}, nil
	}
	if s.registry == nil {
		return nil, nil
	}
	return s.registry.Names(), nil
}

// providerModels serves one provider's list. The cache holds the full list,
// so chat_only and the full view share one vendor call.
func (s *Service) providerModels(ctx context.Context, name llmProvider.LLMProviderName, chatOnly bool) *dto.LLMProviderModels {
	out := &dto.LLMProviderModels{Provider: name.String(), Models: []dto.LLMModel{}}

	entry, cached := s.models.get(name)
	if !cached {
		models, err := s.registry.ListModelsVia(ctx, name)
		if err != nil {
			out.Error = modelsError(ctx, err)
			return out
		}
		entry = s.models.put(name, models)
	}

	out.Cached = cached
	out.FetchedAt = mtime.NewMathTimePtr(entry.fetchedAt)
	for _, m := range entry.models {
		if chatOnly && !llmProvider.IsChatModel(name, m) {
			continue
		}
		out.Models = append(out.Models, dto.LLMModel(m))
	}
	out.ModelCount = len(out.Models)
	out.TotalCount = len(entry.models)
	return out
}

// modelsError keeps the status code and its localised message only. The
// MathError's debug text is dropped on purpose: it carries the vendor's
// error message, which for a bad key can quote part of the key.
func modelsError(ctx context.Context, err error) *dto.LLMModelsError {
	mErr, ok := errs.IsMathError(err)
	if !ok {
		// Not a MathError (a provider bypassing the error contract).
		mErr = errs.NewError(ctx, status.LLM_OP_FAILED, nil, err)
	}
	return &dto.LLMModelsError{MStatus: int(mErr.GetStatusCode()), MMessage: string(mErr.GetStatusMessage())}
}
