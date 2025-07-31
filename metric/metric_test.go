package metric

import (
	"context"
	"testing"
	"time"
)

func TestCounter(t *testing.T) {
	registry := NewDefaultRegistry()
	counter := registry.Counter(Options{
		Name:        "test_counter",
		Description: "Test counter",
		Unit:        "count",
	})

	// Test increment
	counter.Inc()
	counter.Inc()
	counter.Add(3.0)

	// We can't assert the actual value since it's not exposed in the interface
	// A real test would use a test reporter or mock registry to verify values
}

func TestGauge(t *testing.T) {
	registry := NewDefaultRegistry()
	gauge := registry.Gauge(Options{
		Name:        "test_gauge",
		Description: "Test gauge",
		Unit:        "bytes",
	})

	// Test operations
	gauge.Set(100)
	gauge.Inc()
	gauge.Add(20)
	gauge.Dec()

	// We can't assert the actual value since it's not exposed in the interface
	// A real test would use a test reporter or mock registry to verify values
}

func TestHistogram(t *testing.T) {
	registry := NewDefaultRegistry()
	histogram := registry.Histogram(Options{
		Name:        "test_histogram",
		Description: "Test histogram",
		Unit:        "milliseconds",
	})

	// Record some values
	histogram.Observe(10)
	histogram.Observe(20)
	histogram.Observe(30)

	// We can't assert the actual value since it's not exposed in the interface
	// A real test would use a test reporter or mock registry to verify values
}

func TestTimer(t *testing.T) {
	registry := NewDefaultRegistry()
	timer := registry.Timer(Options{
		Name:        "test_timer",
		Description: "Test timer",
		Unit:        "milliseconds",
	})

	// Record a duration
	timer.Record(100 * time.Millisecond)

	// Time since a start time
	start := time.Now().Add(-50 * time.Millisecond) // 50ms ago
	timer.RecordSince(start)

	// Time a function
	duration := timer.Time(func() {
		time.Sleep(10 * time.Millisecond)
	})

	if duration < 10*time.Millisecond {
		t.Errorf("Timer recorded less than 10ms: %v", duration)
	}

	// We can't assert the actual values since they're not exposed in the interface
	// A real test would use a test reporter or mock registry to verify values
}

func TestTagging(t *testing.T) {
	registry := NewDefaultRegistry()
	counter := registry.Counter(Options{
		Name: "tagged_counter",
		Tags: Tags{
			"service": "test",
			"method":  "GET",
		},
	})

	// Create a counter with additional tags
	taggedCounter := counter.With(Tags{
		"status": "200",
		"region": "us-west",
	})

	// Make sure the base counter and the tagged counter are different
	if counter == taggedCounter {
		t.Error("Tagged counter should return a new instance")
	}

	// Check that tags are preserved
	tags := taggedCounter.Tags()
	expectedTags := map[string]string{
		"service": "test",
		"method":  "GET",
		"status":  "200",
		"region":  "us-west",
	}

	for k, v := range expectedTags {
		if tags[k] != v {
			t.Errorf("Expected tag %s=%s, got %s", k, v, tags[k])
		}
	}

	// Both counters should be usable
	counter.Inc()
	taggedCounter.Inc()
}

func TestRegistry(t *testing.T) {
	registry := NewDefaultRegistry()

	// Create metrics of different types
	registry.Counter(Options{Name: "counter1"})
	registry.Gauge(Options{Name: "gauge1"})
	registry.Histogram(Options{Name: "histogram1"})
	registry.Timer(Options{Name: "timer1"})

	// Create metrics with same name but different types
	// Using blank identifiers to avoid unused variable warnings
	_ = registry.Counter(Options{Name: "metric1"})
	_ = registry.Gauge(Options{Name: "metric1"})

	// Should get the same counter back
	_ = registry.Counter(Options{Name: "counter1"})

	// Count metrics
	count := 0
	registry.Each(func(m Metric) {
		count++
	})

	// We should have 6 metrics (4 unique names, 2 duplicated names but different types)
	if count != 6 {
		t.Errorf("Expected 6 metrics, got %d", count)
	}

	// Unregister a metric
	registry.Unregister("counter1")

	// Count again
	count = 0
	registry.Each(func(m Metric) {
		count++
	})

	// We should have 5 metrics after unregistering
	if count != 5 {
		t.Errorf("Expected 5 metrics after unregistering, got %d", count)
	}
}

func TestContext(t *testing.T) {
	registry := NewDefaultRegistry()
	// Use background context instead of nil
	parentCtx := context.Background()
	ctx := NewContext(parentCtx, registry)

	extractedRegistry, ok := FromContext(ctx)
	if !ok {
		t.Error("Expected registry to be found in context")
	}

	if extractedRegistry != registry {
		t.Error("Registry extracted from context is not the same as the original")
	}
}
