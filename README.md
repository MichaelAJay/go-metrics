# go-metrics

A flexible, high-performance metrics collection library for Go applications that supports multiple backends, tagging, and various metric types.

## Features

- **Multiple metric types**: Counters, Gauges, Histograms, and Timers
- **Tagging/labeling system**: Add dimensions to metrics with key-value tags
- **Operational metrics**: High-level convenience API with advanced builder patterns for error tracking, operation timing, and contextual metrics (`operational` package)
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

### Quick Start with Operational Metrics

For more complex applications requiring operational telemetry:

```go
package main

import (
	"time"
	"github.com/MichaelAJay/go-metrics/metric"
	"github.com/MichaelAJay/go-metrics/operational"
)

func main() {
	// Create registry and operational metrics
	registry := metric.NewDefaultRegistry()
	om := operational.New(registry)
	builder := operational.NewMetricsBuilder(om)

	// Record contextual operations
	context := map[string]string{
		"provider": "password",
		"user_type": "premium",
	}
	builder.RecordWithContext("authentication", "success", 150*time.Millisecond, context)

	// Record security events
	securityContext := map[string]string{
		"ip": "192.168.1.100",
		"source": "api",
	}
	builder.RecordSecurityEvent("login_attempt", "allowed", securityContext)

	// The metrics are automatically available for reporting
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

## Operational Metrics

The `operational` package provides high-level convenience methods for common operational telemetry patterns. It's designed to solve the problem of generalizing disparate metrics extensions across different services, particularly useful when you need standardized error tracking and operation timing.

### Features

- **Error Categorization**: Track errors with operation, type, and category labels
- **Operation Timing**: Automatic timing and status tracking for operations
- **Advanced Builder Pattern**: Contextual metric recording for complex scenarios
- **Thread-Safe**: Safe for concurrent use across multiple goroutines
- **Metric Caching**: Efficiently reuses metric instances for performance
- **Testing Support**: Full mock implementation included

### Basic Usage

```go
import (
    "github.com/MichaelAJay/go-metrics/metric"
    "github.com/MichaelAJay/go-metrics/operational"
)

// Create operational metrics
registry := metric.NewDefaultRegistry()
om := operational.New(registry)

// Record errors with categorization
om.RecordError("GenerateNonce", "crypto_error", "random_generation")
om.RecordError("ValidateRequest", "validation_error", "invalid_format")

// Record operations with timing
om.RecordOperation("GenerateNonce", "success", 50*time.Millisecond)
om.RecordOperation("ValidateRequest", "error", 100*time.Millisecond)
```

### MetricsBuilder - Advanced Pattern

The `MetricsBuilder` provides enhanced functionality for recording metrics with additional context, particularly useful for services that need to track domain-specific information.

#### Creating a MetricsBuilder

```go
// Create builder from operational metrics
builder := operational.NewMetricsBuilder(om)
```

#### Recording Operations with Context

Use `RecordWithContext` for operations that need additional contextual information:

```go
// Record authentication with provider context
context := map[string]string{
    "provider":  "password",
    "user_type": "premium",
}
builder.RecordWithContext("authentication", "success", 150*time.Millisecond, context)

// This creates metrics for both the primary operation and contextual dimensions:
// - authentication_total{operation="authentication", status="success"}
// - authentication_duration{operation="authentication"}
// - authentication_provider_total{operation="authentication_provider", status="password"}
// - authentication_user_type_total{operation="authentication_user_type", status="premium"}
```

#### Recording Security Events

Use `RecordSecurityEvent` for security-related telemetry:

```go
// Record security events with context
securityContext := map[string]string{
    "ip":         "192.168.1.100",
    "user_agent": "Mozilla/5.0",
    "source":     "api",
}

// Record a blocked brute force attempt
builder.RecordSecurityEvent("brute_force", "blocked", securityContext)

// Record successful login
builder.RecordSecurityEvent("login_attempt", "allowed", securityContext)
```

#### Recording Business Metrics

Use `RecordBusinessMetric` for business-related measurements:

```go
// Track user conversion with business context
businessContext := map[string]string{
    "source": "organic",
    "tier":   "premium",
    "region": "us-west",
}

// Record conversion value (e.g., revenue in cents)
builder.RecordBusinessMetric("user_conversion", "completed", 2999.0, businessContext)

// Record session duration
builder.RecordBusinessMetric("session_duration", "active", 45.5, businessContext)
```

### Common Integration Patterns

#### Service Integration

```go
type AuthService struct {
    metrics *operational.MetricsBuilder
}

func NewAuthService(registry metric.Registry) *AuthService {
    om := operational.New(registry)
    return &AuthService{
        metrics: operational.NewMetricsBuilder(om),
    }
}

func (s *AuthService) Authenticate(username, password string, clientIP string) error {
    start := time.Now()

    defer func() {
        context := map[string]string{
            "provider": "password",
            "client_ip": clientIP,
        }

        if err := recover(); err != nil {
            s.metrics.RecordWithContext("authentication", "panic", time.Since(start), context)
            panic(err)
        }
    }()

    // Authentication logic here...

    // Record successful authentication
    context := map[string]string{
        "provider": "password",
        "client_ip": clientIP,
    }
    s.metrics.RecordWithContext("authentication", "success", time.Since(start), context)

    return nil
}
```

#### API Middleware Integration

```go
func MetricsMiddleware(builder *operational.MetricsBuilder) http.HandlerFunc {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            start := time.Now()

            // Create context from request
            context := map[string]string{
                "method": r.Method,
                "path":   r.URL.Path,
                "user_agent": r.Header.Get("User-Agent"),
            }

            next.ServeHTTP(w, r)

            // Record API operation
            status := "success" // Determine from response
            builder.RecordWithContext("api_request", status, time.Since(start), context)
        })
    }
}
```

### Testing with Mocks

The operational package includes comprehensive mocking support:

```go
func TestAuthService(t *testing.T) {
    mockMetrics := operational.NewMockOperationalMetrics()
    service := NewAuthService(mockMetrics)

    // Test authentication
    err := service.Authenticate("user", "pass", "127.0.0.1")

    // Verify metrics were recorded
    successCount := mockMetrics.GetOperationCallCount("authentication", "success")
    if successCount != 1 {
        t.Errorf("Expected 1 successful auth, got %d", successCount)
    }

    // Check average duration
    avgDuration := mockMetrics.GetAverageDuration("authentication", "success")
    if avgDuration <= 0 {
        t.Error("Expected non-zero duration")
    }
}
```

### Performance Characteristics

The operational metrics package is optimized for high-throughput scenarios:

- **Metric Caching**: Reuses metric instances across calls
- **Object Pooling**: Uses sync.Pool for tag maps to reduce allocations
- **Lock-Free Operations**: Minimal locking for thread safety
- **Efficient Tagging**: Optimized tag management for contextual metrics

Benchmark results show excellent performance for concurrent operations:
- `RecordError`: ~130ns/op
- `RecordOperation`: ~170ns/op
- `RecordWithContext`: ~250ns/op

## Version Compatibility

This library requires Go 1.20 or later.

## License

MIT

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.