package response

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"math-ai.com/math-ai/internal/shared/constant/status"
	err_svc "math-ai.com/math-ai/internal/shared/error"
	"math-ai.com/math-ai/internal/shared/metadata"
	"math-ai.com/math-ai/internal/shared/utils/message"
)

func WriteJson(w http.ResponseWriter, ctx context.Context, data any, err error, statusCode status.Code) {
	payload := make(map[string]any)

	if data != nil {
		dataBytes, err := json.Marshal(data)
		if err != nil {
			log.Printf("WriteJson: failed to marshal data: %v\n", err)
			return
		}
		var tmp map[string]any
		err = json.Unmarshal(dataBytes, &tmp)
		if err == nil || tmp != nil {
			payload = tmp
		}
		// payload["result"] = data
	}

	if err != nil {
		payload["error"] = err.Error()
		// Check if error is a DynamicError with dynamic arguments
		if dynErr, ok := err_svc.IsDynamicError(err); ok {
			// Use the localized message with dynamic arguments
			payload["message"] = GetMessage(ctx, dynErr.GetStatusCode(), dynErr.GetArgs())
		} else {
			// Use standard localized message
			payload["message"] = GetMessage(ctx, statusCode, nil)
		}
	}

	// Default to not set if not set
	if statusCode != 0 {
		payload["status"] = statusCode
	}

	if (payload["message"] == "Unknown" || payload["message"] == "") && err != nil {
		payload["message"] = err.Error()
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(payload)
}

// GetMessageFromStatusCodeWithArgs returns a localized message with dynamic arguments interpolated
// This supports dynamic error messages like "Email john@example.com already exists"
func GetMessage(ctx context.Context, statusCode status.Code, args map[string]interface{}) string {
	lan := metadata.GetLanguage(ctx)
	return message.GetMessage(message.LanguageType(lan), statusCode, args)
}
