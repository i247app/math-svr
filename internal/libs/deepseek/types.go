package deepseek

// Role enumerates the chat message roles. DeepSeek accepts the full
// OpenAI set, so no narrowing is needed at the boundary (contrast
// libs/gemini, which has to lift system out into systemInstruction).
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
	// JSONMode sets response_format={"type":"json_object"}.
	//
	// DeepSeek's JSON mode constrains the decoder but does not by itself
	// instruct the model — the prompt must still ask for JSON. Every
	// prompt in domain/bot already does, so this is a note for whoever
	// writes the next one, not a gap.
	JSONMode bool
}

// Usage reports tokens billed by the upstream when available.
//
// CacheHitTokens / CacheMissTokens and ReasoningTokens are DeepSeek
// extensions with no equivalent on the adapter's shared Usage type. They
// are logged as operator metadata rather than propagated, because they
// exist to explain a bill, not to change a caller's behaviour.
type Usage struct {
	PromptTokens     int
	CompletionTokens int
	TotalTokens      int

	CacheHitTokens  int
	CacheMissTokens int
	ReasoningTokens int
}

// ChatResponse is the provider-agnostic shape of a non-streamed reply.
//
// ReasoningContent holds the model's deliberation when thinking mode is
// on. It is captured so the client can report on it, but the adapter
// deliberately does NOT forward it: callers of this repo want the answer
// (a quiz JSON blob), and the reasoning is both large and not part of the
// contract the prompts define.
type ChatResponse struct {
	Model            string
	Content          string
	ReasoningContent string
	FinishReason     string
	Usage            Usage
}

// StreamChunk is a single event from GenerateStream. Delta is the
// incremental answer text; reasoning deltas are not forwarded, for the
// reason given on ChatResponse.ReasoningContent. Done is true exactly
// once, on the final chunk, which may carry Usage.
type StreamChunk struct {
	Delta string
	Done  bool
	Usage *Usage
}

// Finish reasons that matter to this client.
const (
	finishReasonStop = "stop"
	// finishReasonLength means the token cap cut the answer off mid-JSON.
	finishReasonLength = "length"
	// finishReasonNoResource is DeepSeek-specific: the platform ran out of
	// capacity mid-generation. Transient, unlike a content filter.
	finishReasonNoResource = "insufficient_system_resource"
	finishReasonFilter     = "content_filter"
)

// ===========================================================================
// Wire types — the JSON actually exchanged with DeepSeek. Kept unexported
// so the public surface above stays stable if the vendor schema shifts.
// ===========================================================================

type wireMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
	Name    string `json:"name,omitempty"`
	// ReasoningContent is response-only; it must never be echoed back on a
	// subsequent turn.
	ReasoningContent string `json:"reasoning_content,omitempty"`
}

type wireResponseFormat struct {
	Type string `json:"type"`
}

type wireStreamOptions struct {
	IncludeUsage bool `json:"include_usage"`
}

type wireThinking struct {
	Type            string `json:"type,omitempty"`
	ReasoningEffort string `json:"reasoning_effort,omitempty"`
}

// wireChatRequest is the POST body. Sampling fields are pointers so an
// unset override is omitted entirely rather than sent as a zero value —
// sending temperature:0 is a real (and very different) instruction.
//
// Note MaxTokens, not MaxCompletionTokens: DeepSeek reads the original
// OpenAI field name. See the package doc.
type wireChatRequest struct {
	Model          string              `json:"model"`
	Messages       []wireMessage       `json:"messages"`
	Stream         bool                `json:"stream,omitempty"`
	StreamOptions  *wireStreamOptions  `json:"stream_options,omitempty"`
	Temperature    *float64            `json:"temperature,omitempty"`
	TopP           *float64            `json:"top_p,omitempty"`
	MaxTokens      *int                `json:"max_tokens,omitempty"`
	Stop           []string            `json:"stop,omitempty"`
	ResponseFormat *wireResponseFormat `json:"response_format,omitempty"`
	Thinking       *wireThinking       `json:"thinking,omitempty"`
}

type wireUsage struct {
	PromptTokens            int `json:"prompt_tokens"`
	CompletionTokens        int `json:"completion_tokens"`
	TotalTokens             int `json:"total_tokens"`
	PromptCacheHitTokens    int `json:"prompt_cache_hit_tokens"`
	PromptCacheMissTokens   int `json:"prompt_cache_miss_tokens"`
	CompletionTokensDetails *struct {
		ReasoningTokens int `json:"reasoning_tokens"`
	} `json:"completion_tokens_details,omitempty"`
}

func (u *wireUsage) toUsage() Usage {
	if u == nil {
		return Usage{}
	}
	out := Usage{
		PromptTokens:     u.PromptTokens,
		CompletionTokens: u.CompletionTokens,
		TotalTokens:      u.TotalTokens,
		CacheHitTokens:   u.PromptCacheHitTokens,
		CacheMissTokens:  u.PromptCacheMissTokens,
	}
	if u.CompletionTokensDetails != nil {
		out.ReasoningTokens = u.CompletionTokensDetails.ReasoningTokens
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
			Role             string `json:"role"`
			Content          string `json:"content"`
			ReasoningContent string `json:"reasoning_content"`
		} `json:"delta"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage *wireUsage `json:"usage,omitempty"`
	Error *wireError `json:"error,omitempty"`
}

// wireError is DeepSeek's error object, which follows OpenAI's shape.
type wireError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Param   string `json:"param,omitempty"`
	Code    string `json:"code,omitempty"`
}
