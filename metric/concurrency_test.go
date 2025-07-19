package metric

import (
	"runtime"
	"sync"
	"testing"
	"time"
)

// TestCounterConcurrency tests that counter operations are thread-safe
func TestCounterConcurrency(t *testing.T) {
	registry := NewRegistry()
	counter := registry.Counter(Options{
		Name:        "concurrent_counter",
		Description: "Test concurrent counter operations",
	})

	const numGoroutines = 100
	const incrementsPerGoroutine = 1000
	const expectedTotal = numGoroutines * incrementsPerGoroutine

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// Start multiple goroutines that increment the counter
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < incrementsPerGoroutine; j++ {
				counter.Inc()
			}
		}()
	}

	// Wait for all goroutines to complete
	wg.Wait()

	// Verify the final value
	finalValue := counter.Value()
	if finalValue != expectedTotal {
		t.Errorf("Expected counter value %d, got %d", expectedTotal, finalValue)
	}
}

// TestGaugeConcurrency tests that gauge operations are thread-safe
func TestGaugeConcurrency(t *testing.T) {
	registry := NewRegistry()
	gauge := registry.Gauge(Options{
		Name:        "concurrent_gauge",
		Description: "Test concurrent gauge operations",
	})

	const numGoroutines = 50
	const operationsPerGoroutine = 100

	var wg sync.WaitGroup
	wg.Add(numGoroutines * 2) // Both increment and decrement goroutines

	// Start goroutines that increment
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < operationsPerGoroutine; j++ {
				gauge.Inc()
			}
		}()
	}

	// Start goroutines that decrement
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < operationsPerGoroutine; j++ {
				gauge.Dec()
			}
		}()
	}

	// Wait for all goroutines to complete
	wg.Wait()

	// The final value should be 0 since we have equal increments and decrements
	finalValue := gauge.Value()
	if finalValue != 0 {
		t.Errorf("Expected gauge value 0, got %d", finalValue)
	}
}

// TestHistogramConcurrency tests that histogram operations are thread-safe
// This is the most critical test since it verifies our race condition fix
func TestHistogramConcurrency(t *testing.T) {
	registry := NewRegistry()
	histogram := registry.Histogram(Options{
		Name:        "concurrent_histogram",
		Description: "Test concurrent histogram operations",
	})

	const numGoroutines = 100
	const observationsPerGoroutine = 100
	const expectedCount = numGoroutines * observationsPerGoroutine

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// Start multiple goroutines that observe values
	for i := 0; i < numGoroutines; i++ {
		go func(goroutineID int) {
			defer wg.Done()
			for j := 0; j < observationsPerGoroutine; j++ {
				// Use different values to test min/max updates
				// Start from 1 to ensure we have a clear expected range
				value := float64(goroutineID*observationsPerGoroutine + j + 1)
				histogram.Observe(value)
			}
		}(i)
	}

	// Wait for all goroutines to complete
	wg.Wait()

	// Verify the histogram state
	snapshot := histogram.Snapshot()
	
	if snapshot.Count != expectedCount {
		t.Errorf("Expected histogram count %d, got %d", expectedCount, snapshot.Count)
	}

	// Verify min and max values are reasonable
	expectedMin := uint64(1) // First value from goroutine 0 (now starts at 1)
	expectedMax := uint64(numGoroutines*observationsPerGoroutine) // Last value from last goroutine

	if snapshot.Min != expectedMin {
		t.Errorf("Expected histogram min %d, got %d", expectedMin, snapshot.Min)
	}

	if snapshot.Max != expectedMax {
		t.Errorf("Expected histogram max %d, got %d", expectedMax, snapshot.Max)
	}

	// Verify sum is correct (sum of 1 to n = n*(n+1)/2)
	n := numGoroutines * observationsPerGoroutine
	expectedSum := uint64(n * (n + 1) / 2)
	if snapshot.Sum != expectedSum {
		t.Errorf("Expected histogram sum %d, got %d", expectedSum, snapshot.Sum)
	}
}

// TestTimerConcurrency tests that timer operations are thread-safe
func TestTimerConcurrency(t *testing.T) {
	registry := NewRegistry()
	timer := registry.Timer(Options{
		Name:        "concurrent_timer",
		Description: "Test concurrent timer operations",
	})

	const numGoroutines = 50
	const recordsPerGoroutine = 20
	const expectedCount = numGoroutines * recordsPerGoroutine

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// Start multiple goroutines that record durations
	for i := 0; i < numGoroutines; i++ {
		go func(goroutineID int) {
			defer wg.Done()
			for j := 0; j < recordsPerGoroutine; j++ {
				// Record various durations
				duration := time.Duration(goroutineID*recordsPerGoroutine+j) * time.Microsecond
				timer.Record(duration)
			}
		}(i)
	}

	// Wait for all goroutines to complete
	wg.Wait()

	// Verify the timer state
	snapshot := timer.Snapshot()
	
	if snapshot.Count != expectedCount {
		t.Errorf("Expected timer count %d, got %d", expectedCount, snapshot.Count)
	}

	// Verify that some data was recorded
	if snapshot.Sum == 0 {
		t.Error("Expected timer sum to be greater than 0")
	}
}

// TestRaceDetection runs a stress test to catch race conditions
// This test is designed to run with -race flag: go test -race
func TestRaceDetection(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping race detection test in short mode")
	}

	registry := NewRegistry()
	
	// Create metrics
	counter := registry.Counter(Options{Name: "race_counter"})
	gauge := registry.Gauge(Options{Name: "race_gauge"})
	histogram := registry.Histogram(Options{Name: "race_histogram"})
	timer := registry.Timer(Options{Name: "race_timer"})

	const duration = 100 * time.Millisecond
	numWorkers := runtime.NumCPU() * 2

	// Channel to signal workers to stop
	stop := make(chan struct{})

	var wg sync.WaitGroup
	wg.Add(numWorkers * 4) // 4 metric types

	// Counter workers
	for i := 0; i < numWorkers; i++ {
		go func() {
			defer wg.Done()
			for {
				select {
				case <-stop:
					return
				default:
					counter.Inc()
					counter.Add(2.5)
					_ = counter.Value() // Read the value
				}
			}
		}()
	}

	// Gauge workers
	for i := 0; i < numWorkers; i++ {
		go func() {
			defer wg.Done()
			for {
				select {
				case <-stop:
					return
				default:
					gauge.Set(42)
					gauge.Inc()
					gauge.Dec()
					gauge.Add(-1.5)
					_ = gauge.Value() // Read the value
				}
			}
		}()
	}

	// Histogram workers
	for i := 0; i < numWorkers; i++ {
		go func() {
			defer wg.Done()
			for {
				select {
				case <-stop:
					return
				default:
					histogram.Observe(float64(time.Now().UnixNano() % 1000))
					_ = histogram.Snapshot() // Read the snapshot
				}
			}
		}()
	}

	// Timer workers
	for i := 0; i < numWorkers; i++ {
		go func() {
			defer wg.Done()
			for {
				select {
				case <-stop:
					return
				default:
					timer.Record(time.Microsecond * time.Duration(time.Now().UnixNano()%1000))
					timer.RecordSince(time.Now().Add(-time.Millisecond))
					_ = timer.Snapshot() // Read the snapshot
				}
			}
		}()
	}

	// Let the workers run for the specified duration
	time.Sleep(duration)
	close(stop)
	wg.Wait()

	// If we reach here without race conditions, the test passes
	t.Log("Race detection test completed successfully")
}

// BenchmarkHistogramConcurrent benchmarks concurrent histogram operations
func BenchmarkHistogramConcurrent(b *testing.B) {
	registry := NewRegistry()
	histogram := registry.Histogram(Options{
		Name:        "benchmark_histogram",
		Description: "Benchmark concurrent histogram operations",
	})

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			histogram.Observe(float64(time.Now().UnixNano() % 1000))
		}
	})
}

// BenchmarkCounterConcurrent benchmarks concurrent counter operations
func BenchmarkCounterConcurrent(b *testing.B) {
	registry := NewRegistry()
	counter := registry.Counter(Options{
		Name:        "benchmark_counter",
		Description: "Benchmark concurrent counter operations",
	})

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			counter.Inc()
		}
	})
}