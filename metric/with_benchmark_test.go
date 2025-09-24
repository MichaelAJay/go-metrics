package metric

import (
	"testing"
)

// BenchmarkCounterWith benchmarks the Counter.With() method to measure tag allocation overhead
func BenchmarkCounterWith(b *testing.B) {
	registry := NewDefaultRegistry()
	counter := registry.Counter(Options{
		Name:        "benchmark_counter_with",
		Description: "Benchmark counter With() operations",
		Tags: Tags{
			"base_tag": "base_value",
		},
	})

	tags := Tags{
		"operation": "test",
		"status":    "success",
	}

	// Warm the pool to avoid measuring pool allocation overhead
	warmTagMapPool(10)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		tagged := counter.With(tags)
		tagged.Inc()
	}
}

// BenchmarkGaugeWith benchmarks the Gauge.With() method to measure tag allocation overhead
func BenchmarkGaugeWith(b *testing.B) {
	registry := NewDefaultRegistry()
	gauge := registry.Gauge(Options{
		Name:        "benchmark_gauge_with",
		Description: "Benchmark gauge With() operations",
		Tags: Tags{
			"base_tag": "base_value",
		},
	})

	tags := Tags{
		"operation": "test",
		"status":    "success",
	}

	// Warm the pool to avoid measuring pool allocation overhead
	warmTagMapPool(10)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		tagged := gauge.With(tags)
		tagged.Set(float64(i))
	}
}

// BenchmarkHistogramWith benchmarks the Histogram.With() method to measure tag allocation overhead
func BenchmarkHistogramWith(b *testing.B) {
	registry := NewDefaultRegistry()
	histogram := registry.Histogram(Options{
		Name:        "benchmark_histogram_with",
		Description: "Benchmark histogram With() operations",
		Tags: Tags{
			"base_tag": "base_value",
		},
	})

	tags := Tags{
		"operation": "test",
		"status":    "success",
	}

	// Warm the pool to avoid measuring pool allocation overhead
	warmTagMapPool(10)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		tagged := histogram.With(tags)
		tagged.Observe(float64(i % 100))
	}
}

// BenchmarkTimerWith benchmarks the Timer.With() method to measure tag allocation overhead
func BenchmarkTimerWith(b *testing.B) {
	registry := NewDefaultRegistry()
	timer := registry.Timer(Options{
		Name:        "benchmark_timer_with",
		Description: "Benchmark timer With() operations",
		Tags: Tags{
			"base_tag": "base_value",
		},
	})

	tags := Tags{
		"operation": "test",
		"status":    "success",
	}

	// Warm the pool to avoid measuring pool allocation overhead
	warmTagMapPool(10)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		tagged := timer.With(tags)
		tagged.Record(1000) // 1 microsecond
	}
}