package httproute

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClassifier_Route(t *testing.T) {
	c := NewClassifier()
	// Mixed spacing on purpose — mirrors real gex specs ("GET  /...").
	c.Add("GET  /users/{id}")
	c.Add("POST /users/create")
	c.Add("GET /ping")
	c.Add("POST /users/create") // duplicate must not panic

	tests := []struct {
		method, path, want string
	}{
		{http.MethodGet, "/users/42", "/users/{id}"},
		{http.MethodGet, "/users/9007199254740991", "/users/{id}"},
		{http.MethodPost, "/users/create", "/users/create"},
		{http.MethodGet, "/ping", "/ping"},
		{http.MethodGet, "/totally/unknown/path", Unmatched},
		{http.MethodDelete, "/users/42", Unmatched}, // method not registered
	}
	for _, tt := range tests {
		req := httptest.NewRequest(tt.method, tt.path, nil)
		if got := c.Route(req); got != tt.want {
			t.Errorf("Route(%s %s) = %q, want %q", tt.method, tt.path, got, tt.want)
		}
	}
}

func TestClassifier_NilSafe(t *testing.T) {
	var c *Classifier
	c.Add("GET /x") // must not panic
	if got := c.Route(httptest.NewRequest(http.MethodGet, "/x", nil)); got != Unmatched {
		t.Errorf("nil classifier Route = %q, want %q", got, Unmatched)
	}
}
