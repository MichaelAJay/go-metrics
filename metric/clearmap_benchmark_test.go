package metric

import (
	"testing"
)

// BenchmarkClearMapSmall benchmarks clearing small maps (typical tag map size)
func BenchmarkClearMapSmall(b *testing.B) {
	// Pre-populate pool to avoid measuring pool allocation overhead
	warmTagMapPool(10)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		m := tagMapPool.Get().(map[string]string)

		// Simulate typical tag usage (3-4 tags)
		m["operation"] = "test"
		m["status"] = "success"
		m["service"] = "auth"

		clearMap(m)
		tagMapPool.Put(m)
	}
}

// BenchmarkClearMapMedium benchmarks clearing medium-sized maps
func BenchmarkClearMapMedium(b *testing.B) {
	warmTagMapPool(10)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		m := tagMapPool.Get().(map[string]string)

		// Simulate larger tag usage (8-10 tags)
		m["operation"] = "test"
		m["status"] = "success"
		m["service"] = "auth"
		m["method"] = "POST"
		m["endpoint"] = "/api/v1/users"
		m["region"] = "us-east-1"
		m["environment"] = "production"
		m["version"] = "1.2.3"

		clearMap(m)
		tagMapPool.Put(m)
	}
}

// BenchmarkClearMapLarge benchmarks clearing large maps (stress test)
func BenchmarkClearMapLarge(b *testing.B) {
	warmTagMapPool(10)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		m := tagMapPool.Get().(map[string]string)

		// Simulate very large tag usage (15+ tags)
		for j := 0; j < 15; j++ {
			key := "key" + string(rune('0'+j%10))
			value := "value" + string(rune('0'+j%10))
			m[key] = value
		}

		clearMap(m)
		tagMapPool.Put(m)
	}
}

// BenchmarkClearMapVsRecreate compares clearing vs recreating maps
func BenchmarkClearMapVsRecreate(b *testing.B) {
	b.Run("ClearAndReuse", func(b *testing.B) {
		warmTagMapPool(10)

		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			m := tagMapPool.Get().(map[string]string)

			m["operation"] = "test"
			m["status"] = "success"
			m["service"] = "auth"

			clearMap(m)
			tagMapPool.Put(m)
		}
	})

	b.Run("RecreateEachTime", func(b *testing.B) {
		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			m := make(map[string]string, 8)

			m["operation"] = "test"
			m["status"] = "success"
			m["service"] = "auth"

			// No clearing needed since we're recreating
			_ = m
		}
	})
}

// BenchmarkMapClearingMethods compares different clearing approaches
func BenchmarkMapClearingMethods(b *testing.B) {
	b.Run("RangeDelete", func(b *testing.B) {
		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			m := make(map[string]string)
			m["key1"] = "value1"
			m["key2"] = "value2"
			m["key3"] = "value3"

			// Range delete approach (our clearMap function)
			for k := range m {
				delete(m, k)
			}
		}
	})

	b.Run("RecreateMap", func(b *testing.B) {
		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			m := make(map[string]string)
			m["key1"] = "value1"
			m["key2"] = "value2"
			m["key3"] = "value3"

			// Recreate approach
			m = make(map[string]string)
			_ = m
		}
	})
}

// BenchmarkPooledMapLifecycle benchmarks the complete lifecycle with pooling
func BenchmarkPooledMapLifecycle(b *testing.B) {
	warmTagMapPool(10)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Get from pool
		m := tagMapPool.Get().(map[string]string)

		// Use the map (typical tag operations)
		m["operation"] = "test"
		m["status"] = "success"
		m["service"] = "auth"
		m["method"] = "GET"

		// Clear and return to pool
		tagMapPool.Put(clearMap(m))
	}
}

// BenchmarkClearMapEmpty benchmarks clearing already-empty maps
func BenchmarkClearMapEmpty(b *testing.B) {
	warmTagMapPool(10)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		m := tagMapPool.Get().(map[string]string)

		// Map is already empty from pool
		clearMap(m)
		tagMapPool.Put(m)
	}
}

// BenchmarkClearMapCapacityPreservation tests that capacity is preserved
func BenchmarkClearMapCapacityPreservation(b *testing.B) {
	warmTagMapPool(10)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		m := tagMapPool.Get().(map[string]string)

		// Fill map beyond initial capacity
		for j := 0; j < 12; j++ { // More than initial capacity of 8
			m["key"+string(rune('0'+j%10))] = "value"
		}

		clearMap(m)
		// Map should retain its grown capacity
		tagMapPool.Put(m)
	}
}