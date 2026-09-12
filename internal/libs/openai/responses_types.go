package openai

// Public and wire types for the Responses API (POST /v1/responses).
//
// The Responses API is the second endpoint this client speaks. Chat
// Completions stays the path for stateless generation (quiz, exam without
// a session); Responses exists for one reason — server-side conversation
// state via previous_response_id — and this file names only the slice of
// the surface that reason needs. Tools, background mode, streaming, file
// and image inputs are deliberately absent.

const responsesPath = "/responses"

// RespondRequest is one turn against POST /responses.
//
// Instructions is the system prompt. It is NOT inherited across
// PreviousResponseID — the vendor documents "resend stable instructions on
// each request" — so callers pass it on every turn, chained or not.
//
// Store must be true whenever PreviousResponseID is set: a response can
// only be chained on if it was retained. The client rejects the
// combination at build time rather than letting the vendor reject it per
// call. Store is a per-request decision here, independent of
// Config.Store, which governs chat-completions logging only.
type RespondRequest struct {
	// Model overrides Config.Model for this call. Optional.
	Model string

	Instructions string

	// Input is the new content for this turn. On a chained turn it is
	// typically a single user message; on a fresh turn it may carry a
	// rebuilt history. Must not be empty.
	Input []Message

	// PreviousResponseID chains this turn onto an earlier response the
	// vendor still holds. Empty means "start a new chain".
	PreviousResponseID string

	// Store retains the response on the vendor side (30 days) so a later
	// turn can chain on it. Required when PreviousResponseID is set.
	Store bool

	// Sampling — same conventions as ChatRequest: negative means "use
	// config default", zero MaxTokens means "no cap".
	Temperature float64
	TopP        float64
	MaxTokens   int

	// JSONMode requests a JSON-object response (text.format).
	JSONMode bool
}

// RespondResponse is the parsed outcome of one Responses-API turn.
type RespondResponse struct {
	// ID is the chain pointer: pass it as PreviousResponseID next turn.
	ID string
	// Model is the id the vendor actually served.
	Model string
	// Content is the assembled text of every message item in output.
	Content string
	// Status is the vendor lifecycle value: "completed" on a full answer,
	// "incomplete" when generation stopped early (see IncompleteReason).
	Status string
	// IncompleteReason is set when Status == "incomplete", e.g.
	// "max_output_tokens".
	IncompleteReason string
	Usage            Usage
}

// --- wire: request -------------------------------------------------------

// wireResponsesRequest mirrors the request body. Pointer sampling fields
// are omitted rather than sent as zero, exactly as in wireChatRequest.
//
// Note the field shapes that differ from chat-completions and would be
// silently ignored if sent the chat way: max_output_tokens (not
// max_completion_tokens), reasoning.effort (nested, not flat
// reasoning_effort), text.format (not response_format).
type wireResponsesRequest struct {
	Model              string             `json:"model"`
	Instructions       string             `json:"instructions,omitempty"`
	Input              []wireInputMessage `json:"input"`
	PreviousResponseID string             `json:"previous_response_id,omitempty"`
	Store              *bool              `json:"store,omitempty"`
	Temperature        *float64           `json:"temperature,omitempty"`
	TopP               *float64           `json:"top_p,omitempty"`
	MaxOutputTokens    *int               `json:"max_output_tokens,omitempty"`
	Reasoning          *wireReasoning     `json:"reasoning,omitempty"`
	Text               *wireTextConfig    `json:"text,omitempty"`
	Metadata           map[string]string  `json:"metadata,omitempty"`
}

// samplingCarrier implementation — see Client.postSampled.
func (r *wireResponsesRequest) samplingModel() string { return r.Model }
func (r *wireResponsesRequest) hasSampling() bool     { return r.Temperature != nil || r.TopP != nil }
func (r *wireResponsesRequest) clearSampling()        { r.Temperature, r.TopP = nil, nil }

// wireInputMessage is the simple role/content form of an input item.
// The API also accepts typed item objects; the plain form is enough for
// text turns and keeps the body readable in the dashboard.
type wireInputMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type wireReasoning struct {
	Effort string `json:"effort,omitempty"`
}

type wireTextConfig struct {
	Format wireTextFormat `json:"format"`
}

type wireTextFormat struct {
	Type string `json:"type"`
}

// --- wire: response ------------------------------------------------------

type wireResponsesResponse struct {
	ID                string                 `json:"id"`
	Model             string                 `json:"model"`
	Status            string                 `json:"status"`
	IncompleteDetails *wireIncompleteDetails `json:"incomplete_details"`
	Output            []wireOutputItem       `json:"output"`
	Usage             *wireResponsesUsage    `json:"usage"`
	Error             *wireError             `json:"error"`
}

type wireIncompleteDetails struct {
	Reason string `json:"reason"`
}

// wireOutputItem is one entry of output[]. Only type "message" carries
// user-visible text; "reasoning" items (present when reasoning effort is
// above none) are skipped, never surfaced.
type wireOutputItem struct {
	Type    string              `json:"type"`
	Role    string              `json:"role,omitempty"`
	Content []wireOutputContent `json:"content,omitempty"`
}

// wireOutputContent is one content part of a message item. Type is
// "output_text" for prose and "refusal" when the model declined.
type wireOutputContent struct {
	Type    string `json:"type"`
	Text    string `json:"text,omitempty"`
	Refusal string `json:"refusal,omitempty"`
}

// wireResponsesUsage uses the Responses vocabulary (input/output) rather
// than chat-completions' prompt/completion. toUsage folds it into the
// shared Usage so the adapter sees one shape.
type wireResponsesUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
	TotalTokens  int `json:"total_tokens"`
}

func (u *wireResponsesUsage) toUsage() Usage {
	if u == nil {
		return Usage{}
	}
	return Usage{
		PromptTokens:     u.InputTokens,
		CompletionTokens: u.OutputTokens,
		TotalTokens:      u.TotalTokens,
	}
}

const (
	responseStatusIncomplete = "incomplete"

	outputItemMessage    = "message"
	outputContentText    = "output_text"
	outputContentRefusal = "refusal"
	textFormatJSONObject = "json_object"
)
