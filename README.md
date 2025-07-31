# go-metrics

A flexible, high-performance metrics collection library for Go applications that supports multiple backends, tagging, and various metric types.

## Features

- **Multiple metric types**: Counters, Gauges, Histograms, and Timers
- **Tagging/labeling system**: Add dimensions to metrics with key-value tags
- **Operational metrics**: High-level convenience API for error tracking and operation timing (`operational` package)
- **Multiple backend support**:
  - Prometheus integration
  - OpenTelemetry compatibility
  - Extensible for additional backends
- **Performance optimized**:
  - Lock-free implementations where possible using atomic operations
  - Thread-safe design for concurrent access
- **Context propagation**: Integration with Go context for tracing
- **Host/container metadata**: Automatic enrichment with service and environment information
- **Testing support**: Comprehensive mocks for unit testing

## Installation

```bash
go get github.com/MichaelAJay/go-metrics@v0.1.0
```

## Quick Start

### Basic Usage

```go
package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/MichaelAJay/go-metrics/metric"
	"github.com/MichaelAJay/go-metrics/metric/prometheus"
)

func main() {
	// Create a registry
	registry := metric.NewDefaultRegistry()

	// Create a counter
	requestCounter := registry.Counter(metric.Options{
		Name:        "http_requests_total",
		Description: "Total number of HTTP requests",
		Tags: metric.Tags{
			"method": "GET",
			"path":   "/api",
		},
	})

	// Increment the counter
	requestCounter.Inc()

	// Create a timer
	requestTimer := registry.Timer(metric.Options{
		Name:        "http_request_duration",
		Description: "HTTP request duration",
		Unit:        "milliseconds",
	})

	// Time a function execution
	requestTimer.Time(func() {
		// Your code here
		time.Sleep(10 * time.Millisecond)
	})

	// Set up Prometheus metrics endpoint
	reporter := prometheus.NewReporter()
	http.Handle("/metrics", reporter.Handler())

	// Report metrics regularly
	go func() {
		for {
			reporter.Report(registry)
			time.Sleep(10 * time.Second)
		}
	}()

	// Start HTTP server
	fmt.Println("Serving metrics at http://localhost:8080/metrics")
	http.ListenAndServe(":8080", nil)
}
```

## Metric Types

### Counter

Counters represent a monotonically increasing numerical value. Typically used for counting events or operations.

```go
counter := registry.Counter(metric.Options{
    Name: "requests_total",
    Tags: metric.Tags{"method": "GET"},
})

counter.Inc()        // Increment by 1
counter.Add(42.0)    // Increment by a specific value (value must be positive)
```

### Gauge

Gauges represent a single numerical value that can go up and down. Typically used for measuring current states.

```go
gauge := registry.Gauge(metric.Options{
    Name: "memory_usage_bytes",
})

gauge.Set(12345)     // Set to specific value
gauge.Inc()          // Increment by 1
gauge.Dec()          // Decrement by 1
gauge.Add(-10.0)     // Add value (can be negative)
```

### Histogram

Histograms track the distribution of a set of values. Useful for measuring things like response sizes.

```go
histogram := registry.Histogram(metric.Options{
    Name: "request_size_bytes",
})

histogram.Observe(42.0)  // Record a value
```

### Timer

Timers are specialized histograms for measuring durations. They provide convenience methods for timing.

```go
timer := registry.Timer(metric.Options{
    Name: "request_duration",
})

// Record a duration directly
timer.Record(150 * time.Millisecond)

// Record time since a starting point
start := time.Now()
// ... do work ...
timer.RecordSince(start)

// Time a function and get the duration
duration := timer.Time(func() {
    // ... function to time ...
})
```

## Tagging

All metrics support tags (or labels) to add dimensions to your metrics:

```go
counter := registry.Counter(metric.Options{
    Name: "http_requests_total",
    Tags: metric.Tags{
        "method": "GET",
        "path": "/api",
        "status": "200",
    },
})

// Adding tags to an existing metric creates a new metric with the combined tags
counterWithRegion := counter.With(metric.Tags{
    "region": "us-west",
})
```

## Backends

### Prometheus

```go
import (
    "net/http"
    "github.com/MichaelAJay/go-metrics/metric/prometheus"
)

// Create a Prometheus reporter
reporter := prometheus.NewReporter(
    prometheus.WithDefaultLabels(map[string]string{
        "service": "my-service",
        "version": "1.0.0",
    }),
)

// Expose HTTP endpoint for Prometheus to scrape
http.Handle("/metrics", reporter.Handler())

// Report metrics periodically
go func() {
    ticker := time.NewTicker(10 * time.Second)
    defer ticker.Stop()
    for range ticker.C {
        reporter.Report(registry)
    }
}()
```

### OpenTelemetry

```go
import (
    "github.com/MichaelAJay/go-metrics/metric/otel"
)

// Create an OpenTelemetry reporter
reporter, err := otel.NewReporter(
    "my-service",  // service name
    "1.0.0",       // version
    otel.WithAttributes(map[string]string{
        "environment": "production",
    }),
)
if err != nil {
    log.Fatal(err)
}
defer reporter.Close()

// Report metrics periodically
go func() {
    ticker := time.NewTicker(10 * time.Second)
    defer ticker.Stop()
    for range ticker.C {
        reporter.Report(registry)
    }
}()
```

## Global Registry and Functions

For convenience, a global registry is provided:

```go
// Using the global registry
counter := metrics.GetCounter(metric.Options{
    Name: "global_counter",
})
counter.Inc()

// Get the global registry directly
registry := metrics.GlobalRegistry()
```

## Thread Safety

All components in this library are designed to be thread-safe:

- Metric operations are implemented using atomic operations when possible
- The registry uses read/write locks for concurrent access
- Reporters are designed for concurrent reporting

## Guidelines for Metric Naming

When naming metrics, follow these best practices:

1. Use lowercase names with underscores separating words
2. Include units in the metric name (e.g., `request_duration_milliseconds`)
3. Use the plural form for counters (e.g., `requests_total`)
4. Use specific, descriptive names that indicate what is being measured

## Design Considerations

The go-metrics library follows these design principles:

1. **Low overhead**: Optimized for high throughput and minimal impact on application performance
2. **Separation of concerns**: 
   - Metric collection (counters, gauges, etc.)
   - Metric aggregation (registry)
   - Metric reporting (reporters)
3. **Extensibility**: Easy to add new reporters for different backends
4. **Consistency**: Consistent API design across different metric types

## Version Compatibility

This library requires Go 1.20 or later.

## License

MIT

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.