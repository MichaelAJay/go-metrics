package metrics

import (
	"testing"

	"github.com/MichaelAJay/go-metrics/metric"
)

func TestGlobalFunctions(t *testing.T) {
	// Test NewRegistry function
	reg := NewRegistry()
	if reg == nil {
		t.Fatal("NewRegistry() returned nil")
	}

	// Test GlobalRegistry function
	globalReg := GlobalRegistry()
	if globalReg == nil {
		t.Fatal("GlobalRegistry() returned nil")
	}
	if globalReg != metric.GlobalRegistry {
		t.Error("GlobalRegistry() did not return the expected registry")
	}

	// Test type constants
	if TypeCounter != metric.TypeCounter {
		t.Error("TypeCounter does not match the expected value")
	}
	if TypeGauge != metric.TypeGauge {
		t.Error("TypeGauge does not match the expected value")
	}
	if TypeHistogram != metric.TypeHistogram {
		t.Error("TypeHistogram does not match the expected value")
	}
	if TypeTimer != metric.TypeTimer {
		t.Error("TypeTimer does not match the expected value")
	}

	// Test creating metrics directly from the global registry
	registry := GlobalRegistry()

	counter := registry.Counter(Options{Name: "test_global_counter"})
	if counter == nil {
		t.Error("registry.Counter() returned nil")
	}

	gauge := registry.Gauge(Options{Name: "test_global_gauge"})
	if gauge == nil {
		t.Error("registry.Gauge() returned nil")
	}

	histogram := registry.Histogram(Options{Name: "test_global_histogram"})
	if histogram == nil {
		t.Error("registry.Histogram() returned nil")
	}

	timer := registry.Timer(Options{Name: "test_global_timer"})
	if timer == nil {
		t.Error("registry.Timer() returned nil")
	}
}
