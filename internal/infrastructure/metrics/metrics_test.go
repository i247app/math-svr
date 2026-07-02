package metrics

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// TestObserveHTTP_EmitsExemplar proves the latency histogram carries the
// trace_id as an OpenMetrics exemplar when a trace id is supplied — this is
// what lets Grafana jump from a latency point to the trace in Tempo.
func TestObserveHTTP_EmitsExemplar(t *testing.T) {
	m := New("t", "t")
	m.ObserveHTTP("GET", "/users/:id", 200, 0.01, "abc123deadbeef456")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	// Exemplars are only rendered in the OpenMetrics exposition format.
	req.Header.Set("Accept", "application/openmetrics-text; version=1.0.0")
	promhttp.HandlerFor(m.Registry, promhttp.HandlerOpts{EnableOpenMetrics: true}).
		ServeHTTP(rec, req)

	body := rec.Body.String()
	if !strings.Contains(body, `trace_id="abc123deadbeef456"`) {
		t.Errorf("expected exemplar with trace_id in exposition, got:\n%s", body)
	}
}

func TestNormalizeRoute(t *testing.T) {
	tests := []struct {
		name string
		path string
		want string
	}{
		{"root", "/", "/"},
		{"empty", "", ""},
		{"static two-seg", "/users/create", "/users/create"},
		{"numeric id", "/users/42", "/users/:id"},
		{"numeric id mid-path", "/classrooms/128/members", "/classrooms/:id/members"},
		{"short word not id", "/auth/login", "/auth/login"},
		{"ping", "/ping", "/ping"},
		{"long token with digit", "/classrooms/9f3a1b2c4d5e6f7089ab12cd", "/classrooms/:id"},
		{"long word no digit stays", "/programs/khoahocphothongmoi", "/programs/khoahocphothongmoi"},
		{"trailing slash id", "/users/7/", "/users/:id/"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NormalizeRoute(tt.path); got != tt.want {
				t.Errorf("NormalizeRoute(%q) = %q, want %q", tt.path, got, tt.want)
			}
		})
	}
}
