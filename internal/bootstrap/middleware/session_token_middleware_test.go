package middleware

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// serveSessionToken runs SessionTokenMiddleware and records the Authorization
// header + body the downstream handler sees.
func serveSessionToken(r *http.Request) (authHeader, body string) {
	next := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		authHeader = req.Header.Get("Authorization")
		b, _ := io.ReadAll(req.Body)
		body = string(b)
	})
	SessionTokenMiddleware(next).ServeHTTP(httptest.NewRecorder(), r)
	return authHeader, body
}

func TestSessionToken_JSONBody(t *testing.T) {
	raw := `{"metadata":{"authorization":"Bearer tok-abc","device_uuid":"d1"},"user_id":42}`
	r := httptest.NewRequest(http.MethodPost, "/users/detail", strings.NewReader(raw))
	r.Header.Set("Content-Type", "application/json")

	auth, body := serveSessionToken(r)

	if want := "Bearer tok-abc"; auth != want {
		t.Fatalf("Authorization = %q, want %q", auth, want)
	}
	if body != raw {
		t.Fatalf("downstream body = %q, want it intact", body)
	}
}

func TestSessionToken_Multipart(t *testing.T) {
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	_ = mw.WriteField("metadata", `{"authorization":"Bearer tok-mp"}`)
	fw, _ := mw.CreateFormFile("avatar", "a.png")
	_, _ = fw.Write([]byte("img"))
	_ = mw.Close()

	r := httptest.NewRequest(http.MethodPost, "/users/upload-avatar", bytes.NewReader(buf.Bytes()))
	r.Header.Set("Content-Type", mw.FormDataContentType())

	auth, body := serveSessionToken(r)

	if want := "Bearer tok-mp"; auth != want {
		t.Fatalf("Authorization = %q, want %q", auth, want)
	}
	// Body must survive so the handler can still parse the upload.
	r2 := httptest.NewRequest(http.MethodPost, "/x", strings.NewReader(body))
	r2.Header.Set("Content-Type", mw.FormDataContentType())
	if err := r2.ParseMultipartForm(1 << 20); err != nil {
		t.Fatalf("downstream ParseMultipartForm: %v", err)
	}
}

// TestSessionToken_HeaderIsDropped is the core "body-only" guarantee: a token
// supplied via the Authorization header must NOT survive when the body has no
// metadata.authorization.
func TestSessionToken_HeaderIsDropped(t *testing.T) {
	raw := `{"metadata":{"device_uuid":"d1"},"user_id":42}`
	r := httptest.NewRequest(http.MethodPost, "/users/detail", strings.NewReader(raw))
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("Authorization", "Bearer header-token")

	auth, _ := serveSessionToken(r)

	if auth != "" {
		t.Fatalf("Authorization = %q, want empty (header must be dropped)", auth)
	}
}

// TestSessionToken_BodyOverridesHeader: even if a header token is present, the
// body's metadata.authorization is the one that wins.
func TestSessionToken_BodyOverridesHeader(t *testing.T) {
	raw := `{"metadata":{"authorization":"Bearer body-token"}}`
	r := httptest.NewRequest(http.MethodPost, "/users/detail", strings.NewReader(raw))
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("Authorization", "Bearer header-token")

	auth, _ := serveSessionToken(r)

	if want := "Bearer body-token"; auth != want {
		t.Fatalf("Authorization = %q, want %q", auth, want)
	}
}

func TestSessionToken_BareTokenGetsBearerPrefix(t *testing.T) {
	raw := `{"metadata":{"authorization":"raw-jwt"}}`
	r := httptest.NewRequest(http.MethodPost, "/users/detail", strings.NewReader(raw))
	r.Header.Set("Content-Type", "application/json")

	auth, _ := serveSessionToken(r)

	if want := "Bearer raw-jwt"; auth != want {
		t.Fatalf("Authorization = %q, want %q", auth, want)
	}
}

// TestSessionToken_WebSocketPassthrough: WS handshakes keep their real Bearer
// header (they have no body).
func TestSessionToken_WebSocketPassthrough(t *testing.T) {
	r := wsUpgradeRequest()
	r.Header.Set("Authorization", "Bearer ws-token")

	auth, _ := serveSessionToken(r)

	if want := "Bearer ws-token"; auth != want {
		t.Fatalf("Authorization = %q, want %q (WS must pass through)", auth, want)
	}
}
