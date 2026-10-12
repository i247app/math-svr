package openai

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"math-ai.com/math-ai/internal/infrastructure/logger"
)

// Respond runs one turn against POST /responses under the configured
// timeout and retry budget.
//
// Chaining: when req.PreviousResponseID is set the vendor prepends the
// whole stored chain to this turn; the caller sends only what is new.
// The returned ID is the pointer for the next turn. Instructions are
// resent every turn because the vendor does not carry them across the
// chain.
//
// Errors:
//   - ErrInvalidConfig      — empty Input, or PreviousResponseID without Store.
//   - ErrPreviousResponseGone — the pointer is no longer usable (expired,
//     wrong, or foreign). The caller decides whether to rebuild.
//   - ErrContentRefused     — 200 whose output is a refusal, or Status
//     "incomplete" with reason "content_filter".
//   - ErrDecodeResponse     — 200 with no message text, or Status
//     "incomplete" for any other reason (max_output_tokens, ...) — the partial
//     response is returned alongside so the caller can log its ID, but
//     it must not be chained on.
//   - everything else       — same classification as Generate.
func (c *Client) Respond(ctx context.Context, req RespondRequest) (*RespondResponse, error) {
	if len(req.Input) == 0 {
		return nil, fmt.Errorf("%w: at least one input message is required", ErrInvalidConfig)
	}
	// A chained turn needs the previous response to have been stored, and
	// the vendor would need to store this one for the NEXT turn. Refusing
	// here turns a per-call 400 into a boot-time-style config error.
	if req.PreviousResponseID != "" && !req.Store {
		return nil, fmt.Errorf("%w: PreviousResponseID requires Store=true", ErrInvalidConfig)
	}

	callCtx, cancel := context.WithTimeout(ctx, c.cfg.Timeout)
	defer cancel()

	wireReq := c.buildResponsesRequest(req)
	chained := req.PreviousResponseID != ""

	body, err := c.postSampled(callCtx, responsesPath, &wireReq)
	if err != nil {
		var api *APIError
		if errors.As(err, &api) && previousResponseGoneShape(api, chained) {
			return nil, fmt.Errorf("%w: %w", ErrPreviousResponseGone, api)
		}
		return nil, err
	}

	var wire wireResponsesResponse
	if err := json.Unmarshal(body, &wire); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDecodeResponse, err)
	}
	if wire.Error != nil {
		return nil, classifyAPIError(wire.Error.toAPIError(200, 0))
	}
	if wire.ID == "" {
		return nil, fmt.Errorf("%w: response carried no id", ErrDecodeResponse)
	}

	resp := &RespondResponse{
		ID:     wire.ID,
		Model:  firstNonEmpty(wire.Model, req.Model, c.cfg.Model),
		Status: wire.Status,
		Usage:  wire.Usage.toUsage(),
	}
	if wire.IncompleteDetails != nil {
		resp.IncompleteReason = wire.IncompleteDetails.Reason
	}

	content, refused := assembleOutput(wire.Output)
	resp.Content = content

	// Operator metadata only — never the prompt or the response body.
	var cached, reasoning int
	if wire.Usage != nil {
		cached = wire.Usage.InputTokensDetails.CachedTokens
		reasoning = wire.Usage.OutputTokensDetails.ReasoningTokens
	}
	logger.From(ctx).Infof("openai.respond id=%s model=%s status=%s chained=%v input_tokens=%d cached_tokens=%d output_tokens=%d reasoning_tokens=%d",
		resp.ID, resp.Model, resp.Status, chained,
		resp.Usage.PromptTokens, cached, resp.Usage.CompletionTokens, reasoning)

	// Order matters below: a refusal is a decision about the content, a
	// truncation is a decision about the budget, an empty body is a
	// decode problem. Each is reported with the response attached so the
	// caller can still see the id that was minted.
	if refused {
		return resp, fmt.Errorf("%w: output contained a refusal item", ErrContentRefused)
	}
	if wire.Status == responseStatusIncomplete && resp.IncompleteReason == incompleteContentFilter {
		return resp, fmt.Errorf("%w: response stopped by the content filter", ErrContentRefused)
	}
	if wire.Status == responseStatusIncomplete {
		// max_output_tokens is the budget case; other reasons (max_messages,
		// steered) are reported as-is so the log does not point at the
		// wrong knob.
		hint := ""
		if resp.IncompleteReason == "max_output_tokens" {
			hint = " — raise MaxTokens"
		}
		return resp, fmt.Errorf("%w: response incomplete (reason=%s, output_tokens=%d)%s",
			ErrDecodeResponse, firstNonEmpty(resp.IncompleteReason, "unknown"), resp.Usage.CompletionTokens, hint)
	}
	if strings.TrimSpace(content) == "" {
		return resp, fmt.Errorf("%w: response carried no message text", ErrDecodeResponse)
	}
	return resp, nil
}

// buildResponsesRequest translates a RespondRequest plus the Client's
// defaults into the wire body. Same precedence as buildRequest: per-call
// wins, then config, then omit so the model default stands.
func (c *Client) buildResponsesRequest(req RespondRequest) wireResponsesRequest {
	model := firstNonEmpty(req.Model, c.cfg.Model)

	input := make([]wireInputMessage, 0, len(req.Input)+1)
	mentionsJSON := false
	for _, m := range req.Input {
		input = append(input, wireInputMessage{Role: string(m.Role), Content: m.Content})
		mentionsJSON = mentionsJSON || strings.Contains(strings.ToLower(m.Content), "json")
	}
	// json_object mode is refused (400, param=input) unless an INPUT
	// message names JSON — instructions do not count, unlike the system
	// message on chat-completions. Moving the system prompt into input
	// would re-bill it on every chained turn, so a one-line developer
	// message is added instead.
	if req.JSONMode && !mentionsJSON {
		input = append(input, wireInputMessage{Role: roleDeveloper, Content: jsonModeInputNudge})
	}

	store := req.Store
	out := wireResponsesRequest{
		Model:              model,
		Instructions:       req.Instructions,
		Input:              input,
		PreviousResponseID: req.PreviousResponseID,
		// Always explicit. The vendor defaults this endpoint to true; the
		// intent must be visible in the body, not inherited from a default.
		Store: &store,
	}
	// Metadata is only retained alongside a stored response, same rule as
	// chat-completions.
	if store && len(c.cfg.Metadata) > 0 {
		out.Metadata = c.cfg.Metadata
	}

	if !c.samplingBlocked(model) {
		temp := c.cfg.Temperature
		if req.Temperature >= 0 {
			temp = req.Temperature
		}
		if temp >= 0 {
			out.Temperature = &temp
		}
		topP := c.cfg.TopP
		if req.TopP >= 0 {
			topP = req.TopP
		}
		if topP >= 0 {
			out.TopP = &topP
		}
	}

	maxTokens := c.cfg.MaxTokens
	if req.MaxTokens > 0 {
		maxTokens = req.MaxTokens
	}
	if maxTokens > 0 {
		out.MaxOutputTokens = &maxTokens
	}

	if c.cfg.ReasoningEffort != "" {
		out.Reasoning = &wireReasoning{Effort: c.cfg.ReasoningEffort}
	}
	if req.JSONMode {
		out.Text = &wireTextConfig{Format: wireTextFormat{Type: textFormatJSONObject}}
	}
	return out
}

// assembleOutput joins the text of every message item, in order, and
// reports whether any content part was a refusal. Reasoning items and
// any other item types are skipped: they carry no user-visible text.
func assembleOutput(items []wireOutputItem) (content string, refused bool) {
	var b strings.Builder
	for _, item := range items {
		if item.Type != outputItemMessage {
			continue
		}
		for _, part := range item.Content {
			switch part.Type {
			case outputContentText:
				b.WriteString(part.Text)
			case outputContentRefusal:
				refused = true
			}
		}
	}
	return b.String(), refused
}
