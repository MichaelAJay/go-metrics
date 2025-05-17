// Package metric provides a flexible, high-performance metrics collection library
// that supports multiple backends, tagging, and various metric types.
package metric

import (
	"context"
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
}

// Histogram represents a statistical distribution of values
type Histogram interface {
	Metric
	// Observe records a value in the histogram
	Observe(value float64)
	// With returns a Histogram with additional tags
	With(tags Tags) Histogram
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
