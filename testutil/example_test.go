package testutil_test

import (
	"testing"
	"time"

	"github.com/MichaelAJay/go-metrics/metric"
	"github.com/MichaelAJay/go-metrics/testutil"
)

// Example service that uses metrics
type ExampleService struct {
	registry        metric.Registry
	requestCounter  metric.Counter
	activeGauge     metric.Gauge
	durationTimer   metric.Timer
}

func NewExampleService(registry metric.Registry) *ExampleService {
	return &ExampleService{
		registry:        registry,
		requestCounter:  registry.Counter(metric.Options{Name: "requests_total"}),
		activeGauge:     registry.Gauge(metric.Options{Name: "active_requests"}),
		durationTimer:   registry.Timer(metric.Options{Name: "request_duration"}),
	}
}

func (s *ExampleService) ProcessRequest() {
	s.requestCounter.Inc()
	s.activeGauge.Inc()
	
	defer func() {
		s.activeGauge.Dec()
	}()
	
	// Simulate processing time
	duration := s.durationTimer.Time(func() {
		time.Sleep(10 * time.Millisecond)
	})
	
	// Also record the duration directly
	s.durationTimer.Record(duration)
}

func (s *ExampleService) ProcessRequestWithTags(method, status string) {
	tags := metric.Tags{
		"method": method,
		"status": status,
	}
	
	s.requestCounter.With(tags).Inc()
	s.durationTimer.With(tags).Record(5 * time.Millisecond)
}

// TestExampleService demonstrates how to test a service using mock metrics
func TestExampleService(t *testing.T) {
	// Create a mock registry
	mockRegistry := testutil.NewMockRegistry()
	
	// Create the service with mock metrics
	service := NewExampleService(mockRegistry)
	
	// Process a request
	service.ProcessRequest()
	
	// Verify metrics were recorded correctly
	requestCounter := mockRegistry.GetCounter("requests_total")
	if requestCounter == nil {
		t.Fatal("Expected requests_total counter to exist")
	}
	
	testutil.AssertCounterValue(t, requestCounter, 1)
	testutil.AssertCounterIncCalls(t, requestCounter, 1)
	
	// Verify gauge operations
	activeGauge := mockRegistry.GetGauge("active_requests")
	if activeGauge == nil {
		t.Fatal("Expected active_requests gauge to exist")
	}
	
	// Gauge should be incremented then decremented (back to 0)
	testutil.AssertGaugeValue(t, activeGauge, 0)
	if activeGauge.IncCalls() != 1 {
		t.Errorf("Expected 1 gauge increment, got %d", activeGauge.IncCalls())
	}
	if activeGauge.DecCalls() != 1 {
		t.Errorf("Expected 1 gauge decrement, got %d", activeGauge.DecCalls())
	}
	
	// Verify timer was used
	durationTimer := mockRegistry.GetTimer("request_duration")
	if durationTimer == nil {
		t.Fatal("Expected request_duration timer to exist")
	}
	
	// Should have 2 record calls (Time() + explicit Record())
	testutil.AssertTimerRecordCalls(t, durationTimer, 2)
	
	// Verify all durations are reasonable (less than 100ms)
	testutil.AssertTimerRecordCallsWithin(t, durationTimer, 0, 100*time.Millisecond)
}

// TestExampleServiceWithTags demonstrates testing tagged metrics
func TestExampleServiceWithTags(t *testing.T) {
	mockRegistry := testutil.NewMockRegistry()
	service := NewExampleService(mockRegistry)
	
	// Process requests with different tags
	service.ProcessRequestWithTags("GET", "200")
	service.ProcessRequestWithTags("POST", "201")
	service.ProcessRequestWithTags("GET", "404")
	
	// Verify the With() calls were made with correct tags
	requestCounter := mockRegistry.GetCounter("requests_total")
	withCalls := requestCounter.WithCalls()
	
	expectedTags := []metric.Tags{
		{"method": "GET", "status": "200"},
		{"method": "POST", "status": "201"},
		{"method": "GET", "status": "404"},
	}
	
	testutil.AssertWithCalls(t, withCalls, expectedTags)
	
	// Verify timer also received tagged calls
	durationTimer := mockRegistry.GetTimer("request_duration")
	timerWithCalls := durationTimer.WithCalls()
	testutil.AssertWithCalls(t, timerWithCalls, expectedTags)
}

// TestRegistryCallTracking demonstrates registry call tracking
func TestRegistryCallTracking(t *testing.T) {
	mockRegistry := testutil.NewMockRegistry()
	
	// Create various metrics
	mockRegistry.Counter(metric.Options{Name: "counter1"})
	mockRegistry.Counter(metric.Options{Name: "counter2"})
	mockRegistry.Gauge(metric.Options{Name: "gauge1"})
	mockRegistry.Histogram(metric.Options{Name: "histogram1"})
	mockRegistry.Timer(metric.Options{Name: "timer1"})
	
	// Verify call counts
	testutil.AssertRegistryCallCounts(t, mockRegistry, 2, 1, 1, 1)
	
	// Test unregister
	mockRegistry.Unregister("counter1")
	if len(mockRegistry.UnregisterCalls) != 1 {
		t.Errorf("Expected 1 unregister call, got %d", len(mockRegistry.UnregisterCalls))
	}
}

// TestCallbacksAndCustomBehavior demonstrates using callbacks for custom test behavior
func TestCallbacksAndCustomBehavior(t *testing.T) {
	mockRegistry := testutil.NewMockRegistry()
	
	// Set up a callback to track when counters are incremented
	var incrementCalls int
	mockRegistry.OnCounterCallback = func(opts metric.Options) metric.Counter {
		counter := testutil.NewMockCounter(opts)
		counter.OnIncCallback = func() {
			incrementCalls++
		}
		return counter
	}
	
	// Create and use a counter
	counter := mockRegistry.Counter(metric.Options{Name: "test_counter"})
	counter.Inc()
	counter.Inc()
	counter.Inc()
	
	// Verify our callback was called
	if incrementCalls != 3 {
		t.Errorf("Expected 3 increment callback calls, got %d", incrementCalls)
	}
}

// TestResetFunctionality demonstrates resetting mock state
func TestResetFunctionality(t *testing.T) {
	mockRegistry := testutil.NewMockRegistry()
	
	// Create and use metrics
	counter := mockRegistry.Counter(metric.Options{Name: "test_counter"})
	mockCounter := counter.(*testutil.MockCounter)
	
	mockCounter.Inc()
	mockCounter.Add(5.0)
	
	// Verify initial state
	testutil.AssertCounterValue(t, mockCounter, 6)
	testutil.AssertCounterIncCalls(t, mockCounter, 1)
	
	// Reset the counter
	mockCounter.Reset()
	
	// Verify reset state
	testutil.AssertCounterValue(t, mockCounter, 0)
	testutil.AssertCounterIncCalls(t, mockCounter, 0)
	if len(mockCounter.AddCalls()) != 0 {
		t.Error("Expected no Add calls after reset")
	}
	
	// Reset the entire registry
	mockRegistry.Reset()
	
	// Verify registry is clean
	testutil.AssertRegistryCallCounts(t, mockRegistry, 0, 0, 0, 0)
}

// BenchmarkMockMetrics demonstrates benchmarking with mock metrics
func BenchmarkMockMetrics(b *testing.B) {
	scenarios := testutil.BenchmarkScenarios()
	
	for _, scenario := range scenarios {
		b.Run(scenario.Name, func(b *testing.B) {
			mockRegistry := testutil.NewMockRegistry()
			metric := scenario.Setup(mockRegistry)
			
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				scenario.Operation(metric)
			}
		})
	}
}