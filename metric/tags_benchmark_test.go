package metric

import (
	"testing"
)

// BenchmarkCounterTags benchmarks the Counter.Tags() method to measure defensive copy allocation overhead
func BenchmarkCounterTags(b *testing.B) {
	registry := NewDefaultRegistry()
	counter := registry.Counter(Options{
		Name:        "benchmark_counter_tags",
		Description: "Benchmark counter Tags() operations",
		Tags: Tags{
			"service":   "api",
			"method":    "POST",
			"endpoint":  "/users",
			"status":    "200",
			"region":    "us-west-2",
		},
	})

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = counter.Tags()
	}
}

// BenchmarkGaugeTags benchmarks the Gauge.Tags() method to measure defensive copy allocation overhead
func BenchmarkGaugeTags(b *testing.B) {
	registry := NewDefaultRegistry()
	gauge := registry.Gauge(Options{
		Name:        "benchmark_gauge_tags",
		Description: "Benchmark gauge Tags() operations",
		Tags: Tags{
			"service":   "api",
			"method":    "POST",
			"endpoint":  "/users",
			"status":    "200",
			"region":    "us-west-2",
		},
	})

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = gauge.Tags()
	}
}

// BenchmarkHistogramTags benchmarks the Histogram.Tags() method to measure defensive copy allocation overhead
func BenchmarkHistogramTags(b *testing.B) {
	registry := NewDefaultRegistry()
	histogram := registry.Histogram(Options{
		Name:        "benchmark_histogram_tags",
		Description: "Benchmark histogram Tags() operations",
		Tags: Tags{
			"service":   "api",
			"method":    "POST",
			"endpoint":  "/users",
			"status":    "200",
			"region":    "us-west-2",
		},
	})

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = histogram.Tags()
	}
}

// BenchmarkTimerTags benchmarks the Timer.Tags() method to measure defensive copy allocation overhead
func BenchmarkTimerTags(b *testing.B) {
	registry := NewDefaultRegistry()
	timer := registry.Timer(Options{
		Name:        "benchmark_timer_tags",
		Description: "Benchmark timer Tags() operations",
		Tags: Tags{
			"service":   "api",
			"method":    "POST",
			"endpoint":  "/users",
			"status":    "200",
			"region":    "us-west-2",
		},
	})

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = timer.Tags()
	}
}