package metric

import (
	"maps"
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

// histogramImpl implements the Histogram interface
type histogramImpl struct {
	baseMetric
	count   uint64
	sum     uint64
	min     uint64
	max     uint64
	buckets []uint64 // Simple fixed bucket implementation
}

func newHistogram(opts Options) Histogram {
	return &histogramImpl{
		baseMetric: baseMetric{
			name:        opts.Name,
			description: opts.Description,
			unit:        opts.Unit,
			metricType:  TypeHistogram,
			tags:        opts.Tags,
		},
		// Simple default buckets - would be configurable in a full implementation
		buckets: make([]uint64, 10),
	}
}

func (h *histogramImpl) Observe(value float64) {
	// This is a simplified implementation; a production version would use more efficient
	// approaches for calculating histograms and handling concurrent updates

	// Convert to uint64 for atomic operations
	v := uint64(value)

	atomic.AddUint64(&h.count, 1)
	atomic.AddUint64(&h.sum, v)

	// Simplified bucket logic - a real implementation would use
	// proper bucketing based on configurable boundaries
	bucket := min(9, int(value/10))
	atomic.AddUint64(&h.buckets[bucket], 1)

	// Update min/max (simplified - not thread-safe without more complex locking)
	if v < atomic.LoadUint64(&h.min) || atomic.LoadUint64(&h.min) == 0 {
		atomic.StoreUint64(&h.min, v)
	}
	if v > atomic.LoadUint64(&h.max) {
		atomic.StoreUint64(&h.max, v)
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

// Helper functions

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
