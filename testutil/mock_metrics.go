package testutil

import (
	"sync"
	"time"

	"github.com/MichaelAJay/go-metrics/metric"
)

// baseMetric provides common metric functionality for all mock metrics.
type baseMetric struct {
	name        string
	description string
	metricType  metric.Type
	tags        metric.Tags
}

func (b *baseMetric) Name() string {
	return b.name
}

func (b *baseMetric) Description() string {
	return b.description
}

func (b *baseMetric) Type() metric.Type {
	return b.metricType
}

func (b *baseMetric) Tags() metric.Tags {
	return b.tags
}

// MockCounter captures counter operations for inspection in tests.
type MockCounter struct {
	baseMetric
	value     uint64
	incCalls  int
	addCalls  []float64
	withCalls []metric.Tags
	
	// Optional callbacks
	OnIncCallback  func()
	OnAddCallback  func(value float64)
	OnWithCallback func(tags metric.Tags) metric.Counter
	
	mu sync.RWMutex
}

// NewMockCounter creates a new MockCounter instance.
func NewMockCounter(opts metric.Options) *MockCounter {
	return &MockCounter{
		baseMetric: baseMetric{
			name:        opts.Name,
			description: opts.Description,
			metricType:  metric.TypeCounter,
			tags:        opts.Tags,
		},
	}
}

func (m *MockCounter) Inc() {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.incCalls++
	m.value++
	
	if m.OnIncCallback != nil {
		m.OnIncCallback()
	}
}

func (m *MockCounter) Add(value float64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.addCalls = append(m.addCalls, value)
	m.value += uint64(value)
	
	if m.OnAddCallback != nil {
		m.OnAddCallback(value)
	}
}

func (m *MockCounter) With(tags metric.Tags) metric.Counter {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.withCalls = append(m.withCalls, tags)
	
	if m.OnWithCallback != nil {
		return m.OnWithCallback(tags)
	}
	
	// For simplicity, return the same instance
	return m
}

func (m *MockCounter) Value() uint64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.value
}

// Test inspection methods
func (m *MockCounter) IncCalls() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.incCalls
}

func (m *MockCounter) AddCalls() []float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return append([]float64(nil), m.addCalls...)
}

func (m *MockCounter) WithCalls() []metric.Tags {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]metric.Tags, len(m.withCalls))
	copy(result, m.withCalls)
	return result
}

func (m *MockCounter) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.value = 0
	m.incCalls = 0
	m.addCalls = nil
	m.withCalls = nil
}

// MockGauge captures gauge operations for inspection in tests.
type MockGauge struct {
	baseMetric
	value     int64
	setCalls  []float64
	addCalls  []float64
	incCalls  int
	decCalls  int
	withCalls []metric.Tags
	
	// Optional callbacks
	OnSetCallback  func(value float64)
	OnAddCallback  func(value float64)
	OnIncCallback  func()
	OnDecCallback  func()
	OnWithCallback func(tags metric.Tags) metric.Gauge
	
	mu sync.RWMutex
}

// NewMockGauge creates a new MockGauge instance.
func NewMockGauge(opts metric.Options) *MockGauge {
	return &MockGauge{
		baseMetric: baseMetric{
			name:        opts.Name,
			description: opts.Description,
			metricType:  metric.TypeGauge,
			tags:        opts.Tags,
		},
	}
}

func (m *MockGauge) Set(value float64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.setCalls = append(m.setCalls, value)
	m.value = int64(value)
	
	if m.OnSetCallback != nil {
		m.OnSetCallback(value)
	}
}

func (m *MockGauge) Add(value float64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.addCalls = append(m.addCalls, value)
	m.value += int64(value)
	
	if m.OnAddCallback != nil {
		m.OnAddCallback(value)
	}
}

func (m *MockGauge) Inc() {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.incCalls++
	m.value++
	
	if m.OnIncCallback != nil {
		m.OnIncCallback()
	}
}

func (m *MockGauge) Dec() {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.decCalls++
	m.value--
	
	if m.OnDecCallback != nil {
		m.OnDecCallback()
	}
}

func (m *MockGauge) With(tags metric.Tags) metric.Gauge {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.withCalls = append(m.withCalls, tags)
	
	if m.OnWithCallback != nil {
		return m.OnWithCallback(tags)
	}
	
	return m
}

func (m *MockGauge) Value() int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.value
}

// Test inspection methods
func (m *MockGauge) SetCalls() []float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return append([]float64(nil), m.setCalls...)
}

func (m *MockGauge) AddCalls() []float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return append([]float64(nil), m.addCalls...)
}

func (m *MockGauge) IncCalls() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.incCalls
}

func (m *MockGauge) DecCalls() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.decCalls
}

func (m *MockGauge) WithCalls() []metric.Tags {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]metric.Tags, len(m.withCalls))
	copy(result, m.withCalls)
	return result
}

func (m *MockGauge) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.value = 0
	m.setCalls = nil
	m.addCalls = nil
	m.incCalls = 0
	m.decCalls = 0
	m.withCalls = nil
}

// MockHistogram captures histogram operations for inspection in tests.
type MockHistogram struct {
	baseMetric
	observeCalls []float64
	withCalls    []metric.Tags
	snapshot     metric.HistogramSnapshot
	
	// Optional callbacks
	OnObserveCallback  func(value float64)
	OnWithCallback     func(tags metric.Tags) metric.Histogram
	OnSnapshotCallback func() metric.HistogramSnapshot
	
	mu sync.RWMutex
}

// NewMockHistogram creates a new MockHistogram instance.
func NewMockHistogram(opts metric.Options) *MockHistogram {
	return &MockHistogram{
		baseMetric: baseMetric{
			name:        opts.Name,
			description: opts.Description,
			metricType:  metric.TypeHistogram,
			tags:        opts.Tags,
		},
		snapshot: metric.HistogramSnapshot{
			Buckets: make([]uint64, 10), // Default bucket count
		},
	}
}

func (m *MockHistogram) Observe(value float64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.observeCalls = append(m.observeCalls, value)
	
	// Update snapshot
	m.snapshot.Count++
	m.snapshot.Sum += uint64(value)
	if m.snapshot.Min == 0 || uint64(value) < m.snapshot.Min {
		m.snapshot.Min = uint64(value)
	}
	if uint64(value) > m.snapshot.Max {
		m.snapshot.Max = uint64(value)
	}
	
	if m.OnObserveCallback != nil {
		m.OnObserveCallback(value)
	}
}

func (m *MockHistogram) With(tags metric.Tags) metric.Histogram {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.withCalls = append(m.withCalls, tags)
	
	if m.OnWithCallback != nil {
		return m.OnWithCallback(tags)
	}
	
	return m
}

func (m *MockHistogram) Snapshot() metric.HistogramSnapshot {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	if m.OnSnapshotCallback != nil {
		return m.OnSnapshotCallback()
	}
	
	return m.snapshot
}

// Test inspection methods
func (m *MockHistogram) ObserveCalls() []float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return append([]float64(nil), m.observeCalls...)
}

func (m *MockHistogram) WithCalls() []metric.Tags {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]metric.Tags, len(m.withCalls))
	copy(result, m.withCalls)
	return result
}

func (m *MockHistogram) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.observeCalls = nil
	m.withCalls = nil
	m.snapshot = metric.HistogramSnapshot{
		Buckets: make([]uint64, 10),
	}
}

// MockTimer captures timer operations for inspection in tests.
type MockTimer struct {
	baseMetric
	recordCalls      []time.Duration
	recordSinceCalls []time.Time
	timeCalls        int
	withCalls        []metric.Tags
	snapshot         metric.HistogramSnapshot
	
	// Optional callbacks
	OnRecordCallback      func(d time.Duration)
	OnRecordSinceCallback func(t time.Time)
	OnTimeCallback        func(fn func()) time.Duration
	OnWithCallback        func(tags metric.Tags) metric.Timer
	OnSnapshotCallback    func() metric.HistogramSnapshot
	
	mu sync.RWMutex
}

// NewMockTimer creates a new MockTimer instance.
func NewMockTimer(opts metric.Options) *MockTimer {
	return &MockTimer{
		baseMetric: baseMetric{
			name:        opts.Name,
			description: opts.Description,
			metricType:  metric.TypeTimer,
			tags:        opts.Tags,
		},
		snapshot: metric.HistogramSnapshot{
			Buckets: make([]uint64, 10),
		},
	}
}

func (m *MockTimer) Record(d time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.recordCalls = append(m.recordCalls, d)
	
	// Update snapshot
	m.snapshot.Count++
	duration := uint64(d.Nanoseconds())
	m.snapshot.Sum += duration
	if m.snapshot.Min == 0 || duration < m.snapshot.Min {
		m.snapshot.Min = duration
	}
	if duration > m.snapshot.Max {
		m.snapshot.Max = duration
	}
	
	if m.OnRecordCallback != nil {
		m.OnRecordCallback(d)
	}
}

func (m *MockTimer) RecordSince(t time.Time) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.recordSinceCalls = append(m.recordSinceCalls, t)
	
	duration := time.Since(t)
	m.recordCalls = append(m.recordCalls, duration)
	
	if m.OnRecordSinceCallback != nil {
		m.OnRecordSinceCallback(t)
	}
}

func (m *MockTimer) Time(fn func()) time.Duration {
	m.mu.Lock()
	m.timeCalls++
	m.mu.Unlock()
	
	if m.OnTimeCallback != nil {
		return m.OnTimeCallback(fn)
	}
	
	start := time.Now()
	fn()
	duration := time.Since(start)
	
	m.Record(duration)
	return duration
}

func (m *MockTimer) With(tags metric.Tags) metric.Timer {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.withCalls = append(m.withCalls, tags)
	
	if m.OnWithCallback != nil {
		return m.OnWithCallback(tags)
	}
	
	return m
}

func (m *MockTimer) Snapshot() metric.HistogramSnapshot {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	if m.OnSnapshotCallback != nil {
		return m.OnSnapshotCallback()
	}
	
	return m.snapshot
}

// Test inspection methods
func (m *MockTimer) RecordCalls() []time.Duration {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return append([]time.Duration(nil), m.recordCalls...)
}

func (m *MockTimer) RecordSinceCalls() []time.Time {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return append([]time.Time(nil), m.recordSinceCalls...)
}

func (m *MockTimer) TimeCalls() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.timeCalls
}

func (m *MockTimer) WithCalls() []metric.Tags {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]metric.Tags, len(m.withCalls))
	copy(result, m.withCalls)
	return result
}

func (m *MockTimer) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.recordCalls = nil
	m.recordSinceCalls = nil
	m.timeCalls = 0
	m.withCalls = nil
	m.snapshot = metric.HistogramSnapshot{
		Buckets: make([]uint64, 10),
	}
}

// Compile-time interface compliance checks
var _ metric.Counter = (*MockCounter)(nil)
var _ metric.Gauge = (*MockGauge)(nil)
var _ metric.Histogram = (*MockHistogram)(nil)
var _ metric.Timer = (*MockTimer)(nil)