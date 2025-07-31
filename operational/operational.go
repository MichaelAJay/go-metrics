// Package operational provides high-level operational metrics functionality
// built on top of the core metrics package. This package offers convenient
// methods for common operational patterns like error tracking and operation timing.
package operational

import (
	"fmt"
	"sync"
	"time"

	"github.com/MichaelAJay/go-metrics/metric"
)

// OperationalMetrics provides high-level operational metrics functionality
type OperationalMetrics interface {
	// RecordError records an error event with categorization
	// operation: the operation that failed (e.g., "GenerateNonce", "ValidateRequest")
	// errorType: the type of error (e.g., "crypto_error", "validation_error")
	// errorCategory: additional categorization (e.g., "random_generation", "timeout")
	RecordError(operation, errorType, errorCategory string)

	// RecordOperation records an operation with its status and duration
	// operation: the operation name (e.g., "GenerateNonce", "ValidateRequest")
	// status: the operation status (e.g., "success", "error", "timeout")
	// duration: how long the operation took
	RecordOperation(operation, status string, duration time.Duration)
}

// operationalMetrics implements the OperationalMetrics interface
type operationalMetrics struct {
	registry metric.Registry
	
	// Cached metric instances for performance
	errorCounters     map[string]metric.Counter
	operationTimers   map[string]metric.Timer
	operationCounters map[string]metric.Counter
	
	// Mutex for thread-safe metric caching
	mu sync.RWMutex
}

// New creates a new OperationalMetrics instance
func New(registry metric.Registry) OperationalMetrics {
	return &operationalMetrics{
		registry:          registry,
		errorCounters:     make(map[string]metric.Counter),
		operationTimers:   make(map[string]metric.Timer),
		operationCounters: make(map[string]metric.Counter),
	}
}

// RecordError implements the OperationalMetrics interface
func (om *operationalMetrics) RecordError(operation, errorType, errorCategory string) {
	// Create error counter with tags for categorization
	counter := om.getOrCreateErrorCounter(operation, errorType, errorCategory)
	counter.Inc()
}

// RecordOperation implements the OperationalMetrics interface
func (om *operationalMetrics) RecordOperation(operation, status string, duration time.Duration) {
	// Record timing information
	timer := om.getOrCreateOperationTimer(operation)
	timer.Record(duration)
	
	// Record operation count with status
	counter := om.getOrCreateOperationCounter(operation, status)
	counter.Inc()
}

// getOrCreateErrorCounter creates or retrieves a cached error counter
func (om *operationalMetrics) getOrCreateErrorCounter(operation, errorType, errorCategory string) metric.Counter {
	// Create a unique key for this error counter
	key := fmt.Sprintf("error:%s:%s:%s", operation, errorType, errorCategory)
	
	// Try to get from cache first
	om.mu.RLock()
	if counter, exists := om.errorCounters[key]; exists {
		om.mu.RUnlock()
		return counter
	}
	om.mu.RUnlock()
	
	// Create new counter if not in cache
	om.mu.Lock()
	defer om.mu.Unlock()
	
	// Double-check after acquiring write lock
	if counter, exists := om.errorCounters[key]; exists {
		return counter
	}
	
	// Create the counter with appropriate name and tags
	metricName := fmt.Sprintf("%s_errors_total", operation)
	counter := om.registry.Counter(metric.Options{
		Name:        metricName,
		Description: fmt.Sprintf("Total number of errors for %s operation", operation),
		Unit:        "count",
		Tags: metric.Tags{
			"operation":      operation,
			"error_type":     errorType,
			"error_category": errorCategory,
		},
	})
	
	// Cache for future use
	om.errorCounters[key] = counter
	return counter
}

// getOrCreateOperationTimer creates or retrieves a cached operation timer
func (om *operationalMetrics) getOrCreateOperationTimer(operation string) metric.Timer {
	key := fmt.Sprintf("timer:%s", operation)
	
	// Try to get from cache first
	om.mu.RLock()
	if timer, exists := om.operationTimers[key]; exists {
		om.mu.RUnlock()
		return timer
	}
	om.mu.RUnlock()
	
	// Create new timer if not in cache
	om.mu.Lock()
	defer om.mu.Unlock()
	
	// Double-check after acquiring write lock
	if timer, exists := om.operationTimers[key]; exists {
		return timer
	}
	
	// Create the timer
	metricName := fmt.Sprintf("%s_duration", operation)
	timer := om.registry.Timer(metric.Options{
		Name:        metricName,
		Description: fmt.Sprintf("Duration of %s operation", operation),
		Unit:        "nanoseconds",
		Tags: metric.Tags{
			"operation": operation,
		},
	})
	
	// Cache for future use
	om.operationTimers[key] = timer
	return timer
}

// getOrCreateOperationCounter creates or retrieves a cached operation counter
func (om *operationalMetrics) getOrCreateOperationCounter(operation, status string) metric.Counter {
	key := fmt.Sprintf("counter:%s:%s", operation, status)
	
	// Try to get from cache first
	om.mu.RLock()
	if counter, exists := om.operationCounters[key]; exists {
		om.mu.RUnlock()
		return counter
	}
	om.mu.RUnlock()
	
	// Create new counter if not in cache
	om.mu.Lock()
	defer om.mu.Unlock()
	
	// Double-check after acquiring write lock
	if counter, exists := om.operationCounters[key]; exists {
		return counter
	}
	
	// Create the counter
	metricName := fmt.Sprintf("%s_total", operation)
	counter := om.registry.Counter(metric.Options{
		Name:        metricName,
		Description: fmt.Sprintf("Total number of %s operations", operation),
		Unit:        "count",
		Tags: metric.Tags{
			"operation": operation,
			"status":    status,
		},
	})
	
	// Cache for future use
	om.operationCounters[key] = counter
	return counter
}