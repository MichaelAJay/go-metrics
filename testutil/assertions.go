package testutil

import (
	"testing"
	"time"

	"github.com/MichaelAJay/go-metrics/metric"
)

// AssertCounterValue verifies that a counter has the expected value.
func AssertCounterValue(t *testing.T, counter *MockCounter, expected uint64) {
	t.Helper()
	if actual := counter.Value(); actual != expected {
		t.Errorf("Expected counter value %d, got %d", expected, actual)
	}
}

// AssertCounterIncCalls verifies the number of Inc() calls on a counter.
func AssertCounterIncCalls(t *testing.T, counter *MockCounter, expected int) {
	t.Helper()
	if actual := counter.IncCalls(); actual != expected {
		t.Errorf("Expected %d Inc() calls, got %d", expected, actual)
	}
}

// AssertCounterAddCalls verifies the Add() calls on a counter.
func AssertCounterAddCalls(t *testing.T, counter *MockCounter, expected []float64) {
	t.Helper()
	actual := counter.AddCalls()
	if len(actual) != len(expected) {
		t.Errorf("Expected %d Add() calls, got %d", len(expected), len(actual))
		return
	}
	for i, exp := range expected {
		if actual[i] != exp {
			t.Errorf("Add() call %d: expected %f, got %f", i, exp, actual[i])
		}
	}
}

// AssertGaugeValue verifies that a gauge has the expected value.
func AssertGaugeValue(t *testing.T, gauge *MockGauge, expected int64) {
	t.Helper()
	if actual := gauge.Value(); actual != expected {
		t.Errorf("Expected gauge value %d, got %d", expected, actual)
	}
}

// AssertGaugeSetCalls verifies the Set() calls on a gauge.
func AssertGaugeSetCalls(t *testing.T, gauge *MockGauge, expected []float64) {
	t.Helper()
	actual := gauge.SetCalls()
	if len(actual) != len(expected) {
		t.Errorf("Expected %d Set() calls, got %d", len(expected), len(actual))
		return
	}
	for i, exp := range expected {
		if actual[i] != exp {
			t.Errorf("Set() call %d: expected %f, got %f", i, exp, actual[i])
		}
	}
}

// AssertHistogramObserveCalls verifies the Observe() calls on a histogram.
func AssertHistogramObserveCalls(t *testing.T, histogram *MockHistogram, expected []float64) {
	t.Helper()
	actual := histogram.ObserveCalls()
	if len(actual) != len(expected) {
		t.Errorf("Expected %d Observe() calls, got %d", len(expected), len(actual))
		return
	}
	for i, exp := range expected {
		if actual[i] != exp {
			t.Errorf("Observe() call %d: expected %f, got %f", i, exp, actual[i])
		}
	}
}

// AssertHistogramSnapshot verifies histogram statistics.
func AssertHistogramSnapshot(t *testing.T, histogram *MockHistogram, expectedCount uint64, expectedSum uint64) {
	t.Helper()
	snapshot := histogram.Snapshot()
	if snapshot.Count != expectedCount {
		t.Errorf("Expected histogram count %d, got %d", expectedCount, snapshot.Count)
	}
	if snapshot.Sum != expectedSum {
		t.Errorf("Expected histogram sum %d, got %d", expectedSum, snapshot.Sum)
	}
}

// AssertTimerRecordCalls verifies the Record() calls on a timer.
func AssertTimerRecordCalls(t *testing.T, timer *MockTimer, expectedCount int) {
	t.Helper()
	actual := timer.RecordCalls()
	if len(actual) != expectedCount {
		t.Errorf("Expected %d Record() calls, got %d", expectedCount, len(actual))
	}
}

// AssertTimerRecordCallsWithin verifies that all Record() calls are within expected duration ranges.
func AssertTimerRecordCallsWithin(t *testing.T, timer *MockTimer, min, max time.Duration) {
	t.Helper()
	calls := timer.RecordCalls()
	for i, duration := range calls {
		if duration < min || duration > max {
			t.Errorf("Record() call %d: duration %v not within range [%v, %v]", i, duration, min, max)
		}
	}
}

// AssertMetricTags verifies that a metric has the expected tags.
func AssertMetricTags(t *testing.T, m metric.Metric, expected metric.Tags) {
	t.Helper()
	actual := m.Tags()
	if len(actual) != len(expected) {
		t.Errorf("Expected %d tags, got %d", len(expected), len(actual))
		return
	}
	for key, expectedValue := range expected {
		if actualValue, exists := actual[key]; !exists {
			t.Errorf("Expected tag %s not found", key)
		} else if actualValue != expectedValue {
			t.Errorf("Tag %s: expected %s, got %s", key, expectedValue, actualValue)
		}
	}
}

// AssertWithCalls verifies the With() calls and their tags.
func AssertWithCalls(t *testing.T, withCalls []metric.Tags, expected []metric.Tags) {
	t.Helper()
	if len(withCalls) != len(expected) {
		t.Errorf("Expected %d With() calls, got %d", len(expected), len(withCalls))
		return
	}
	for i, expectedTags := range expected {
		actualTags := withCalls[i]
		if len(actualTags) != len(expectedTags) {
			t.Errorf("With() call %d: expected %d tags, got %d", i, len(expectedTags), len(actualTags))
			continue
		}
		for key, expectedValue := range expectedTags {
			if actualValue, exists := actualTags[key]; !exists {
				t.Errorf("With() call %d: expected tag %s not found", i, key)
			} else if actualValue != expectedValue {
				t.Errorf("With() call %d: tag %s expected %s, got %s", i, key, expectedValue, actualValue)
			}
		}
	}
}

// AssertRegistryCallCounts verifies the number of metric creation calls on a registry.
func AssertRegistryCallCounts(t *testing.T, registry *MockRegistry, expectedCounters, expectedGauges, expectedHistograms, expectedTimers int) {
	t.Helper()
	if len(registry.CounterCalls) != expectedCounters {
		t.Errorf("Expected %d Counter() calls, got %d", expectedCounters, len(registry.CounterCalls))
	}
	if len(registry.GaugeCalls) != expectedGauges {
		t.Errorf("Expected %d Gauge() calls, got %d", expectedGauges, len(registry.GaugeCalls))
	}
	if len(registry.HistogramCalls) != expectedHistograms {
		t.Errorf("Expected %d Histogram() calls, got %d", expectedHistograms, len(registry.HistogramCalls))
	}
	if len(registry.TimerCalls) != expectedTimers {
		t.Errorf("Expected %d Timer() calls, got %d", expectedTimers, len(registry.TimerCalls))
	}
}

// MetricExists checks if a metric exists in the registry and returns it.
func MetricExists(t *testing.T, registry *MockRegistry, name string, metricType metric.Type) metric.Metric {
	t.Helper()
	switch metricType {
	case metric.TypeCounter:
		if counter := registry.GetCounter(name); counter != nil {
			return counter
		}
	case metric.TypeGauge:
		if gauge := registry.GetGauge(name); gauge != nil {
			return gauge
		}
	case metric.TypeHistogram:
		if histogram := registry.GetHistogram(name); histogram != nil {
			return histogram
		}
	case metric.TypeTimer:
		if timer := registry.GetTimer(name); timer != nil {
			return timer
		}
	}
	t.Errorf("Metric %s of type %s not found", name, metricType)
	return nil
}

// PrintMetricSummary prints a summary of all metrics in the registry for debugging.
func PrintMetricSummary(t *testing.T, registry *MockRegistry) {
	t.Helper()
	t.Logf("Registry Summary:")
	t.Logf("  Counters: %d", len(registry.CounterCalls))
	t.Logf("  Gauges: %d", len(registry.GaugeCalls))
	t.Logf("  Histograms: %d", len(registry.HistogramCalls))
	t.Logf("  Timers: %d", len(registry.TimerCalls))
	t.Logf("  Unregister calls: %d", len(registry.UnregisterCalls))
	t.Logf("  Each calls: %d", registry.EachCalls)
}
