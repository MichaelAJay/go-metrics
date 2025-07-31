package operational_test

import (
	"fmt"
	"time"

	"github.com/MichaelAJay/go-metrics/metric"
	"github.com/MichaelAJay/go-metrics/operational"
)

// ExampleOperationalMetrics demonstrates basic usage of operational metrics
func ExampleOperationalMetrics() {
	// Create a metrics registry
	registry := metric.NewDefaultRegistry()
	
	// Create operational metrics instance
	om := operational.New(registry)
	
	// Record successful operations
	om.RecordOperation("GenerateNonce", "success", 50*time.Millisecond)
	om.RecordOperation("ValidateRequest", "success", 100*time.Millisecond)
	
	// Record errors
	om.RecordError("GenerateNonce", "crypto_error", "random_generation")
	om.RecordError("ValidateRequest", "validation_error", "invalid_format")
	
	// The metrics are now available in the registry and can be reported
	// by any metrics reporter (Prometheus, OpenTelemetry, etc.)
	
	fmt.Println("Operational metrics recorded successfully")
	// Output: Operational metrics recorded successfully
}

// ExampleMockOperationalMetrics demonstrates testing with mocks
func ExampleMockOperationalMetrics() {
	// Create a mock for testing
	mock := operational.NewMockOperationalMetrics()
	
	// Use mock in your code under test
	simulateService(mock)
	
	// Verify the expected calls were made
	errorCount := mock.GetErrorCallCount("ProcessData", "validation_error", "missing_field")
	successCount := mock.GetOperationCallCount("ProcessData", "success")
	
	fmt.Printf("Errors recorded: %d\n", errorCount)
	fmt.Printf("Successful operations: %d\n", successCount)
	
	// Check average duration
	avgDuration := mock.GetAverageDuration("ProcessData", "success")
	fmt.Printf("Average processing time: %v\n", avgDuration)
	
	// Output:
	// Errors recorded: 1
	// Successful operations: 2
	// Average processing time: 75ms
}

// simulateService simulates a service that uses operational metrics
func simulateService(om operational.OperationalMetrics) {
	// Simulate some successful operations
	om.RecordOperation("ProcessData", "success", 50*time.Millisecond)
	om.RecordOperation("ProcessData", "success", 100*time.Millisecond)
	
	// Simulate an error
	om.RecordError("ProcessData", "validation_error", "missing_field")
}

// ExampleOperationalMetrics_patterns demonstrates common usage patterns
func ExampleOperationalMetrics_patterns() {
	registry := metric.NewDefaultRegistry()
	om := operational.New(registry)
	
	// Pattern 1: Timing operations with defer
	func() {
		start := time.Now()
		defer func() {
			om.RecordOperation("DatabaseQuery", "success", time.Since(start))
		}()
		
		// Simulate database work
		time.Sleep(10 * time.Millisecond)
	}()
	
	// Pattern 2: Error handling with context
	func() {
		start := time.Now()
		defer func() {
			if r := recover(); r != nil {
				om.RecordError("CriticalOperation", "panic", "unexpected_panic")
				om.RecordOperation("CriticalOperation", "error", time.Since(start))
			}
		}()
		
		// Simulate operation that might panic
		// In real code, this would be actual business logic
		om.RecordOperation("CriticalOperation", "success", time.Since(start))
	}()
	
	// Pattern 3: Categorizing errors by source
	func() {
		// Network-related errors
		om.RecordError("APICall", "network_error", "timeout")
		om.RecordError("APICall", "network_error", "connection_refused")
		
		// Validation errors
		om.RecordError("APICall", "validation_error", "invalid_json")
		om.RecordError("APICall", "validation_error", "missing_required_field")
		
		// Business logic errors
		om.RecordError("APICall", "business_error", "insufficient_permissions")
	}()
	
	fmt.Println("Common patterns demonstrated")
	// Output: Common patterns demonstrated
}