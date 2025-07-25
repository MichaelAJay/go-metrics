package operational

import (
	"sync"
	"time"
)

// MockOperationalMetrics is a mock implementation of OperationalMetrics for testing
type MockOperationalMetrics struct {
	// Call tracking
	ErrorCalls     []ErrorCall
	OperationCalls []OperationCall
	
	// Mutex for thread-safe access
	mu sync.Mutex
}

// ErrorCall represents a call to RecordError
type ErrorCall struct {
	Operation     string
	ErrorType     string
	ErrorCategory string
	Timestamp     time.Time
}

// OperationCall represents a call to RecordOperation
type OperationCall struct {
	Operation string
	Status    string
	Duration  time.Duration
	Timestamp time.Time
}

// NewMockOperationalMetrics creates a new mock implementation
func NewMockOperationalMetrics() *MockOperationalMetrics {
	return &MockOperationalMetrics{
		ErrorCalls:     make([]ErrorCall, 0),
		OperationCalls: make([]OperationCall, 0),
	}
}

// RecordError implements the OperationalMetrics interface
func (m *MockOperationalMetrics) RecordError(operation, errorType, errorCategory string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.ErrorCalls = append(m.ErrorCalls, ErrorCall{
		Operation:     operation,
		ErrorType:     errorType,
		ErrorCategory: errorCategory,
		Timestamp:     time.Now(),
	})
}

// RecordOperation implements the OperationalMetrics interface
func (m *MockOperationalMetrics) RecordOperation(operation, status string, duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.OperationCalls = append(m.OperationCalls, OperationCall{
		Operation: operation,
		Status:    status,
		Duration:  duration,
		Timestamp: time.Now(),
	})
}

// GetErrorCallCount returns the number of error calls for a specific operation/type/category
func (m *MockOperationalMetrics) GetErrorCallCount(operation, errorType, errorCategory string) int {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	count := 0
	for _, call := range m.ErrorCalls {
		if call.Operation == operation && call.ErrorType == errorType && call.ErrorCategory == errorCategory {
			count++
		}
	}
	return count
}

// GetOperationCallCount returns the number of operation calls for a specific operation/status
func (m *MockOperationalMetrics) GetOperationCallCount(operation, status string) int {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	count := 0
	for _, call := range m.OperationCalls {
		if call.Operation == operation && call.Status == status {
			count++
		}
	}
	return count
}

// GetTotalErrorCalls returns the total number of error calls
func (m *MockOperationalMetrics) GetTotalErrorCalls() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	return len(m.ErrorCalls)
}

// GetTotalOperationCalls returns the total number of operation calls
func (m *MockOperationalMetrics) GetTotalOperationCalls() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	return len(m.OperationCalls)
}

// GetErrorCallsForOperation returns all error calls for a specific operation
func (m *MockOperationalMetrics) GetErrorCallsForOperation(operation string) []ErrorCall {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	var calls []ErrorCall
	for _, call := range m.ErrorCalls {
		if call.Operation == operation {
			calls = append(calls, call)
		}
	}
	return calls
}

// GetOperationCallsForOperation returns all operation calls for a specific operation
func (m *MockOperationalMetrics) GetOperationCallsForOperation(operation string) []OperationCall {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	var calls []OperationCall
	for _, call := range m.OperationCalls {
		if call.Operation == operation {
			calls = append(calls, call)
		}
	}
	return calls
}

// Reset clears all recorded calls
func (m *MockOperationalMetrics) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.ErrorCalls = make([]ErrorCall, 0)
	m.OperationCalls = make([]OperationCall, 0)
}

// GetLastErrorCall returns the most recent error call, or nil if none
func (m *MockOperationalMetrics) GetLastErrorCall() *ErrorCall {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if len(m.ErrorCalls) == 0 {
		return nil
	}
	
	call := m.ErrorCalls[len(m.ErrorCalls)-1]
	return &call
}

// GetLastOperationCall returns the most recent operation call, or nil if none
func (m *MockOperationalMetrics) GetLastOperationCall() *OperationCall {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if len(m.OperationCalls) == 0 {
		return nil
	}
	
	call := m.OperationCalls[len(m.OperationCalls)-1]
	return &call
}

// GetAverageDuration returns the average duration for operations with the given operation/status
func (m *MockOperationalMetrics) GetAverageDuration(operation, status string) time.Duration {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	var total time.Duration
	var count int
	
	for _, call := range m.OperationCalls {
		if call.Operation == operation && call.Status == status {
			total += call.Duration
			count++
		}
	}
	
	if count == 0 {
		return 0
	}
	
	return total / time.Duration(count)
}