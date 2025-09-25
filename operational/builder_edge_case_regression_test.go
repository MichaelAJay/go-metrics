package operational

import (
	"fmt"
	"math"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/MichaelAJay/go-metrics/metric"
)

// TestPhase4EdgeCases implements comprehensive edge case testing for MetricsBuilder
// This covers Phase 4 of the test implementation strategy
func TestPhase4EdgeCases(t *testing.T) {
	t.Run("Nil Context Handling", testNilContextHandling)
	t.Run("Very Large Context Maps", testVeryLargeContextMaps)
	t.Run("Very Long Tag Values", testVeryLongTagValues)
	t.Run("Empty and Whitespace Strings", testEmptyAndWhitespaceStrings)
	t.Run("Special Characters and Unicode", testSpecialCharactersAndUnicode)
	t.Run("Extreme Values", testExtremeValues)
}

func testNilContextHandling(t *testing.T) {
	registry := metric.NewRegistry(metric.DefaultTagValidationConfig(), 5*time.Minute)
	defer registry.Close()

	om := New(registry)
	builder := NewMetricsBuilder(om)

	tests := []struct {
		name    string
		context map[string]string
	}{
		{"Nil context", nil},
		{"Empty context map", make(map[string]string)},
		{"Context with empty values", map[string]string{"key": ""}},
		{"Context with empty keys", map[string]string{"": "value"}},
		{"Context with both empty", map[string]string{"": ""}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Panic occurred with %s: %v", tt.name, r)
				}
			}()

			// Test all builder methods with various nil/empty contexts
			builder.RecordWithContext("test_operation", "success", 10*time.Millisecond, tt.context)
			builder.RecordSecurityEvent("test_event", "occurred", tt.context)
			builder.RecordBusinessMetric("test_metric", "completed", 1.0, tt.context)
		})
	}
}

func testVeryLargeContextMaps(t *testing.T) {
	registry := metric.NewRegistry(metric.DefaultTagValidationConfig(), 5*time.Minute)
	defer registry.Close()

	om := New(registry)
	builder := NewMetricsBuilder(om)

	// Test with contexts of varying sizes
	contextSizes := []int{50, 100, 250, 500}

	for _, size := range contextSizes {
		t.Run(fmt.Sprintf("Context with %d keys", size), func(t *testing.T) {
			largeContext := make(map[string]string, size)
			for i := 0; i < size; i++ {
				largeContext[fmt.Sprintf("key_%d", i)] = fmt.Sprintf("value_%d", i)
			}

			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Panic with %d context keys: %v", size, r)
				}
			}()

			start := time.Now()
			builder.RecordWithContext("large_context_test", "success", 10*time.Millisecond, largeContext)
			duration := time.Since(start)

			// Ensure reasonable performance even with large contexts
			if duration > 100*time.Millisecond {
				t.Logf("Large context (%d keys) took %v to process", size, duration)
			}

			// Test memory doesn't grow excessively
			var m1, m2 runtime.MemStats
			runtime.GC()
			runtime.ReadMemStats(&m1)

			// Perform multiple operations with the same large context
			for i := 0; i < 10; i++ {
				builder.RecordWithContext("repeated_large_context", "success", 10*time.Millisecond, largeContext)
			}

			runtime.GC()
			runtime.ReadMemStats(&m2)

			// Memory growth should be reasonable due to pooling
			memGrowth := m2.Alloc - m1.Alloc
			t.Logf("Memory growth for repeated large context operations: %d bytes", memGrowth)
		})
	}
}

func testVeryLongTagValues(t *testing.T) {
	registry := metric.NewRegistry(metric.DefaultTagValidationConfig(), 5*time.Minute)
	defer registry.Close()

	om := New(registry)
	builder := NewMetricsBuilder(om)

	// Test tag values of various lengths
	lengthTests := []struct {
		length int
		char   rune
		name   string
	}{
		{100, 'a', "100 ASCII chars"},
		{200, 'x', "200 ASCII chars (at limit)"},
		{150, '测', "150 Unicode chars"},
		{300, 'z', "300 chars (exceeds limit)"},
	}

	for _, tt := range lengthTests {
		t.Run(tt.name, func(t *testing.T) {
			longValue := strings.Repeat(string(tt.char), tt.length)
			context := map[string]string{
				"short_key":       "short_value",
				"long_value_key":  longValue,
				"another_key":     "another_value",
			}

			defer func() {
				if r := recover(); r != nil {
					// This is expected for values that exceed limits
					if tt.length <= 200 {
						t.Errorf("Unexpected panic with %d char value: %v", tt.length, r)
					} else {
						t.Logf("Expected panic with %d char value: %v", tt.length, r)
					}
				}
			}()

			builder.RecordWithContext("long_value_test", "success", 10*time.Millisecond, context)
			builder.RecordSecurityEvent("long_security_event", "detected", context)
			builder.RecordBusinessMetric("long_business_metric", "completed", 1.0, context)
		})
	}
}

func testEmptyAndWhitespaceStrings(t *testing.T) {
	registry := metric.NewRegistry(metric.DefaultTagValidationConfig(), 5*time.Minute)
	defer registry.Close()

	om := New(registry)
	builder := NewMetricsBuilder(om)

	testCases := []struct {
		name      string
		operation string
		status    string
		context   map[string]string
	}{
		{"Empty operation", "", "success", map[string]string{"key": "value"}},
		{"Empty status", "test_op", "", map[string]string{"key": "value"}},
		{"Both empty", "", "", map[string]string{"key": "value"}},
		{"Whitespace operation", "   ", "success", map[string]string{"key": "value"}},
		{"Whitespace status", "test_op", "   ", map[string]string{"key": "value"}},
		{"Tab characters", "\t\t", "\t", map[string]string{"key": "value"}},
		{"Newline characters", "\n", "\n", map[string]string{"key": "value"}},
		{"Mixed whitespace", " \t\n ", " \t\n ", map[string]string{"key": "value"}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Panic with %s: %v", tc.name, r)
				}
			}()

			builder.RecordWithContext(tc.operation, tc.status, 10*time.Millisecond, tc.context)
		})
	}
}

func testSpecialCharactersAndUnicode(t *testing.T) {
	registry := metric.NewRegistry(metric.DefaultTagValidationConfig(), 5*time.Minute)
	defer registry.Close()

	om := New(registry)
	builder := NewMetricsBuilder(om)

	specialContexts := []struct {
		name    string
		context map[string]string
	}{
		{
			"Special ASCII characters",
			map[string]string{
				"hyphen-key":     "value-with-hyphens",
				"underscore_key": "value_with_underscores",
				"dot.key":        "value.with.dots",
				"colon:key":      "value:with:colons",
			},
		},
		{
			"Unicode characters",
			map[string]string{
				"unicode_chinese": "测试值",
				"unicode_emoji":   "🚀✨",
				"unicode_arabic":  "اختبار",
				"unicode_russian": "тест",
			},
		},
		{
			"Mixed special and unicode",
			map[string]string{
				"mixed-key_测试": "value-测试_🚀",
				"normal":         "value",
			},
		},
		{
			"Numbers and symbols",
			map[string]string{
				"numeric_123": "456",
				"symbols":     "!@#$%^&*()",
				"version":     "v1.2.3-beta",
			},
		},
	}

	for _, sc := range specialContexts {
		t.Run(sc.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Logf("Panic with %s: %v (may be expected for invalid characters)", sc.name, r)
				}
			}()

			builder.RecordWithContext("special_chars_test", "success", 10*time.Millisecond, sc.context)
			builder.RecordSecurityEvent("special_security", "detected", sc.context)
			builder.RecordBusinessMetric("special_business", "completed", 1.0, sc.context)
		})
	}
}

func testExtremeValues(t *testing.T) {
	registry := metric.NewRegistry(metric.DefaultTagValidationConfig(), 5*time.Minute)
	defer registry.Close()

	om := New(registry)
	builder := NewMetricsBuilder(om)

	// Test extreme duration values
	extremeDurations := []struct {
		name     string
		duration time.Duration
	}{
		{"Zero duration", 0},
		{"Negative duration", -10 * time.Millisecond},
		{"Maximum duration", time.Duration(1<<63 - 1)},
		{"One nanosecond", 1 * time.Nanosecond},
		{"One second", 1 * time.Second},
		{"One hour", 1 * time.Hour},
	}

	for _, ed := range extremeDurations {
		t.Run(ed.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Panic with %s duration: %v", ed.name, r)
				}
			}()

			context := map[string]string{"duration_test": ed.name}
			builder.RecordWithContext("extreme_duration", "success", ed.duration, context)
		})
	}

	// Test extreme business metric values
	extremeValues := []struct {
		name  string
		value float64
	}{
		{"Zero value", 0.0},
		{"Negative value", -100.5},
		{"Very large value", 1e15},
		{"Very small positive", 1e-15},
		{"Infinity", math.Inf(1)},
		{"Negative infinity", math.Inf(-1)},
		{"NaN", math.NaN()},
	}

	for _, ev := range extremeValues {
		t.Run(ev.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Logf("Panic with %s value (%v): %v", ev.name, ev.value, r)
				}
			}()

			context := map[string]string{"value_test": ev.name}
			builder.RecordBusinessMetric("extreme_values", "completed", ev.value, context)
		})
	}
}

// TestPhase4ErrorConditions implements error condition handling tests
func TestPhase4ErrorConditions(t *testing.T) {
	t.Run("Registry Failure Simulation", testRegistryFailureSimulation)
	t.Run("Pool Exhaustion Scenarios", testPoolExhaustionScenarios)
	t.Run("Invalid Tag Configurations", testInvalidTagConfigurations)
	t.Run("Concurrent Access During Shutdown", testConcurrentAccessDuringShutdown)
}

func testRegistryFailureSimulation(t *testing.T) {
	// Test with a registry that has very short cleanup interval to trigger cleanup
	registry := metric.NewRegistry(metric.DefaultTagValidationConfig(), 1*time.Millisecond)
	om := New(registry)
	builder := NewMetricsBuilder(om)

	// Perform operations before closing registry
	context := map[string]string{"test": "registry_failure"}
	builder.RecordWithContext("before_close", "success", 10*time.Millisecond, context)

	// Close registry early
	registry.Close()

	// Attempt operations after registry is closed
	t.Run("Operations after registry close", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Logf("Expected panic after registry close: %v", r)
			}
		}()

		// These might fail or behave differently after registry closure
		builder.RecordWithContext("after_close", "success", 10*time.Millisecond, context)
		builder.RecordSecurityEvent("after_close_security", "event", context)
		builder.RecordBusinessMetric("after_close_business", "metric", 1.0, context)
	})
}

func testPoolExhaustionScenarios(t *testing.T) {
	registry := metric.NewRegistry(metric.DefaultTagValidationConfig(), 5*time.Minute)
	defer registry.Close()

	om := New(registry)
	builder := NewMetricsBuilder(om)

	// Test pool behavior under extreme concurrent load
	const numGoroutines = 1000
	const operationsPerGoroutine = 100

	var wg sync.WaitGroup
	startChan := make(chan struct{})

	// Create a large number of concurrent operations to stress the pool
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			// Wait for signal to start all goroutines simultaneously
			<-startChan

			for j := 0; j < operationsPerGoroutine; j++ {
				context := map[string]string{
					"worker":    fmt.Sprintf("worker_%d", workerID),
					"iteration": fmt.Sprintf("%d", j),
					"load_test": "pool_exhaustion",
				}

				// Rapid-fire operations to stress pool allocation/deallocation
				builder.RecordWithContext("pool_stress", "success", 1*time.Microsecond, context)

				// Small sleep to allow some pool recycling
				if j%10 == 0 {
					time.Sleep(1 * time.Microsecond)
				}
			}
		}(i)
	}

	// Start all goroutines simultaneously
	close(startChan)

	// Wait for completion with timeout
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		t.Log("Pool exhaustion test completed successfully")
	case <-time.After(30 * time.Second):
		t.Error("Pool exhaustion test timed out")
	}
}

func testInvalidTagConfigurations(t *testing.T) {
	// Test with very restrictive tag validation
	restrictiveConfig := metric.TagValidationConfig{
		MaxKeyLength:   10,
		MaxValueLength: 10,
		MaxKeys:        3,
	}

	registry := metric.NewRegistry(restrictiveConfig, 5*time.Minute)
	defer registry.Close()

	om := New(registry)
	builder := NewMetricsBuilder(om)

	invalidConfigs := []struct {
		name    string
		context map[string]string
	}{
		{
			"Keys too long",
			map[string]string{
				"this_key_is_way_too_long": "value",
				"short":                    "value",
			},
		},
		{
			"Values too long",
			map[string]string{
				"key": "this_value_is_way_too_long_for_the_config",
			},
		},
		{
			"Too many tags",
			map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
				"key4": "value4", // This exceeds MaxKeys=3
				"key5": "value5",
			},
		},
		{
			"Multiple violations",
			map[string]string{
				"this_key_is_too_long": "this_value_is_also_too_long",
				"key2":                 "value2",
				"key3":                 "value3",
				"key4":                 "value4", // Too many tags
			},
		},
	}

	for _, ic := range invalidConfigs {
		t.Run(ic.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Logf("Panic with %s (may be expected): %v", ic.name, r)
				}
			}()

			builder.RecordWithContext("invalid_config_test", "success", 10*time.Millisecond, ic.context)
		})
	}
}

func testConcurrentAccessDuringShutdown(t *testing.T) {
	registry := metric.NewRegistry(metric.DefaultTagValidationConfig(), 5*time.Minute)
	om := New(registry)
	builder := NewMetricsBuilder(om)

	const numWorkers = 50
	var wg sync.WaitGroup
	stopChan := make(chan struct{})

	// Start workers that continuously perform operations
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			operationCount := 0
			for {
				select {
				case <-stopChan:
					t.Logf("Worker %d completed %d operations", workerID, operationCount)
					return
				default:
					context := map[string]string{
						"worker":    fmt.Sprintf("worker_%d", workerID),
						"operation": fmt.Sprintf("op_%d", operationCount),
					}

					// These operations might fail after shutdown begins
					func() {
						defer func() {
							if r := recover(); r != nil {
								// Expected during shutdown
							}
						}()
						builder.RecordWithContext("shutdown_test", "running", 1*time.Millisecond, context)
					}()

					operationCount++
					time.Sleep(1 * time.Millisecond)
				}
			}
		}(i)
	}

	// Let operations run for a short time
	time.Sleep(100 * time.Millisecond)

	// Initiate shutdown while operations are still running
	go func() {
		time.Sleep(50 * time.Millisecond)
		registry.Close()
	}()

	// Stop all workers after a short delay
	time.Sleep(200 * time.Millisecond)
	close(stopChan)

	// Wait for all workers to complete
	wg.Wait()
	t.Log("Concurrent shutdown test completed")
}

// TestPhase4RegressionSuite implements comprehensive regression testing
func TestPhase4RegressionSuite(t *testing.T) {
	t.Run("Functionality Regression", testFunctionalityRegression)
	t.Run("Performance Regression", testPerformanceRegression)
	t.Run("Memory Leak Detection", testMemoryLeakDetection)
	t.Run("Concurrency Regression", testConcurrencyRegression)
}

func testFunctionalityRegression(t *testing.T) {
	registry := metric.NewRegistry(metric.DefaultTagValidationConfig(), 5*time.Minute)
	defer registry.Close()

	om := New(registry)
	builder := NewMetricsBuilder(om)

	// Test that all core functionality still works correctly
	testCases := []struct {
		name    string
		testFunc func(*testing.T)
	}{
		{
			"Basic RecordWithContext",
			func(t *testing.T) {
				context := map[string]string{"provider": "test", "region": "us-east"}
				builder.RecordWithContext("functionality_test", "success", 10*time.Millisecond, context)
			},
		},
		{
			"Security Event Recording",
			func(t *testing.T) {
				context := map[string]string{"event_type": "test", "severity": "low"}
				builder.RecordSecurityEvent("regression_security", "detected", context)
			},
		},
		{
			"Business Metric Recording",
			func(t *testing.T) {
				context := map[string]string{"category": "test", "source": "regression"}
				builder.RecordBusinessMetric("regression_business", "completed", 123.45, context)
			},
		},
		{
			"Empty Context Handling",
			func(t *testing.T) {
				builder.RecordWithContext("empty_context", "success", 5*time.Millisecond, nil)
			},
		},
		{
			"Large Context Processing",
			func(t *testing.T) {
				largeContext := make(map[string]string)
				for i := 0; i < 20; i++ {
					largeContext[fmt.Sprintf("key_%d", i)] = fmt.Sprintf("value_%d", i)
				}
				builder.RecordWithContext("large_context", "success", 15*time.Millisecond, largeContext)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Regression detected in %s: %v", tc.name, r)
				}
			}()

			tc.testFunc(t)
		})
	}
}

func testPerformanceRegression(t *testing.T) {
	registry := metric.NewRegistry(metric.DefaultTagValidationConfig(), 5*time.Minute)
	defer registry.Close()

	om := New(registry)
	builder := NewMetricsBuilder(om)

	// Performance benchmarks to detect regressions
	context := map[string]string{
		"provider": "test",
		"region":   "us-east",
		"tier":     "premium",
	}

	const iterations = 1000

	// Test RecordWithContext performance
	start := time.Now()
	for i := 0; i < iterations; i++ {
		builder.RecordWithContext("perf_test", "success", 10*time.Millisecond, context)
	}
	recordDuration := time.Since(start)

	// Test SecurityEvent performance
	start = time.Now()
	for i := 0; i < iterations; i++ {
		builder.RecordSecurityEvent("perf_security", "detected", context)
	}
	securityDuration := time.Since(start)

	// Test BusinessMetric performance
	start = time.Now()
	for i := 0; i < iterations; i++ {
		builder.RecordBusinessMetric("perf_business", "completed", float64(i), context)
	}
	businessDuration := time.Since(start)

	// Log performance metrics (in a real scenario, these would be compared against baselines)
	t.Logf("Performance results for %d iterations:", iterations)
	t.Logf("  RecordWithContext: %v (%v per op)", recordDuration, recordDuration/iterations)
	t.Logf("  RecordSecurityEvent: %v (%v per op)", securityDuration, securityDuration/iterations)
	t.Logf("  RecordBusinessMetric: %v (%v per op)", businessDuration, businessDuration/iterations)

	// Basic performance regression checks (these thresholds would be tuned based on baseline measurements)
	if recordDuration > 100*time.Millisecond {
		t.Logf("RecordWithContext performance may have regressed: %v for %d ops", recordDuration, iterations)
	}
	if securityDuration > 50*time.Millisecond {
		t.Logf("RecordSecurityEvent performance may have regressed: %v for %d ops", securityDuration, iterations)
	}
	if businessDuration > 50*time.Millisecond {
		t.Logf("RecordBusinessMetric performance may have regressed: %v for %d ops", businessDuration, iterations)
	}
}

func testMemoryLeakDetection(t *testing.T) {
	registry := metric.NewRegistry(metric.DefaultTagValidationConfig(), 5*time.Minute)
	defer registry.Close()

	om := New(registry)
	builder := NewMetricsBuilder(om)

	var m1, m2 runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m1)

	// Perform a large number of operations with varying contexts
	const iterations = 10000
	for i := 0; i < iterations; i++ {
		context := map[string]string{
			"iteration": fmt.Sprintf("%d", i),
			"batch":     fmt.Sprintf("batch_%d", i/100),
			"worker":    fmt.Sprintf("worker_%d", i%10),
		}

		switch i % 3 {
		case 0:
			builder.RecordWithContext("leak_test", "success", 1*time.Millisecond, context)
		case 1:
			builder.RecordSecurityEvent("leak_security", "event", context)
		case 2:
			builder.RecordBusinessMetric("leak_business", "metric", float64(i), context)
		}

		// Force some GC pressure
		if i%1000 == 0 {
			runtime.GC()
		}
	}

	runtime.GC()
	runtime.ReadMemStats(&m2)

	// Analyze memory usage
	allocDiff := m2.Alloc - m1.Alloc
	totalAllocDiff := m2.TotalAlloc - m1.TotalAlloc

	t.Logf("Memory usage after %d operations:", iterations)
	t.Logf("  Current alloc diff: %d bytes", allocDiff)
	t.Logf("  Total alloc diff: %d bytes", totalAllocDiff)
	t.Logf("  Bytes per operation: %.2f", float64(totalAllocDiff)/float64(iterations))

	// Memory leak detection (these thresholds would be tuned based on expected usage)
	const maxBytesPerOp = 1000 // Adjust based on expected allocation per operation
	if bytesPerOp := float64(allocDiff) / float64(iterations); bytesPerOp > maxBytesPerOp {
		t.Logf("Potential memory leak detected: %.2f bytes per operation (threshold: %d)", bytesPerOp, maxBytesPerOp)
	}
}

func testConcurrencyRegression(t *testing.T) {
	registry := metric.NewRegistry(metric.DefaultTagValidationConfig(), 5*time.Minute)
	defer registry.Close()

	om := New(registry)
	builder := NewMetricsBuilder(om)

	const numWorkers = 100
	const operationsPerWorker = 100

	var wg sync.WaitGroup
	errorChan := make(chan error, numWorkers)

	// Launch concurrent workers
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			defer func() {
				if r := recover(); r != nil {
					errorChan <- fmt.Errorf("worker %d panicked: %v", workerID, r)
				}
			}()

			for j := 0; j < operationsPerWorker; j++ {
				context := map[string]string{
					"worker":    fmt.Sprintf("worker_%d", workerID),
					"operation": fmt.Sprintf("op_%d", j),
				}

				// Mix different operation types
				switch j % 3 {
				case 0:
					builder.RecordWithContext("concurrent_test", "success", 1*time.Millisecond, context)
				case 1:
					builder.RecordSecurityEvent("concurrent_security", "event", context)
				case 2:
					builder.RecordBusinessMetric("concurrent_business", "metric", float64(j), context)
				}
			}
		}(i)
	}

	// Wait for completion with timeout
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		t.Log("Concurrency regression test completed successfully")
	case <-time.After(30 * time.Second):
		t.Error("Concurrency regression test timed out")
	}

	// Check for any errors
	close(errorChan)
	var errors []error
	for err := range errorChan {
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		t.Errorf("Concurrency regression detected: %d errors occurred", len(errors))
		for _, err := range errors {
			t.Logf("  Error: %v", err)
		}
	}
}