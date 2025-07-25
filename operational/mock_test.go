package operational

import (
	"testing"
	"time"
)

func TestMockOperationalMetrics(t *testing.T) {
	mock := NewMockOperationalMetrics()
	
	// Verify it implements the interface
	var _ OperationalMetrics = mock
	
	// Test initial state
	if mock.GetTotalErrorCalls() != 0 {
		t.Errorf("Expected 0 initial error calls, got %d", mock.GetTotalErrorCalls())
	}
	
	if mock.GetTotalOperationCalls() != 0 {
		t.Errorf("Expected 0 initial operation calls, got %d", mock.GetTotalOperationCalls())
	}
}

func TestMockRecordError(t *testing.T) {
	mock := NewMockOperationalMetrics()
	
	// Record some errors
	mock.RecordError("TestOp", "validation_error", "invalid_input")
	mock.RecordError("TestOp", "validation_error", "invalid_input")
	mock.RecordError("AnotherOp", "timeout_error", "network_timeout")
	
	// Verify call tracking
	if mock.GetTotalErrorCalls() != 3 {
		t.Errorf("Expected 3 total error calls, got %d", mock.GetTotalErrorCalls())
	}
	
	// Verify specific call counts
	count := mock.GetErrorCallCount("TestOp", "validation_error", "invalid_input")
	if count != 2 {
		t.Errorf("Expected 2 calls for TestOp/validation_error/invalid_input, got %d", count)
	}
	
	count = mock.GetErrorCallCount("AnotherOp", "timeout_error", "network_timeout")
	if count != 1 {
		t.Errorf("Expected 1 call for AnotherOp/timeout_error/network_timeout, got %d", count)
	}
	
	// Verify last call
	lastCall := mock.GetLastErrorCall()
	if lastCall == nil {
		t.Fatal("Expected last error call, got nil")
	}
	
	if lastCall.Operation != "AnotherOp" {
		t.Errorf("Expected last call operation 'AnotherOp', got '%s'", lastCall.Operation)
	}
}

func TestMockRecordOperation(t *testing.T) {
	mock := NewMockOperationalMetrics()
	
	// Record some operations
	mock.RecordOperation("TestOp", "success", 100*time.Millisecond)
	mock.RecordOperation("TestOp", "success", 150*time.Millisecond)
	mock.RecordOperation("TestOp", "error", 50*time.Millisecond)
	mock.RecordOperation("AnotherOp", "success", 200*time.Millisecond)
	
	// Verify call tracking
	if mock.GetTotalOperationCalls() != 4 {
		t.Errorf("Expected 4 total operation calls, got %d", mock.GetTotalOperationCalls())
	}
	
	// Verify specific call counts
	count := mock.GetOperationCallCount("TestOp", "success")
	if count != 2 {
		t.Errorf("Expected 2 calls for TestOp/success, got %d", count)
	}
	
	count = mock.GetOperationCallCount("TestOp", "error")
	if count != 1 {
		t.Errorf("Expected 1 call for TestOp/error, got %d", count)
	}
	
	// Test average duration
	avgDuration := mock.GetAverageDuration("TestOp", "success")
	expectedAvg := (100*time.Millisecond + 150*time.Millisecond) / 2
	if avgDuration != expectedAvg {
		t.Errorf("Expected average duration %v, got %v", expectedAvg, avgDuration)
	}
}

func TestMockGetCallsForOperation(t *testing.T) {
	mock := NewMockOperationalMetrics()
	
	// Record calls for multiple operations
	mock.RecordError("TestOp", "error1", "category1")
	mock.RecordError("AnotherOp", "error2", "category2")
	mock.RecordError("TestOp", "error3", "category3")
	
	mock.RecordOperation("TestOp", "success", 100*time.Millisecond)
	mock.RecordOperation("AnotherOp", "error", 200*time.Millisecond)
	mock.RecordOperation("TestOp", "error", 150*time.Millisecond)
	
	// Test error calls for specific operation
	errorCalls := mock.GetErrorCallsForOperation("TestOp")
	if len(errorCalls) != 2 {
		t.Errorf("Expected 2 error calls for TestOp, got %d", len(errorCalls))
	}
	
	// Test operation calls for specific operation
	opCalls := mock.GetOperationCallsForOperation("TestOp")
	if len(opCalls) != 2 {
		t.Errorf("Expected 2 operation calls for TestOp, got %d", len(opCalls))
	}
}

func TestMockReset(t *testing.T) {
	mock := NewMockOperationalMetrics()
	
	// Record some calls
	mock.RecordError("TestOp", "error", "category")
	mock.RecordOperation("TestOp", "success", 100*time.Millisecond)
	
	// Verify calls exist
	if mock.GetTotalErrorCalls() == 0 || mock.GetTotalOperationCalls() == 0 {
		t.Fatal("Expected calls to be recorded before reset")
	}
	
	// Reset and verify
	mock.Reset()
	
	if mock.GetTotalErrorCalls() != 0 {
		t.Errorf("Expected 0 error calls after reset, got %d", mock.GetTotalErrorCalls())
	}
	
	if mock.GetTotalOperationCalls() != 0 {
		t.Errorf("Expected 0 operation calls after reset, got %d", mock.GetTotalOperationCalls())
	}
	
	if mock.GetLastErrorCall() != nil {
		t.Error("Expected nil last error call after reset")
	}
	
	if mock.GetLastOperationCall() != nil {
		t.Error("Expected nil last operation call after reset")
	}
}

func TestMockConcurrentAccess(t *testing.T) {
	mock := NewMockOperationalMetrics()
	
	const numGoroutines = 50
	const numOperations = 100
	
	// Use a channel to synchronize goroutine completion
	done := make(chan bool, numGoroutines)
	
	// Launch multiple goroutines that record metrics concurrently
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer func() { done <- true }()
			
			for j := 0; j < numOperations; j++ {
				mock.RecordError("ConcurrentOp", "test_error", "concurrent_test")
				mock.RecordOperation("ConcurrentOp", "success", time.Millisecond)
			}
		}(i)
	}
	
	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}
	
	// Verify all operations were recorded
	expectedCount := numGoroutines * numOperations
	
	errorCount := mock.GetErrorCallCount("ConcurrentOp", "test_error", "concurrent_test")
	if errorCount != expectedCount {
		t.Errorf("Expected %d error calls, got %d", expectedCount, errorCount)
	}
	
	operationCount := mock.GetOperationCallCount("ConcurrentOp", "success")
	if operationCount != expectedCount {
		t.Errorf("Expected %d operation calls, got %d", expectedCount, operationCount)
	}
}

func TestMockGetAverageDurationEdgeCases(t *testing.T) {
	mock := NewMockOperationalMetrics()
	
	// Test average duration with no calls
	avgDuration := mock.GetAverageDuration("NonExistentOp", "success")
	if avgDuration != 0 {
		t.Errorf("Expected 0 average duration for non-existent operation, got %v", avgDuration)
	}
	
	// Test with single call
	mock.RecordOperation("SingleOp", "success", 500*time.Millisecond)
	avgDuration = mock.GetAverageDuration("SingleOp", "success")
	if avgDuration != 500*time.Millisecond {
		t.Errorf("Expected 500ms average duration for single call, got %v", avgDuration)
	}
}