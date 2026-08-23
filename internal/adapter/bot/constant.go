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
