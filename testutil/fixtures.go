package testutil

import (
	"time"

	"github.com/MichaelAJay/go-metrics/metric"
)

// Common test fixtures and utilities

// DefaultCounterOptions returns standard counter options for testing.
func DefaultCounterOptions() metric.Options {
	return metric.Options{
		Name:        "test_counter",
		Description: "A test counter metric",
		Unit:        "requests",
		Tags: metric.Tags{
			"service": "test",
			"env":     "test",
		},
	}
}

// DefaultGaugeOptions returns standard gauge options for testing.
func DefaultGaugeOptions() metric.Options {
	return metric.Options{
		Name:        "test_gauge",
		Description: "A test gauge metric",
		Unit:        "connections",
		Tags: metric.Tags{
			"service": "test",
			"env":     "test",
		},
	}
}

// DefaultHistogramOptions returns standard histogram options for testing.
func DefaultHistogramOptions() metric.Options {
	return metric.Options{
		Name:        "test_histogram",
		Description: "A test histogram metric",
		Unit:        "bytes",
		Tags: metric.Tags{
			"service": "test",
			"env":     "test",
		},
	}
}

// DefaultTimerOptions returns standard timer options for testing.
func DefaultTimerOptions() metric.Options {
	return metric.Options{
		Name:        "test_timer",
		Description: "A test timer metric",
		Unit:        "milliseconds",
		Tags: metric.Tags{
			"service": "test",
			"env":     "test",
		},
	}
}

// CommonTags returns a set of common tags for testing.
func CommonTags() metric.Tags {
	return metric.Tags{
		"service": "test-service",
		"env":     "test",
		"version": "1.0.0",
	}
}

// AdditionalTags returns additional tags for With() testing.
func AdditionalTags() metric.Tags {
	return metric.Tags{
		"method": "GET",
		"status": "200",
	}
}

// SetupTestRegistry creates a new MockRegistry with some pre-configured metrics.
func SetupTestRegistry() *MockRegistry {
	registry := NewMockRegistry()

	// Pre-create some metrics
	registry.Counter(DefaultCounterOptions())
	registry.Gauge(DefaultGaugeOptions())
	registry.Histogram(DefaultHistogramOptions())
	registry.Timer(DefaultTimerOptions())

	return registry
}

// SimulateCounterActivity performs typical counter operations for testing.
func SimulateCounterActivity(counter *MockCounter) {
	counter.Inc()
	counter.Add(5.0)
	counter.Add(10.0)
	counter.With(AdditionalTags())
}

// SimulateGaugeActivity performs typical gauge operations for testing.
func SimulateGaugeActivity(gauge *MockGauge) {
	gauge.Set(100.0)
	gauge.Add(50.0)
	gauge.Inc()
	gauge.Dec()
	gauge.With(AdditionalTags())
}

// SimulateHistogramActivity performs typical histogram operations for testing.
func SimulateHistogramActivity(histogram *MockHistogram) {
	histogram.Observe(10.0)
	histogram.Observe(25.0)
	histogram.Observe(50.0)
	histogram.Observe(100.0)
	histogram.With(AdditionalTags())
}

// SimulateTimerActivity performs typical timer operations for testing.
func SimulateTimerActivity(timer *MockTimer) {
	timer.Record(10 * time.Millisecond)
	timer.Record(25 * time.Millisecond)
	timer.RecordSince(time.Now().Add(-50 * time.Millisecond))
	timer.Time(func() {
		time.Sleep(1 * time.Millisecond)
	})
	timer.With(AdditionalTags())
}

// TestScenario represents a complete test scenario with expected outcomes.
type TestScenario struct {
	Name        string
	Setup       func(*MockRegistry)
	Assertions  func(*MockRegistry)
	Description string
}

// CommonTestScenarios returns a set of common test scenarios.
func CommonTestScenarios() []TestScenario {
	return []TestScenario{
		{
			Name: "Basic Counter Operations",
			Setup: func(registry *MockRegistry) {
				counter := registry.Counter(DefaultCounterOptions())
				mockCounter := counter.(*MockCounter)
				SimulateCounterActivity(mockCounter)
			},
			Description: "Tests basic counter increment and add operations",
		},
		{
			Name: "Basic Gauge Operations",
			Setup: func(registry *MockRegistry) {
				gauge := registry.Gauge(DefaultGaugeOptions())
				mockGauge := gauge.(*MockGauge)
				SimulateGaugeActivity(mockGauge)
			},
			Description: "Tests basic gauge set, add, increment, and decrement operations",
		},
		{
			Name: "Basic Histogram Operations",
			Setup: func(registry *MockRegistry) {
				histogram := registry.Histogram(DefaultHistogramOptions())
				mockHistogram := histogram.(*MockHistogram)
				SimulateHistogramActivity(mockHistogram)
			},
			Description: "Tests basic histogram observation operations",
		},
		{
			Name: "Basic Timer Operations",
			Setup: func(registry *MockRegistry) {
				timer := registry.Timer(DefaultTimerOptions())
				mockTimer := timer.(*MockTimer)
				SimulateTimerActivity(mockTimer)
			},
			Description: "Tests basic timer recording operations",
		},
		{
			Name: "Multiple Metrics",
			Setup: func(registry *MockRegistry) {
				counter := registry.Counter(metric.Options{Name: "requests_total"})
				gauge := registry.Gauge(metric.Options{Name: "active_connections"})
				histogram := registry.Histogram(metric.Options{Name: "request_duration"})
				timer := registry.Timer(metric.Options{Name: "processing_time"})

				counter.(*MockCounter).Inc()
				gauge.(*MockGauge).Set(50)
				histogram.(*MockHistogram).Observe(100)
				timer.(*MockTimer).Record(10 * time.Millisecond)
			},
			Description: "Tests multiple different metrics working together",
		},
	}
}

// BenchmarkScenario represents a scenario for performance testing.
type BenchmarkScenario struct {
	Name        string
	Setup       func(*MockRegistry) any
	Operation   func(any)
	Description string
}

// BenchmarkScenarios returns scenarios for benchmarking mock performance.
func BenchmarkScenarios() []BenchmarkScenario {
	return []BenchmarkScenario{
		{
			Name: "Counter Increment",
			Setup: func(registry *MockRegistry) any {
				return registry.Counter(DefaultCounterOptions())
			},
			Operation: func(m any) {
				m.(metric.Counter).Inc()
			},
			Description: "Benchmark counter increment operations",
		},
		{
			Name: "Gauge Set",
			Setup: func(registry *MockRegistry) any {
				return registry.Gauge(DefaultGaugeOptions())
			},
			Operation: func(m any) {
				m.(metric.Gauge).Set(100.0)
			},
			Description: "Benchmark gauge set operations",
		},
		{
			Name: "Histogram Observe",
			Setup: func(registry *MockRegistry) any {
				return registry.Histogram(DefaultHistogramOptions())
			},
			Operation: func(m any) {
				m.(metric.Histogram).Observe(50.0)
			},
			Description: "Benchmark histogram observe operations",
		},
		{
			Name: "Timer Record",
			Setup: func(registry *MockRegistry) any {
				return registry.Timer(DefaultTimerOptions())
			},
			Operation: func(m any) {
				m.(metric.Timer).Record(10 * time.Millisecond)
			},
			Description: "Benchmark timer record operations",
		},
	}
}
