# Operational Metrics Package

The operational metrics package provides high-level convenience methods for common operational telemetry patterns. It builds on top of the core `github.com/MichaelAJay/go-metrics` package to offer simple APIs for error tracking and operation timing.

## Features

- **Error Categorization**: Track errors with operation, type, and category labels
- **Operation Timing**: Automatic timing and status tracking for operations  
- **Thread-Safe**: Safe for concurrent use across multiple goroutines
- **Metric Caching**: Efficiently reuses metric instances for performance
- **Mock Support**: Full mock implementation for testing
- **Tag-Based**: Uses metric tags for rich categorization and filtering

## Installation

```go
import "github.com/MichaelAJay/go-metrics/operational"
```

## Basic Usage

### Creating Operational Metrics

```go
import (
    "github.com/MichaelAJay/go-metrics"
    "github.com/MichaelAJay/go-metrics/operational"
)

// Create a registry
registry := metrics.NewRegistry()

// Create operational metrics
om := operational.New(registry)
```

### Recording Errors

```go
// Record different types of errors with categorization
om.RecordError("GenerateNonce", "crypto_error", "random_generation")
om.RecordError("ValidateRequest", "validation_error", "invalid_format") 
om.RecordError("DatabaseQuery", "network_error", "timeout")
```

This creates counter metrics like:
- `GenerateNonce_errors_total{operation="GenerateNonce", error_type="crypto_error", error_category="random_generation"}`
- `ValidateRequest_errors_total{operation="ValidateRequest", error_type="validation_error", error_category="invalid_format"}`

### Recording Operations

```go
// Record operation timing and status
om.RecordOperation("GenerateNonce", "success", 50*time.Millisecond)
om.RecordOperation("ValidateRequest", "error", 100*time.Millisecond)
```

This creates:
- **Timers**: `GenerateNonce_duration{operation="GenerateNonce"}` 
- **Counters**: `GenerateNonce_total{operation="GenerateNonce", status="success"}`

## Common Patterns

### Pattern 1: Operation Timing with Defer

```go
func (s *Service) ProcessData(data []byte) error {
    start := time.Now()
    defer func() {
        if err := recover(); err != nil {
            s.metrics.RecordError("ProcessData", "panic", "unexpected_panic")
            s.metrics.RecordOperation("ProcessData", "error", time.Since(start))
            panic(err) // re-panic
        }
    }()
    
    // ... business logic ...
    
    s.metrics.RecordOperation("ProcessData", "success", time.Since(start))
    return nil
}
```

### Pattern 2: Error Categorization

```go
func (s *Service) MakeAPICall() error {
    start := time.Now()
    
    resp, err := http.Get("https://api.example.com/data")
    if err != nil {
        // Categorize network errors
        if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
            s.metrics.RecordError("APICall", "network_error", "timeout")
        } else {
            s.metrics.RecordError("APICall", "network_error", "connection_failed")
        }
        s.metrics.RecordOperation("APICall", "error", time.Since(start))
        return err
    }
    defer resp.Body.Close()
    
    if resp.StatusCode >= 400 {
        // Categorize HTTP errors
        if resp.StatusCode >= 500 {
            s.metrics.RecordError("APICall", "server_error", fmt.Sprintf("status_%d", resp.StatusCode))
        } else {
            s.metrics.RecordError("APICall", "client_error", fmt.Sprintf("status_%d", resp.StatusCode))
        }
        s.metrics.RecordOperation("APICall", "error", time.Since(start))
        return fmt.Errorf("API call failed with status %d", resp.StatusCode)
    }
    
    s.metrics.RecordOperation("APICall", "success", time.Since(start))
    return nil
}
```

## Testing with Mocks

The package includes a full mock implementation for testing:

```go
import "github.com/MichaelAJay/go-metrics/operational"

func TestMyService(t *testing.T) {
    // Create mock
    mockMetrics := operational.NewMockOperationalMetrics()
    
    // Inject into service
    service := NewMyService(mockMetrics)
    
    // Exercise service
    err := service.ProcessData([]byte("test"))
    
    // Verify metrics were recorded
    errorCount := mockMetrics.GetErrorCallCount("ProcessData", "validation_error", "invalid_input")
    if errorCount != 1 {
        t.Errorf("Expected 1 validation error, got %d", errorCount)
    }
    
    successCount := mockMetrics.GetOperationCallCount("ProcessData", "success")
    if successCount != 0 {
        t.Errorf("Expected 0 successful operations, got %d", successCount)
    }
    
    // Check timing
    avgDuration := mockMetrics.GetAverageDuration("ProcessData", "error")
    if avgDuration <= 0 {
        t.Error("Expected non-zero average duration")
    }
}
```

### Mock API

The mock provides these methods for testing:

- `GetErrorCallCount(operation, errorType, errorCategory string) int`
- `GetOperationCallCount(operation, status string) int`
- `GetTotalErrorCalls() int`
- `GetTotalOperationCalls() int`
- `GetAverageDuration(operation, status string) time.Duration`
- `GetLastErrorCall() *ErrorCall`
- `GetLastOperationCall() *OperationCall`
- `Reset()` - Clear all recorded calls

## Integration with Reporters

The operational metrics work seamlessly with any metrics reporter:

```go
import (
    "github.com/MichaelAJay/go-metrics"
    "github.com/MichaelAJay/go-metrics/operational"
    "github.com/MichaelAJay/go-metrics/metric/prometheus"
)

// Create registry and operational metrics
registry := metrics.NewRegistry()
om := operational.New(registry)

// Record some metrics
om.RecordError("MyOperation", "validation_error", "missing_field")
om.RecordOperation("MyOperation", "success", 100*time.Millisecond)

// Export to Prometheus
reporter := prometheus.NewReporter()
reporter.Report(registry)

// Metrics are now available at /metrics endpoint
```

## Performance

The operational metrics package is designed for high performance:

- **Metric Caching**: Metrics are cached after first creation
- **Lock-Free Operations**: Minimal locking for concurrent access  
- **Efficient Tagging**: Reuses metric instances with different tag combinations

Benchmark results (Apple M1 Pro):

```
BenchmarkRecordError-10              	 9129495	       130.4 ns/op
BenchmarkRecordOperation-10          	 7404250	       170.6 ns/op
BenchmarkConcurrentRecordError-10    	 8115967	       145.3 ns/op
```

## Metric Schema

### Error Metrics

Errors are recorded as counters with these tags:

```
{operation}_errors_total{
    operation="{operation}",
    error_type="{errorType}",
    error_category="{errorCategory}"
}
```

### Operation Metrics  

Operations create two metrics:

**Duration (Timer):**
```
{operation}_duration{
    operation="{operation}"
}
```

**Count (Counter):**
```
{operation}_total{
    operation="{operation}",
    status="{status}"
}
```

## Best Practices

1. **Use Consistent Naming**: Keep operation names consistent across your application
2. **Categorize Errors**: Use meaningful error types and categories for better observability
3. **Standard Status Values**: Use standard status values like "success", "error", "timeout"
4. **Avoid High Cardinality**: Don't use dynamic values (like user IDs) in tags
5. **Mock in Tests**: Always use mocks in unit tests to verify metric recording

## Thread Safety

All methods are safe for concurrent use. The package uses minimal locking and efficient caching for high-performance concurrent access.