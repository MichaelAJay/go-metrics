package operational

import (
	"testing"
	"time"

	"github.com/MichaelAJay/go-metrics/metric"
)

func TestNew(t *testing.T) {
	registry := metric.NewDefaultRegistry()
	om := New(registry)
	
	if om == nil {
		t.Fatal("New() returned nil")
	}
	
	// Verify it implements the interface
	var _ OperationalMetrics = om
}

func TestRecordError(t *testing.T) {
	registry := metric.NewDefaultRegistry()
	om := New(registry)
	
	// Record some errors
	om.RecordError("GenerateNonce", "crypto_error", "random_generation")
	om.RecordError("GenerateNonce", "crypto_error", "random_generation")
	om.RecordError("ValidateRequest", "validation_error", "invalid_format")
	
	// Verify counters were created and incremented
	var errorCount uint64
	registry.Each(func(m metric.Metric) {
		if m.Type() == metric.TypeCounter {
			if counter, ok := m.(metric.Counter); ok {
				errorCount += counter.Value()
			}
		}
	})
	
	if errorCount != 3 {
		t.Errorf("Expected 3 total error counts, got %d", errorCount)
	}
}

func TestRecordOperation(t *testing.T) {
	registry := metric.NewDefaultRegistry()
	om := New(registry)
	
	// Record some operations
	om.RecordOperation("GenerateNonce", "success", 100*time.Millisecond)
	om.RecordOperation("GenerateNonce", "success", 150*time.Millisecond)
	om.RecordOperation("GenerateNonce", "error", 50*time.Millisecond)
	om.RecordOperation("ValidateRequest", "success", 200*time.Millisecond)
	
	// Count metrics created
	var counterCount, timerCount int
	var totalOperations uint64
	
	registry.Each(func(m metric.Metric) {
		switch m.Type() {
		case metric.TypeCounter:
			counterCount++
			if counter, ok := m.(metric.Counter); ok {
				totalOperations += counter.Value()
			}
		case metric.TypeTimer:
			timerCount++
		}
	})
	
	// Should have created:
	// - 2 counters (GenerateNonce_total, ValidateRequest_total) with different tags
	// - 2 timers (GenerateNonce, ValidateRequest)
	if counterCount != 2 {
		t.Errorf("Expected 2 counters, got %d", counterCount)
	}
	
	if timerCount != 2 {
		t.Errorf("Expected 2 timers, got %d", timerCount)
	}
	
	if totalOperations != 4 {
		t.Errorf("Expected 4 total operations, got %d", totalOperations)
	}
}

func TestMetricCaching(t *testing.T) {
	registry := metric.NewDefaultRegistry()
	om := New(registry)
	
	// Record the same error multiple times
	for i := 0; i < 10; i++ {
		om.RecordError("TestOp", "test_error", "test_category")
	}
	
	// Should only create one counter
	var counterCount int
	var totalCount uint64
	
	registry.Each(func(m metric.Metric) {
		if m.Type() == metric.TypeCounter {
			counterCount++
			if counter, ok := m.(metric.Counter); ok {
				totalCount += counter.Value()
			}
		}
	})
	
	if counterCount != 1 {
		t.Errorf("Expected 1 counter (cached), got %d", counterCount)
	}
	
	if totalCount != 10 {
		t.Errorf("Expected total count of 10, got %d", totalCount)
	}
}

func TestConcurrentAccess(t *testing.T) {
	registry := metric.NewDefaultRegistry()
	om := New(registry)
	
	// Test concurrent access to ensure thread safety
	const numGoroutines = 50
	const numOperations = 100
	
	// Use a channel to synchronize goroutine completion
	done := make(chan bool, numGoroutines)
	
	// Launch multiple goroutines that record metrics concurrently
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer func() { done <- true }()
			
			for j := 0; j < numOperations; j++ {
				om.RecordError("ConcurrentOp", "test_error", "concurrent_test")
				om.RecordOperation("ConcurrentOp", "success", time.Millisecond)
			}
		}(i)
	}
	
	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}
	
	// Verify all operations were recorded
	var errorCount, operationCount uint64
	
	registry.Each(func(m metric.Metric) {
		if m.Type() == metric.TypeCounter {
			if counter, ok := m.(metric.Counter); ok {
				tags := counter.Tags()
				if errorType, exists := tags["error_type"]; exists && errorType == "test_error" {
					errorCount += counter.Value()
				} else if status, exists := tags["status"]; exists && status == "success" {
					operationCount += counter.Value()
				}
			}
		}
	})
	
	expectedCount := uint64(numGoroutines * numOperations)
	
	if errorCount != expectedCount {
		t.Errorf("Expected %d error counts, got %d", expectedCount, errorCount)
	}
	
	if operationCount != expectedCount {
		t.Errorf("Expected %d operation counts, got %d", expectedCount, operationCount)
	}
}

func TestMetricTags(t *testing.T) {
	registry := metric.NewDefaultRegistry()
	om := New(registry)
	
	// Record operations with different parameters
	om.RecordError("TestOp", "validation_error", "invalid_input")
	om.RecordOperation("TestOp", "success", 100*time.Millisecond)
	
	// Verify tags are set correctly
	registry.Each(func(m metric.Metric) {
		tags := m.Tags()
		
		if m.Type() == metric.TypeCounter {
			// Check that operation tag exists
			if operation, exists := tags["operation"]; !exists || operation != "TestOp" {
				t.Errorf("Expected operation tag 'TestOp', got '%s'", operation)
			}
			
			// Check for error-specific tags
			if errorType, exists := tags["error_type"]; exists {
				if errorType != "validation_error" {
					t.Errorf("Expected error_type 'validation_error', got '%s'", errorType)
				}
				
				if errorCategory, exists := tags["error_category"]; !exists || errorCategory != "invalid_input" {
					t.Errorf("Expected error_category 'invalid_input', got '%s'", errorCategory)
				}
			}
			
			// Check for status tags
			if status, exists := tags["status"]; exists {
				if status != "success" {
					t.Errorf("Expected status 'success', got '%s'", status)
				}
			}
		}
		
		if m.Type() == metric.TypeTimer {
			// Check timer tags
			if operation, exists := tags["operation"]; !exists || operation != "TestOp" {
				t.Errorf("Expected timer operation tag 'TestOp', got '%s'", operation)
			}
		}
	})
}

func BenchmarkRecordError(b *testing.B) {
	registry := metric.NewDefaultRegistry()
	om := New(registry)
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		om.RecordError("BenchOp", "test_error", "benchmark")
	}
}

func BenchmarkRecordOperation(b *testing.B) {
	registry := metric.NewDefaultRegistry()
	om := New(registry)
	
	duration := 100 * time.Millisecond
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		om.RecordOperation("BenchOp", "success", duration)
	}
}

func BenchmarkConcurrentRecordError(b *testing.B) {
	registry := metric.NewDefaultRegistry()
	om := New(registry)
	
	b.ResetTimer()
	
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			om.RecordError("ConcurrentBenchOp", "test_error", "benchmark")
		}
	})
}