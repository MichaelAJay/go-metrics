// Package metrics provides a flexible metrics collection and reporting system
package metrics

// Re-export all types from the metric package
import (
	"github.com/MichaelAJay/go-metrics/metric"
)

// Type represents the type of a metric
type Type = metric.Type

// Metric type constants
const (
	TypeCounter   = metric.TypeCounter
	TypeGauge     = metric.TypeGauge
	TypeHistogram = metric.TypeHistogram
	TypeTimer     = metric.TypeTimer
)

// Tags represents metric tags as string key-value pairs
type Tags = metric.Tags

// Options for creating metrics
type Options = metric.Options

// Registry is an interface for a metrics registry
type Registry = metric.Registry

// Reporter is an interface for a metrics reporter
type Reporter = metric.Reporter

// Metric is the base interface for all metrics
type Metric = metric.Metric

// Counter is a monotonically increasing value
type Counter = metric.Counter

// Gauge is a value that can go up and down
type Gauge = metric.Gauge

// Histogram records observations in buckets
type Histogram = metric.Histogram

// Timer is specialized histogram for time durations
type Timer = metric.Timer

// NewRegistry creates a new metrics registry
func NewRegistry() Registry {
	return metric.NewRegistry()
}

// GlobalRegistry returns the global default registry
func GlobalRegistry() Registry {
	return metric.GlobalRegistry
}
