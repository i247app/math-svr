package middleware

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net"
	"net/http"
	"strings"
	"time"

	"math-ai.com/math-ai/internal/shared/colors"
	"math-ai.com/math-ai/internal/shared/utils"

	"math-ai.com/math-ai/internal/infrastructure/logger"
)

// LogRequestMiddleware logs every HTTP request/response pair through the
// project logger. Sensitive headers and JSON body fields are masked before
// they reach the log output (and therefore the persistent log file).
//
// Placement: between LoggerMiddleware and RecoveryMiddleware. That gives
// it three properties at once:
//   - logger.From(r.Context()) returns the per-request AppLogger that
//     LoggerMiddleware injected, so log lines carry the same [token]/[uid]
//     prefix as everything else.
//   - Recovery sits inside this middleware, so a panic-rendered 500
//     response still flows through the response recorder and gets logged.
//   - Gzip sits inside this middleware as well — when a client negotiated
//     gzip, the response bytes reaching the recorder are already
//     compressed. decodeResponseBody gunzips them so the logged BODY
//     matches what Postman / curl actually display.
func LogRequestMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if shouldSkipLogging(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		started := time.Now()
		lg := logger.From(r.Context())

		// 1. Inbound: read & restore request body when it's safe to do so.
		reqBody, reqBodyNote := snapshotRequestBody(r)

		lg.Info(colors.Colorize(colors.FGReset, colors.BGYellow, "INCOMING REQUEST"))
		lg.Infof("Client: %s", clientAddr(r))
		lg.Infof("HEADERS: %s", utils.ToIndentedJson(maskHeaders(r.Header)))
		lg.Infof("BODY: %s", maskBodyForLog(reqBody, r.Header.Get("Content-Type"), reqBodyNote))

		// 2. Outbound: wrap the writer to capture status & a bounded
		// slice of the response body.
		rec := newResponseRecorder(w)
		next.ServeHTTP(rec, r)

		// Decode compressed bodies (currently gzip; other encodings fall
		// back to a placeholder note). When the response stream was cut
		// short by responseCaptureCap, decoding may yield partial bytes
		// — we log what's there and tag a truncation marker.
		decoded, respNote := decodeResponseBody(rec.body.Bytes(), rec.Header().Get("Content-Encoding"))
		if respNote == "" && rec.truncated {
			respNote = bodyTruncatedNote
		}

		lg.Info(colors.Colorize(colors.FGReset, colors.BGYellow, "OUTGOING RESPONSE"))
		lg.Infof("STATUS: %d", rec.statusCode)
		lg.Infof("HEADERS: %s", utils.ToIndentedJson(maskHeaders(rec.Header())))
		lg.Infof("BODY: %s", maskBodyForLog(decoded, rec.Header().Get("Content-Type"), respNote))
		lg.Infof("DURATION: %dms", time.Since(started).Milliseconds())
	})
}

// ──────────────────────────────────────────────────────────────────────────
// Configuration (package-level for now; promote to Options if needed).
// ──────────────────────────────────────────────────────────────────────────

const (
	// logBodyCap is the upper bound on how much body content reaches the
	// final log line. maskBodyForLog truncates to this size. Keep it
	// small — log lines past a few KB are unwieldy in tail / grep.
	logBodyCap = 4 * 1024

	// responseCaptureCap is the upper bound on bytes the response
	// recorder buffers for later inspection. It's larger than logBodyCap
	// because compressed responses (gzip) may need a multi-KB capture
	// window to decode into a few KB of readable text. Memory is bounded
	// per request.
	responseCaptureCap = 64 * 1024

	// Hard cap on request body sizes we'll buffer at all. Above this we
	// pass the body through untouched and log a marker. Prevents large
	// uploads from being copied into memory just for logging.
	requestBodyHardCap = 1 << 20 // 1 MiB

	// Sentinels inserted into the logged body when content was clipped.
	bodyTruncatedNote = "[truncated]"
	bodySkippedNote   = "[skipped]"
	bodyRedactedValue = "[REDACTED]"
)

// skipLoggingPaths are never inspected by the middleware — the request
// flows straight through with no logging. Health and probe paths fire
// often enough that logging them adds noise without value.
var skipLoggingPaths = map[string]struct{}{
	"/health":  {},
	"/healthz": {},
	"/metrics": {},
	"/ready":   {},
}

// sensitiveHeaders are masked before being logged. Lookup keys are
// canonical (http.CanonicalHeaderKey).
var sensitiveHeaders = map[string]struct{}{
	"Authorization": {},
	"Cookie":        {},
	"Set-Cookie":    {},
	"X-Admin-Token": {},
	"X-Api-Key":     {},
	"X-Auth-Token":  {},
}

// sensitiveBodyKeys are JSON object keys whose values get redacted
// recursively. Comparison is case-insensitive.
var sensitiveBodyKeys = map[string]struct{}{
	"password":         {},
	"current_password": {},
	"new_password":     {},
	"token":            {},
	"access_token":     {},
	"refresh_token":    {},
	"id_token":         {},
	"jwt":              {},
	"secret":           {},
	"client_secret":    {},
	"api_key":          {},
	"apikey":           {},
	"otp":              {},
	"code":             {},
	"mfa":              {},
	"totp":             {},
	"ssn":              {},
	"credit_card":      {},
	"card_number":      {},
	"cvv":              {},
	"private_key":      {},
	"signature":        {},
}

func shouldSkipLogging(path string) bool {
	_, skip := skipLoggingPaths[path]
	return skip
}

// ──────────────────────────────────────────────────────────────────────────
// Response recorder.
// ──────────────────────────────────────────────────────────────────────────

// responseRecorder wraps http.ResponseWriter to capture status code and a
// bounded prefix of the response body without breaking downstream
// streaming behaviour. It implements Flusher / Hijacker via passthrough so
// SSE and websocket upgrades still work when this middleware is in place.
type responseRecorder struct {
	http.ResponseWriter
	statusCode    int
	body          *bytes.Buffer
	truncated     bool
	headerWritten bool
}

func newResponseRecorder(w http.ResponseWriter) *responseRecorder {
	return &responseRecorder{
		ResponseWriter: w,
		statusCode:     http.StatusOK, // default per net/http when WriteHeader isn't called
		body:           &bytes.Buffer{},
	}
}

func (r *responseRecorder) WriteHeader(code int) {
	if !r.headerWritten {
		r.statusCode = code
		r.headerWritten = true
	}
	r.ResponseWriter.WriteHeader(code)
}

func (r *responseRecorder) Write(b []byte) (int, error) {
	if !r.headerWritten {
		r.headerWritten = true // implicit per http spec
	}
	if r.body.Len() < responseCaptureCap {
		room := responseCaptureCap - r.body.Len()
		if len(b) > room {
			r.body.Write(b[:room])
			r.truncated = true
		} else {
			r.body.Write(b)
		}
	} else if len(b) > 0 {
		r.truncated = true
	}
	return r.ResponseWriter.Write(b)
}

func (r *responseRecorder) Flush() {
	if f, ok := r.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

func (r *responseRecorder) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if h, ok := r.ResponseWriter.(http.Hijacker); ok {
		return h.Hijack()
	}
	return nil, nil, errors.New("middleware: underlying ResponseWriter does not support Hijack")
}

// ──────────────────────────────────────────────────────────────────────────
// Request body snapshot — read once, restore unread bytes.
// ──────────────────────────────────────────────────────────────────────────

// snapshotRequestBody returns a copy of the request body (capped) and an
// optional note explaining why a copy is missing. It always restores
// r.Body so downstream handlers can read it. Cases that bypass capture:
//   - no body / Content-Length 0
//   - chunked-encoding (Content-Length unknown)
//   - body larger than requestBodyHardCap
//   - multipart/form-data or application/octet-stream (uploads)
func snapshotRequestBody(r *http.Request) ([]byte, string) {
	if r.Body == nil || r.Body == http.NoBody {
		return nil, ""
	}

	switch ct := strings.ToLower(parseMediaType(r.Header.Get("Content-Type"))); {
	case strings.HasPrefix(ct, "multipart/"),
		strings.HasPrefix(ct, "application/octet-stream"):
		return nil, bodySkippedNote + " (binary)"
	}

	if r.ContentLength < 0 {
		return nil, bodySkippedNote + " (chunked)"
	}
	if r.ContentLength == 0 {
		return nil, ""
	}
	if r.ContentLength > requestBodyHardCap {
		return nil, bodySkippedNote + " (oversize)"
	}

	raw, err := io.ReadAll(r.Body)
	_ = r.Body.Close()
	// Restore body even on read error so the handler isn't left with a
	// half-drained reader.
	r.Body = io.NopCloser(bytes.NewReader(raw))
	if err != nil {
		return raw, "[read-error: " + err.Error() + "]"
	}
	return raw, ""
}

// ──────────────────────────────────────────────────────────────────────────
// Masking.
// ──────────────────────────────────────────────────────────────────────────

func maskHeaders(h http.Header) http.Header {
	out := make(http.Header, len(h))
	for k, vs := range h {
		canon := http.CanonicalHeaderKey(k)
		if _, sensitive := sensitiveHeaders[canon]; !sensitive {
			out[canon] = vs
			continue
		}
		switch canon {
		case "Authorization":
			out[canon] = maskBearerValues(vs)
		default:
			out[canon] = []string{bodyRedactedValue}
		}
	}
	return out
}

// maskBearerValues preserves enough of a Bearer token to correlate log
// lines (last 6 chars — same length as the [token] prefix our logger
// already prints) without leaking enough to reuse the credential.
func maskBearerValues(vs []string) []string {
	const prefix = "Bearer "
	out := make([]string, len(vs))
	for i, v := range vs {
		if !strings.HasPrefix(v, prefix) {
			out[i] = bodyRedactedValue
			continue
		}
		tok := strings.TrimPrefix(v, prefix)
		if len(tok) <= 6 {
			out[i] = prefix + bodyRedactedValue
			continue
		}
		out[i] = prefix + strings.Repeat("*", 6) + tok[len(tok)-6:]
	}
	return out
}

// maskBodyForLog returns the body suitable for a log attribute. If an
// upstream step supplied a note (`[skipped]`, `[truncated]`, …) it takes
// precedence — we never log the raw bytes in that case.
func maskBodyForLog(body []byte, contentType, note string) string {
	if note != "" {
		return note
	}
	if len(body) == 0 {
		return ""
	}
	mt := parseMediaType(contentType)
	if isJSON(mt) {
		return maskJSONBody(body)
	}
	// Non-JSON: log up to the cap as a string. Useful for plain text /
	// form-urlencoded; binary content was already filtered out upstream.
	return truncate(string(body), logBodyCap)
}

// maskJSONBody parses, redacts, and pretty-prints a JSON body so it lands
// in the log as multi-line indented text matching what Postman / curl
// would display. Non-parseable input falls back to the truncated raw
// string so we never crash the logger over a malformed payload.
func maskJSONBody(body []byte) string {
	var v any
	if err := json.Unmarshal(body, &v); err != nil {
		return truncate(string(body), logBodyCap)
	}
	redactJSONValue(v)
	out, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return truncate(string(body), logBodyCap)
	}
	return truncate(string(out), logBodyCap)
}

// redactJSONValue mutates v in place, replacing values for sensitive keys
// with bodyRedactedValue. Walks both maps and slices recursively.
func redactJSONValue(v any) {
	switch x := v.(type) {
	case map[string]any:
		for k, val := range x {
			if _, ok := sensitiveBodyKeys[strings.ToLower(k)]; ok {
				x[k] = bodyRedactedValue
				continue
			}
			redactJSONValue(val)
		}
	case []any:
		for _, item := range x {
			redactJSONValue(item)
		}
	}
}

// ──────────────────────────────────────────────────────────────────────────
// Response body decoding.
// ──────────────────────────────────────────────────────────────────────────

// decodeResponseBody returns the bytes that should be passed to
// maskBodyForLog. When Content-Encoding is empty / "identity" the input
// is returned untouched. When it is "gzip" the bytes are gunzipped so the
// log captures the actual JSON the client receives in Postman / curl
// (after Accept-Encoding: gzip negotiates compression). Other encodings
// fall through to a placeholder note — adding brotli or deflate is a
// few-line extension here.
//
// Truncation handling: the recorder caps captured bytes at
// responseCaptureCap, so a large compressed response may yield an
// incomplete gzip stream. ReadAll then returns io.ErrUnexpectedEOF —
// instead of discarding what was decoded so far, we keep the partial
// result and tag a truncation note. Real decode errors (bad magic,
// corrupt header) fall back to the placeholder.
func decodeResponseBody(body []byte, encoding string) ([]byte, string) {
	enc := strings.ToLower(strings.TrimSpace(encoding))
	switch enc {
	case "", "identity":
		return body, ""
	case "gzip":
		return decodeGzip(body)
	default:
		return nil, "[" + enc + "-encoded]"
	}
}

func decodeGzip(body []byte) ([]byte, string) {
	if len(body) == 0 {
		return nil, ""
	}
	gr, err := gzip.NewReader(bytes.NewReader(body))
	if err != nil {
		return nil, "[gzip-decode-failed: " + err.Error() + "]"
	}
	defer gr.Close()

	decoded, err := io.ReadAll(gr)
	switch {
	case err == nil:
		return decoded, ""
	case errors.Is(err, io.ErrUnexpectedEOF):
		// Partial stream — capture cap clipped the recorder before the
		// gzip footer arrived. Whatever was decoded is still useful; tag
		// it so consumers know the tail is missing.
		return decoded, bodyTruncatedNote + " (gzip stream incomplete)"
	default:
		return decoded, "[gzip-decode-failed: " + err.Error() + "]"
	}
}

// ──────────────────────────────────────────────────────────────────────────
// Misc helpers.
// ──────────────────────────────────────────────────────────────────────────

func parseMediaType(ct string) string {
	if ct == "" {
		return ""
	}
	mt, _, err := mime.ParseMediaType(ct)
	if err != nil {
		return strings.ToLower(strings.TrimSpace(ct))
	}
	return mt
}

func isJSON(mediaType string) bool {
	return mediaType == "application/json" || strings.HasSuffix(mediaType, "+json")
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + bodyTruncatedNote
}

func clientAddr(r *http.Request) string {
	if v := r.Header.Get("X-Forwarded-For"); v != "" {
		// First IP wins (the original client; the rest are proxies).
		first, _, _ := strings.Cut(v, ",")
		return strings.TrimSpace(first)
	}
	if v := r.Header.Get("X-Real-Ip"); v != "" {
		return v
	}
	return r.RemoteAddr
}
