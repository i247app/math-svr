package openrouter

// Role enumerates the chat message roles supported by the OpenRouter
// chat-completions schema (which is OpenAI-compatible).
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
	// JSONMode sets response_format={"type":"json_object"}. OpenRouter
	// forwards it to backends that support it; models that do not simply
	// ignore it, matching the posture of the other two bot providers.
	JSONMode bool
}

// Usage reports tokens billed by the upstream when available. Responses
// that omit usage leave all three fields zero.
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
// once, on the final chunk, which may carry Usage when the backend
// reports it.
type StreamChunk struct {
	Delta string
	Done  bool
	Usage *Usage
}

// ===========================================================================
// Wire types — the JSON actually exchanged with OpenRouter. Kept unexported
// so the public surface above stays stable if the vendor schema shifts.
// ===========================================================================

type wireMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
	Name    string `json:"name,omitempty"`
}

type wireResponseFormat struct {
	Type string `json:"type"`
}

// wireChatRequest is the POST body. Sampling fields are pointers so an
// unset override is omitted entirely rather than sent as a zero value —
// sending temperature:0 is a real (and very different) instruction.
type wireChatRequest struct {
	Model          string              `json:"model"`
	Messages       []wireMessage       `json:"messages"`
	Stream         bool                `json:"stream,omitempty"`
	Temperature    *float64            `json:"temperature,omitempty"`
	TopP           *float64            `json:"top_p,omitempty"`
	MaxTokens      *int                `json:"max_tokens,omitempty"`
	Stop           []string            `json:"stop,omitempty"`
	ResponseFormat *wireResponseFormat `json:"response_format,omitempty"`
	// StreamOptions asks OpenAI-compatible backends to emit a final usage
	// chunk on streamed responses. Only sent when Stream is true.
	StreamOptions *wireStreamOptions `json:"stream_options,omitempty"`
}

type wireStreamOptions struct {
	IncludeUsage bool `json:"include_usage"`
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

// wireChatResponse is a non-streamed completion. Error is populated on the
// error envelope, which OpenRouter may also return with a 200 status on a
// stream that failed after the headers were written.
type wireChatResponse struct {
	Model   string `json:"model"`
	Choices []struct {
		Index        int         `json:"index"`
		Message      wireMessage `json:"message"`
		FinishReason string      `json:"finish_reason"`
		Error        *wireError  `json:"error,omitempty"`
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
		FinishReason string     `json:"finish_reason"`
		Error        *wireError `json:"error,omitempty"`
	} `json:"choices"`
	Usage *wireUsage `json:"usage,omitempty"`
	Error *wireError `json:"error,omitempty"`
}

// wireError is OpenRouter's error object. Code is `any` because the API
// documents it as a number but OpenAI-compatible upstreams sometimes
// forward a string; see codeString.
type wireError struct {
	Code     any            `json:"code"`
	Message  string         `json:"message"`
	Metadata map[string]any `json:"metadata,omitempty"`
}
