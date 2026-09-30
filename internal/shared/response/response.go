// Package response writes the JSON envelope every HTTP handler answers with.
//
// The HTTP status line is always 200; the outcome travels in the body's
// "mstatus" field, which is what the mobile client reads. A JSON-object
// payload is flattened into the envelope; anything else (array, scalar,
// null) is nested under "result".
package response

import (
	"encoding/json"
	"fmt"
	"net/http"

	errs "math-ai.com/math-ai/internal/domain/shared/error"
	"math-ai.com/math-ai/internal/domain/shared/status"
)

// Envelope keys — part of the wire contract, the client reads them by name.
const (
	keyStatus  = "status"
	keyMStatus = "mstatus"
	keyMessage = "mmessage"
	keyDebug   = "debug"
	keyResult  = "result"

	statusSuccess = "Success"
)

// WriteJson writes data and err as one envelope. On success it adds
// {"status":"Success","mstatus":200}; on error it adds mstatus/mmessage
// (plus "debug" for a MathError) and keeps any data beside them. Envelope
// keys always win over data fields of the same name.
func WriteJson(w http.ResponseWriter, data any, err error) {
	writeEnvelope(w, data, err, status.SUCCESS)
}

// WriteJsonNoContent writes the "deleted / nothing to return" outcome —
// {"status":"Success","mstatus":204}, still with HTTP 200. data is optional
// (nil for a bare envelope) and is merged exactly as WriteJson merges it.
func WriteJsonNoContent(w http.ResponseWriter, data any) {
	writeEnvelope(w, data, nil, status.NO_CONTENT)
}

func WriteJsonNoData(w http.ResponseWriter) {
	writeEnvelope(w, nil, nil, status.NO_DATA)
}

// writeEnvelope is the one path both writers share: data first, then the
// error half or the success half (with successCode) on top of it.
func writeEnvelope(w http.ResponseWriter, data any, err error, successCode status.StatusCode) {
	payload, encErr := flatten(data)
	if err == nil {
		err = encErr
	}

	if err != nil {
		setError(payload, err)
	} else {
		setSuccess(payload, successCode)
	}
	write(w, payload)
}

// flatten encodes data once and returns the envelope seeded with it. Object
// fields are carried as json.RawMessage, so they are never decoded into
// float64 — an int64 above 2^53 would otherwise lose its low digits.
//
// On an encode failure it returns an empty envelope and the error, so the
// caller reports it instead of sending a blank body.
func flatten(data any) (map[string]any, error) {
	if data == nil {
		return make(map[string]any, 3), nil
	}

	raw, err := json.Marshal(data)
	if err != nil {
		return make(map[string]any, 2), fmt.Errorf("response: encode data: %w", err)
	}

	// json.Marshal output is compact, so the first byte identifies the kind.
	if string(raw) == "null" {
		return make(map[string]any, 3), nil // nil pointer/slice/map: nothing to add
	}
	if raw[0] != '{' {
		return map[string]any{keyResult: json.RawMessage(raw)}, nil
	}

	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil {
		return make(map[string]any, 2), fmt.Errorf("response: decode data: %w", err)
	}
	payload := make(map[string]any, len(fields)+3)
	for k, v := range fields {
		payload[k] = v
	}
	return payload, nil
}

func setSuccess(payload map[string]any, code status.StatusCode) {
	payload[keyStatus] = statusSuccess
	payload[keyMStatus] = code
}

// setError fills the error half of the envelope. A MathError carries its own
// code and localized message; any other error is an internal error.
func setError(payload map[string]any, err error) {
	mathErr, ok := errs.IsMathError(err)
	if !ok {
		payload[keyMStatus] = status.INTERNAL_SERVER_ERROR
		payload[keyMessage] = err.Error()
		return
	}

	msg := string(mathErr.GetStatusMessage())
	if msg == "" {
		msg = mathErr.Error()
	}
	payload[keyMStatus] = mathErr.GetStatusCode()
	payload[keyMessage] = msg
	payload[keyDebug] = mathErr.Error()
}

func write(w http.ResponseWriter, payload map[string]any) {
	h := w.Header()
	h.Set("Content-Type", "application/json; charset=utf-8")
	h.Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(http.StatusOK)
	// Every value is a string, a status code, or already-valid JSON, so this
	// can only fail on the write itself — the client is gone, nothing to do.
	_ = json.NewEncoder(w).Encode(payload)
}
