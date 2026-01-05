package telemetry

import (
	"os"
	"strconv"
	"time"
)

// Config holds OpenTelemetry configuration
type Config struct {
	// Feature flags
	EnableTracing bool
	EnableMetrics bool

	// Service identification
	ServiceName    string
	ServiceVersion string
	Environment    string // dev, staging, prod

	// OTLP Exporter
	OTLPEndpoint string
	OTLPInsecure bool

	// Trace sampling
	TraceSampleRate float64 // 0.0 to 1.0

	// Metric export
	MetricExportInterval time.Duration

	// Prometheus
	PrometheusPort string
}

// NewConfig creates a new OTel configuration with default values
func NewConfig() *Config {
	return &Config{
		EnableTracing:        true,
		EnableMetrics:        true,
		ServiceName:          "math-srv",
		ServiceVersion:       "1.0.0",
		Environment:          "development",
		OTLPEndpoint:         "localhost:4318",
		OTLPInsecure:         true,
		TraceSampleRate:      1.0,
		MetricExportInterval: time.Second * 30,
		PrometheusPort:       "9090",
	}
}

// NewConfigFromEnv creates OTel configuration from environment variables
func NewConfigFromEnv() *Config {
	return &Config{
		EnableTracing:        getBoolEnv("OTEL_ENABLE_TRACING", true),
		EnableMetrics:        getBoolEnv("OTEL_ENABLE_METRICS", true),
		ServiceName:          getEnv("OTEL_SERVICE_NAME", "math-srv"),
		ServiceVersion:       getEnv("OTEL_SERVICE_VERSION", "1.0.0"),
		Environment:          getEnv("OTEL_ENVIRONMENT", "development"),
		OTLPEndpoint:         getEnv("OTEL_EXPORTER_OTLP_ENDPOINT", "localhost:4318"),
		OTLPInsecure:         getBoolEnv("OTEL_EXPORTER_OTLP_INSECURE", true),
		TraceSampleRate:      getFloatEnv("OTEL_TRACE_SAMPLE_RATE", 1.0),
		MetricExportInterval: time.Second * 30,
		PrometheusPort:       getEnv("OTEL_PROMETHEUS_PORT", "9090"),
	}
}

// Helper functions for environment variable parsing
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getBoolEnv(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if b, err := strconv.ParseBool(value); err == nil {
			return b
		}
	}
	return defaultValue
}

func getFloatEnv(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if f, err := strconv.ParseFloat(value, 64); err == nil {
			return f
		}
	}
	return defaultValue
}
