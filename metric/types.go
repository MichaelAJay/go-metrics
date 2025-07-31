// Package metric provides a flexible, high-performance metrics collection library
// that supports multiple backends, tagging, and various metric types.
package metric

import (
	"context"
	"fmt"
	"time"
)

// Type represents the available metric types
type Type string

const (
	// TypeCounter is for monotonically increasing values
	TypeCounter Type = "counter"
	// TypeGauge is for current point-in-time measurements
	TypeGauge Type = "gauge"
	// TypeHistogram is for statistical distributions of values
	TypeHistogram Type = "histogram"
	// TypeTimer is a specialized metric for duration measurements
	TypeTimer Type = "timer"
)

// Tags represents a map of key-value pairs associated with a metric
type Tags map[string]string

// TagValidationConfig contains configuration for tag validation
type TagValidationConfig struct {
	// MaxKeys is the maximum number of tags allowed per metric
	MaxKeys int
	// MaxKeyLength is the maximum length of a tag key
	MaxKeyLength int
	// MaxValueLength is the maximum length of a tag value
	MaxValueLength int
	// MaxCardinality is the maximum number of unique tag combinations per metric name
	MaxCardinality int
	// DisallowedKeys is a list of tag keys that are not allowed
	DisallowedKeys []string
}

// DefaultTagValidationConfig returns a sensible default tag validation configuration
func DefaultTagValidationConfig() TagValidationConfig {
	return TagValidationConfig{
		MaxKeys:         10,
		MaxKeyLength:    100,
		MaxValueLength:  200,
		MaxCardinality:  1000,
		DisallowedKeys:  []string{},
	}
}

// ValidateTags validates tags against the given configuration
func ValidateTags(tags Tags, config TagValidationConfig) error {
	if len(tags) > config.MaxKeys {
		return fmt.Errorf("too many tags: %d exceeds maximum of %d", len(tags), config.MaxKeys)
	}

	for key, value := range tags {
		// Check key length
		if len(key) > config.MaxKeyLength {
			return fmt.Errorf("tag key '%s' exceeds maximum length of %d", key, config.MaxKeyLength)
		}

		// Check value length
		if len(value) > config.MaxValueLength {
			return fmt.Errorf("tag value for key '%s' exceeds maximum length of %d", key, config.MaxValueLength)
		}

		// Check disallowed keys
		for _, disallowed := range config.DisallowedKeys {
			if key == disallowed {
				return fmt.Errorf("tag key '%s' is not allowed", key)
			}
		}

		// Basic validation: keys and values should not be empty
		if key == "" {
			return fmt.Errorf("tag keys cannot be empty")
		}
	}

	return nil
}

// BucketType represents the type of histogram bucket distribution
type BucketType int

const (
	// BucketTypeExponential creates exponentially-sized buckets
	BucketTypeExponential BucketType = iota
	// BucketTypeLinear creates linearly-sized buckets
	BucketTypeLinear
	// BucketTypeCustom uses user-provided bucket boundaries
	BucketTypeCustom
)

// GenerateLinearBuckets creates linearly spaced bucket boundaries
func GenerateLinearBuckets(start, width float64, count int) []float64 {
	if count <= 0 {
		return nil
	}
	
	buckets := make([]float64, count)
	for i := 0; i < count; i++ {
		buckets[i] = start + float64(i)*width
	}
	return buckets
}

// GenerateExponentialBuckets creates exponentially spaced bucket boundaries
func GenerateExponentialBuckets(start, factor float64, count int) []float64 {
	if count <= 0 || start <= 0 || factor <= 1 {
		return nil
	}
	
	buckets := make([]float64, count)
	current := start
	for i := 0; i < count; i++ {
		buckets[i] = current
		current *= factor
	}
	return buckets
}

// ValidateBuckets ensures bucket boundaries are valid and sorted
func ValidateBuckets(buckets []float64) error {
	if len(buckets) == 0 {
		return nil // Empty buckets are allowed (will use defaults)
	}
	
	// Check for non-positive values
	for i, bucket := range buckets {
		if bucket <= 0 {
			return fmt.Errorf("bucket boundary at index %d must be positive, got %f", i, bucket)
		}
	}
	
	// Check if sorted in ascending order
	for i := 1; i < len(buckets); i++ {
		if buckets[i] <= buckets[i-1] {
			return fmt.Errorf("bucket boundaries must be in ascending order: bucket[%d]=%f <= bucket[%d]=%f", 
				i, buckets[i], i-1, buckets[i-1])
		}
	}
	
	return nil
}

// Options contains configuration options for a metric
type Options struct {
	// Name is the unique identifier for the metric
	Name string
	// Description provides additional information about what the metric measures
	Description string
	// Unit specifies the unit of measurement (e.g. "milliseconds", "bytes")
	Unit string
	// Tags are key-value pairs for adding dimensions to metrics
	Tags Tags
	// Buckets defines custom histogram bucket boundaries (optional, for histograms only)
	// If not specified, default buckets will be used
	Buckets []float64
	// TTL defines how long the metric should be kept in the registry (optional)
	// If zero, the metric will not expire
	TTL time.Duration
}

// Metric is the base interface that all metric types implement
type Metric interface {
	// Name returns the unique identifier for the metric
	Name() string
	// Description provides additional information about what the metric measures
	Description() string
	// Type returns the metric type (counter, gauge, etc.)
	Type() Type
	// Tags returns the key-value pairs associated with this metric
	Tags() Tags
}

// Counter represents a monotonically increasing value
type Counter interface {
	Metric
	// Inc increments the counter by 1
	Inc()
	// Add increases the counter by the given value
	Add(value float64)
	// With returns a Counter with additional tags
	With(tags Tags) Counter
	// Value returns the current counter value
	Value() uint64
}

// Gauge represents a current point-in-time measurement
type Gauge interface {
	Metric
	// Set sets the gauge to the given value
	Set(value float64)
	// Add adds the given value to the gauge (can be negative)
	Add(value float64)
	// Inc increments the gauge by 1
	Inc()
	// Dec decrements the gauge by 1
	Dec()
	// With returns a Gauge with additional tags
	With(tags Tags) Gauge
	// Value returns the current gauge value
	Value() int64
}

// HistogramSnapshot represents the current state of a histogram
type HistogramSnapshot struct {
	Count   uint64
	Sum     uint64
	Min     uint64
	Max     uint64
	Buckets []uint64
}

// Histogram represents a statistical distribution of values
type Histogram interface {
	Metric
	// Observe records a value in the histogram
	Observe(value float64)
	// With returns a Histogram with additional tags
	With(tags Tags) Histogram
	// Snapshot returns the current histogram statistics
	Snapshot() HistogramSnapshot
}

// Timer is a specialized metric for measuring durations
type Timer interface {
	Metric
	// Record records a duration
	Record(d time.Duration)
	// RecordSince records the duration since the provided time
	RecordSince(t time.Time)
	// Time is a convenience method for timing a function
	Time(fn func()) time.Duration
	// With returns a Timer with additional tags
	With(tags Tags) Timer
	// Snapshot returns the underlying histogram statistics
	Snapshot() HistogramSnapshot
}

// Registry manages a collection of metrics
type Registry interface {
	// Counter creates or retrieves a Counter
	Counter(opts Options) Counter
	// Gauge creates or retrieves a Gauge
	Gauge(opts Options) Gauge
	// Histogram creates or retrieves a Histogram
	Histogram(opts Options) Histogram
	// Timer creates or retrieves a Timer
	Timer(opts Options) Timer
	// Unregister removes a metric from the registry
	Unregister(name string)
	// Each iterates over all registered metrics
	Each(fn func(Metric))
	// ManualCleanup removes all expired metrics immediately
	ManualCleanup()
	// Close stops background cleanup and releases resources
	Close() error
}

// Reporter is the interface for reporting metrics to a backend system
type Reporter interface {
	// Report sends metrics to a backend system
	Report(Registry) error
	// Flush ensures all buffered metrics are sent
	Flush() error
	// Close shuts down the reporter, flushing any remaining metrics
	Close() error
}

// ContextKey is a type for context keys
type ContextKey string

const (
	// RegistryContextKey is the context key for the metric registry
	RegistryContextKey ContextKey = "metrics-registry"
)

// FromContext extracts a Registry from a context.Context
func FromContext(ctx context.Context) (Registry, bool) {
	reg, ok := ctx.Value(RegistryContextKey).(Registry)
	return reg, ok
}

// NewContext creates a new context.Context with the given Registry
func NewContext(ctx context.Context, registry Registry) context.Context {
	return context.WithValue(ctx, RegistryContextKey, registry)
}
