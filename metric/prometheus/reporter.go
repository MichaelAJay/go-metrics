// Package prometheus provides integration with Prometheus metrics system
package prometheus

import (
	"fmt"
	"sync"

	"net/http"

	"github.com/MichaelAJay/go-metrics/metric"
	prom "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Reporter implements the metric.Reporter interface for Prometheus
type Reporter struct {
	registry      *prom.Registry
	counters      map[string]prom.Counter
	gauges        map[string]prom.Gauge
	histograms    map[string]prom.Observer
	mutex         sync.Mutex
	defaultLabels prom.Labels
	registered    map[string]bool
}

// NewReporter creates a new Prometheus reporter
func NewReporter(opts ...Option) *Reporter {
	r := &Reporter{
		registry:      prom.NewRegistry(),
		counters:      make(map[string]prom.Counter),
		gauges:        make(map[string]prom.Gauge),
		histograms:    make(map[string]prom.Observer),
		defaultLabels: prom.Labels{},
		registered:    make(map[string]bool),
	}

	// Apply options
	for _, opt := range opts {
		opt(r)
	}

	return r
}

// Option is a functional option for configuring the Prometheus reporter
type Option func(*Reporter)

// WithDefaultLabels adds default labels to all metrics
func WithDefaultLabels(labels map[string]string) Option {
	return func(r *Reporter) {
		for k, v := range labels {
			r.defaultLabels[k] = v
		}
	}
}

// WithRegistry uses a custom Prometheus registry
func WithRegistry(registry *prom.Registry) Option {
	return func(r *Reporter) {
		r.registry = registry
	}
}

// Handler returns an HTTP handler for the Prometheus metrics
func (r *Reporter) Handler() http.Handler {
	return promhttp.HandlerFor(r.registry, promhttp.HandlerOpts{})
}

// Report implements the metric.Reporter interface
func (r *Reporter) Report(registry metric.Registry) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	registry.Each(func(m metric.Metric) {
		name := sanitizeName(m.Name())
		tags := m.Tags()

		// Create label set with default labels plus metric tags
		labelNames := make([]string, 0, len(tags))
		labelValues := make([]string, 0, len(tags))

		for k, v := range tags {
			labelNames = append(labelNames, k)
			labelValues = append(labelValues, v)
		}

		switch m.Type() {
		case metric.TypeCounter:
			if counter, ok := m.(metric.Counter); ok {
				r.reportCounter(name, labelNames, labelValues, counter)
			}
		case metric.TypeGauge:
			if gauge, ok := m.(metric.Gauge); ok {
				r.reportGauge(name, labelNames, labelValues, gauge)
			}
		case metric.TypeHistogram:
			if histogram, ok := m.(metric.Histogram); ok {
				r.reportHistogram(name, labelNames, labelValues, histogram)
			}
		case metric.TypeTimer:
			if timer, ok := m.(metric.Timer); ok {
				r.reportTimer(name, labelNames, labelValues, timer)
			}
		}
	})

	return nil
}

func (r *Reporter) reportCounter(name string, labelNames, labelValues []string, counter metric.Counter) {
	key := fmt.Sprintf("%s:%v", name, labelNames)
	if _, exists := r.counters[key]; !exists {
		// Only register if we haven't seen this counter before
		if !r.registered[key] {
			c := prom.NewCounterVec(
				prom.CounterOpts{
					Name: name,
					Help: getMetricHelp(counter),
				},
				labelNames,
			)

			// Use MustRegister and handle potential panics for duplicate registrations
			try(func() {
				r.registry.MustRegister(c)
				r.registered[key] = true
			})

			// Only set the counter if registration was successful
			if r.registered[key] {
				r.counters[key] = c.WithLabelValues(labelValues...)
			}
		}
	}
}

func (r *Reporter) reportGauge(name string, labelNames, labelValues []string, gauge metric.Gauge) {
	key := fmt.Sprintf("%s:%v", name, labelNames)
	if _, exists := r.gauges[key]; !exists {
		// Only register if we haven't seen this gauge before
		if !r.registered[key] {
			g := prom.NewGaugeVec(
				prom.GaugeOpts{
					Name: name,
					Help: getMetricHelp(gauge),
				},
				labelNames,
			)

			// Use MustRegister and handle potential panics for duplicate registrations
			try(func() {
				r.registry.MustRegister(g)
				r.registered[key] = true
			})

			// Only set the gauge if registration was successful
			if r.registered[key] {
				r.gauges[key] = g.WithLabelValues(labelValues...)
			}
		}
	}
}

func (r *Reporter) reportHistogram(name string, labelNames, labelValues []string, histogram metric.Histogram) {
	key := fmt.Sprintf("%s:%v", name, labelNames)
	if _, exists := r.histograms[key]; !exists {
		// Only register if we haven't seen this histogram before
		if !r.registered[key] {
			h := prom.NewHistogramVec(
				prom.HistogramOpts{
					Name:    name,
					Help:    getMetricHelp(histogram),
					Buckets: prom.DefBuckets, // Default buckets
				},
				labelNames,
			)

			// Use MustRegister and handle potential panics for duplicate registrations
			try(func() {
				r.registry.MustRegister(h)
				r.registered[key] = true
			})

			// Only set the histogram if registration was successful
			if r.registered[key] {
				r.histograms[key] = h.WithLabelValues(labelValues...)
			}
		}
	}
}

func (r *Reporter) reportTimer(name string, labelNames, labelValues []string, timer metric.Timer) {
	// Timers are histograms in Prometheus
	// We use Observer interface which is implemented by both Histogram and Summary
	// Instead of using a type assertion, use the timer's properties to create a histogram
	timerName := fmt.Sprintf("%s_seconds", name)
	key := fmt.Sprintf("%s:%v", timerName, labelNames)

	if _, exists := r.histograms[key]; !exists {
		// Only register if we haven't seen this timer before
		if !r.registered[key] {
			h := prom.NewHistogramVec(
				prom.HistogramOpts{
					Name:    timerName,
					Help:    getMetricHelp(timer),
					Buckets: prom.DefBuckets, // Default buckets
				},
				labelNames,
			)

			// Use MustRegister and handle potential panics for duplicate registrations
			try(func() {
				r.registry.MustRegister(h)
				r.registered[key] = true
			})

			// Only set the histogram if registration was successful
			if r.registered[key] {
				r.histograms[key] = h.WithLabelValues(labelValues...)
			}
		}
	}
}

// Flush implements the metric.Reporter interface
func (r *Reporter) Flush() error {
	// No-op for Prometheus as it's a pull-based system
	return nil
}

// Close implements the metric.Reporter interface
func (r *Reporter) Close() error {
	// Not much to do for Prometheus
	return nil
}

// Helper functions

func sanitizeName(name string) string {
	// @TODO ensure the name follows Prometheus naming conventions
	return name
}

func getMetricHelp(m metric.Metric) string {
	// Use description if available, or generate a default help text
	if desc := m.Description(); desc != "" {
		return desc
	}
	return "No description provided"
}

// try executes a function and recovers from panics
func try(f func()) {
	defer func() {
		if r := recover(); r != nil {
			// @TODO handle panics
		}
	}()
	f()
}
