package operational

import (
	"testing"
	"time"

	"github.com/MichaelAJay/go-metrics/metric"
)

// warmOperationalTagPool pre-populates the operational tag pool for benchmarking
func warmOperationalTagPool(count int) {
	for i := 0; i < count; i++ {
		operationalTagPool.Put(make(map[string]string, 8))
	}
}

// BenchmarkRecordErrorColdStart benchmarks RecordError with fresh OperationalMetrics each time
// This simulates the worst-case scenario where metrics are not cached
func BenchmarkRecordErrorColdStart(b *testing.B) {
	registry := metric.NewDefaultRegistry()

	// Warm the tag pool to avoid measuring pool allocation overhead
	warmOperationalTagPool(10)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Create fresh operational metrics each time (no caching benefit)
		om := New(registry)
		om.RecordError("auth", "validation_error", "invalid_token")
	}
}

// BenchmarkRecordErrorCachedMetrics benchmarks RecordError with reused OperationalMetrics
// This simulates production usage where metrics are cached after first creation
func BenchmarkRecordErrorCachedMetrics(b *testing.B) {
	registry := metric.NewDefaultRegistry()
	om := New(registry)

	// Pre-warm the cache by calling once before benchmark
	om.RecordError("auth", "validation_error", "invalid_token")

	// Warm the tag pool to avoid measuring pool allocation overhead
	warmOperationalTagPool(10)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Same operation - should hit cached metric
		om.RecordError("auth", "validation_error", "invalid_token")
	}
}

// BenchmarkRecordErrorVariedOperations benchmarks RecordError with different operations
// This simulates mixed workload where some operations are cached, others are new
func BenchmarkRecordErrorVariedOperations(b *testing.B) {
	registry := metric.NewDefaultRegistry()
	om := New(registry)

	operations := []string{"auth", "query", "cache", "storage"}
	errorTypes := []string{"validation_error", "timeout_error", "connection_error"}
	categories := []string{"invalid_token", "expired", "unavailable"}

	// Warm the tag pool to avoid measuring pool allocation overhead
	warmOperationalTagPool(15)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		op := operations[i%len(operations)]
		errType := errorTypes[i%len(errorTypes)]
		category := categories[i%len(categories)]
		om.RecordError(op, errType, category)
	}
}

// BenchmarkRecordOperationColdStart benchmarks RecordOperation with fresh OperationalMetrics
func BenchmarkRecordOperationColdStart(b *testing.B) {
	registry := metric.NewDefaultRegistry()

	// Warm the tag pool (need more as RecordOperation uses 2 pooled maps)
	warmOperationalTagPool(20)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Create fresh operational metrics each time
		om := New(registry)
		om.RecordOperation("auth", "success", time.Millisecond)
	}
}

// BenchmarkRecordOperationCachedMetrics benchmarks RecordOperation with cached metrics
func BenchmarkRecordOperationCachedMetrics(b *testing.B) {
	registry := metric.NewDefaultRegistry()
	om := New(registry)

	// Pre-warm the cache
	om.RecordOperation("auth", "success", time.Millisecond)

	// Warm the tag pool (need more as RecordOperation uses 2 pooled maps)
	warmOperationalTagPool(20)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Same operation - should hit cached metrics
		om.RecordOperation("auth", "success", time.Millisecond)
	}
}

// BenchmarkRecordOperationVariedOperations benchmarks mixed RecordOperation calls
func BenchmarkRecordOperationVariedOperations(b *testing.B) {
	registry := metric.NewDefaultRegistry()
	om := New(registry)

	operations := []string{"auth", "query", "cache", "storage"}
	statuses := []string{"success", "error", "timeout"}

	// Warm the tag pool (need more as RecordOperation uses 2 pooled maps)
	warmOperationalTagPool(25)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		op := operations[i%len(operations)]
		status := statuses[i%len(statuses)]
		om.RecordOperation(op, status, time.Duration(i%1000)*time.Microsecond)
	}
}