package gemini

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"math-ai.com/math-ai/internal/infrastructure/logger"
	"math-ai.com/math-ai/internal/shared/http_client"
)

const (
	modelsPath = "/models"
	// modelsPageSize is the vendor's maximum page size.
	modelsPageSize = "1000"
	// maxModelPages bounds the paging loop against a vendor that keeps
	// returning a nextPageToken.
	maxModelPages = 10
)

// Model is one entry of GET /models. ID drops the "models/" prefix — it is
// the value llm_model (and LLM_GEMINI_MODEL) takes. SupportedMethods tells
// a chat model ("generateContent") from an embedding one ("embedContent").
type Model struct {
	ID               string
	Name             string
	InputTokenLimit  int
	SupportedMethods []string
	OutputTokenLimit int
}

type wireModelList struct {
	Models []struct {
		Name                       string   `json:"name"`
		DisplayName                string   `json:"displayName"`
		InputTokenLimit            int      `json:"inputTokenLimit"`
		OutputTokenLimit           int      `json:"outputTokenLimit"`
		SupportedGenerationMethods []string `json:"supportedGenerationMethods"`
	} `json:"models"`
	NextPageToken string `json:"nextPageToken"`
}

// ListModels returns every model the API key can use, following
// nextPageToken (at most maxModelPages pages), all pages under ONE
// configured timeout. One attempt per page, no retry: the list is
// informational.
func (c *Client) ListModels(ctx context.Context) ([]Model, error) {
	callCtx, cancel := context.WithTimeout(ctx, c.cfg.Timeout)
	defer cancel()

	var out []Model
	token := ""
	for page := range maxModelPages {
		opts := []http_client.RequestOption{http_client.WithRequestQueryParam("pageSize", modelsPageSize)}
		if token != "" {
			opts = append(opts, http_client.WithRequestQueryParam("pageToken", token))
		}
		resp, err := c.http.Get(callCtx, modelsPath, opts...)
		if err != nil {
			if ctxErr := callCtx.Err(); ctxErr != nil {
				return nil, ctxErr
			}
			return nil, fmt.Errorf("gemini: request failed: %w", err)
		}
		// The key travels in a header (never ?key=), so the logged path is safe.
		logger.From(ctx).Infof("gemini.http path=%s page=%d status=%d bytes=%d", modelsPath, page+1, resp.StatusCode, len(resp.Body))
		if !resp.IsSuccess() {
			return nil, classifyAPIError(errorFromResponse(resp))
		}

		var wire wireModelList
		if err := json.Unmarshal(resp.Body, &wire); err != nil {
			return nil, fmt.Errorf("%w: %v", ErrDecodeResponse, err)
		}
		for _, m := range wire.Models {
			id := strings.TrimPrefix(m.Name, "models/")
			if id == "" {
				continue
			}
			out = append(out, Model{
				ID:               id,
				Name:             m.DisplayName,
				InputTokenLimit:  m.InputTokenLimit,
				OutputTokenLimit: m.OutputTokenLimit,
				SupportedMethods: m.SupportedGenerationMethods,
			})
		}
		if wire.NextPageToken == "" {
			return out, nil
		}
		token = wire.NextPageToken
	}
	logger.From(ctx).Warnf("gemini.list_models stopped after %d pages with a page token left", maxModelPages)
	return out, nil
}
