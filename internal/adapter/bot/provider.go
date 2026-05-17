package bot

import "context"

// BotProvider is the contract every concrete AI provider implements.
// Adapter dispatches via the registered default or a named provider.
//
// Streaming and embedding are optional capabilities: providers that do
// not support them return errs.NewError(BOT_UNSUPPORTED_OP, ...). The
// adapter passes the request through unchanged; callers should
// nil-check or feature-flag based on the registered provider.
type BotProvider interface {
	// Name returns the registry name. Stable across the process lifetime.
	Name() BotProviderName

	// Chat runs a non-streamed completion.
	Chat(ctx context.Context, req ChatRequest) (*ChatResult, error)

	// Stream runs a streaming completion. The provider invokes onChunk
	// once per token delta and exactly once with Done=true at the end.
	// Returns the assembled ChatResult on success.
	Stream(ctx context.Context, req ChatRequest, onChunk func(StreamChunk) error) (*ChatResult, error)

	// Embed produces embeddings for each input string, preserving order.
	Embed(ctx context.Context, req EmbedRequest) (*EmbedResult, error)
}
