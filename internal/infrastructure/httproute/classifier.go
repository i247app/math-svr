// Package httproute provides a Classifier that maps an incoming request to
// the registered route TEMPLATE (e.g. "/users/{id}") rather than its raw path
// (e.g. "/users/42"). Observability middleware use this to label metrics and
// name spans with a bounded set of routes — arbitrary/unknown paths (404s,
// path-spraying) collapse to "unmatched", so the `route` label / span-name
// cardinality can never explode.
//
// It works by mirroring every gex route spec into a private *http.ServeMux and
// asking it (via ServeMux.Handler, which is read-only and does NOT serve the
// request) which pattern a request matches.
package httproute

import (
	"net/http"
	"strings"
	"sync"
)

// Unmatched is the label used for requests that match no registered route.
const Unmatched = "unmatched"

// Classifier resolves requests to their registered route template. Add is
// called at boot (single-threaded) while routes are registered; Route is
// called per request afterwards. The underlying ServeMux is safe for
// concurrent Handler lookups once population is done.
type Classifier struct {
	mux  *http.ServeMux
	mu   sync.Mutex
	seen map[string]struct{} // dedupe specs so re-registration can't panic
}

func NewClassifier() *Classifier {
	return &Classifier{
		mux:  http.NewServeMux(),
		seen: make(map[string]struct{}),
	}
}

// Add registers a gex route spec ("METHOD  /path/{param}", possibly with
// irregular spacing) into the classifier. Duplicate specs are ignored so a
// second registration never panics the process. A no-op handler is stored —
// the mux is only ever consulted via Handler(), never actually served.
func (c *Classifier) Add(spec string) {
	if c == nil {
		return
	}
	pattern := normalizeSpec(spec)
	if pattern == "" {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	if _, ok := c.seen[pattern]; ok {
		return
	}
	c.seen[pattern] = struct{}{}
	c.mux.HandleFunc(pattern, func(http.ResponseWriter, *http.Request) {})
}

// Route returns the matched route template (method stripped, e.g.
// "/users/{id}") for r, or Unmatched when nothing matches. It never serves r.
func (c *Classifier) Route(r *http.Request) string {
	if c == nil {
		return Unmatched
	}
	_, pattern := c.mux.Handler(r)
	if pattern == "" {
		return Unmatched
	}
	return stripMethod(pattern)
}

// normalizeSpec collapses irregular whitespace ("GET  /x" → "GET /x") and
// keeps only the "METHOD /path" (or "/path") shape the ServeMux understands.
func normalizeSpec(spec string) string {
	fields := strings.Fields(spec)
	switch len(fields) {
	case 0:
		return ""
	case 1:
		return fields[0] // path only
	default:
		return fields[0] + " " + fields[1] // METHOD path
	}
}

// stripMethod drops a leading "METHOD " from a pattern, leaving just the path
// template (the metrics middleware carries method as its own label, and span
// names prepend the method themselves).
func stripMethod(pattern string) string {
	if _, path, found := strings.Cut(pattern, " "); found {
		return path
	}
	return pattern
}
