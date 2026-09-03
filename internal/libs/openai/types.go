package openai

// Role enumerates the chat message roles in the chat-completions schema.
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
	// messages. May be empty.
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
	// JSONMode sets response_format={"type":"json_object"} — OpenAI's
	// legacy JSON mode. The prompt still has to ask for JSON; the flag only
	// constrains the decoder. Structured Outputs (json_schema) is
	// deliberately not wired: the quiz/exercise prompts define their own
	// contract and every other provider in this repo speaks json_object,
	// so results stay comparable.
	JSONMode bool
}

// Usage reports tokens billed by the upstream when available.
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

// StreamChunk is a single token / partial-message event from
// GenerateStream. Delta is the incremental text. Done is true exactly
// once, on the final chunk, which may carry Usage.
type StreamChunk struct {
	Delta string
	Done  bool
	Usage *Usage
}

// EmbedRequest carries a batch of strings to embed. Empty Model falls back
// to Config.EmbedModel, then Config.Model.
type EmbedRequest struct {
	Model  string
	Inputs []string
}

// EmbedResponse pairs each input with its vector, preserving INPUT order
// (the client re-sorts on the response's index field — see Embed).
type EmbedResponse struct {
	Model   string
	Vectors [][]float32
	Usage   Usage
}

// ===========================================================================
// Wire types — the JSON actually exchanged with OpenAI. Kept unexported so
// the public surface above stays stable if the vendor schema shifts.
// ===========================================================================

type wireMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
	Name    string `json:"name,omitempty"`
}

type wireResponseFormat struct {
	Type string `json:"type"`
}

type wireStreamOptions struct {
	IncludeUsage bool `json:"include_usage"`
}

// wireChatRequest is the POST body. Sampling fields are pointers so an
// unset override is omitted entirely rather than sent as a zero value —
// sending temperature:0 is a real (and very different) instruction.
//
// Note MaxCompletionTokens, not MaxTokens: `max_tokens` is deprecated
// upstream and rejected outright by the reasoning models.
type wireChatRequest struct {
	Model               string              `json:"model"`
	Messages            []wireMessage       `json:"messages"`
	Stream              bool                `json:"stream,omitempty"`
	StreamOptions       *wireStreamOptions  `json:"stream_options,omitempty"`
	Temperature         *float64            `json:"temperature,omitempty"`
	TopP                *float64            `json:"top_p,omitempty"`
	MaxCompletionTokens *int                `json:"max_completion_tokens,omitempty"`
	Stop                []string            `json:"stop,omitempty"`
	ResponseFormat      *wireResponseFormat `json:"response_format,omitempty"`
	Store               bool                `json:"store,omitempty"`
	Metadata            map[string]string   `json:"metadata,omitempty"`
}

type wireUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

func (u *wireUsage) toUsage() Usage {
	if u == nil {
		return Usage{}
	}
	out := Usage{
		PromptTokens:     u.PromptTokens,
		CompletionTokens: u.CompletionTokens,
		TotalTokens:      u.TotalTokens,
	}
	if out.TotalTokens == 0 {
		out.TotalTokens = out.PromptTokens + out.CompletionTokens
	}
	return out
}

type wireChatResponse struct {
	Model   string `json:"model"`
	Choices []struct {
		Index        int         `json:"index"`
		Message      wireMessage `json:"message"`
		FinishReason string      `json:"finish_reason"`
	} `json:"choices"`
	Usage *wireUsage `json:"usage,omitempty"`
	Error *wireError `json:"error,omitempty"`
}

// wireStreamChunk is one `data: {...}` SSE payload.
type wireStreamChunk struct {
	Model   string `json:"model"`
	Choices []struct {
		Index int `json:"index"`
		Delta struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"delta"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage *wireUsage `json:"usage,omitempty"`
	Error *wireError `json:"error,omitempty"`
}

type wireEmbedRequest struct {
	Model          string   `json:"model"`
	Input          []string `json:"input"`
	EncodingFormat string   `json:"encoding_format,omitempty"`
}

type wireEmbedResponse struct {
	Model string `json:"model"`
	Data  []struct {
		Index     int       `json:"index"`
		Embedding []float32 `json:"embedding"`
	} `json:"data"`
	Usage *wireUsage `json:"usage,omitempty"`
	Error *wireError `json:"error,omitempty"`
}

// wireError is OpenAI's error object. Unlike OpenRouter's, Code is a
// STRING identifier ("credit_balance_exhausted", "context_length_exceeded")
// and Type carries the category ("rate_limit_error",
// "invalid_authentication_error"). Both drive classifyAPIError.
type wireError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Param   string `json:"param,omitempty"`
	Code    string `json:"code,omitempty"`
}
