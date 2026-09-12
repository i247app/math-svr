package bot

import (
	"errors"
	"fmt"
	"strings"
)

// Message is a single turn in a chat conversation. The provider maps
// Role onto the upstream vendor's role vocabulary.
type Message struct {
	Role    Role
	Content string
	// Name is optional metadata some backends use for tool-routing.
	Name string
}

// ChatRequest is the provider-agnostic chat invocation payload. Sampling
// fields fall back to provider defaults when left at their zero value
// (Temperature < 0, TopP < 0, MaxTokens == 0).
type ChatRequest struct {
	Provider BotProviderName

	// Model overrides the provider's configured default model id.
	Model string

	Messages []Message

	Temperature float64 // <0 means "use provider default"
	TopP        float64 // <0 means "use provider default"
	MaxTokens   int     // 0 means "use provider default"

	// Stop is the vendor-pass-through list of stop sequences. Optional.
	Stop []string

	// JSONMode requests the upstream return a JSON-shaped string. Not
	// every backend supports this; providers that do not silently ignore.
	JSONMode bool
}

// Usage reports tokens billed by the upstream when available.
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// ChatResult is the provider-agnostic outcome of a successful Chat.
type ChatResult struct {
	Provider     BotProviderName `json:"provider"`
	Model        string          `json:"model"`
	Content      string          `json:"content"`
	FinishReason string          `json:"finish_reason,omitempty"`
	Usage        Usage           `json:"usage"`
}

// StreamChunk is a single event delivered to a streaming consumer. Done
// is true exactly once, on the final chunk, which may carry Usage when
// the backend reports it.
type StreamChunk struct {
	Delta string
	Done  bool
	Usage *Usage
}

// EmbedRequest carries a batch of strings to embed.
type EmbedRequest struct {
	Model  string
	Inputs []string
}

// EmbedResult pairs each input with its vector, preserving order.
type EmbedResult struct {
	Provider BotProviderName `json:"provider"`
	Model    string          `json:"model"`
	Vectors  [][]float32     `json:"vectors"`
	Usage    Usage           `json:"usage"`
}

// RespondRequest is one turn of a SERVER-SIDE conversation: the vendor
// keeps the transcript and the caller sends only what is new, chained by
// PreviousResponseID. This is a distinct capability from Chat — see
// ResponderProvider — and deliberately not a field on ChatRequest, because
// only some vendors can honour it and the five stateless providers must
// stay unaware of it.
//
// Instructions (the system prompt) are NOT carried across the chain by
// any known vendor; callers pass them on every turn. Store must be true
// whenever PreviousResponseID is set — a response can only be chained on
// if it was retained — and the provider rejects the combination before
// any network call.
type RespondRequest struct {
	Provider BotProviderName

	// Model overrides the provider's configured default model id.
	Model string

	Instructions string

	// Input is the new content for this turn. Never empty.
	Input []Message

	// PreviousResponseID chains this turn onto the vendor's stored
	// response. Empty starts a new chain.
	PreviousResponseID string

	// Store asks the vendor to retain this response so the next turn can
	// chain on it. Retention is the vendor's policy (OpenAI: 30 days).
	Store bool

	Temperature float64 // <0 means "use provider default"
	TopP        float64 // <0 means "use provider default"
	MaxTokens   int     // 0 means "use provider default"

	JSONMode bool
}

// RespondResult is the outcome of one server-side conversation turn.
// ResponseID is the pointer for the next turn; persisting it is the
// caller's job.
type RespondResult struct {
	Provider   BotProviderName `json:"provider"`
	Model      string          `json:"model"`
	ResponseID string          `json:"response_id"`
	Content    string          `json:"content"`
	Usage      Usage           `json:"usage"`
}

// Validate enforces the documented invariants on ChatRequest. Returns
// plain errors; the Adapter.Chat path wraps them in a
// MathError(BOT_INVALID_PROMPT) so callers can surface the cause without
// leaking the conversation body.
func (r ChatRequest) Validate() error {
	if len(r.Messages) == 0 {
		return errors.New("bot: Messages is required")
	}
	if len(r.Messages) > maxMessages {
		return fmt.Errorf("bot: Messages count %d exceeds limit %d", len(r.Messages), maxMessages)
	}
	for i, m := range r.Messages {
		if strings.TrimSpace(m.Content) == "" {
			return fmt.Errorf("bot: Messages[%d].Content is empty", i)
		}
		if len(m.Content) > maxMessageBytes {
			return fmt.Errorf("bot: Messages[%d].Content exceeds %d bytes", i, maxMessageBytes)
		}
		switch m.Role {
		case RoleSystem, RoleUser, RoleAssistant, RoleTool:
		default:
			return fmt.Errorf("bot: Messages[%d].Role %q is invalid", i, m.Role)
		}
	}
	if r.MaxTokens < 0 {
		return errors.New("bot: MaxTokens must be >= 0")
	}
	return nil
}

// Validate enforces the documented invariants on RespondRequest. Same
// message rules as ChatRequest, plus the chaining precondition.
func (r RespondRequest) Validate() error {
	if len(r.Input) == 0 {
		return errors.New("bot: Input is required")
	}
	if len(r.Input) > maxMessages {
		return fmt.Errorf("bot: Input count %d exceeds limit %d", len(r.Input), maxMessages)
	}
	for i, m := range r.Input {
		if strings.TrimSpace(m.Content) == "" {
			return fmt.Errorf("bot: Input[%d].Content is empty", i)
		}
		if len(m.Content) > maxMessageBytes {
			return fmt.Errorf("bot: Input[%d].Content exceeds %d bytes", i, maxMessageBytes)
		}
		switch m.Role {
		case RoleSystem, RoleUser, RoleAssistant, RoleTool:
		default:
			return fmt.Errorf("bot: Input[%d].Role %q is invalid", i, m.Role)
		}
	}
	if r.PreviousResponseID != "" && !r.Store {
		return errors.New("bot: PreviousResponseID requires Store=true")
	}
	if r.MaxTokens < 0 {
		return errors.New("bot: MaxTokens must be >= 0")
	}
	return nil
}

// Validate enforces the documented invariants on EmbedRequest.
func (r EmbedRequest) Validate() error {
	if len(r.Inputs) == 0 {
		return errors.New("bot: Inputs is required")
	}
	if len(r.Inputs) > maxEmbedInputs {
		return fmt.Errorf("bot: Inputs count %d exceeds limit %d", len(r.Inputs), maxEmbedInputs)
	}
	for i, s := range r.Inputs {
		if strings.TrimSpace(s) == "" {
			return fmt.Errorf("bot: Inputs[%d] is empty", i)
		}
		if len(s) > maxMessageBytes {
			return fmt.Errorf("bot: Inputs[%d] exceeds %d bytes", i, maxMessageBytes)
		}
	}
	return nil
}
