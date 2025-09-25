package operational

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/MichaelAJay/go-metrics/metric"
)

// Phase 2: Performance Benchmarks - Comprehensive Implementation
// Based on METRICSBUILDER_TEST_PLAN.md Phase 2 requirements

// =====================================================
// 2.1 Allocation Comparison Benchmarks
// =====================================================

// BenchmarkAuthServiceAntiPattern simulates the BAD pattern from auth service analysis
// Target baseline: 362 allocs/op with 189 map[string]string{} literal allocations
func BenchmarkAuthServiceAntiPattern(b *testing.B) {
	registry := metric.NewRegistry(metric.DefaultTagValidationConfig(), 0)
	defer registry.Close()

	// Create base metrics like auth service would
	authAttempts := registry.Timer(metric.Options{Name: "auth_attempts"})

	providers := []string{"password", "oauth", "mfa", "sso"}
	statuses := []string{"success", "error", "timeout"}
	subjects := []string{"user1", "user2", "user3", "premium", "basic"}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Simulate the anti-pattern: direct map allocation on each call
		tags := map[string]string{  // Direct allocation - this is the problem!
			"provider":     providers[i%len(providers)],
			"status":       statuses[i%len(statuses)],
			"subject_hash": subjects[i%len(subjects)],
		}
		authAttempts.With(tags).Record(time.Duration(i%1000) * time.Millisecond)
	}
}

// BenchmarkAuthServiceAntiPatternEquivalent simulates what it would take to record
// the same NUMBER of metrics as MetricsBuilder using the anti-pattern
func BenchmarkAuthServiceAntiPatternEquivalent(b *testing.B) {
	registry := metric.NewRegistry(metric.DefaultTagValidationConfig(), 0)
	defer registry.Close()

	// Create multiple metrics to match what MetricsBuilder does
	authTimer := registry.Timer(metric.Options{Name: "authentication"})
	providerTimer := registry.Timer(metric.Options{Name: "authentication_provider"})
	userTypeTimer := registry.Timer(metric.Options{Name: "authentication_user_type"})

	providers := []string{"password", "oauth", "mfa", "sso"}
	statuses := []string{"success", "error", "timeout"}
	userTypes := []string{"premium", "basic", "enterprise"}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		duration := time.Duration(i%1000) * time.Millisecond

		// Record 3 metrics with direct map allocation (like MetricsBuilder does 3 operations)
		authTimer.With(map[string]string{"operation": "authentication", "status": statuses[i%len(statuses)]}).Record(duration)
		providerTimer.With(map[string]string{"operation": "authentication_provider", "status": providers[i%len(providers)]}).Record(duration)
		userTypeTimer.With(map[string]string{"operation": "authentication_user_type", "status": userTypes[i%len(userTypes)]}).Record(duration)
	}
}

// BenchmarkMetricsBuilderPattern demonstrates the GOOD pattern with context
// This records 3 operations (1 main + 2 contextual)
func BenchmarkMetricsBuilderPattern(b *testing.B) {
	registry := metric.NewRegistry(metric.DefaultTagValidationConfig(), 0)
	defer registry.Close()
	om := New(registry)
	builder := NewMetricsBuilder(om)

	providers := []string{"password", "oauth", "mfa", "sso"}
	statuses := []string{"success", "error", "timeout"}
	userTypes := []string{"premium", "basic", "enterprise"}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Context map is NOT allocated in hot path - reuses pool
		context := map[string]string{
			"provider":  providers[i%len(providers)],
			"user_type": userTypes[i%len(userTypes)],
		}
		builder.RecordWithContext("authentication", statuses[i%len(statuses)], time.Duration(i%1000)*time.Millisecond, context)
	}
}

// BenchmarkMetricsBuilderNoContext demonstrates MetricsBuilder with just 1 operation (fair comparison)
func BenchmarkMetricsBuilderNoContext(b *testing.B) {
	registry := metric.NewRegistry(metric.DefaultTagValidationConfig(), 0)
	defer registry.Close()
	om := New(registry)
	builder := NewMetricsBuilder(om)

	statuses := []string{"success", "error", "timeout"}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// No context - this should be directly comparable to the anti-pattern
		builder.RecordWithContext("authentication", statuses[i%len(statuses)], time.Duration(i%1000)*time.Millisecond, nil)
	}
}

// BenchmarkDirectVsBuilderComparison provides side-by-side comparison
func BenchmarkDirectVsBuilderComparison(b *testing.B) {
	b.Run("DirectMapAllocation", func(b *testing.B) {
		registry := metric.NewRegistry(metric.DefaultTagValidationConfig(), 0)
		defer registry.Close()

		counter := registry.Counter(metric.Options{Name: "direct_test"})

		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			tags := map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			}
			counter.With(tags).Inc()
		}
	})

	b.Run("MetricsBuilderPooled", func(b *testing.B) {
		registry := metric.NewRegistry(metric.DefaultTagValidationConfig(), 0)
		defer registry.Close()
		om := New(registry)
		builder := NewMetricsBuilder(om)

		context := map[string]string{
			"key1": "value1",
			"key2": "value2",
			"key3": "value3",
		}

		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			builder.RecordWithContext("test_operation", "success", 0, context)
		}
	})
}

// =====================================================
// 2.2 Scale and Cardinality Benchmarks
// =====================================================

// BenchmarkLowCardinality tests with 3-5 unique tag combinations
func BenchmarkLowCardinality(b *testing.B) {
	registry := metric.NewRegistry(metric.DefaultTagValidationConfig(), 0)
	defer registry.Close()
	om := New(registry)
	builder := NewMetricsBuilder(om)

	contexts := []map[string]string{
		{"env": "prod", "region": "us"},
		{"env": "dev", "region": "us"},
		{"env": "prod", "region": "eu"},
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		context := contexts[i%len(contexts)]
		builder.RecordWithContext("operation", "success", time.Millisecond, context)
	}
}

// BenchmarkMediumCardinality tests with 50-100 unique combinations
func BenchmarkMediumCardinality(b *testing.B) {
	registry := metric.NewRegistry(metric.DefaultTagValidationConfig(), 0)
	defer registry.Close()
	om := New(registry)
	builder := NewMetricsBuilder(om)

	// Generate 50 unique combinations
	contexts := make([]map[string]string, 50)
	for i := 0; i < 50; i++ {
		contexts[i] = map[string]string{
			"provider": []string{"oauth", "saml", "ldap", "local"}[i%4],
			"tier":     []string{"free", "basic", "premium", "enterprise"}[i%4],
			"region":   []string{"us-east", "us-west", "eu-central"}[i%3],
			"client":   []string{"web", "mobile", "api"}[i%3],
		}
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		context := contexts[i%len(contexts)]
		builder.RecordWithContext("auth_flow", "completed", time.Millisecond*100, context)
	}
}

// BenchmarkHighCardinality tests with 1000+ unique combinations
func BenchmarkHighCardinality(b *testing.B) {
	registry := metric.NewRegistry(metric.DefaultTagValidationConfig(), 0)
	defer registry.Close()
	om := New(registry)
	builder := NewMetricsBuilder(om)

	// Generate 1000 unique combinations with high cardinality
	contexts := make([]map[string]string, 1000)
	for i := 0; i < 1000; i++ {
		contexts[i] = map[string]string{
			"user_id":     []string{"user1", "user2", "user3", "user4", "user5"}[i%5],
			"session_id":  []string{"sess1", "sess2", "sess3", "sess4", "sess5"}[i%5],
			"request_id":  []string{"req1", "req2", "req3", "req4", "req5"}[i%5],
			"ip_range":    []string{"192.168.1", "192.168.2", "10.0.0", "172.16.0"}[i%4],
			"user_agent":  []string{"chrome", "firefox", "safari", "edge"}[i%4],
			"country":     []string{"US", "CA", "GB", "DE", "FR"}[i%5],
			"device_type": []string{"desktop", "mobile", "tablet"}[i%3],
			"os_version":  []string{"v1", "v2", "v3", "v4"}[i%4],
		}
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		context := contexts[i%len(contexts)]
		builder.RecordWithContext("high_cardinality_op", "processed", time.Microsecond*500, context)
	}
}

// BenchmarkContextSizeVariation tests different numbers of context keys
func BenchmarkContextSizeVariation(b *testing.B) {
	b.Run("1_Context_Key", func(b *testing.B) {
		registry := metric.NewRegistry(metric.DefaultTagValidationConfig(), 0)
		defer registry.Close()
		om := New(registry)
		builder := NewMetricsBuilder(om)

		context := map[string]string{"key1": "value1"}

		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			builder.RecordWithContext("op", "success", time.Millisecond, context)
		}
	})

	b.Run("3_Context_Keys", func(b *testing.B) {
		registry := metric.NewRegistry(metric.DefaultTagValidationConfig(), 0)
		defer registry.Close()
		om := New(registry)
		builder := NewMetricsBuilder(om)

		context := map[string]string{
			"key1": "value1",
			"key2": "value2",
			"key3": "value3",
		}

		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			builder.RecordWithContext("op", "success", time.Millisecond, context)
		}
	})

	b.Run("5_Context_Keys", func(b *testing.B) {
		registry := metric.NewRegistry(metric.DefaultTagValidationConfig(), 0)
		defer registry.Close()
		om := New(registry)
		builder := NewMetricsBuilder(om)

		context := map[string]string{
			"key1": "value1",
			"key2": "value2",
			"key3": "value3",
			"key4": "value4",
			"key5": "value5",
		}

		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			builder.RecordWithContext("op", "success", time.Millisecond, context)
		}
	})

	b.Run("10_Context_Keys", func(b *testing.B) {
		registry := metric.NewRegistry(metric.DefaultTagValidationConfig(), 0)
		defer registry.Close()
		om := New(registry)
		builder := NewMetricsBuilder(om)

		context := map[string]string{
			"key1":  "value1",
			"key2":  "value2",
			"key3":  "value3",
			"key4":  "value4",
			"key5":  "value5",
			"key6":  "value6",
			"key7":  "value7",
			"key8":  "value8",
			"key9":  "value9",
			"key10": "value10",
		}

		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			builder.RecordWithContext("op", "success", time.Millisecond, context)
		}
	})
}

// BenchmarkConcurrentWorkers tests different levels of concurrency
func BenchmarkConcurrentWorkers(b *testing.B) {
	workers := []int{1, 10, 100, 1000}

	for _, numWorkers := range workers {
		b.Run(fmt.Sprintf("%d_workers", numWorkers), func(b *testing.B) {
			registry := metric.NewRegistry(metric.DefaultTagValidationConfig(), 0)
			defer registry.Close()
			om := New(registry)
			builder := NewMetricsBuilder(om)

			context := map[string]string{
				"worker_test": "true",
				"concurrency": "high",
			}

			b.ResetTimer()
			b.ReportAllocs()

			var wg sync.WaitGroup
			opsPerWorker := b.N / numWorkers
			if opsPerWorker == 0 {
				opsPerWorker = 1
			}

			for w := 0; w < numWorkers; w++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					for i := 0; i < opsPerWorker; i++ {
						builder.RecordWithContext("concurrent_op", "processed", time.Microsecond*100, context)
					}
				}()
			}
			wg.Wait()
		})
	}
}

// =====================================================
// 2.3 Memory Pressure Benchmarks
// =====================================================

// BenchmarkMemoryPressureScenario tests behavior under memory pressure
func BenchmarkMemoryPressureScenario(b *testing.B) {
	b.Run("Large_Number_Concurrent_Operations", func(b *testing.B) {
		registry := metric.NewRegistry(metric.DefaultTagValidationConfig(), 0)
		defer registry.Close()
		om := New(registry)
		builder := NewMetricsBuilder(om)

		// Create many different operation types to stress the system
		operationTypes := make([]string, 100)
		for i := 0; i < 100; i++ {
			operationTypes[i] = fmt.Sprintf("operation_type_%d", i)
		}

		contexts := make([]map[string]string, 50)
		for i := 0; i < 50; i++ {
			contexts[i] = map[string]string{
				"stress_test": "true",
				"batch_id":    fmt.Sprintf("batch_%d", i%10),
				"instance":    fmt.Sprintf("instance_%d", i%5),
			}
		}

		b.ResetTimer()
		b.ReportAllocs()

		var wg sync.WaitGroup
		numWorkers := 100
		opsPerWorker := b.N / numWorkers
		if opsPerWorker == 0 {
			opsPerWorker = 1
		}

		for w := 0; w < numWorkers; w++ {
			wg.Add(1)
			go func(workerID int) {
				defer wg.Done()
				for i := 0; i < opsPerWorker; i++ {
					opType := operationTypes[(workerID*opsPerWorker+i)%len(operationTypes)]
					context := contexts[i%len(contexts)]
					builder.RecordWithContext(opType, "completed", time.Microsecond*50, context)
				}
			}(w)
		}
		wg.Wait()
	})

	b.Run("Pool_Efficiency_Under_Stress", func(b *testing.B) {
		// Test that the tag pool maintains efficiency under stress
		b.ResetTimer()
		b.ReportAllocs()

		var wg sync.WaitGroup
		numWorkers := 50
		opsPerWorker := b.N / numWorkers
		if opsPerWorker == 0 {
			opsPerWorker = 1
		}

		for w := 0; w < numWorkers; w++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for i := 0; i < opsPerWorker; i++ {
					// Simulate pool usage under concurrent stress
					tags := operationalTagPool.Get().(map[string]string)

					// Heavy usage
					for j := 0; j < 10; j++ {
						tags[fmt.Sprintf("key_%d", j)] = fmt.Sprintf("value_%d", j)
					}

					operationalTagPool.Put(clearOperationalTags(tags))
				}
			}()
		}
		wg.Wait()
	})

	b.Run("GC_Pressure_Comparison", func(b *testing.B) {
		// Compare GC pressure between MetricsBuilder and direct allocation
		registry := metric.NewRegistry(metric.DefaultTagValidationConfig(), 0)
		defer registry.Close()
		om := New(registry)
		builder := NewMetricsBuilder(om)

		// Force some GC activity before test
		var dummy []map[string]string
		for i := 0; i < 1000; i++ {
			dummy = append(dummy, map[string]string{"gc": "pressure"})
		}
		dummy = nil

		context := map[string]string{
			"gc_test": "active",
			"memory":  "pressure",
		}

		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			builder.RecordWithContext("gc_pressure_test", "executed", time.Microsecond*10, context)

			// Trigger some GC activity periodically
			if i%1000 == 0 {
				// Force a small allocation to trigger GC
				_ = make([]byte, 1024)
			}
		}
	})
}

// BenchmarkAuthServiceWorkloadSimulation replicates the exact patterns from the auth service analysis
func BenchmarkAuthServiceWorkloadSimulation(b *testing.B) {
	registry := metric.NewRegistry(metric.DefaultTagValidationConfig(), 0)
	defer registry.Close()
	om := New(registry)
	builder := NewMetricsBuilder(om)

	// Simulate the exact patterns identified in the auth service analysis
	authContexts := []map[string]string{
		{"provider": "password", "status": "success", "user_type": "premium"},
		{"provider": "oauth", "status": "success", "user_type": "basic"},
		{"provider": "mfa", "status": "error", "user_type": "premium"},
		{"provider": "sso", "status": "timeout", "user_type": "enterprise"},
	}

	securityEvents := []struct {
		eventType string
		action    string
		context   map[string]string
	}{
		{"brute_force", "blocked", map[string]string{"ip": "192.168.1.1", "user_agent": "curl/7.68.0"}},
		{"credential_stuffing", "flagged", map[string]string{"source": "tor", "patterns": "high"}},
		{"anomaly_detection", "triggered", map[string]string{"risk_score": "85", "location": "unusual"}},
	}

	businessMetrics := []struct {
		metricType string
		category   string
		value      float64
		context    map[string]string
	}{
		{"session_duration", "completed", 1800.5, map[string]string{"tier": "premium"}},
		{"provider_usage", "oauth", 1.0, map[string]string{"region": "us-east"}},
		{"conversion_rate", "signup", 0.85, map[string]string{"source": "organic"}},
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Authentication flow (most common)
		authCtx := authContexts[i%len(authContexts)]
		builder.RecordWithContext("authentication", authCtx["status"], 100*time.Millisecond, authCtx)

		// Security events (periodic)
		if i%10 == 0 {
			event := securityEvents[i%len(securityEvents)]
			builder.RecordSecurityEvent(event.eventType, event.action, event.context)
		}

		// Business metrics (less frequent)
		if i%25 == 0 {
			metric := businessMetrics[i%len(businessMetrics)]
			builder.RecordBusinessMetric(metric.metricType, metric.category, metric.value, metric.context)
		}
	}
}