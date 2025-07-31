package otel

import (
	"testing"
	"time"

	"github.com/MichaelAJay/go-metrics/metric"
)

func TestNewReporter(t *testing.T) {
	// Test with basic parameters
	reporter, err := NewReporter("test-service", "v1.0.0")
	if err != nil {
		t.Fatalf("NewReporter() returned error: %v", err)
	}
	if reporter == nil {
		t.Fatal("NewReporter() returned nil")
	}

	// Clean up
	if err := reporter.Close(); err != nil {
		t.Errorf("Close() returned error: %v", err)
	}

	// Test with custom options
	customReporter, err := NewReporter(
		"custom-service",
		"v2.0.0",
		WithAttributes(map[string]string{
			"environment": "test",
			"region":      "us-west-2",
		}),
	)
	if err != nil {
		t.Fatalf("NewReporter() with custom options returned error: %v", err)
	}
	if customReporter == nil {
		t.Fatal("NewReporter() with custom options returned nil")
	}

	// Clean up
	if err := customReporter.Close(); err != nil {
		t.Errorf("Close() returned error: %v", err)
	}
}

func TestReporterImplementsInterface(t *testing.T) {
	reporter, err := NewReporter("test-service", "v1.0.0")
	if err != nil {
		t.Fatalf("NewReporter() returned error: %v", err)
	}
	defer reporter.Close()

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
	reporter, err := NewReporter("test-service", "v1.0.0")
	if err != nil {
		t.Fatalf("NewReporter() returned error: %v", err)
	}
	defer reporter.Close()

	// Call Report - we're mostly checking that it doesn't panic
	err = reporter.Report(registry)
	if err != nil {
		t.Errorf("Report() returned error: %v", err)
	}

	// Test Flush
	err = reporter.Flush()
	if err != nil {
		t.Errorf("Flush() returned error: %v", err)
	}
}

func TestReportCounter(t *testing.T) {
	registry := metric.NewDefaultRegistry()
	reporter, err := NewReporter("test-service", "v1.0.0")
	if err != nil {
		t.Fatalf("NewReporter() returned error: %v", err)
	}
	defer reporter.Close()

	// Create a counter and increment it
	counter := registry.Counter(metric.Options{
		Name:        "test_counter",
		Description: "Test counter for unit testing",
		Unit:        "count",
		Tags:        metric.Tags{"test": "true"},
	})
	
	// Increment the counter multiple times
	counter.Inc()
	counter.Add(5)
	
	// Verify the counter has the expected value
	if counter.Value() != 6 {
		t.Errorf("Expected counter value 6, got %d", counter.Value())
	}

	// Report the metrics
	err = reporter.Report(registry)
	if err != nil {
		t.Errorf("Report() returned error: %v", err)
	}

	// Verify the counter is tracked
	reporter.mutex.RLock()
	_, exists := reporter.counters["test_counter"]
	reporter.mutex.RUnlock()
	
	if !exists {
		t.Error("Counter was not created in reporter")
	}
}

func TestReportGauge(t *testing.T) {
	registry := metric.NewDefaultRegistry()
	reporter, err := NewReporter("test-service", "v1.0.0")
	if err != nil {
		t.Fatalf("NewReporter() returned error: %v", err)
	}
	defer reporter.Close()

	// Create a gauge and set values
	gauge := registry.Gauge(metric.Options{
		Name:        "test_gauge",
		Description: "Test gauge for unit testing",
		Unit:        "bytes",
		Tags:        metric.Tags{"test": "true"},
	})
	
	// Set and modify the gauge
	gauge.Set(100)
	gauge.Add(50)
	gauge.Add(-25) // Use Add with negative value instead of Sub
	
	// Verify the gauge has the expected value
	if gauge.Value() != 125 {
		t.Errorf("Expected gauge value 125, got %d", gauge.Value())
	}

	// Report the metrics
	err = reporter.Report(registry)
	if err != nil {
		t.Errorf("Report() returned error: %v", err)
	}

	// Verify the gauge is tracked
	reporter.mutex.RLock()
	_, exists := reporter.gauges["test_gauge"]
	reporter.mutex.RUnlock()
	
	if !exists {
		t.Error("Gauge was not created in reporter")
	}
}

func TestReportHistogram(t *testing.T) {
	registry := metric.NewDefaultRegistry()
	reporter, err := NewReporter("test-service", "v1.0.0")
	if err != nil {
		t.Fatalf("NewReporter() returned error: %v", err)
	}
	defer reporter.Close()

	// Create a histogram and add observations
	histogram := registry.Histogram(metric.Options{
		Name:        "test_histogram",
		Description: "Test histogram for unit testing",
		Unit:        "ms",
		Tags:        metric.Tags{"test": "true"},
	})
	
	// Add multiple observations
	histogram.Observe(10)
	histogram.Observe(20)
	histogram.Observe(30)
	
	// Verify the histogram has observations
	snapshot := histogram.Snapshot()
	if snapshot.Count != 3 {
		t.Errorf("Expected histogram count 3, got %d", snapshot.Count)
	}
	if snapshot.Sum != 60 {
		t.Errorf("Expected histogram sum 60, got %d", snapshot.Sum)
	}

	// Report the metrics
	err = reporter.Report(registry)
	if err != nil {
		t.Errorf("Report() returned error: %v", err)
	}

	// Verify the histogram is tracked
	reporter.mutex.RLock()
	_, exists := reporter.histograms["test_histogram"]
	reporter.mutex.RUnlock()
	
	if !exists {
		t.Error("Histogram was not created in reporter")
	}
}

func TestReportTimer(t *testing.T) {
	registry := metric.NewDefaultRegistry()
	reporter, err := NewReporter("test-service", "v1.0.0")
	if err != nil {
		t.Fatalf("NewReporter() returned error: %v", err)
	}
	defer reporter.Close()

	// Create a timer and record durations
	timer := registry.Timer(metric.Options{
		Name:        "test_timer",
		Description: "Test timer for unit testing",
		Unit:        "ms",
		Tags:        metric.Tags{"test": "true"},
	})
	
	// Record multiple durations
	timer.Record(time.Millisecond * 10)
	timer.Record(time.Millisecond * 20)
	timer.Record(time.Millisecond * 30)
	
	// Verify the timer has recordings
	snapshot := timer.Snapshot()
	if snapshot.Count != 3 {
		t.Errorf("Expected timer count 3, got %d", snapshot.Count)
	}

	// Report the metrics
	err = reporter.Report(registry)
	if err != nil {
		t.Errorf("Report() returned error: %v", err)
	}

	// Verify the timer histogram is tracked (timers create histograms with "_seconds" suffix)
	reporter.mutex.RLock()
	_, exists := reporter.histograms["test_timer_seconds"]
	reporter.mutex.RUnlock()
	
	if !exists {
		t.Error("Timer histogram was not created in reporter")
	}
}

func TestReporterClose(t *testing.T) {
	reporter, err := NewReporter("test-service", "v1.0.0")
	if err != nil {
		t.Fatalf("NewReporter() returned error: %v", err)
	}

	// Add some metrics to create callbacks
	registry := metric.NewDefaultRegistry()
	gauge := registry.Gauge(metric.Options{
		Name: "test_gauge",
		Tags: metric.Tags{"test": "true"},
	})
	gauge.Set(42)

	// Report to create callbacks
	err = reporter.Report(registry)
	if err != nil {
		t.Errorf("Report() returned error: %v", err)
	}

	// Close should not return error and should clean up resources
	err = reporter.Close()
	if err != nil {
		t.Errorf("Close() returned error: %v", err)
	}

	// Context should be cancelled
	select {
	case <-reporter.ctx.Done():
		// Expected
	default:
		t.Error("Context was not cancelled after Close()")
	}
}

func TestConvertTags(t *testing.T) {
	reporter, err := NewReporter("test-service", "v1.0.0",
		WithAttributes(map[string]string{
			"default": "value",
		}))
	if err != nil {
		t.Fatalf("NewReporter() returned error: %v", err)
	}
	defer reporter.Close()

	// Test with empty tags
	attrs := reporter.convertTags(metric.Tags{})
	if len(attrs) != 1 {
		t.Errorf("Expected 1 attribute (default), got %d", len(attrs))
	}

	// Test with some tags
	tags := metric.Tags{
		"service": "test",
		"env":     "testing",
	}
	attrs = reporter.convertTags(tags)
	expectedLen := 3 // 1 default + 2 tags
	if len(attrs) != expectedLen {
		t.Errorf("Expected %d attributes, got %d", expectedLen, len(attrs))
	}
}

func TestMultipleReports(t *testing.T) {
	registry := metric.NewDefaultRegistry()
	reporter, err := NewReporter("test-service", "v1.0.0")
	if err != nil {
		t.Fatalf("NewReporter() returned error: %v", err)
	}
	defer reporter.Close()

	// Create metrics
	counter := registry.Counter(metric.Options{Name: "test_counter"})
	gauge := registry.Gauge(metric.Options{Name: "test_gauge"})

	// Multiple reports should not cause issues
	for i := 0; i < 5; i++ {
		counter.Inc()
		gauge.Set(float64(i * 10))

		err = reporter.Report(registry)
		if err != nil {
			t.Errorf("Report() iteration %d returned error: %v", i, err)
		}
	}

	// Verify final values
	if counter.Value() != 5 {
		t.Errorf("Expected counter value 5, got %d", counter.Value())
	}
	if gauge.Value() != 40 {
		t.Errorf("Expected gauge value 40, got %d", gauge.Value())
	}
}
