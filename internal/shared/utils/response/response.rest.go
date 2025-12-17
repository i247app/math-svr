package response

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"math-ai.com/math-ai/internal/shared/constant/status"
	appctx "math-ai.com/math-ai/internal/shared/utils/context"
	"math-ai.com/math-ai/internal/shared/utils/locales"
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
		payload["message"] = GetMessageFromStatusCode(ctx, statusCode)
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

func GetMessageFromStatusCode(ctx context.Context, statusCode status.Code) string {
	lan := appctx.GetLocale(ctx)

	switch locales.LanguageType(lan) {
	case locales.EN:
		return locales.GetMessageENFromStatus(statusCode)
	case locales.VN:
		return locales.GetMessageVNFromStatus(statusCode)
	case locales.FR:
		return locales.GetMessageFRFromStatus(statusCode)
	default:
		return locales.GetMessageENFromStatus(statusCode)
	}
}
