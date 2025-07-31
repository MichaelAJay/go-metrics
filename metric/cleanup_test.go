package metric

import (
	"testing"
	"time"
)

func TestMetricTTL(t *testing.T) {
	// Create registry with short cleanup interval for testing
	registry := NewRegistry(DefaultTagValidationConfig(), 100*time.Millisecond)
	defer registry.Close()

	// Create a metric with TTL
	counter := registry.Counter(Options{
		Name: "ttl_counter",
		TTL:  200 * time.Millisecond,
	})

	counter.Inc()
	if counter.Value() != 1 {
		t.Errorf("Expected counter value 1, got %d", counter.Value())
	}

	// Metric should still exist before TTL expires
	time.Sleep(100 * time.Millisecond)
	counter2 := registry.Counter(Options{Name: "ttl_counter"})
	if counter2.Value() != 1 {
		t.Errorf("Expected counter to still exist with value 1, got %d", counter2.Value())
	}

	// Wait for TTL to expire and cleanup to run
	time.Sleep(300 * time.Millisecond)

	// Create new counter with same name - should be a fresh instance
	counter3 := registry.Counter(Options{Name: "ttl_counter"})
	if counter3.Value() != 0 {
		t.Errorf("Expected fresh counter with value 0, got %d", counter3.Value())
	}
}

func TestMetricWithoutTTL(t *testing.T) {
	registry := NewRegistry(DefaultTagValidationConfig(), 50*time.Millisecond)
	defer registry.Close()

	// Create a metric without TTL
	counter := registry.Counter(Options{Name: "persistent_counter"})
	counter.Inc()

	// Wait longer than cleanup interval
	time.Sleep(200 * time.Millisecond)

	// Counter should still exist
	counter2 := registry.Counter(Options{Name: "persistent_counter"})
	if counter2.Value() != 1 {
		t.Errorf("Expected persistent counter to retain value 1, got %d", counter2.Value())
	}
}

func TestManualCleanup(t *testing.T) {
	registry := NewRegistry(DefaultTagValidationConfig(), time.Hour) // Long interval
	defer registry.Close()

	// Create metric with short TTL
	counter := registry.Counter(Options{
		Name: "manual_cleanup_counter",
		TTL:  50 * time.Millisecond,
	})
	counter.Inc()

	// Wait for TTL to expire
	time.Sleep(100 * time.Millisecond)

	// Counter should still exist because cleanup hasn't run
	counter2 := registry.Counter(Options{Name: "manual_cleanup_counter"})
	if counter2.Value() != 1 {
		t.Errorf("Expected counter to still exist before manual cleanup, got %d", counter2.Value())
	}

	// Run manual cleanup
	registry.ManualCleanup()

	// Now counter should be gone
	counter3 := registry.Counter(Options{Name: "manual_cleanup_counter"})
	if counter3.Value() != 0 {
		t.Errorf("Expected fresh counter after manual cleanup, got %d", counter3.Value())
	}
}

func TestRegistryClose(t *testing.T) {
	registry := NewRegistry(DefaultTagValidationConfig(), 50*time.Millisecond)

	// Create some metrics
	counter := registry.Counter(Options{Name: "test_counter"})
	counter.Inc()

	// Close should not return error
	err := registry.Close()
	if err != nil {
		t.Errorf("Expected Close() to succeed, got error: %v", err)
	}

	// Registry should still be functional after close (but no background cleanup)
	counter2 := registry.Counter(Options{Name: "test_counter2"})
	counter2.Inc()
	if counter2.Value() != 1 {
		t.Errorf("Expected counter to work after Close(), got %d", counter2.Value())
	}
}

func TestCardinalityWithTTL(t *testing.T) {
	config := TagValidationConfig{
		MaxKeys:        10,
		MaxKeyLength:   100,
		MaxValueLength: 200,
		MaxCardinality: 2,
	}
	registry := NewRegistry(config, 50*time.Millisecond)
	defer registry.Close()

	// Create two metrics (at cardinality limit)
	counter1 := registry.Counter(Options{
		Name: "cardinality_test",
		TTL:  100 * time.Millisecond,
	})
	gauge1 := registry.Gauge(Options{
		Name: "cardinality_test",
		TTL:  100 * time.Millisecond,
	})

	counter1.Inc()
	gauge1.Set(10)

	// Wait for TTL to expire and cleanup
	time.Sleep(200 * time.Millisecond)

	// Should be able to create new metrics after cleanup reduced cardinality
	counter2 := registry.Counter(Options{Name: "cardinality_test"})
	gauge2 := registry.Gauge(Options{Name: "cardinality_test"})

	counter2.Inc()
	gauge2.Set(20)

	if counter2.Value() != 1 {
		t.Errorf("Expected fresh counter value 1, got %d", counter2.Value())
	}
	if gauge2.Value() != 20 {
		t.Errorf("Expected fresh gauge value 20, got %d", gauge2.Value())
	}
}

func TestMixedTTLMetrics(t *testing.T) {
	registry := NewRegistry(DefaultTagValidationConfig(), 50*time.Millisecond)
	defer registry.Close()

	// Create metrics with different TTLs
	shortTTL := registry.Counter(Options{
		Name: "short_ttl",
		TTL:  50 * time.Millisecond,
	})
	longTTL := registry.Counter(Options{
		Name: "long_ttl", 
		TTL:  300 * time.Millisecond,
	})
	noTTL := registry.Counter(Options{
		Name: "no_ttl",
	})

	shortTTL.Inc()
	longTTL.Inc()
	noTTL.Inc()

	// Wait for short TTL to expire
	time.Sleep(150 * time.Millisecond)

	// Short TTL should be gone, others should remain
	shortTTL2 := registry.Counter(Options{Name: "short_ttl"})
	longTTL2 := registry.Counter(Options{Name: "long_ttl"})
	noTTL2 := registry.Counter(Options{Name: "no_ttl"})

	if shortTTL2.Value() != 0 {
		t.Errorf("Expected short TTL counter to be reset, got %d", shortTTL2.Value())
	}
	if longTTL2.Value() != 1 {
		t.Errorf("Expected long TTL counter to persist, got %d", longTTL2.Value())
	}
	if noTTL2.Value() != 1 {
		t.Errorf("Expected no TTL counter to persist, got %d", noTTL2.Value())
	}

	// Wait for long TTL to expire
	time.Sleep(200 * time.Millisecond)

	longTTL3 := registry.Counter(Options{Name: "long_ttl"})
	noTTL3 := registry.Counter(Options{Name: "no_ttl"})

	if longTTL3.Value() != 0 {
		t.Errorf("Expected long TTL counter to be reset after expiration, got %d", longTTL3.Value())
	}
	if noTTL3.Value() != 1 {
		t.Errorf("Expected no TTL counter to still persist, got %d", noTTL3.Value())
	}
}

func TestEachWithTTL(t *testing.T) {
	registry := NewRegistry(DefaultTagValidationConfig(), 50*time.Millisecond)
	defer registry.Close()

	// Create metrics with TTL
	counter := registry.Counter(Options{
		Name: "each_counter",
		TTL:  50 * time.Millisecond,
	})
	gauge := registry.Gauge(Options{
		Name: "each_gauge",
		TTL:  300 * time.Millisecond,
	})

	counter.Inc()
	gauge.Set(42)

	// Count metrics before expiration
	count := 0
	registry.Each(func(m Metric) {
		count++
	})

	if count != 2 {
		t.Errorf("Expected 2 metrics before expiration, got %d", count)
	}

	// Wait for first metric to expire
	time.Sleep(150 * time.Millisecond)

	// Count metrics  after first expiration
	count = 0
	registry.Each(func(m Metric) {
		count++
	})

	if count != 1 {
		t.Errorf("Expected 1 metric after first expiration, got %d", count)
	}
}