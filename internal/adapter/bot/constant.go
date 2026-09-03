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
// The first wired provider is "langchain", backed by
// `internal/libs/langchain` which itself dispatches to one of several
// LLM vendors (googleai, openai, anthropic, ollama) selected at boot.
package bot

// BotProviderName is the registered name of a concrete provider.
type BotProviderName string

const (
	// ProviderLangChain dispatches via internal/libs/langchain. The
	// backend vendor (Gemini, OpenAI, etc.) is selected inside that
	// library by BotConfig.LangChainBackend.
	ProviderLangChain BotProviderName = "langchain"

	// ProviderEino dispatches via internal/libs/eino (cloudwego/eino +
	// eino-ext). The backend vendor is selected inside that library by
	// BotConfig.EinoBackend. Registered alongside langchain when
	// configured; BOT_PROVIDER picks which one is the default.
	ProviderEino BotProviderName = "eino"

	// ProviderOpenRouter dispatches via internal/libs/openrouter, which
	// calls the OpenRouter REST API directly over the shared http_client
	// (no vendor SDK). There is no backend selector: OpenRouter is itself
	// the router and the vendor is chosen by the model id
	// ("vendor/model-name") in BotConfig.OpenRouterModel. Registered
	// alongside the other frameworks when configured; BOT_PROVIDER picks
	// which one is the default.
	ProviderOpenRouter BotProviderName = "openrouter"

	// ProviderOpenAI dispatches via internal/libs/openai, which calls
	// api.openai.com directly over the shared http_client (no vendor SDK,
	// no broker). Like openrouter it has no backend selector — it talks to
	// OpenAI and only OpenAI, and BotConfig.OpenAIModel names a bare model
	// id ("gpt-4.1", not "openai/gpt-4.1").
	//
	// Naming caution: "openai" is ALSO a valid value for
	// BOT_LANGCHAIN_BACKEND / BOT_EINO_BACKEND, where it selects the vendor
	// an SDK talks to. As a BOT_PROVIDER value it means the direct client
	// below. The two never collide because they are read from different
	// env keys, but read a config twice before assuming which one is meant.
	//
	// This is the only provider besides langchain that implements Embed.
	ProviderOpenAI BotProviderName = "openai"

	// ProviderGemini dispatches via internal/libs/gemini, which calls
	// generativelanguage.googleapis.com directly over the shared
	// http_client (no vendor SDK, no broker). Like openai it has no
	// backend selector, and BotConfig.GeminiModel names a bare model id
	// ("gemini-2.0-flash"); the "models/" URL prefix is added by the
	// client.
	//
	// Same naming caution as ProviderOpenAI: "googleai" is the BACKEND
	// name langchain and eino use to reach the same vendor through an SDK.
	// This provider is the direct path.
	//
	// Supports Chat, Stream and Embed.
	ProviderGemini BotProviderName = "gemini"
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
