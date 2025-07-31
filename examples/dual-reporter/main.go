// Example demonstrating usage of both Prometheus and OpenTelemetry reporters
package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/MichaelAJay/go-metrics/metric"
	"github.com/MichaelAJay/go-metrics/metric/host"
	"github.com/MichaelAJay/go-metrics/metric/otel"
	"github.com/MichaelAJay/go-metrics/metric/prometheus"
)

func main() {
	// Create a new registry
	registry := metric.NewDefaultRegistry()

	// Add host information to metrics
	if err := host.InjectHostInfo(registry); err != nil {
		log.Printf("Warning: Failed to inject host info: %v", err)
	}

	// Create a Prometheus reporter
	promReporter := prometheus.NewReporter(
		prometheus.WithDefaultLabels(map[string]string{
			"service": "example-dual-reporter",
			"version": "1.0.0",
		}),
	)

	// Create an OpenTelemetry reporter
	otelReporter, err := otel.NewReporter(
		"example-dual-reporter",
		"1.0.0",
		otel.WithAttributes(map[string]string{
			"environment": "development",
		}),
	)
	if err != nil {
		log.Fatalf("Failed to create OpenTelemetry reporter: %v", err)
	}
	defer otelReporter.Close()

	// Register metrics handlers - Prometheus endpoint
	http.Handle("/metrics", promReporter.Handler())

	// Create some metrics
	requestCounter := registry.Counter(metric.Options{
		Name:        "http_requests_total",
		Description: "Total number of HTTP requests",
		Unit:        "requests",
		Tags: metric.Tags{
			"method": "GET",
			"path":   "/",
		},
	})

	requestLatency := registry.Histogram(metric.Options{
		Name:        "http_request_duration",
		Description: "HTTP request latency",
		Unit:        "milliseconds",
		Tags: metric.Tags{
			"method": "GET",
			"path":   "/",
		},
	})

	activeConnections := registry.Gauge(metric.Options{
		Name:        "active_connections",
		Description: "Number of active connections",
		Unit:        "connections",
	})

	// Simulate some metrics
	go simulateMetrics(requestCounter, requestLatency, activeConnections)

	// Report metrics to both systems periodically
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			// Report to Prometheus
			if err := promReporter.Report(registry); err != nil {
				log.Printf("Error reporting to Prometheus: %v", err)
			}

			// Report to OpenTelemetry
			if err := otelReporter.Report(registry); err != nil {
				log.Printf("Error reporting to OpenTelemetry: %v", err)
			}

			log.Println("Metrics reported to both Prometheus and OpenTelemetry")
		}
	}()

	// Start HTTP server
	fmt.Println("Starting server on :8080, metrics available at /metrics")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}

func simulateMetrics(requestCounter metric.Counter, requestLatency metric.Histogram, activeConnections metric.Gauge) {
	// Initialize random number generator
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	// Simulate metrics every second
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		// Simulate requests
		numRequests := r.Intn(5) + 1
		for i := 0; i < numRequests; i++ {
			requestCounter.Inc()

			// Simulate latency between 10-200ms
			latency := 10.0 + r.Float64()*190.0
			requestLatency.Observe(latency)
		}

		// Simulate connections (oscillating between 1-100)
		connections := r.Intn(100) + 1
		activeConnections.Set(float64(connections))

		fmt.Printf("Processed %d requests, current connections: %d\n", numRequests, connections)
	}
}
