package langchain

// Role enumerates the chat message roles supported across backends.
// The Client maps these to langchaingo's ChatMessageType constants in
// toLLMRole.
type Role string

const (
	RoleSystem    Role = "system"
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
	RoleTool      Role = "tool"
)

// Message is a single turn in a chat conversation.
type Message struct {
	Role    Role
	Content string
	// Name is an optional discriminator some backends use for tool
	// messages (e.g. OpenAI function-name routing). May be empty.
	Name string
}

// ChatRequest carries the LLM invocation parameters. Empty Model falls
// back to Config.Model; sampling overrides outside their valid ranges
// (Temperature/TopP < 0, MaxTokens == 0) fall back to Config defaults.
type ChatRequest struct {
	Model       string
	Messages    []Message
	Temperature float64 // <0 means "use config default"
	TopP        float64 // <0 means "use config default"
	MaxTokens   int     // 0 means "use config default"
	Stop        []string
	JSONMode    bool
}

// Usage reports tokens billed by the upstream when available. Backends
// that do not surface usage leave all three fields zero.
type Usage struct {
	PromptTokens     int
	CompletionTokens int
	TotalTokens      int
}

// ChatResponse is the provider-agnostic shape of a non-streamed reply.
type ChatResponse struct {
	Model        string
	Content      string
	FinishReason string
	Usage        Usage
}

// StreamChunk is a single token / partial-message event from GenerateStream.
// Delta is the incremental text. Done is true exactly once, on the final
// chunk, which may carry Usage when the backend reports it.
type StreamChunk struct {
	Delta string
	Done  bool
	Usage *Usage
}

// EmbedRequest carries a batch of inputs to embed.
type EmbedRequest struct {
	Model  string
	Inputs []string
}

// EmbedResponse pairs each input with its vector, preserving order.
type EmbedResponse struct {
	Model   string
	Vectors [][]float32
	Usage   Usage
}
