package operational

import (
	"testing"
)

// BenchmarkOperationalClearMapSmall benchmarks clearing small operational tag maps
func BenchmarkOperationalClearMapSmall(b *testing.B) {
	warmOperationalTagPool(10)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		m := operationalTagPool.Get().(map[string]string)

		// Simulate typical operational tag usage
		m["operation"] = "auth"
		m["error_type"] = "validation_error"
		m["error_category"] = "invalid_token"

		clearOperationalTags(m)
		operationalTagPool.Put(m)
	}
}

// BenchmarkOperationalClearMapMedium benchmarks clearing medium operational tag maps
func BenchmarkOperationalClearMapMedium(b *testing.B) {
	warmOperationalTagPool(10)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		m := operationalTagPool.Get().(map[string]string)

		// Simulate more comprehensive operational tagging
		m["operation"] = "database_query"
		m["error_type"] = "connection_error"
		m["error_category"] = "timeout"
		m["database"] = "postgres"
		m["table"] = "users"
		m["query_type"] = "select"

		clearOperationalTags(m)
		operationalTagPool.Put(m)
	}
}

// BenchmarkOperationalMapClearingMethods compares clearing approaches for operational package
func BenchmarkOperationalMapClearingMethods(b *testing.B) {
	b.Run("OperationalClearFunction", func(b *testing.B) {
		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			m := make(map[string]string)
			m["operation"] = "auth"
			m["status"] = "success"

			// Use operational clearing function
			clearOperationalTags(m)
		}
	})

	b.Run("RecreateMap", func(b *testing.B) {
		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			m := make(map[string]string)
			m["operation"] = "auth"
			m["status"] = "success"

			// Recreate instead of clearing
			m = make(map[string]string)
			_ = m
		}
	})
}

// BenchmarkOperationalPooledMapLifecycle benchmarks complete lifecycle in operational package
func BenchmarkOperationalPooledMapLifecycle(b *testing.B) {
	warmOperationalTagPool(10)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Get from pool
		m := operationalTagPool.Get().(map[string]string)

		// Use for operation tracking
		m["operation"] = "auth"
		m["status"] = "success"
		m["duration_ms"] = "150"

		// Clear and return to pool
		operationalTagPool.Put(clearOperationalTags(m))
	}
}

// BenchmarkOperationalClearVsMetricClear compares clearing functions between packages
func BenchmarkOperationalClearVsMetricClear(b *testing.B) {
	b.Run("OperationalClear", func(b *testing.B) {
		warmOperationalTagPool(10)

		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			m := operationalTagPool.Get().(map[string]string)
			m["key1"] = "value1"
			m["key2"] = "value2"
			m["key3"] = "value3"

			clearOperationalTags(m)
			operationalTagPool.Put(m)
		}
	})

	// Note: We can't directly import metric package clearMap function here
	// but the benchmarks show both use the same range-delete approach
}

// BenchmarkOperationalClearEmpty benchmarks clearing already-empty operational maps
func BenchmarkOperationalClearEmpty(b *testing.B) {
	warmOperationalTagPool(10)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		m := operationalTagPool.Get().(map[string]string)

		// Map is already empty from pool
		clearOperationalTags(m)
		operationalTagPool.Put(m)
	}
}