package metrics

import (
	"testing"
	"time"
)

func TestNewNoop(t *testing.T) {
	registry := NewNoop()
	
	// Verify it implements the Registry interface
	if registry == nil {
		t.Fatal("NewNoop() returned nil")
	}
	
	// Test Counter operations
	counter := registry.Counter(Options{Name: "test_counter"})
	if counter == nil {
		t.Error("Counter() returned nil")
	}
	
	counter.Inc()
	counter.Add(5.0)
	if counter.Value() != 0 { // Should always return 0 for noop
		t.Error("Noop counter should always return 0")
	}
	
	// Test Gauge operations
	gauge := registry.Gauge(Options{Name: "test_gauge"})
	if gauge == nil {
		t.Error("Gauge() returned nil")
	}
	
	gauge.Set(10.0)
	gauge.Inc()
	gauge.Dec()
	gauge.Add(-5.0)
	if gauge.Value() != 0 { // Should always return 0 for noop
		t.Error("Noop gauge should always return 0")
	}
	
	// Test Histogram operations
	histogram := registry.Histogram(Options{Name: "test_histogram"})
	if histogram == nil {
		t.Error("Histogram() returned nil")
	}
	
	histogram.Observe(15.0)
	snapshot := histogram.Snapshot()
	if snapshot.Count != 0 || snapshot.Sum != 0 { // Should always return empty snapshot
		t.Error("Noop histogram should always return empty snapshot")
	}
	
	// Test Timer operations
	timer := registry.Timer(Options{Name: "test_timer"})
	if timer == nil {
		t.Error("Timer() returned nil")
	}
	
	timer.Record(time.Second)
	timer.RecordSince(time.Now())
	duration := timer.Time(func() {
		// Do something
		time.Sleep(time.Millisecond)
	})
	if duration != 0 {
		t.Error("Noop timer Time() should still execute function but return 0")
	}
	
	timerSnapshot := timer.Snapshot()
	if timerSnapshot.Count != 0 { // Should always return empty snapshot
		t.Error("Noop timer should always return empty snapshot")
	}
	
	// Test registry operations
	registry.Unregister("test_counter") // Should not panic
	
	callCount := 0
	registry.Each(func(Metric) {
		callCount++
	})
	if callCount != 0 { // Should never call the function
		t.Error("Noop registry Each() should never call the function")
	}
}

func TestNoopMetricsWith(t *testing.T) {
	registry := NewNoop()
	tags := Tags{"env": "test", "service": "auth"}
	
	// Test that With() methods return new instances but still noop
	counter := registry.Counter(Options{Name: "test"})
	counterWithTags := counter.With(tags)
	if counterWithTags == nil {
		t.Error("Counter With() returned nil")
	}
	
	gauge := registry.Gauge(Options{Name: "test"})
	gaugeWithTags := gauge.With(tags)
	if gaugeWithTags == nil {
		t.Error("Gauge With() returned nil")
	}
	
	histogram := registry.Histogram(Options{Name: "test"})
	histogramWithTags := histogram.With(tags)
	if histogramWithTags == nil {
		t.Error("Histogram With() returned nil")
	}
	
	timer := registry.Timer(Options{Name: "test"})
	timerWithTags := timer.With(tags)
	if timerWithTags == nil {
		t.Error("Timer With() returned nil")
	}
	
	// All operations should still be noop
	counterWithTags.Inc()
	gaugeWithTags.Set(100)
	histogramWithTags.Observe(50)
	timerWithTags.Record(time.Second)
	
	// Values should still be zero/empty
	if counterWithTags.Value() != 0 {
		t.Error("Counter with tags should still return 0")
	}
	if gaugeWithTags.Value() != 0 {
		t.Error("Gauge with tags should still return 0")
	}
	if histogramWithTags.Snapshot().Count != 0 {
		t.Error("Histogram with tags should still return empty snapshot")
	}
	if timerWithTags.Snapshot().Count != 0 {
		t.Error("Timer with tags should still return empty snapshot")
	}
}

func TestNoopRealWorldUsage(t *testing.T) {
	// Test realistic usage patterns from auth package tests
	registry := NewNoop()
	
	// Simulate creating metrics like in session management
	counter := registry.Counter(Options{
		Name:        "session_operations_total",
		Description: "Total session operations",
		Tags:        Tags{"operation": "create"},
	})
	
	timer := registry.Timer(Options{
		Name:        "session_operation_duration",
		Description: "Duration of session operations",
		Tags:        Tags{"operation": "create"},
	})
	
	histogram := registry.Histogram(Options{
		Name:        "request_size_bytes",
		Description: "Size of requests",
	})
	
	// Simulate usage patterns from auth code
	counter.Inc()
	counter.Add(5)
	
	timer.Record(time.Millisecond * 150)
	timer.RecordSince(time.Now().Add(-time.Second))
	
	histogram.Observe(1024)
	
	// Test With() chaining that's common in metrics code
	errorCounter := counter.With(Tags{"status": "error"})
	errorCounter.Inc()
	
	slowTimer := timer.With(Tags{"slow": "true"})
	slowTimer.Record(time.Second * 5)
	
	// All operations should complete without issues and return zero values
	if counter.Value() != 0 {
		t.Error("Noop counter should return 0")
	}
	if errorCounter.Value() != 0 {
		t.Error("Noop counter with tags should return 0") 
	}
	if timer.Snapshot().Count != 0 {
		t.Error("Noop timer should return empty snapshot")
	}
	if slowTimer.Snapshot().Count != 0 {
		t.Error("Noop timer with tags should return empty snapshot")
	}
	if histogram.Snapshot().Sum != 0 {
		t.Error("Noop histogram should return empty snapshot")
	}
}

func TestNoopMetricProperties(t *testing.T) {
	registry := NewNoop()
	
	opts := Options{
		Name:        "test_metric",
		Description: "Test metric description",
		Unit:        "milliseconds",
		Tags:        Tags{"type": "test"},
	}
	
	counter := registry.Counter(opts)
	if counter.Name() != opts.Name {
		t.Errorf("Counter Name() = %s, want %s", counter.Name(), opts.Name)
	}
	if counter.Type() != TypeCounter {
		t.Errorf("Counter Type() = %s, want %s", counter.Type(), TypeCounter)
	}
	
	gauge := registry.Gauge(opts)
	if gauge.Name() != opts.Name {
		t.Errorf("Gauge Name() = %s, want %s", gauge.Name(), opts.Name)
	}  
	if gauge.Type() != TypeGauge {
		t.Errorf("Gauge Type() = %s, want %s", gauge.Type(), TypeGauge)
	}
	
	histogram := registry.Histogram(opts)
	if histogram.Name() != opts.Name {
		t.Errorf("Histogram Name() = %s, want %s", histogram.Name(), opts.Name)
	}
	if histogram.Type() != TypeHistogram {
		t.Errorf("Histogram Type() = %s, want %s", histogram.Type(), TypeHistogram)
	}
	
	timer := registry.Timer(opts)
	if timer.Name() != opts.Name {
		t.Errorf("Timer Name() = %s, want %s", timer.Name(), opts.Name)
	}
	if timer.Type() != TypeTimer {
		t.Errorf("Timer Type() = %s, want %s", timer.Type(), TypeTimer)
	}
}