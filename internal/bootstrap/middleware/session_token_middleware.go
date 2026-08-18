package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"strings"
)

// sessionTokenMetadataField is the multipart form field carrying the JSON
// `metadata` blob for file-upload requests. For JSON requests the same object
// is a top-level `metadata` key.
const sessionTokenMetadataField = "metadata"

// sessionTokenMaxBuffer bounds how much of a request body the middleware will
// buffer to find the token. Detail requests are tiny JSON; avatar uploads
// (multipart) can be several MB. Above this cap the body is left unread — an
// auth-gated route then fails downstream as unauthorized.
const sessionTokenMaxBuffer = 16 << 20 // 16 MiB

// SessionTokenMiddleware makes the request body's `metadata.authorization`
// the ONLY accepted source of the session token, then hands it to
// GexSessionMiddleware via the Authorization header (the header the gex
// session provider reads).
//
// It must run just OUTSIDE GexSessionMiddleware. For every non-WebSocket
// request it:
//  1. deletes any client-supplied Authorization header, so a header-carried
//     token can never authenticate, and
//  2. re-sets the Authorization header from metadata.authorization when the
//     body carries one — restoring r.Body so LogRequestMiddleware,
//     MetadataMiddleware, and the handler can read it again.
//
// WebSocket upgrades have no body and authenticate via the real Bearer header
// on the handshake (see rules/socket.md); they pass through untouched.
func SessionTokenMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if isWebSocketUpgrade(r) {
			next.ServeHTTP(w, r)
			return
		}

		// Header is never trusted for auth — drop it up front, before the body
		// is inspected. A request with no metadata.authorization therefore
		// reaches GexSessionMiddleware with NO Authorization header at all, and
		// an auth-gated route rejects it. Do not add a header fallback here:
		// that reopens header-carried auth for every route.
		r.Header.Del("Authorization")

		token := extractBodyAuthorization(r)
		if token != "" {
			if !strings.HasPrefix(token, "Bearer ") {
				token = "Bearer " + token
			}
			r.Header.Set("Authorization", token)
		}
		next.ServeHTTP(w, r)
	})
}

// extractBodyAuthorization returns the `metadata.authorization` value from the
// request body (JSON body or multipart `metadata` form field), always
// restoring r.Body. Returns "" when absent or unreadable.
func extractBodyAuthorization(r *http.Request) string {
	if r.Body == nil || r.Body == http.NoBody || r.ContentLength == 0 {
		return ""
	}
	if r.ContentLength > sessionTokenMaxBuffer {
		return ""
	}

	mediaType, params, _ := mime.ParseMediaType(r.Header.Get("Content-Type"))
	mediaType = strings.ToLower(mediaType)

	switch {
	case mediaType == "application/json" || strings.HasSuffix(mediaType, "+json"):
		return authFromJSONBody(r)
	case strings.HasPrefix(mediaType, "multipart/"):
		return authFromMultipartBody(r, params["boundary"])
	}
	return ""
}

// metadataAuthProbe is the minimal shape needed to read the token out of the
// body without decoding the whole payload.
type metadataAuthProbe struct {
	Metadata struct {
		Authorization string `json:"authorization"`
	} `json:"metadata"`
}

// authFromJSONBody reads the JSON body, pulls metadata.authorization, and
// restores r.Body so downstream consumers read it unchanged.
func authFromJSONBody(r *http.Request) string {
	raw, err := io.ReadAll(io.LimitReader(r.Body, sessionTokenMaxBuffer))
	_ = r.Body.Close()
	r.Body = io.NopCloser(bytes.NewReader(raw))
	if err != nil {
		return ""
	}

	var probe metadataAuthProbe
	if err := json.Unmarshal(raw, &probe); err != nil {
		return ""
	}
	return strings.TrimSpace(probe.Metadata.Authorization)
}

// authFromMultipartBody buffers the multipart body, reads the JSON `metadata`
// form field's authorization without copying file parts into memory, and
// restores r.Body so both the logger and the handler can re-parse the upload.
func authFromMultipartBody(r *http.Request, boundary string) string {
	if boundary == "" {
		return ""
	}
	raw, err := io.ReadAll(io.LimitReader(r.Body, sessionTokenMaxBuffer))
	_ = r.Body.Close()
	r.Body = io.NopCloser(bytes.NewReader(raw))
	if err != nil {
		return ""
	}

	mr := multipart.NewReader(bytes.NewReader(raw), boundary)
	for {
		part, err := mr.NextPart()
		if err != nil {
			return ""
		}
		if part.FileName() == "" && part.FormName() == sessionTokenMetadataField {
			val, _ := io.ReadAll(io.LimitReader(part, 64<<10))
			_ = part.Close()
			var probe struct {
				Authorization string `json:"authorization"`
			}
			if err := json.Unmarshal(val, &probe); err != nil {
				return ""
			}
			return strings.TrimSpace(probe.Authorization)
		}
		_ = part.Close()
	}
}
