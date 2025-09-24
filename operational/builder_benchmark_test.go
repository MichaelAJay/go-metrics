package operational

import (
	"testing"
	"time"

	"github.com/MichaelAJay/go-metrics/metric"
)

// BenchmarkMetricsBuilder_RecordWithContext benchmarks the new MetricsBuilder approach
func BenchmarkMetricsBuilder_RecordWithContext(b *testing.B) {
	registry := metric.NewRegistry(metric.DefaultTagValidationConfig(), 0) // No cleanup for benchmarks
	defer registry.Close()
	om := New(registry)
	builder := NewMetricsBuilder(om)

	context := map[string]string{
		"provider":  "password",
		"user_type": "premium",
		"region":    "us-east-1",
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		builder.RecordWithContext("authentication", "success", 100*time.Millisecond, context)
	}
}

// BenchmarkMetricsBuilder_RecordWithContext_NoContext benchmarks without additional context
func BenchmarkMetricsBuilder_RecordWithContext_NoContext(b *testing.B) {
	registry := metric.NewRegistry(metric.DefaultTagValidationConfig(), 0)
	defer registry.Close()
	om := New(registry)
	builder := NewMetricsBuilder(om)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		builder.RecordWithContext("authentication", "success", 100*time.Millisecond, nil)
	}
}

// BenchmarkMetricsBuilder_RecordSecurityEvent benchmarks security event recording
func BenchmarkMetricsBuilder_RecordSecurityEvent(b *testing.B) {
	registry := metric.NewRegistry(metric.DefaultTagValidationConfig(), 0)
	defer registry.Close()
	om := New(registry)
	builder := NewMetricsBuilder(om)

	context := map[string]string{
		"ip":         "192.168.1.1",
		"user_agent": "Mozilla/5.0",
		"country":    "US",
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		builder.RecordSecurityEvent("brute_force", "blocked", context)
	}
}

// BenchmarkMetricsBuilder_RecordBusinessMetric benchmarks business metric recording
func BenchmarkMetricsBuilder_RecordBusinessMetric(b *testing.B) {
	registry := metric.NewRegistry(metric.DefaultTagValidationConfig(), 0)
	defer registry.Close()
	om := New(registry)
	builder := NewMetricsBuilder(om)

	context := map[string]string{
		"source": "organic",
		"tier":   "premium",
		"plan":   "annual",
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		builder.RecordBusinessMetric("user_conversion", "completed", 250.5, context)
	}
}

// BenchmarkDirectMapAllocation shows the allocation cost of the anti-pattern
func BenchmarkDirectMapAllocation(b *testing.B) {
	registry := metric.NewRegistry(metric.DefaultTagValidationConfig(), 0)
	defer registry.Close()

	// Get a base counter to measure just the map allocation overhead
	counter := registry.Counter(metric.Options{
		Name: "direct_map_test",
	})

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Simulate the anti-pattern: direct map allocation
		tags := map[string]string{
			"provider":  "password",
			"user_type": "premium",
			"region":    "us-east-1",
		}
		_ = tags // Use the tags to prevent optimization
		counter.Inc()
	}
}

// BenchmarkOperationalMetrics_DirectUsage benchmarks the underlying operational metrics directly
func BenchmarkOperationalMetrics_DirectUsage(b *testing.B) {
	registry := metric.NewRegistry(metric.DefaultTagValidationConfig(), 0)
	defer registry.Close()
	om := New(registry)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		om.RecordOperation("authentication", "success", 100*time.Millisecond)
	}
}

// BenchmarkTagPoolUsage demonstrates the efficiency of the tag pool
func BenchmarkTagPoolUsage(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Get from pool
		tags := operationalTagPool.Get().(map[string]string)

		// Use it
		tags["key1"] = "value1"
		tags["key2"] = "value2"
		tags["key3"] = "value3"

		// Return to pool
		operationalTagPool.Put(clearOperationalTags(tags))
	}
}