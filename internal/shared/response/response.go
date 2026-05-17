package response

import (
	"encoding/json"
	"log"
	"net/http"

	"math-ai.com/math-ai/internal/domain/shared/status"

	errs "math-ai.com/math-ai/internal/domain/shared/error"
)

func WriteJson(w http.ResponseWriter, data any, err error) {
	payload := make(map[string]any)

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
		// case for error with result data
		if payload == nil {
			payload = make(map[string]any)
		}

		// Check if error is a binbaseError
		if mathErr, ok := errs.IsMathError(err); ok {
			payload["mstatus"] = mathErr.GetStatusCode()
			payload["mmessage"] = mathErr.GetStatusMessage()
			payload["debug"] = mathErr.Error()
		} else {
			payload["mstatus"] = status.INTERNAL_SERVER_ERROR
			payload["mmessage"] = err.Error()
		}
	} else {
		payload["status"] = "Success"
		payload["mstatus"] = status.SUCCESS
	}

	if payload["mmessage"] == "" && err != nil {
		payload["mmessage"] = err.Error()
	}

	if _, exists := payload["mstatus"]; !exists {
		payload["mstatus"] = status.INTERNAL_SERVER_ERROR
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(payload)
}

// WriteJsonNoContent emits a "deleted / no content" success envelope:
// {"status":"Success","mstatus":204} with HTTP 200 (matching the WriteJson
// convention where the JSON body's mstatus, not the HTTP status line, carries
// the semantic outcome).
func WriteJsonNoContent(w http.ResponseWriter) {
	payload := map[string]any{
		"status":  "Success",
		"mstatus": status.NO_CONTENT,
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(payload)
}
