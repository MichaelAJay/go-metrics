# go-metrics Test Utilities

This package provides comprehensive mock implementations and testing utilities for the go-metrics library, enabling easy unit testing of applications that use metrics.

## Features

- **Complete Mock Implementation**: Full mock versions of all metric types (Counter, Gauge, Histogram, Timer) and Registry
- **Call Tracking**: Records all method calls with parameters for inspection
- **Thread Safety**: All mocks are safe for concurrent use
- **Value Inspection**: Access to internal metric values for verification
- **Callback Support**: Optional callbacks for custom test behavior
- **Reset Functionality**: Clean state between tests
- **Assertion Helpers**: Convenient functions for common test assertions
- **Test Fixtures**: Pre-configured scenarios for common testing patterns

## Quick Start

```go
import (
    "testing"
    "github.com/MichaelAJay/go-metrics/metric"
    "github.com/MichaelAJay/go-metrics/testutil"
)

func TestMyService(t *testing.T) {
    // Create a mock registry
    mockRegistry := testutil.NewMockRegistry()
    
    // Create your service with the mock registry
    service := NewMyService(mockRegistry)
    
    // Execute the code under test
    service.ProcessRequest()
    
    // Verify metrics were recorded correctly
    counter := mockRegistry.GetCounter("requests_total")
    testutil.AssertCounterValue(t, counter, 1)
    testutil.AssertCounterIncCalls(t, counter, 1)
}
```

## Mock Types

### MockRegistry

The `MockRegistry` captures all metric creation and management operations:

```go
mockRegistry := testutil.NewMockRegistry()

// Track metric creation calls
counter := mockRegistry.Counter(metric.Options{Name: "test_counter"})
gauge := mockRegistry.Gauge(metric.Options{Name: "test_gauge"})

// Verify call counts
testutil.AssertRegistryCallCounts(t, mockRegistry, 1, 1, 0, 0) // counters, gauges, histograms, timers

// Access created metrics
testCounter := mockRegistry.GetCounter("test_counter")
```

### MockCounter

Tracks counter operations with value inspection:

```go
counter := mockRegistry.Counter(metric.Options{Name: "requests"})
mockCounter := counter.(*testutil.MockCounter)

counter.Inc()
counter.Add(5.0)

// Verify operations
testutil.AssertCounterValue(t, mockCounter, 6)
testutil.AssertCounterIncCalls(t, mockCounter, 1)
testutil.AssertCounterAddCalls(t, mockCounter, []float64{5.0})
```

### MockGauge

Tracks gauge operations with state inspection:

```go
gauge := mockRegistry.Gauge(metric.Options{Name: "connections"})
mockGauge := gauge.(*testutil.MockGauge)

gauge.Set(100.0)
gauge.Add(50.0)
gauge.Inc()
gauge.Dec()

// Verify operations
testutil.AssertGaugeValue(t, mockGauge, 150) // 100 + 50 + 1 - 1
testutil.AssertGaugeSetCalls(t, mockGauge, []float64{100.0})
```

### MockHistogram

Tracks histogram observations with snapshot inspection:

```go
histogram := mockRegistry.Histogram(metric.Options{Name: "durations"})
mockHistogram := histogram.(*testutil.MockHistogram)

histogram.Observe(10.0)
histogram.Observe(25.0)
histogram.Observe(50.0)

// Verify observations
testutil.AssertHistogramObserveCalls(t, mockHistogram, []float64{10.0, 25.0, 50.0})
testutil.AssertHistogramSnapshot(t, mockHistogram, 3, 85) // count, sum
```

### MockTimer

Tracks timer operations with duration inspection:

```go
timer := mockRegistry.Timer(metric.Options{Name: "processing_time"})
mockTimer := timer.(*testutil.MockTimer)

timer.Record(10 * time.Millisecond)
timer.RecordSince(time.Now().Add(-50 * time.Millisecond))
timer.Time(func() {
    time.Sleep(1 * time.Millisecond)
})

// Verify timing operations
testutil.AssertTimerRecordCalls(t, mockTimer, 3) // Record + RecordSince + Time
testutil.AssertTimerRecordCallsWithin(t, mockTimer, 0, 100*time.Millisecond)
```

## Tagged Metrics Testing

Test metrics with tags using the `With()` method:

```go
counter := mockRegistry.Counter(metric.Options{Name: "requests"})
mockCounter := counter.(*testutil.MockCounter)

// Create tagged versions
counter.With(metric.Tags{"method": "GET", "status": "200"}).Inc()
counter.With(metric.Tags{"method": "POST", "status": "201"}).Inc()

// Verify tag usage
expectedTags := []metric.Tags{
    {"method": "GET", "status": "200"},
    {"method": "POST", "status": "201"},
}
testutil.AssertWithCalls(t, mockCounter.WithCalls(), expectedTags)
```

## Custom Behavior with Callbacks

Use callbacks to inject custom behavior during testing:

```go
mockRegistry := testutil.NewMockRegistry()

// Set up callback to track counter increments
var incrementCount int
mockRegistry.OnCounterCallback = func(opts metric.Options) metric.Counter {
    counter := testutil.NewMockCounter(opts)
    counter.OnIncCallback = func() {
        incrementCount++
    }
    return counter
}

counter := mockRegistry.Counter(metric.Options{Name: "test"})
counter.Inc() // incrementCount is now 1
```

## Test Fixtures and Scenarios

Use pre-built fixtures for common testing patterns:

```go
// Use default options
opts := testutil.DefaultCounterOptions()
counter := mockRegistry.Counter(opts)

// Set up a complete test registry
registry := testutil.SetupTestRegistry()

// Run common test scenarios
scenarios := testutil.CommonTestScenarios()
for _, scenario := range scenarios {
    t.Run(scenario.Name, func(t *testing.T) {
        mockRegistry := testutil.NewMockRegistry()
        scenario.Setup(mockRegistry)
        // Add your assertions here
    })
}
```

## Reset Between Tests

Clean state between tests:

```go
func TestSomething(t *testing.T) {
    mockRegistry := testutil.NewMockRegistry()
    counter := mockRegistry.Counter(metric.Options{Name: "test"})
    mockCounter := counter.(*testutil.MockCounter)
    
    // Use the counter
    counter.Inc()
    
    // Reset for clean state
    mockCounter.Reset()
    testutil.AssertCounterValue(t, mockCounter, 0)
    
    // Or reset the entire registry
    mockRegistry.Reset()
}
```

## Benchmarking

Test performance with mock metrics:

```go
func BenchmarkMyService(b *testing.B) {
    mockRegistry := testutil.NewMockRegistry()
    service := NewMyService(mockRegistry)
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        service.ProcessRequest()
    }
}

// Or use built-in benchmark scenarios
func BenchmarkMockMetrics(b *testing.B) {
    scenarios := testutil.BenchmarkScenarios()
    for _, scenario := range scenarios {
        b.Run(scenario.Name, func(b *testing.B) {
            mockRegistry := testutil.NewMockRegistry()
            metric := scenario.Setup(mockRegistry)
            
            b.ResetTimer()
            for i := 0; i < b.N; i++ {
                scenario.Operation(metric)
            }
        })
    }
}
```

## Assertion Helpers

The package provides numerous assertion helpers for common verification patterns:

- `AssertCounterValue()` - Verify counter values
- `AssertCounterIncCalls()` - Verify increment call counts
- `AssertCounterAddCalls()` - Verify add operations with values
- `AssertGaugeValue()` - Verify gauge values
- `AssertGaugeSetCalls()` - Verify gauge set operations
- `AssertHistogramObserveCalls()` - Verify histogram observations
- `AssertHistogramSnapshot()` - Verify histogram statistics
- `AssertTimerRecordCalls()` - Verify timer record operations
- `AssertTimerRecordCallsWithin()` - Verify timer durations are within range
- `AssertMetricTags()` - Verify metric tags
- `AssertWithCalls()` - Verify tagged metric usage
- `AssertRegistryCallCounts()` - Verify registry usage patterns

## Thread Safety

All mock implementations are thread-safe and can be used in concurrent tests:

```go
func TestConcurrentAccess(t *testing.T) {
    mockRegistry := testutil.NewMockRegistry()
    counter := mockRegistry.Counter(metric.Options{Name: "concurrent_counter"})
    
    var wg sync.WaitGroup
    for i := 0; i < 100; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            counter.Inc()
        }()
    }
    wg.Wait()
    
    mockCounter := counter.(*testutil.MockCounter)
    testutil.AssertCounterValue(t, mockCounter, 100)
}
```

## Integration with Real Metrics

The mocks implement the same interfaces as real metrics, making it easy to swap them in tests:

```go
type MyService struct {
    registry metric.Registry
}

func NewMyService(registry metric.Registry) *MyService {
    return &MyService{registry: registry}
}

// In production
service := NewMyService(metric.NewDefaultRegistry())

// In tests
service := NewMyService(testutil.NewMockRegistry())
```

This provides complete test coverage of your metrics usage without actually sending metrics to external systems during testing.