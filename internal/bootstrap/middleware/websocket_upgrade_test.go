package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/i247app/gex/sessionprovider"

	"math-ai.com/math-ai/internal/infrastructure/metrics"
	"math-ai.com/math-ai/internal/infrastructure/session"
)

// wsUpgradeRequest builds a minimal WebSocket handshake request.
func wsUpgradeRequest() *http.Request {
	r := httptest.NewRequest(http.MethodGet, "/ws/connect", nil)
	r.Header.Set("Upgrade", "websocket")
	r.Header.Set("Connection", "Upgrade")
	return r
}

// captureHandler records the ResponseWriter and context it is served with, so a
// test can assert whether an upgrade request was passed through unwrapped.
type captureHandler struct {
	got http.ResponseWriter
	ctx context.Context
}

func (h *captureHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.got = w
	h.ctx = r.Context()
}

func TestIsWebSocketUpgrade(t *testing.T) {
	cases := []struct {
		name       string
		upgrade    string
		connection string
		want       bool
	}{
		{"standard", "websocket", "Upgrade", true},
		{"case insensitive + token list", "WebSocket", "keep-alive, Upgrade", true},
		{"missing upgrade header", "", "Upgrade", false},
		{"missing connection header", "websocket", "", false},
		{"not websocket", "h2c", "Upgrade", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "/x", nil)
			if c.upgrade != "" {
				r.Header.Set("Upgrade", c.upgrade)
			}
			if c.connection != "" {
				r.Header.Set("Connection", c.connection)
			}
			if got := isWebSocketUpgrade(r); got != c.want {
				t.Fatalf("isWebSocketUpgrade = %v, want %v", got, c.want)
			}
		})
	}
}

func TestGzipMiddleware_WebSocketPassThrough(t *testing.T) {
	h := &captureHandler{}
	rec := httptest.NewRecorder()
	req := wsUpgradeRequest()
	req.Header.Set("Accept-Encoding", "gzip")

	GzipMiddleware(h).ServeHTTP(rec, req)

	if h.got != rec {
		t.Fatal("upgrade request was wrapped by gzip (must pass through)")
	}
	if enc := rec.Header().Get("Content-Encoding"); enc != "" {
		t.Fatalf("Content-Encoding = %q, want empty on upgrade", enc)
	}
}

func TestGzipMiddleware_NormalRequestStillWraps(t *testing.T) {
	h := &captureHandler{}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	req.Header.Set("Accept-Encoding", "gzip")

	GzipMiddleware(h).ServeHTTP(rec, req)

	if h.got == rec {
		t.Fatal("normal gzip request was not wrapped (regression)")
	}
}

func TestLogRequestMiddleware_WebSocketPassThrough(t *testing.T) {
	h := &captureHandler{}
	rec := httptest.NewRecorder()

	LogRequestMiddleware(h).ServeHTTP(rec, wsUpgradeRequest())

	if h.got != rec {
		t.Fatal("upgrade request was wrapped by log-request (must pass through)")
	}
}

func TestMetricsMiddleware_WebSocketPassThrough(t *testing.T) {
	m := metrics.New("test", "0")
	h := &captureHandler{}
	rec := httptest.NewRecorder()

	MetricsMiddleware(m, nil)(h).ServeHTTP(rec, wsUpgradeRequest())

	if h.got != rec {
		t.Fatal("upgrade request was wrapped by metrics (must pass through)")
	}
}

func TestTracingMiddleware_WebSocketPassThrough(t *testing.T) {
	h := &captureHandler{}
	rec := httptest.NewRecorder()

	TracingMiddleware(true, nil)(h).ServeHTTP(rec, wsUpgradeRequest())

	if h.got != rec {
		t.Fatal("upgrade request was wrapped by tracing (must pass through)")
	}
}

// fakeProvider is a stub SessionProvider for the session-middleware test.
type fakeProvider struct {
	result *sessionprovider.SessionResult
	err    error
}

func (f fakeProvider) GetSessionFromRequest(*http.Request) (*sessionprovider.SessionResult, error) {
	return f.result, f.err
}

func TestGexSessionMiddleware_WebSocketResolvesButDoesNotWrap(t *testing.T) {
	sess := session.NewSession()
	prov := fakeProvider{result: &sessionprovider.SessionResult{Session: sess, AuthToken: "tok"}}
	h := &captureHandler{}
	rec := httptest.NewRecorder()

	GexSessionMiddleware(prov, session.SessionContextKey)(h).ServeHTTP(rec, wsUpgradeRequest())

	if h.got != rec {
		t.Fatal("upgrade request was wrapped by session middleware (breaks hijack)")
	}
	// The session must still be bound to the context for downstream auth.
	if h.ctx == nil || h.ctx.Value(session.SessionContextKey) == nil {
		t.Fatal("session was not resolved onto the request context")
	}
}
