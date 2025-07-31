package prometheus

import (
	"testing"

	"github.com/MichaelAJay/go-metrics/metric"
	prom "github.com/prometheus/client_golang/prometheus"
)

func TestNewReporter(t *testing.T) {
	// Test with no options
	reporter := NewReporter()
	if reporter == nil {
		t.Fatal("NewReporter() returned nil")
	}

	// Test with custom options
	customRegistry := prom.NewRegistry()
	customReporter := NewReporter(
		WithRegistry(customRegistry),
		WithDefaultLabels(map[string]string{
			"service": "test",
			"env":     "testing",
		}),
	)
	if customReporter == nil {
		t.Fatal("NewReporter() with custom options returned nil")
	}
}

func TestReporterImplementsInterface(t *testing.T) {
	reporter := NewReporter()

	// Assert that our reporter implements the Reporter interface
	var _ metric.Reporter = reporter
}

func TestReportWithMetrics(t *testing.T) {
	// Create a registry with some metrics
	registry := metric.NewDefaultRegistry()

	// Add some metrics
	counter := registry.Counter(metric.Options{
		Name:        "test_counter",
		Description: "Test counter",
		Unit:        "count",
		Tags:        metric.Tags{"service": "test"},
	})
	counter.Inc()

	gauge := registry.Gauge(metric.Options{
		Name:        "test_gauge",
		Description: "Test gauge",
		Unit:        "bytes",
		Tags:        metric.Tags{"service": "test"},
	})
	gauge.Set(100)

	histogram := registry.Histogram(metric.Options{
		Name:        "test_histogram",
		Description: "Test histogram",
		Unit:        "ms",
		Tags:        metric.Tags{"service": "test"},
	})
	histogram.Observe(50)

	timer := registry.Timer(metric.Options{
		Name:        "test_timer",
		Description: "Test timer",
		Unit:        "ms",
		Tags:        metric.Tags{"service": "test"},
	})
	timer.Record(100)

	// Create a reporter
	reporter := NewReporter()

	// Call Report - we're mostly checking that it doesn't panic
	err := reporter.Report(registry)
	if err != nil {
		t.Errorf("Report() returned error: %v", err)
	}

	// Test Flush
	err = reporter.Flush()
	if err != nil {
		t.Errorf("Flush() returned error: %v", err)
	}

	// Test Close
	err = reporter.Close()
	if err != nil {
		t.Errorf("Close() returned error: %v", err)
	}
}
