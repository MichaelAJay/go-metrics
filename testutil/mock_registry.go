package testutil

import (
	"sync"

	"github.com/MichaelAJay/go-metrics/metric"
)

// MockRegistry captures metric operations for inspection in tests.
type MockRegistry struct {
	counters   map[string]*MockCounter
	gauges     map[string]*MockGauge
	histograms map[string]*MockHistogram
	timers     map[string]*MockTimer
	
	// Call tracking
	CounterCalls   []metric.Options
	GaugeCalls     []metric.Options
	HistogramCalls []metric.Options
	TimerCalls     []metric.Options
	UnregisterCalls []string
	EachCalls      int
	
	// Optional callbacks for custom test behavior
	OnCounterCallback   func(opts metric.Options) metric.Counter
	OnGaugeCallback     func(opts metric.Options) metric.Gauge
	OnHistogramCallback func(opts metric.Options) metric.Histogram
	OnTimerCallback     func(opts metric.Options) metric.Timer
	OnUnregisterCallback func(name string)
	OnEachCallback      func(fn func(metric.Metric))
	
	mu sync.RWMutex
}

// NewMockRegistry creates a new MockRegistry instance.
func NewMockRegistry() *MockRegistry {
	return &MockRegistry{
		counters:   make(map[string]*MockCounter),
		gauges:     make(map[string]*MockGauge),
		histograms: make(map[string]*MockHistogram),
		timers:     make(map[string]*MockTimer),
	}
}

// Counter creates or retrieves a MockCounter.
func (m *MockRegistry) Counter(opts metric.Options) metric.Counter {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.CounterCalls = append(m.CounterCalls, opts)
	
	if m.OnCounterCallback != nil {
		return m.OnCounterCallback(opts)
	}
	
	if counter, exists := m.counters[opts.Name]; exists {
		return counter
	}
	
	counter := NewMockCounter(opts)
	m.counters[opts.Name] = counter
	return counter
}

// Gauge creates or retrieves a MockGauge.
func (m *MockRegistry) Gauge(opts metric.Options) metric.Gauge {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.GaugeCalls = append(m.GaugeCalls, opts)
	
	if m.OnGaugeCallback != nil {
		return m.OnGaugeCallback(opts)
	}
	
	if gauge, exists := m.gauges[opts.Name]; exists {
		return gauge
	}
	
	gauge := NewMockGauge(opts)
	m.gauges[opts.Name] = gauge
	return gauge
}

// Histogram creates or retrieves a MockHistogram.
func (m *MockRegistry) Histogram(opts metric.Options) metric.Histogram {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.HistogramCalls = append(m.HistogramCalls, opts)
	
	if m.OnHistogramCallback != nil {
		return m.OnHistogramCallback(opts)
	}
	
	if histogram, exists := m.histograms[opts.Name]; exists {
		return histogram
	}
	
	histogram := NewMockHistogram(opts)
	m.histograms[opts.Name] = histogram
	return histogram
}

// Timer creates or retrieves a MockTimer.
func (m *MockRegistry) Timer(opts metric.Options) metric.Timer {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.TimerCalls = append(m.TimerCalls, opts)
	
	if m.OnTimerCallback != nil {
		return m.OnTimerCallback(opts)
	}
	
	if timer, exists := m.timers[opts.Name]; exists {
		return timer
	}
	
	timer := NewMockTimer(opts)
	m.timers[opts.Name] = timer
	return timer
}

// Unregister removes a metric from the registry.
func (m *MockRegistry) Unregister(name string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.UnregisterCalls = append(m.UnregisterCalls, name)
	
	if m.OnUnregisterCallback != nil {
		m.OnUnregisterCallback(name)
	}
	
	delete(m.counters, name)
	delete(m.gauges, name)
	delete(m.histograms, name)
	delete(m.timers, name)
}

// Each iterates over all registered metrics.
func (m *MockRegistry) Each(fn func(metric.Metric)) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	m.EachCalls++
	
	if m.OnEachCallback != nil {
		m.OnEachCallback(fn)
		return
	}
	
	for _, counter := range m.counters {
		fn(counter)
	}
	for _, gauge := range m.gauges {
		fn(gauge)
	}
	for _, histogram := range m.histograms {
		fn(histogram)
	}
	for _, timer := range m.timers {
		fn(timer)
	}
}

// GetCounter retrieves a counter by name for test inspection.
func (m *MockRegistry) GetCounter(name string) *MockCounter {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.counters[name]
}

// GetGauge retrieves a gauge by name for test inspection.
func (m *MockRegistry) GetGauge(name string) *MockGauge {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.gauges[name]
}

// GetHistogram retrieves a histogram by name for test inspection.
func (m *MockRegistry) GetHistogram(name string) *MockHistogram {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.histograms[name]
}

// GetTimer retrieves a timer by name for test inspection.
func (m *MockRegistry) GetTimer(name string) *MockTimer {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.timers[name]
}

// Reset clears all metrics and call history.
func (m *MockRegistry) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.counters = make(map[string]*MockCounter)
	m.gauges = make(map[string]*MockGauge)
	m.histograms = make(map[string]*MockHistogram)
	m.timers = make(map[string]*MockTimer)
	
	m.CounterCalls = nil
	m.GaugeCalls = nil
	m.HistogramCalls = nil
	m.TimerCalls = nil
	m.UnregisterCalls = nil
	m.EachCalls = 0
}

// Compile-time interface compliance check
var _ metric.Registry = (*MockRegistry)(nil)