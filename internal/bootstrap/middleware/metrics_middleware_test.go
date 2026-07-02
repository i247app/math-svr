package middleware

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"math-ai.com/math-ai/internal/infrastructure/metrics"
)

// TestMetricsMiddleware_RecordsAndExposes drives a request through the
// middleware, then scrapes the registry the way Prometheus would, proving the
// full record → expose path works end to end (no app boot / DB required).
func TestMetricsMiddleware_RecordsAndExposes(t *testing.T) {
	m := metrics.New("math-svr-test", "test")

	// A handler that behaves like the project's writers: HTTP 200 body.
	h := MetricsMiddleware(m, nil)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"mstatus":200}`))
	}))

	// Fire a request against a route with a numeric id — should normalize.
	req := httptest.NewRequest(http.MethodGet, "/users/42", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("handler status = %d, want 200", rec.Code)
	}

	// Scrape the registry exactly like Prometheus does.
	scrapeRec := httptest.NewRecorder()
	promhttp.HandlerFor(m.Registry, promhttp.HandlerOpts{}).
		ServeHTTP(scrapeRec, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	body, _ := io.ReadAll(scrapeRec.Body)
	out := string(body)

	// The counter must exist, be labelled with the NORMALIZED route, and be 1.
	wantCounter := `http_requests_total{method="GET",route="/users/:id",status="200"} 1`
	if !strings.Contains(out, wantCounter) {
		t.Errorf("scrape missing counter line %q\n---\n%s", wantCounter, out)
	}

	// The duration histogram and in-flight gauge must be present too.
	for _, want := range []string{
		`http_request_duration_seconds_count{method="GET",route="/users/:id",status="200"} 1`,
		`http_requests_in_flight 0`, // incremented then decremented back to 0
		`math_svr_build_info{service="math-svr-test",version="test"} 1`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("scrape missing line %q", want)
		}
	}
}

// TestMetricsMiddleware_NilIsPassthrough ensures a disabled metrics config
// (nil *Metrics) does not break the chain.
func TestMetricsMiddleware_NilIsPassthrough(t *testing.T) {
	called := false
	h := MetricsMiddleware(nil, nil)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusTeapot)
	}))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/ping", nil))
	if !called {
		t.Fatal("next handler was not called through nil-metrics middleware")
	}
	if rec.Code != http.StatusTeapot {
		t.Fatalf("status = %d, want 418", rec.Code)
	}
}
