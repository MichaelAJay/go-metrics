// This example demonstrates basic usage of the metrics package
package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/MichaelAJay/go-metrics/metric"
	"github.com/MichaelAJay/go-metrics/metric/host"
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
	reporter := prometheus.NewReporter(
		prometheus.WithDefaultLabels(map[string]string{
			"service": "example-service",
			"version": "1.0.0",
		}),
	)

	// Register metrics handlers
	http.Handle("/metrics", reporter.Handler())

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

	// Start reporting metrics
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			if err := reporter.Report(registry); err != nil {
				log.Printf("Error reporting metrics: %v", err)
			}
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
