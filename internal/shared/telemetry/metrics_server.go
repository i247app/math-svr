package telemetry

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	promhttp "github.com/prometheus/client_golang/prometheus/promhttp"
)

// StartMetricsServer starts an HTTP server to expose Prometheus metrics
func StartMetricsServer(port string) *http.Server {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}

	go func() {
		log.Printf("Starting Prometheus metrics server on :%s/metrics", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("Metrics server error: %v", err)
		}
	}()

	return server
}

// StopMetricsServer gracefully stops the metrics HTTP server
func StopMetricsServer(server *http.Server) error {
	if server == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	log.Println("Stopping Prometheus metrics server...")
	if err := server.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to shutdown metrics server: %w", err)
	}

	log.Println("Prometheus metrics server stopped")
	return nil
}
