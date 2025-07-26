// Package metrics provides a flexible metrics collection and reporting system
package metrics

// Re-export all types from the metric package
import (
	"time"

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

// HistogramSnapshot represents the current state of a histogram
type HistogramSnapshot = metric.HistogramSnapshot

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

// noopRegistry implements Registry by discarding all metrics
// This is useful for testing and scenarios where metrics are not needed
type noopRegistry struct{}

// NewNoop returns a registry that discards all metrics
func NewNoop() Registry {
	return &noopRegistry{}
}

func (n *noopRegistry) Counter(opts Options) Counter {
	return &noopCounter{name: opts.Name, metricType: TypeCounter, tags: opts.Tags}
}

func (n *noopRegistry) Gauge(opts Options) Gauge {
	return &noopGauge{name: opts.Name, metricType: TypeGauge, tags: opts.Tags}
}

func (n *noopRegistry) Histogram(opts Options) Histogram {
	return &noopHistogram{name: opts.Name, metricType: TypeHistogram, tags: opts.Tags}
}

func (n *noopRegistry) Timer(opts Options) Timer {
	return &noopTimer{name: opts.Name, metricType: TypeTimer, tags: opts.Tags}
}

func (n *noopRegistry) Unregister(name string) {}

func (n *noopRegistry) Each(fn func(Metric)) {}

// Noop metric implementations
type noopCounter struct {
	name       string
	metricType Type
	tags       Tags
}

func (n *noopCounter) Name() string        { return n.name }
func (n *noopCounter) Description() string { return "" }
func (n *noopCounter) Type() Type          { return n.metricType }
func (n *noopCounter) Tags() Tags          { return n.tags }
func (n *noopCounter) Inc()                {}
func (n *noopCounter) Add(value float64)   {}
func (n *noopCounter) Value() uint64       { return 0 }
func (n *noopCounter) With(tags Tags) Counter {
	return &noopCounter{name: n.name, metricType: n.metricType, tags: tags}
}

type noopGauge struct {
	name       string
	metricType Type
	tags       Tags
}

func (n *noopGauge) Name() string        { return n.name }
func (n *noopGauge) Description() string { return "" }
func (n *noopGauge) Type() Type          { return n.metricType }
func (n *noopGauge) Tags() Tags          { return n.tags }
func (n *noopGauge) Set(value float64)   {}
func (n *noopGauge) Add(value float64)   {}
func (n *noopGauge) Inc()                {}
func (n *noopGauge) Dec()                {}
func (n *noopGauge) Value() int64        { return 0 }
func (n *noopGauge) With(tags Tags) Gauge {
	return &noopGauge{name: n.name, metricType: n.metricType, tags: tags}
}

type noopHistogram struct {
	name       string
	metricType Type
	tags       Tags
}

func (n *noopHistogram) Name() string              { return n.name }
func (n *noopHistogram) Description() string       { return "" }
func (n *noopHistogram) Type() Type                { return n.metricType }
func (n *noopHistogram) Tags() Tags                { return n.tags }
func (n *noopHistogram) Observe(value float64)     {}
func (n *noopHistogram) Snapshot() HistogramSnapshot {
	return HistogramSnapshot{}
}
func (n *noopHistogram) With(tags Tags) Histogram {
	return &noopHistogram{name: n.name, metricType: n.metricType, tags: tags}
}

type noopTimer struct {
	name       string
	metricType Type
	tags       Tags
}

func (n *noopTimer) Name() string                   { return n.name }
func (n *noopTimer) Description() string            { return "" }
func (n *noopTimer) Type() Type                     { return n.metricType }
func (n *noopTimer) Tags() Tags                     { return n.tags }
func (n *noopTimer) Record(d time.Duration)         {}
func (n *noopTimer) RecordSince(t time.Time)        {}
func (n *noopTimer) Time(fn func()) time.Duration   { fn(); return 0 }
func (n *noopTimer) Snapshot() HistogramSnapshot { return HistogramSnapshot{} }
func (n *noopTimer) With(tags Tags) Timer {
	return &noopTimer{name: n.name, metricType: n.metricType, tags: tags}
}
