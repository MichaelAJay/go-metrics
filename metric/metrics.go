package metric

import (
	"fmt"
	"maps"
	"sort"
	"sync/atomic"
	"time"
)

// baseMetric implements the common Metric functionality
type baseMetric struct {
	name        string
	description string
	unit        string
	metricType  Type
	tags        Tags
}

func (m *baseMetric) Name() string {
	return m.name
}

func (m *baseMetric) Description() string {
	return m.description
}

func (m *baseMetric) Type() Type {
	return m.metricType
}

func (m *baseMetric) Tags() Tags {
	// Return a copy to prevent modification
	tags := make(Tags, len(m.tags))
	for k, v := range m.tags {
		tags[k] = v
	}
	return tags
}

// Copy tags and add new ones
func copyTags(originalTags, newTags Tags) Tags {
	// If both are nil or empty, return empty Tags
	if (originalTags == nil || len(originalTags) == 0) &&
		(newTags == nil || len(newTags) == 0) {
		return Tags{}
	}

	// Make a copy with capacity for both sets of tags
	tagsCopy := make(Tags, len(originalTags)+len(newTags))

	// Copy original tags
	maps.Copy(tagsCopy, originalTags)

	// Add new tags, overwriting if keys overlap
	maps.Copy(tagsCopy, newTags)

	return tagsCopy
}

// counterImpl implements the Counter interface
type counterImpl struct {
	baseMetric
	value uint64
}

func newCounter(opts Options) Counter {
	return &counterImpl{
		baseMetric: baseMetric{
			name:        opts.Name,
			description: opts.Description,
			unit:        opts.Unit,
			metricType:  TypeCounter,
			tags:        opts.Tags,
		},
	}
}

func (c *counterImpl) Inc() {
	atomic.AddUint64(&c.value, 1)
}

func (c *counterImpl) Add(value float64) {
	// Only add if positive (counters should never decrease)
	if value > 0 {
		atomic.AddUint64(&c.value, uint64(value))
	}
}

func (c *counterImpl) With(tags Tags) Counter {
	return &counterImpl{
		baseMetric: baseMetric{
			name:        c.name,
			description: c.description,
			unit:        c.unit,
			metricType:  c.metricType,
			tags:        copyTags(c.tags, tags),
		},
	}
}

func (c *counterImpl) Value() uint64 {
	return atomic.LoadUint64(&c.value)
}

// gaugeImpl implements the Gauge interface
type gaugeImpl struct {
	baseMetric
	value int64
}

func newGauge(opts Options) Gauge {
	return &gaugeImpl{
		baseMetric: baseMetric{
			name:        opts.Name,
			description: opts.Description,
			unit:        opts.Unit,
			metricType:  TypeGauge,
			tags:        opts.Tags,
		},
	}
}

func (g *gaugeImpl) Set(value float64) {
	atomic.StoreInt64(&g.value, int64(value))
}

func (g *gaugeImpl) Add(value float64) {
	atomic.AddInt64(&g.value, int64(value))
}

func (g *gaugeImpl) Inc() {
	atomic.AddInt64(&g.value, 1)
}

func (g *gaugeImpl) Dec() {
	atomic.AddInt64(&g.value, -1)
}

func (g *gaugeImpl) With(tags Tags) Gauge {
	return &gaugeImpl{
		baseMetric: baseMetric{
			name:        g.name,
			description: g.description,
			unit:        g.unit,
			metricType:  g.metricType,
			tags:        copyTags(g.tags, tags),
		},
	}
}

func (g *gaugeImpl) Value() int64 {
	return atomic.LoadInt64(&g.value)
}

// histogramImpl implements the Histogram interface
type histogramImpl struct {
	baseMetric
	count         uint64
	sum           uint64
	min           uint64
	max           uint64
	buckets       []uint64  // Bucket counts
	boundaries    []float64 // Bucket boundaries
}

func newHistogram(opts Options) Histogram {
	// Use provided buckets or default ones
	boundaries := opts.Buckets
	if len(boundaries) == 0 {
		// Default buckets: exponential buckets from 0.001 to 10000
		boundaries = []float64{0.001, 0.01, 0.1, 1, 10, 100, 1000, 10000}
	}
	
	// Validate bucket boundaries
	if err := ValidateBuckets(boundaries); err != nil {
		// In production, you might want to log this and use default buckets
		panic(fmt.Sprintf("invalid histogram buckets: %v", err))
	}

	return &histogramImpl{
		baseMetric: baseMetric{
			name:        opts.Name,
			description: opts.Description,
			unit:        opts.Unit,
			metricType:  TypeHistogram,
			tags:        opts.Tags,
		},
		boundaries: boundaries,
		buckets:    make([]uint64, len(boundaries)+1), // +1 for the +Inf bucket
	}
}

func (h *histogramImpl) Observe(value float64) {
	// Convert to uint64 for atomic operations
	v := uint64(value)

	atomic.AddUint64(&h.count, 1)
	atomic.AddUint64(&h.sum, v)

	// Find the appropriate bucket using binary search for O(log n) performance
	bucketIndex := h.findBucket(value)
	atomic.AddUint64(&h.buckets[bucketIndex], 1)

	// Update min/max using compare-and-swap to avoid race conditions
	h.updateMin(v)
	h.updateMax(v)
}

// findBucket uses binary search to find the appropriate bucket for the given value
func (h *histogramImpl) findBucket(value float64) int {
	// Use binary search to find the first boundary that is >= value
	// sort.SearchFloat64s returns the index where value would be inserted
	index := sort.SearchFloat64s(h.boundaries, value)
	
	// If value is exactly equal to a boundary, we want that bucket
	// If value is between boundaries, we want the next bucket
	// If value is larger than all boundaries, we want the +Inf bucket
	if index < len(h.boundaries) && value <= h.boundaries[index] {
		return index
	}
	
	// Value is larger than the boundary at index, so use the +Inf bucket
	return len(h.boundaries)
}

// updateMin safely updates the minimum value using compare-and-swap
func (h *histogramImpl) updateMin(v uint64) {
	for {
		current := atomic.LoadUint64(&h.min)
		// If current is 0 (uninitialized) or v is smaller, update it
		if current == 0 || v < current {
			if atomic.CompareAndSwapUint64(&h.min, current, v) {
				break
			}
			// If CAS failed, another goroutine updated it, try again
		} else {
			// v is not smaller than current, no update needed
			break
		}
	}
}

// updateMax safely updates the maximum value using compare-and-swap
func (h *histogramImpl) updateMax(v uint64) {
	for {
		current := atomic.LoadUint64(&h.max)
		// If v is larger than current, update it
		if v > current {
			if atomic.CompareAndSwapUint64(&h.max, current, v) {
				break
			}
			// If CAS failed, another goroutine updated it, try again
		} else {
			// v is not larger than current, no update needed
			break
		}
	}
}

func (h *histogramImpl) With(tags Tags) Histogram {
	return &histogramImpl{
		baseMetric: baseMetric{
			name:        h.name,
			description: h.description,
			unit:        h.unit,
			metricType:  h.metricType,
			tags:        copyTags(h.tags, tags),
		},
		buckets: make([]uint64, len(h.buckets)),
	}
}

func (h *histogramImpl) Snapshot() HistogramSnapshot {
	// Create a copy of buckets to avoid concurrent modification
	buckets := make([]uint64, len(h.buckets))
	for i := range h.buckets {
		buckets[i] = atomic.LoadUint64(&h.buckets[i])
	}
	
	return HistogramSnapshot{
		Count:   atomic.LoadUint64(&h.count),
		Sum:     atomic.LoadUint64(&h.sum),
		Min:     atomic.LoadUint64(&h.min),
		Max:     atomic.LoadUint64(&h.max),
		Buckets: buckets,
	}
}

// timerImpl implements the Timer interface
type timerImpl struct {
	histogram Histogram
}

func newTimer(opts Options) Timer {
	return &timerImpl{
		histogram: newHistogram(opts),
	}
}

func (t *timerImpl) Name() string {
	return t.histogram.Name()
}

func (t *timerImpl) Description() string {
	return t.histogram.Description()
}

func (t *timerImpl) Type() Type {
	return TypeTimer
}

func (t *timerImpl) Tags() Tags {
	return t.histogram.Tags()
}

func (t *timerImpl) Record(d time.Duration) {
	t.histogram.Observe(float64(d.Nanoseconds()))
}

func (t *timerImpl) RecordSince(start time.Time) {
	t.Record(time.Since(start))
}

func (t *timerImpl) Time(fn func()) time.Duration {
	start := time.Now()
	fn()
	d := time.Since(start)
	t.Record(d)
	return d
}

func (t *timerImpl) With(tags Tags) Timer {
	return &timerImpl{
		histogram: t.histogram.With(tags),
	}
}

func (t *timerImpl) Snapshot() HistogramSnapshot {
	return t.histogram.Snapshot()
}

// Helper functions

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
