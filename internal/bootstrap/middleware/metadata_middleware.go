package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"strings"

	"math-ai.com/math-ai/internal/infrastructure/logger"
	"math-ai.com/math-ai/internal/infrastructure/metadata"
)

const (
	// maxMetadataBody bounds how much of a request body is buffered to find
	// its metadata. JSON requests are small; avatar uploads (multipart) can
	// be several MB. A larger body is passed through unread: it gets empty
	// metadata, so an auth-gated route rejects it for lack of a token.
	maxMetadataBody = 16 << 20 // 16 MiB

	// metadataFieldName is the JSON key, or the multipart form field, that
	// carries the metadata object.
	metadataFieldName = "metadata"

	// maxMetadataField bounds the multipart metadata field.
	maxMetadataField = 64 << 10 // 64 KiB
)

// MetadataMiddleware parses the request's `metadata` object — the top-level
// JSON key, or the multipart form field of that name — into the context,
// where metadata.FromContext and the metadata.Get* helpers read it.
//
// It runs before GexSessionMiddleware, which takes the session token from
// metadata.authorization. The body is read once and restored byte for byte:
// downstream middleware and handlers see exactly what the client sent.
func MetadataMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			md := readRequestMetadata(r)
			next.ServeHTTP(w, r.WithContext(metadata.WithMetadata(r.Context(), md)))
		})
	}
}

// readRequestMetadata returns the request's metadata, or empty metadata when
// the body has none or cannot be read.
func readRequestMetadata(r *http.Request) *metadata.RequestMetadata {
	md := metadata.NewRequestMetadata()

	raw, ok := bufferBody(r)
	if !ok {
		return md
	}

	var field []byte
	mediaType, params, _ := mime.ParseMediaType(r.Header.Get("Content-Type"))
	switch {
	case mediaType == "application/json" || strings.HasSuffix(mediaType, "+json"):
		// Decode only the metadata key. The body is never re-encoded, so
		// large int64 ids and key order survive untouched.
		var body struct {
			Metadata json.RawMessage `json:"metadata"`
		}
		if json.Unmarshal(raw, &body) == nil {
			field = body.Metadata
		}
	case mediaType == "multipart/form-data":
		field = multipartTextField(raw, params["boundary"], metadataFieldName)
	}
	if len(field) == 0 || string(field) == "null" {
		return md
	}

	if err := json.Unmarshal(field, md); err != nil {
		logger.From(r.Context()).Warnf("metadata.parse_failed err=%v", err)
		return metadata.NewRequestMetadata()
	}
	return md
}

// bufferBody reads the whole body (up to maxMetadataBody) and puts an
// identical copy back on r. It reports false — with the body left readable —
// when there is no body, it is too large, or it cannot be read.
func bufferBody(r *http.Request) ([]byte, bool) {
	if r.Body == nil || r.Body == http.NoBody || r.ContentLength > maxMetadataBody {
		return nil, false
	}

	raw, err := io.ReadAll(io.LimitReader(r.Body, maxMetadataBody+1))
	if int64(len(raw)) > maxMetadataBody {
		// Unknown length and over the cap: hand back what was read followed
		// by the unread rest, so downstream still sees the full body.
		r.Body = struct {
			io.Reader
			io.Closer
		}{io.MultiReader(bytes.NewReader(raw), r.Body), r.Body}
		return nil, false
	}
	_ = r.Body.Close()
	r.Body = io.NopCloser(bytes.NewReader(raw))
	return raw, err == nil
}

// multipartTextField returns the value of the non-file form field name,
// without copying file parts into memory.
func multipartTextField(raw []byte, boundary, name string) []byte {
	if boundary == "" {
		return nil
	}
	mr := multipart.NewReader(bytes.NewReader(raw), boundary)
	for {
		part, err := mr.NextPart()
		if err != nil {
			return nil
		}
		if part.FileName() == "" && part.FormName() == name {
			val, _ := io.ReadAll(io.LimitReader(part, maxMetadataField))
			_ = part.Close()
			return val
		}
		_ = part.Close()
	}
}
