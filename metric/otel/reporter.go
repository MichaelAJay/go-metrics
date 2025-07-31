// Package otel provides OpenTelemetry integration for the metrics package
package otel

import (
	"context"
	"fmt"
	"sync"

	metricpkg "github.com/MichaelAJay/go-metrics/metric"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/prometheus"
	otelmetric "go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
)

// Reporter implements the metric.Reporter interface for OpenTelemetry
type Reporter struct {
	provider       *sdkmetric.MeterProvider
	meter          otelmetric.Meter
	counters       map[string]otelmetric.Int64Counter
	gauges         map[string]otelmetric.Int64ObservableGauge
	histograms     map[string]otelmetric.Float64Histogram
	mutex          sync.RWMutex
	defaultAttrs   []attribute.KeyValue
	ctx            context.Context
	cancel         context.CancelFunc
	observing      map[string]bool
	gaugeCallbacks map[string]otelmetric.Registration
}

// NewReporter creates a new OpenTelemetry reporter
func NewReporter(serviceName, version string, options ...Option) (*Reporter, error) {
	ctx, cancel := context.WithCancel(context.Background())

	// Create a new Prometheus exporter
	exporter, err := prometheus.New()
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create Prometheus exporter: %w", err)
	}

	// Create resource with service information
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(serviceName),
			semconv.ServiceVersion(version),
		),
	)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Create the MeterProvider
	provider := sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(res),
		sdkmetric.WithReader(exporter),
	)

	// Set the global MeterProvider
	otel.SetMeterProvider(provider)

	// Create the Reporter
	r := &Reporter{
		provider:       provider,
		meter:          provider.Meter(serviceName),
		counters:       make(map[string]otelmetric.Int64Counter),
		gauges:         make(map[string]otelmetric.Int64ObservableGauge),
		histograms:     make(map[string]otelmetric.Float64Histogram),
		defaultAttrs:   []attribute.KeyValue{},
		ctx:            ctx,
		cancel:         cancel,
		observing:      make(map[string]bool),
		gaugeCallbacks: make(map[string]otelmetric.Registration),
	}

	// Apply options
	for _, opt := range options {
		opt(r)
	}

	return r, nil
}

// Option is a functional option for configuring the OpenTelemetry reporter
type Option func(*Reporter)

// WithAttributes adds default attributes to all metrics
func WithAttributes(attrs map[string]string) Option {
	return func(r *Reporter) {
		for k, v := range attrs {
			r.defaultAttrs = append(r.defaultAttrs, attribute.String(k, v))
		}
	}
}

// Report implements the metric.Reporter interface
func (r *Reporter) Report(registry metricpkg.Registry) error {
	// Process each metric in the registry
	registry.Each(func(m metricpkg.Metric) {
		name := m.Name()

		// Convert metric.Tags to OpenTelemetry attributes
		attrs := r.convertTags(m.Tags())

		// Handle each metric type
		switch m.Type() {
		case metricpkg.TypeCounter:
			if counter, ok := m.(metricpkg.Counter); ok {
				r.reportCounter(name, counter)
			}
		case metricpkg.TypeGauge:
			if gauge, ok := m.(metricpkg.Gauge); ok {
				r.reportGauge(name, attrs, gauge)
			}
		case metricpkg.TypeHistogram:
			if histogram, ok := m.(metricpkg.Histogram); ok {
				r.reportHistogram(name, attrs, histogram)
			}
		case metricpkg.TypeTimer:
			if timer, ok := m.(metricpkg.Timer); ok {
				r.reportTimer(name, attrs, timer)
			}
		}
	})

	return nil
}


func (r *Reporter) reportCounter(name string, counter metricpkg.Counter) {
	// Create or get the counter
	otelCounter := r.getOrCreateCounter(name, counter.Description())

	// Get the value from our counter using the safe Value() method
	value := int64(counter.Value())

	// Record the value - convert []attribute.KeyValue to an option list
	// In OpenTelemetry, options need to be passed as variadic parameters
	otelCounter.Add(r.ctx, value)
}

func (r *Reporter) reportGauge(name string, attrs []attribute.KeyValue, gauge metricpkg.Gauge) {
	// Create the gauge if it doesn't exist and set up observation
	otelGauge := r.getOrCreateGauge(name, gauge.Description())

	// Set up a gauge callback if we haven't already
	key := fmt.Sprintf("%s:%v", name, attrs)
	if _, exists := r.gaugeCallbacks[key]; !exists {
		// Save a reference to our gauge for the callback
		// This creates a closure over our gauge instance
		metricGauge := gauge

		// Register a callback for this gauge
		callback, err := r.meter.RegisterCallback(
			func(_ context.Context, o otelmetric.Observer) error {
				// Get current value using the safe Value() method
				value := metricGauge.Value()
				// Report to OpenTelemetry
				o.ObserveInt64(otelGauge, value)
				return nil
			},
			otelGauge,
		)

		if err == nil {
			r.gaugeCallbacks[key] = callback
		}
	}
}

func (r *Reporter) reportHistogram(name string, _ []attribute.KeyValue, histogram metricpkg.Histogram) {
	// Create or get the histogram
	otelHistogram := r.getOrCreateHistogram(name, histogram.Description())

	// Get the current histogram snapshot using the safe Snapshot() method
	snapshot := histogram.Snapshot()

	// Record observations based on the histogram buckets
	// This is a simplified approach - in a full implementation, we'd record
	// individual observations or use OpenTelemetry's histogram directly
	if snapshot.Count > 0 {
		// Record the average value as a representative sample
		avgValue := float64(snapshot.Sum) / float64(snapshot.Count)
		otelHistogram.Record(r.ctx, avgValue)
	}
}

func (r *Reporter) reportTimer(name string, _ []attribute.KeyValue, timer metricpkg.Timer) {
	// Create a histogram for the timer
	otelHistogram := r.getOrCreateHistogram(name+"_seconds", timer.Description())

	// Get the current timer snapshot using the safe Snapshot() method
	snapshot := timer.Snapshot()

	// Record observations based on the timer's histogram data
	// Convert from nanoseconds to seconds for better OpenTelemetry compatibility
	if snapshot.Count > 0 {
		// Record the average duration in seconds
		avgDurationNanos := float64(snapshot.Sum) / float64(snapshot.Count)
		avgDurationSeconds := avgDurationNanos / 1e9 // Convert nanoseconds to seconds
		otelHistogram.Record(r.ctx, avgDurationSeconds)
	}
}

func (r *Reporter) getOrCreateCounter(name, help string) otelmetric.Int64Counter {
	r.mutex.RLock()
	counter, exists := r.counters[name]
	r.mutex.RUnlock()

	if exists {
		return counter
	}

	r.mutex.Lock()
	defer r.mutex.Unlock()

	// Double-check after acquiring write lock
	if counter, exists = r.counters[name]; exists {
		return counter
	}

	// Create the counter
	counter, err := r.meter.Int64Counter(
		name,
		otelmetric.WithDescription(help),
		otelmetric.WithUnit("1"),
	)
	if err == nil {
		r.counters[name] = counter
	}

	return counter
}

func (r *Reporter) getOrCreateGauge(name, help string) otelmetric.Int64ObservableGauge {
	r.mutex.RLock()
	gauge, exists := r.gauges[name]
	r.mutex.RUnlock()

	if exists {
		return gauge
	}

	r.mutex.Lock()
	defer r.mutex.Unlock()

	// Double-check after acquiring write lock
	if gauge, exists = r.gauges[name]; exists {
		return gauge
	}

	// Create the gauge
	gauge, err := r.meter.Int64ObservableGauge(
		name,
		otelmetric.WithDescription(help),
		otelmetric.WithUnit("1"),
	)
	if err == nil {
		r.gauges[name] = gauge
	}

	return gauge
}

func (r *Reporter) getOrCreateHistogram(name, help string) otelmetric.Float64Histogram {
	r.mutex.RLock()
	histogram, exists := r.histograms[name]
	r.mutex.RUnlock()

	if exists {
		return histogram
	}

	r.mutex.Lock()
	defer r.mutex.Unlock()

	// Double-check after acquiring write lock
	if histogram, exists = r.histograms[name]; exists {
		return histogram
	}

	// Create the histogram
	histogram, err := r.meter.Float64Histogram(
		name,
		otelmetric.WithDescription(help),
		otelmetric.WithUnit("1"),
	)
	if err == nil {
		r.histograms[name] = histogram
	}

	return histogram
}

// Flush implements the metric.Reporter interface
func (r *Reporter) Flush() error {
	// OpenTelemetry has background collection, so explicit flushing isn't needed
	return nil
}

// Close implements the metric.Reporter interface
func (r *Reporter) Close() error {
	// Cancel the context and shut down the provider
	r.cancel()

	// Unregister all callbacks
	for _, callback := range r.gaugeCallbacks {
		callback.Unregister()
	}

	// Shutdown the provider
	return r.provider.Shutdown(context.Background())
}

// Helper functions

func (r *Reporter) convertTags(tags metricpkg.Tags) []attribute.KeyValue {
	if len(tags) == 0 {
		return r.defaultAttrs
	}

	attrs := make([]attribute.KeyValue, 0, len(r.defaultAttrs)+len(tags))

	// Copy default attributes
	attrs = append(attrs, r.defaultAttrs...)

	// Add tags as attributes
	for k, v := range tags {
		attrs = append(attrs, attribute.String(k, v))
	}

	return attrs
}
