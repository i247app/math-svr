// Package bot is the AI / chat adapter. It exposes a provider-agnostic
// chat / stream / embed surface to module services (e.g. quiz grading,
// assessment review) while hiding the underlying LLM SDK.
//
// Provider registry pattern mirrors `internal/adapter/sms` and
// `internal/adapter/storage`:
//   - Adapter holds a map of registered providers and a default name.
//   - factory.go is the only place that knows which vendor.
//   - Adding a vendor = a new bot_<vendor>_provider.go file + a new
//     constant here + a new case in factory.go.
//
// Every provider is a direct REST client over the shared http_client
// (internal/libs/<vendor>) — no vendor SDK.
package bot

// BotProviderName is the registered name of a concrete provider.
type BotProviderName string

const (
	// ProviderOpenRouter dispatches via internal/libs/openrouter, which
	// calls the OpenRouter REST API directly over the shared http_client
	// (no vendor SDK). There is no backend selector: OpenRouter is itself
	// the router and the vendor is chosen by the model id
	// ("vendor/model-name") in BotConfig.OpenRouterModel. Registered
	// alongside the other providers when configured; BOT_PROVIDER picks
	// which one is the default.
	ProviderOpenRouter BotProviderName = "openrouter"

	// ProviderOpenAI dispatches via internal/libs/openai, which calls
	// api.openai.com directly over the shared http_client (no vendor SDK,
	// no broker). Like openrouter it has no backend selector — it talks to
	// OpenAI and only OpenAI, and BotConfig.OpenAIModel names a bare model
	// id ("gpt-4.1", not "openai/gpt-4.1").
	//
	// Supports Chat, Stream, Embed and Respond (Responses API).
	ProviderOpenAI BotProviderName = "openai"

	// ProviderGemini dispatches via internal/libs/gemini, which calls
	// generativelanguage.googleapis.com directly over the shared
	// http_client (no vendor SDK, no broker). Like openai it has no
	// backend selector, and BotConfig.GeminiModel names a bare model id
	// ("gemini-2.0-flash"); the "models/" URL prefix is added by the
	// client.
	//
	// Supports Chat, Stream and Embed.
	ProviderGemini BotProviderName = "gemini"

	// ProviderDeepSeek dispatches via internal/libs/deepseek, which calls
	// api.deepseek.com directly over the shared http_client (no vendor
	// SDK, no broker). Like openai and gemini it has no backend selector.
	//
	// DeepSeek's API is OpenAI-compatible but not identical: it reads
	// `max_tokens` rather than `max_completion_tokens`, signals an empty
	// balance with HTTP 402, and adds a thinking mode. See the libs
	// package doc for why that warranted its own client.
	//
	// Supports Chat and Stream; Embed is unsupported (no endpoint).
	ProviderDeepSeek BotProviderName = "deepseek"
)

// Role enumerates the chat message roles the adapter recognises. The
// provider layer is responsible for mapping these onto the upstream
// vendor's role vocabulary.
type Role string

const (
	RoleSystem    Role = "system"
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
	RoleTool      Role = "tool"
)

// Conservative payload limits enforced by Message.Validate. They exist
// to fail fast at the adapter boundary rather than burning vendor quota
// on a payload that will be rejected upstream anyway. Tune per deploy
// once we have real usage data.
const (
	// maxMessages caps the conversation history in a single request.
	maxMessages = 200

	// maxMessageBytes caps a single message body. ~64 KB is well above
	// any reasonable prompt while keeping us under every vendor's hard
	// limit.
	maxMessageBytes = 64 * 1024

	// maxEmbedInputs caps how many strings can be embedded in one call.
	// Vendors typically allow 96–2048; we pick the smaller value to keep
	// upstream batch errors at bay.
	maxEmbedInputs = 96
)
