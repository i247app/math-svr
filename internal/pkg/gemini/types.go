package gemini

// Role enumerates the roles the CALLER may use. Gemini itself only knows
// "user" and "model"; the mapping happens in buildRequest.
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
	// Name is an optional discriminator carried by the adapter. Gemini has
	// no equivalent on a text part, so it is dropped.
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
	// JSONMode sets generationConfig.responseMimeType to
	// "application/json". Gemini's native JSON mode; no schema is attached
	// because the quiz/exercise prompts define their own contract and every
	// other provider in this repo behaves the same way.
	JSONMode bool
}

// Usage reports tokens billed by the upstream when available.
type Usage struct {
	PromptTokens     int
	CompletionTokens int
	TotalTokens      int
}

// ChatResponse is the provider-agnostic shape of a non-streamed reply.
// FinishReason carries Gemini's own enum (STOP, MAX_TOKENS, SAFETY, ...)
// unchanged, so callers and logs see what the vendor actually said.
type ChatResponse struct {
	Model        string
	Content      string
	FinishReason string
	Usage        Usage
}

// StreamChunk is a single event from GenerateStream. Delta is the
// incremental text. Done is true exactly once, on the final chunk, which
// may carry Usage.
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

// EmbedResponse pairs each input with its vector, preserving input order.
type EmbedResponse struct {
	Model   string
	Vectors [][]float32
	Usage   Usage
}

// Finish reasons that matter to this client. Gemini may return others
// (OTHER, LANGUAGE, ...); those pass through untouched.
const (
	finishReasonStop       = "STOP"
	finishReasonMaxTokens  = "MAX_TOKENS"
	finishReasonSafety     = "SAFETY"
	finishReasonRecitation = "RECITATION"
	finishReasonProhibited = "PROHIBITED_CONTENT"
	finishReasonBlocklist  = "BLOCKLIST"
)

// isBlockedReason reports whether a finishReason or a
// promptFeedback.blockReason means a content filter refused the request.
// These never succeed on retry — the prompt has to change.
func isBlockedReason(reason string) bool {
	switch reason {
	case finishReasonSafety, finishReasonRecitation,
		finishReasonProhibited, finishReasonBlocklist, "IMAGE_SAFETY":
		return true
	}
	return false
}

// ===========================================================================
// Wire types — the JSON actually exchanged with Gemini. Kept unexported so
// the public surface above stays stable if the vendor schema shifts.
// ===========================================================================

type wirePart struct {
	Text string `json:"text"`
}

type wireContent struct {
	// Role is "user" or "model". Omitted on systemInstruction, which has
	// no role.
	Role  string     `json:"role,omitempty"`
	Parts []wirePart `json:"parts"`
}

// wireGenerationConfig holds the sampling knobs. Pointers so an unset
// override is omitted entirely rather than sent as a zero value — sending
// temperature:0 is a real (and very different) instruction.
type wireGenerationConfig struct {
	Temperature      *float64 `json:"temperature,omitempty"`
	TopP             *float64 `json:"topP,omitempty"`
	MaxOutputTokens  *int     `json:"maxOutputTokens,omitempty"`
	StopSequences    []string `json:"stopSequences,omitempty"`
	ResponseMimeType string   `json:"responseMimeType,omitempty"`
}

// wireGenerateRequest is the :generateContent body. Note the model is NOT
// here — it is part of the URL path.
type wireGenerateRequest struct {
	Contents          []wireContent         `json:"contents"`
	SystemInstruction *wireContent          `json:"systemInstruction,omitempty"`
	GenerationConfig  *wireGenerationConfig `json:"generationConfig,omitempty"`
}

type wireUsageMetadata struct {
	PromptTokenCount     int `json:"promptTokenCount"`
	CandidatesTokenCount int `json:"candidatesTokenCount"`
	TotalTokenCount      int `json:"totalTokenCount"`
}

func (u *wireUsageMetadata) toUsage() Usage {
	if u == nil {
		return Usage{}
	}
	out := Usage{
		PromptTokens:     u.PromptTokenCount,
		CompletionTokens: u.CandidatesTokenCount,
		TotalTokens:      u.TotalTokenCount,
	}
	if out.TotalTokens == 0 {
		out.TotalTokens = out.PromptTokens + out.CompletionTokens
	}
	return out
}

type wireCandidate struct {
	Content      wireContent `json:"content"`
	FinishReason string      `json:"finishReason"`
	Index        int         `json:"index"`
}

// text joins every text part of the candidate. Gemini may split one reply
// across several parts, so taking parts[0] alone would silently truncate.
func (c wireCandidate) text() string {
	if len(c.Content.Parts) == 1 {
		return c.Content.Parts[0].Text
	}
	var b []byte
	for _, p := range c.Content.Parts {
		b = append(b, p.Text...)
	}
	return string(b)
}

type wirePromptFeedback struct {
	BlockReason string `json:"blockReason"`
}

type wireGenerateResponse struct {
	Candidates     []wireCandidate     `json:"candidates"`
	PromptFeedback *wirePromptFeedback `json:"promptFeedback,omitempty"`
	UsageMetadata  *wireUsageMetadata  `json:"usageMetadata,omitempty"`
	ModelVersion   string              `json:"modelVersion,omitempty"`
	Error          *wireError          `json:"error,omitempty"`
}

// ---------------------------------------------------------------------------
// Embeddings
// ---------------------------------------------------------------------------

// wireEmbedContentRequest is one element of a batchEmbedContents body. Each
// sub-request repeats the model in the "models/<id>" form.
type wireEmbedContentRequest struct {
	Model   string      `json:"model"`
	Content wireContent `json:"content"`
}

type wireBatchEmbedRequest struct {
	Requests []wireEmbedContentRequest `json:"requests"`
}

type wireBatchEmbedResponse struct {
	Embeddings []struct {
		Values []float32 `json:"values"`
	} `json:"embeddings"`
	UsageMetadata *wireUsageMetadata `json:"usageMetadata,omitempty"`
	Error         *wireError         `json:"error,omitempty"`
}

// ---------------------------------------------------------------------------
// Errors
// ---------------------------------------------------------------------------

// wireError covers BOTH documented Gemini error shapes:
//
//	google.rpc style: {"code": 429, "status": "RESOURCE_EXHAUSTED", ...}
//	newer style:      {"code": "rate_limit_exceeded", ...}
//
// Code is therefore `any` and rendered by codeString. Details carries the
// RetryInfo hint on a 429.
type wireError struct {
	Code    any    `json:"code"`
	Message string `json:"message"`
	Status  string `json:"status,omitempty"`
	Details []struct {
		Type       string `json:"@type"`
		RetryDelay string `json:"retryDelay"`
	} `json:"details,omitempty"`
}
