package controller

import (
	"encoding/json"
	"fmt"
	"net/http"

	"math-ai.com/math-ai/internal/app/resources"
	"math-ai.com/math-ai/internal/applications/dto"
	di "math-ai.com/math-ai/internal/core/di/services"
	"math-ai.com/math-ai/internal/shared/constant/status"
	"math-ai.com/math-ai/internal/shared/logger"
	"math-ai.com/math-ai/internal/shared/utils/response"
)

type ChatBoxController struct {
	appResources *resources.AppResource
	service      di.IChatBoxService
}

func NewChatBoxController(appResources *resources.AppResource, service di.IChatBoxService) *ChatBoxController {
	return &ChatBoxController{
		appResources: appResources,
		service:      service,
	}
}

// HandleGenerateQuiz handles POST /generate-quiz requests
func (c *ChatBoxController) HandleGenerateQuiz(w http.ResponseWriter, r *http.Request) {
	var req dto.GenerateQuizRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, r.Context(), nil, fmt.Errorf("invalid request body"), status.BAD_REQUEST)
		return
	}

	// Check if streaming is requested
	if req.Stream {
		c.handleStreamingResponse(w, r, &req)
		return
	}

	// Send message to service
	statusCode, chatResponse, err := c.service.GenerateQuiz(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, r.Context(), nil, err, statusCode)
		return
	}

	res := dto.GenerateQuizResponse{
		Result: chatResponse,
	}

	response.WriteJson(w, r.Context(), res, nil, statusCode)
}

// HandleSubmitQuizAnswer handles POST /submit-quiz requests
func (c *ChatBoxController) HandleSubmitQuizAnswer(w http.ResponseWriter, r *http.Request) {
	var req dto.SubmitQuizRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJson(w, r.Context(), nil, fmt.Errorf("invalid request body"), status.BAD_REQUEST)
		return
	}

	// Send message to service
	statusCode, result, err := c.service.SubmitQuiz(r.Context(), &req)
	if err != nil {
		response.WriteJson(w, r.Context(), nil, err, statusCode)
		return
	}

	res := dto.SubmitQuizResponse{
		Result: result,
	}

	response.WriteJson(w, r.Context(), res, nil, statusCode)
}

// handleStreamingResponse handles streaming chat responses
func (c *ChatBoxController) handleStreamingResponse(w http.ResponseWriter, r *http.Request, req *dto.GenerateQuizRequest) {
	// Set headers for SSE (Server-Sent Events)
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	// Get the flusher
	flusher, ok := w.(http.Flusher)
	if !ok {
		response.WriteJson(w, r.Context(), nil, fmt.Errorf("streaming not supported"), status.INTERNAL)
		return
	}

	// Send message to service with streaming
	statusCode, streamChan, err := c.service.SendMessageStream(r.Context(), req)
	if err != nil {
		response.WriteJson(w, r.Context(), nil, err, statusCode)
		return
	}

	// Stream the response
	encoder := json.NewEncoder(w)
	for chunk := range streamChan {
		if chunk.Error != nil {
			logger.Errorf("Stream error: %v", chunk.Error)
			// Send error chunk
			fmt.Fprintf(w, "data: %s\n\n", mustMarshalJSON(map[string]interface{}{
				"error": chunk.Error.Error(),
				"done":  true,
			}))
			flusher.Flush()
			break
		}

		// Send chunk as SSE
		chunkData := map[string]interface{}{
			"delta":         chunk.Delta,
			"finish_reason": chunk.FinishReason,
			"done":          chunk.Done,
		}

		if err := encoder.Encode(chunkData); err != nil {
			logger.Errorf("Failed to encode chunk: %v", err)
			break
		}
		flusher.Flush()

		if chunk.Done {
			break
		}
	}
}

// mustMarshalJSON marshals a value to JSON, panicking on error
func mustMarshalJSON(v interface{}) string {
	data, err := json.Marshal(v)
	if err != nil {
		logger.Errorf("Failed to marshal JSON: %v", err)
		return "{\"error\": \"internal server error\"}"
	}
	return string(data)
}
