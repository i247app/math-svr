package openrouter

import (
	"context"
	"encoding/json"
	"fmt"

	"math-ai.com/math-ai/internal/infrastructure/logger"
)

const modelsPath = "/models"

// Model is one entry of GET /models. ID is "vendor/model" — the value
// llm_model takes on this provider. OutputModalities tells a text model
// ("text") from an image one ("image"); ContextLength is in tokens.
type Model struct {
	ID               string
	Name             string
	ContextLength    int
	InputModalities  []string
	OutputModalities []string
}

type wireModelList struct {
	Data []struct {
		ID            string `json:"id"`
		Name          string `json:"name"`
		ContextLength int    `json:"context_length"`
		Architecture  struct {
			InputModalities  []string `json:"input_modalities"`
			OutputModalities []string `json:"output_modalities"`
		} `json:"architecture"`
	} `json:"data"`
}

// ListModels returns OpenRouter's model catalog (several hundred entries),
// under the configured timeout. One attempt, no retry: the list is
// informational.
func (c *Client) ListModels(ctx context.Context) ([]Model, error) {
	callCtx, cancel := context.WithTimeout(ctx, c.cfg.Timeout)
	defer cancel()

	resp, err := c.http.Get(callCtx, modelsPath)
	if err != nil {
		if ctxErr := callCtx.Err(); ctxErr != nil {
			return nil, ctxErr
		}
		return nil, fmt.Errorf("openrouter: request failed: %w", err)
	}
	logger.From(ctx).Infof("openrouter.http path=%s status=%d bytes=%d", modelsPath, resp.StatusCode, len(resp.Body))
	if !resp.IsSuccess() {
		return nil, classifyAPIError(errorFromResponse(resp))
	}

	var wire wireModelList
	if err := json.Unmarshal(resp.Body, &wire); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDecodeResponse, err)
	}
	out := make([]Model, 0, len(wire.Data))
	for _, m := range wire.Data {
		if m.ID == "" {
			continue
		}
		out = append(out, Model{
			ID:               m.ID,
			Name:             m.Name,
			ContextLength:    m.ContextLength,
			InputModalities:  m.Architecture.InputModalities,
			OutputModalities: m.Architecture.OutputModalities,
		})
	}
	return out, nil
}
