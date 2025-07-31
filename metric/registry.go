package metric

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// metricEntry holds a metric and its expiration information
type metricEntry struct {
	metric    Metric
	expiresAt time.Time
	ttl       time.Duration
}

// defaultRegistry is a thread-safe implementation of Registry
type defaultRegistry struct {
	mu                  sync.RWMutex
	metrics             map[string]*metricEntry
	cardinality         map[string]int // tracks cardinality per metric name
	tagValidationConfig TagValidationConfig
	ctx                 context.Context
	cancel              context.CancelFunc
	cleanupInterval     time.Duration
}

// NewRegistry creates a new Registry instance with full configuration
func NewRegistry(tagConfig TagValidationConfig, cleanupInterval time.Duration) Registry {
	ctx, cancel := context.WithCancel(context.Background())
	
	r := &defaultRegistry{
		metrics:             make(map[string]*metricEntry),
		cardinality:         make(map[string]int),
		tagValidationConfig: tagConfig,
		ctx:                 ctx,
		cancel:              cancel,
		cleanupInterval:     cleanupInterval,
	}
	
	// Start cleanup goroutine only if cleanup interval is > 0
	if cleanupInterval > 0 {
		go r.cleanupLoop()
	}
	
	return r
}

// NewDefaultRegistry creates a registry with sensible defaults
func NewDefaultRegistry() Registry {
	return NewRegistry(DefaultTagValidationConfig(), 5*time.Minute)
}

// NewNoCleanupRegistry creates a registry that never expires metrics
func NewNoCleanupRegistry() Registry {
	return NewRegistry(DefaultTagValidationConfig(), 0) // 0 means no cleanup
}

// lookup retrieves a metric by name and type or creates it using the factory if it doesn't exist
func (r *defaultRegistry) lookup(opts Options, metricType Type, factory func() Metric) Metric {
	// Validate tags before proceeding
	if err := ValidateTags(opts.Tags, r.tagValidationConfig); err != nil {
		// In production, you might want to log this error and return a no-op metric
		// For now, we'll panic to make the error visible during development
		panic(fmt.Sprintf("tag validation failed: %v", err))
	}

	key := fmt.Sprintf("%s:%s", metricType, opts.Name)

	r.mu.RLock()
	entry, ok := r.metrics[key]
	r.mu.RUnlock()

	if ok {
		return entry.metric
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// Double-check after acquiring write lock
	if entry, ok = r.metrics[key]; ok {
		return entry.metric
	}

	// Check cardinality limit for this metric name
	if r.cardinality[opts.Name] >= r.tagValidationConfig.MaxCardinality {
		// In production, you might want to log this and return a no-op metric
		panic(fmt.Sprintf("cardinality limit exceeded for metric '%s': %d >= %d", 
			opts.Name, r.cardinality[opts.Name], r.tagValidationConfig.MaxCardinality))
	}

	// Create new metric
	m := factory()
	entry = &metricEntry{
		metric: m,
		ttl:    opts.TTL,
	}
	
	// Set expiration time if TTL is specified
	if opts.TTL > 0 {
		entry.expiresAt = time.Now().Add(opts.TTL)
	}
	
	r.metrics[key] = entry
	r.cardinality[opts.Name]++
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

	for _, entry := range r.metrics {
		fn(entry.metric)
	}
}

// cleanupLoop runs in the background and periodically removes expired metrics
func (r *defaultRegistry) cleanupLoop() {
	ticker := time.NewTicker(r.cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-r.ctx.Done():
			return
		case <-ticker.C:
			r.cleanupExpired()
		}
	}
}

// cleanupExpired removes expired metrics from the registry
func (r *defaultRegistry) cleanupExpired() {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	for key, entry := range r.metrics {
		// Skip metrics without TTL
		if entry.ttl == 0 {
			continue
		}

		// Remove expired metrics
		if now.After(entry.expiresAt) {
			delete(r.metrics, key)
			// Decrease cardinality count
			metricName := entry.metric.Name()
			r.cardinality[metricName]--
			if r.cardinality[metricName] <= 0 {
				delete(r.cardinality, metricName)
			}
		}
	}
}

// ManualCleanup removes all expired metrics immediately
func (r *defaultRegistry) ManualCleanup() {
	r.cleanupExpired()
}

// Close stops the cleanup goroutine and cleans up resources
func (r *defaultRegistry) Close() error {
	r.cancel()
	return nil
}

// GlobalRegistry is the default registry used when no registry is specified
var GlobalRegistry = NewDefaultRegistry()

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
