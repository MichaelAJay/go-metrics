package metric

import (
	"fmt"
	"sync"
)

// defaultRegistry is a thread-safe implementation of Registry
type defaultRegistry struct {
	mu      sync.RWMutex
	metrics map[string]Metric
}

// NewRegistry creates a new Registry instance
func NewRegistry() Registry {
	return &defaultRegistry{
		metrics: make(map[string]Metric),
	}
}

// lookup retrieves a metric by name and type or creates it using the factory if it doesn't exist
func (r *defaultRegistry) lookup(opts Options, metricType Type, factory func() Metric) Metric {
	key := fmt.Sprintf("%s:%s", metricType, opts.Name)

	r.mu.RLock()
	m, ok := r.metrics[key]
	r.mu.RUnlock()

	if ok {
		return m
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// Double-check after acquiring write lock
	if m, ok = r.metrics[key]; ok {
		return m
	}

	// Create new metric
	m = factory()
	r.metrics[key] = m
	return m
}

// Counter creates or retrieves a Counter
func (r *defaultRegistry) Counter(opts Options) Counter {
	m := r.lookup(opts, TypeCounter, func() Metric {
		return newCounter(opts)
	})
	return m.(Counter)
}

// Gauge creates or retrieves a Gauge
func (r *defaultRegistry) Gauge(opts Options) Gauge {
	m := r.lookup(opts, TypeGauge, func() Metric {
		return newGauge(opts)
	})
	return m.(Gauge)
}

// Histogram creates or retrieves a Histogram
func (r *defaultRegistry) Histogram(opts Options) Histogram {
	m := r.lookup(opts, TypeHistogram, func() Metric {
		return newHistogram(opts)
	})
	return m.(Histogram)
}

// Timer creates or retrieves a Timer
func (r *defaultRegistry) Timer(opts Options) Timer {
	m := r.lookup(opts, TypeTimer, func() Metric {
		return newTimer(opts)
	})
	return m.(Timer)
}

// Unregister removes a metric from the registry
func (r *defaultRegistry) Unregister(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Delete all metric types with this name
	for key := range r.metrics {
		if fmt.Sprintf("%s:%s", TypeCounter, name) == key ||
			fmt.Sprintf("%s:%s", TypeGauge, name) == key ||
			fmt.Sprintf("%s:%s", TypeHistogram, name) == key ||
			fmt.Sprintf("%s:%s", TypeTimer, name) == key {
			delete(r.metrics, key)
		}
	}
}

// Each iterates over all registered metrics
func (r *defaultRegistry) Each(fn func(Metric)) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, m := range r.metrics {
		fn(m)
	}
}

// GlobalRegistry is the default registry used when no registry is specified
var GlobalRegistry = NewRegistry()

// GetCounter creates or retrieves a Counter from the global registry
func GetCounter(opts Options) Counter {
	return GlobalRegistry.Counter(opts)
}

// GetGauge creates or retrieves a Gauge from the global registry
func GetGauge(opts Options) Gauge {
	return GlobalRegistry.Gauge(opts)
}

// GetHistogram creates or retrieves a Histogram from the global registry
func GetHistogram(opts Options) Histogram {
	return GlobalRegistry.Histogram(opts)
}

// GetTimer creates or retrieves a Timer from the global registry
func GetTimer(opts Options) Timer {
	return GlobalRegistry.Timer(opts)
}
