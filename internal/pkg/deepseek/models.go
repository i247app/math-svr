package deepseek

import (
	"context"
	"encoding/json"
	"fmt"

	"math-ai.com/math-ai/internal/infrastructure/logger"
)

const modelsPath = "/models"

// Model is one entry of GET /models (OpenAI-compatible shape: id +
// owned_by, no display name, chat models only today).
type Model struct {
	ID      string
	OwnedBy string
}

type wireModelList struct {
	Data []struct {
		ID      string `json:"id"`
		OwnedBy string `json:"owned_by"`
	} `json:"data"`
}

// ListModels returns the models the API key can use, under the configured
// timeout. One attempt, no retry: the list is informational.
func (c *Client) ListModels(ctx context.Context) ([]Model, error) {
	callCtx, cancel := context.WithTimeout(ctx, c.cfg.Timeout)
	defer cancel()

	resp, err := c.http.Get(callCtx, modelsPath)
	if err != nil {
		if ctxErr := callCtx.Err(); ctxErr != nil {
			return nil, ctxErr
		}
		return nil, fmt.Errorf("deepseek: request failed: %w", err)
	}
	logger.From(ctx).Infof("deepseek.http path=%s status=%d bytes=%d", modelsPath, resp.StatusCode, len(resp.Body))
	if !resp.IsSuccess() {
		return nil, classifyAPIError(errorFromResponse(resp))
	}

	var wire wireModelList
	if err := json.Unmarshal(resp.Body, &wire); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDecodeResponse, err)
	}
	out := make([]Model, 0, len(wire.Data))
	for _, m := range wire.Data {
		if m.ID != "" {
			out = append(out, Model{ID: m.ID, OwnedBy: m.OwnedBy})
		}
	}
	return out, nil
}
