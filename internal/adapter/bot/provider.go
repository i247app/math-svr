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

// ResponderProvider is the OPTIONAL capability of holding conversation
// state on the vendor side and chaining turns by id. It is a separate
// interface rather than a method on BotProvider so that adding it did not
// touch the five providers that cannot implement it; the adapter
// discovers support by type assertion (Adapter.SupportsRespond) and
// answers BOT_UNSUPPORTED_OP for the rest — the same shape Embed already
// has on eino.
//
// Today only the OpenAI provider implements it (Responses API,
// previous_response_id).
type ResponderProvider interface {
	BotProvider

	// Respond runs one chained turn. The returned RespondResult.ResponseID
	// is the pointer the caller must persist for the next turn.
	Respond(ctx context.Context, req RespondRequest) (*RespondResult, error)
}
