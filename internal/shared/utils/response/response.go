package response

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"math-ai.com/math-ai/internal/shared/constant/status"
	appctx "math-ai.com/math-ai/internal/shared/utils/context"
)

func WriteJson(w http.ResponseWriter, ctx context.Context, data any, err error, statusCode status.Code, args ...any) {
	payload := make(map[string]any)

	// If there's data, try to unmarshal data into being the payload
	if data != nil {
		dataBytes, err := json.Marshal(data)
		if err != nil {
			log.Printf("WriteJson: failed to marshal data: %v\n", err)
			return
		}
		var tmp map[string]any
		err = json.Unmarshal(dataBytes, &tmp)
		if err != nil || tmp == nil {
			// If this fails, just add the data to an empty payload as "result"
			payload["result"] = data
		} else {
			payload = tmp
		}
	}

	if err != nil {
		payload["error"] = err.Error()
	}

	// Default to not set if not set
	if statusCode != 0 {
		payload["status"] = statusCode
		payload["message"] = GetMessageFromStatusCode(ctx, statusCode, args...)
	} else {
		payload["status"] = status.INTERNAL
	}

	if payload["message"] == "Unknown" && err != nil {
		payload["message"] = err.Error()
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(payload)
}

func GetMessageFromStatusCode(ctx context.Context, statusCode status.Code, args ...any) string {
	lan := appctx.GetLocale(ctx)

	return lan
}
