package tracing

import (
	"context"
	"os"
	"testing"
	"time"

	"go.opentelemetry.io/otel"

	"math-ai.com/math-ai/internal/infrastructure/config"
)

// TestTempoSmoke ships a real parent+child trace to Tempo. It is skipped
// unless TEMPO_SMOKE=1, so `go test ./...` stays hermetic. Run it against the
// local stack to verify the OTLP export path end to end:
//
//	TEMPO_SMOKE=1 go test -run TestTempoSmoke -v ./internal/infrastructure/tracing/
//
// The logged TEMPO_SMOKE_TRACE_ID can then be looked up in Tempo/Grafana.
func TestTempoSmoke(t *testing.T) {
	if os.Getenv("TEMPO_SMOKE") == "" {
		t.Skip("set TEMPO_SMOKE=1 to run (requires Tempo at OBS_OTLP_ENDPOINT)")
	}

	endpoint := os.Getenv("OBS_OTLP_ENDPOINT")
	if endpoint == "" {
		endpoint = "localhost:4318"
	}

	cfg := config.ObservabilityConfig{
		ServiceName:      "math-svr-smoke",
		ServiceVersion:   "test",
		Environment:      "local",
		TracingEnabled:   true,
		OTLPEndpoint:     endpoint,
		OTLPInsecure:     true,
		TraceSampleRatio: 1.0,
	}

	shutdown, err := Init(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Init: %v", err)
	}

	ctx, parent := Tracer().Start(context.Background(), "smoke.parent")
	traceID := parent.SpanContext().TraceID().String()

	_, child := otel.Tracer("math-svr/db").Start(ctx, "db.mysql SELECT")
	time.Sleep(5 * time.Millisecond)
	child.End()
	parent.End()

	// Flush — this is where spans are actually pushed to Tempo.
	if err := shutdown(context.Background()); err != nil {
		t.Fatalf("shutdown/flush: %v", err)
	}

	t.Logf("TEMPO_SMOKE_TRACE_ID=%s", traceID)
}
