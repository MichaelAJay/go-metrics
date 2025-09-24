// Package operational provides high-level operational metrics functionality
// built on top of the core metrics package. This package offers convenient
// methods for common operational patterns like error tracking and operation timing.
package operational

import (
	"fmt"
	"maps"
	"sync"
	"time"

	"github.com/MichaelAJay/go-metrics/metric"
)

// Internal tag map pool for operational metrics to reduce allocations
var operationalTagPool = sync.Pool{
	New: func() any {
		return make(map[string]string, 8)
	},
}

// clearOperationalTags safely clears a tag map and returns it for reuse
func clearOperationalTags(m map[string]string) map[string]string {
	for k := range m {
		delete(m, k)
	}
	return m
}

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
	tags := operationalTagPool.Get().(map[string]string)
	defer operationalTagPool.Put(clearOperationalTags(tags))

	tags["operation"] = operation
	tags["error_type"] = errorType
	tags["error_category"] = errorCategory

	// Create error counter with tags for categorization
	counter := om.getOrCreateErrorCounterWithTags(operation, tags)
	counter.Inc()
}

// RecordOperation implements the OperationalMetrics interface
func (om *operationalMetrics) RecordOperation(operation, status string, duration time.Duration) {
	timerTags := operationalTagPool.Get().(map[string]string)
	defer operationalTagPool.Put(clearOperationalTags(timerTags))

	timerTags["operation"] = operation

	// Record timing information
	timer := om.getOrCreateOperationTimerWithTags(operation, timerTags)
	timer.Record(duration)

	counterTags := operationalTagPool.Get().(map[string]string)
	defer operationalTagPool.Put(clearOperationalTags(counterTags))

	counterTags["operation"] = operation
	counterTags["status"] = status

	// Record operation count with status
	counter := om.getOrCreateOperationCounterWithTags(operation, counterTags)
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

// getOrCreateErrorCounterWithTags creates or retrieves a cached error counter using pooled tags
func (om *operationalMetrics) getOrCreateErrorCounterWithTags(operation string, tags map[string]string) metric.Counter {
	// Create a unique key for this error counter
	key := fmt.Sprintf("error:%s:%s:%s", operation, tags["error_type"], tags["error_category"])

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

	// Create final tags map for the counter
	finalTags := make(metric.Tags)
	maps.Copy(finalTags, tags)

	// Create the counter with appropriate name and tags
	metricName := fmt.Sprintf("%s_errors_total", operation)
	counter := om.registry.Counter(metric.Options{
		Name:        metricName,
		Description: fmt.Sprintf("Total number of errors for %s operation", operation),
		Unit:        "count",
		Tags:        finalTags,
	})

	// Cache for future use
	om.errorCounters[key] = counter
	return counter
}

// getOrCreateOperationTimerWithTags creates or retrieves a cached operation timer using pooled tags
func (om *operationalMetrics) getOrCreateOperationTimerWithTags(operation string, tags map[string]string) metric.Timer {
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

	// Create final tags map for the timer
	finalTags := make(metric.Tags)
	for k, v := range tags {
		finalTags[k] = v
	}

	// Create the timer
	metricName := fmt.Sprintf("%s_duration", operation)
	timer := om.registry.Timer(metric.Options{
		Name:        metricName,
		Description: fmt.Sprintf("Duration of %s operation", operation),
		Unit:        "nanoseconds",
		Tags:        finalTags,
	})

	// Cache for future use
	om.operationTimers[key] = timer
	return timer
}

// getOrCreateOperationCounterWithTags creates or retrieves a cached operation counter using pooled tags
func (om *operationalMetrics) getOrCreateOperationCounterWithTags(operation string, tags map[string]string) metric.Counter {
	key := fmt.Sprintf("counter:%s:%s", operation, tags["status"])

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

	// Create final tags map for the counter
	finalTags := make(metric.Tags)
	maps.Copy(finalTags, tags)

	// Create the counter
	metricName := fmt.Sprintf("%s_total", operation)
	counter := om.registry.Counter(metric.Options{
		Name:        metricName,
		Description: fmt.Sprintf("Total number of %s operations", operation),
		Unit:        "count",
		Tags:        finalTags,
	})

	// Cache for future use
	om.operationCounters[key] = counter
	return counter
}

// MetricsBuilder provides general-purpose operational metric recording
// that can be used by any service to record domain-specific metrics
// while leveraging the pooled tag infrastructure for performance
type MetricsBuilder struct {
	om OperationalMetrics
}

// NewMetricsBuilder creates a new MetricsBuilder instance
func NewMetricsBuilder(om OperationalMetrics) *MetricsBuilder {
	return &MetricsBuilder{
		om: om,
	}
}

// RecordWithContext records an operation with additional contextual information
// operation: the operation name (e.g., "authentication", "payment_processing")
// status: the operation status (e.g., "success", "error", "timeout")
// duration: how long the operation took
// context: additional contextual tags (e.g., map[string]string{"provider": "password", "user_type": "premium"})
func (b *MetricsBuilder) RecordWithContext(operation, status string, duration time.Duration, context map[string]string) {
	// Record the primary operation using the existing pooled implementation
	b.om.RecordOperation(operation, status, duration)

	// If no additional context, we're done
	if len(context) == 0 {
		return
	}

	// Record contextual metrics efficiently using the existing infrastructure
	// We'll create contextual operation metrics for each key-value pair
	for key, value := range context {
		contextualOperation := fmt.Sprintf("%s_%s", operation, key)
		b.om.RecordOperation(contextualOperation, value, duration)
	}
}

// RecordSecurityEvent records a security-related event with contextual information
// eventType: the type of security event (e.g., "brute_force", "credential_stuffing", "login_attempt")
// action: the action taken (e.g., "blocked", "allowed", "flagged")
// context: additional contextual information (e.g., map[string]string{"ip": clientIP, "user_agent": userAgent})
func (b *MetricsBuilder) RecordSecurityEvent(eventType, action string, context map[string]string) {
	operation := fmt.Sprintf("security_%s", eventType)
	// Security events are recorded with zero duration as they are typically point-in-time events
	b.om.RecordOperation(operation, action, 0)

	// Record additional contextual metrics for security analysis
	if len(context) > 0 {
		for key, value := range context {
			contextualOperation := fmt.Sprintf("security_%s_%s", eventType, key)
			b.om.RecordOperation(contextualOperation, value, 0)
		}
	}
}

// RecordBusinessMetric records a business-related metric with contextual information
// metricType: the type of business metric (e.g., "user_conversion", "payment_processing", "session_duration")
// category: the category or status (e.g., "completed", "organic", "premium")
// value: the numeric value associated with the metric (converted to duration for compatibility)
// context: additional contextual information (e.g., map[string]string{"source": "organic", "tier": "premium"})
func (b *MetricsBuilder) RecordBusinessMetric(metricType, category string, value float64, context map[string]string) {
	operation := fmt.Sprintf("business_%s", metricType)
	// Convert float64 value to duration (nanoseconds) for timer compatibility
	duration := time.Duration(value * float64(time.Millisecond))
	b.om.RecordOperation(operation, category, duration)

	// Record additional contextual metrics for business analysis
	if len(context) > 0 {
		for key, contextValue := range context {
			contextualOperation := fmt.Sprintf("business_%s_%s", metricType, key)
			b.om.RecordOperation(contextualOperation, contextValue, duration)
		}
	}
}
