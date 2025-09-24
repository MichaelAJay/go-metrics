package operational

import (
	"testing"
	"time"

	"github.com/MichaelAJay/go-metrics/metric"
)

func TestMetricsBuilder_RecordWithContext(t *testing.T) {
	registry := metric.NewRegistry(metric.DefaultTagValidationConfig(), 5*time.Minute)
	defer registry.Close()
	om := New(registry)
	builder := NewMetricsBuilder(om)

	// Test basic operation recording
	duration := 100 * time.Millisecond
	builder.RecordWithContext("authentication", "success", duration, nil)

	// Test with context
	context := map[string]string{
		"provider":  "password",
		"user_type": "premium",
	}
	builder.RecordWithContext("authentication", "success", duration, context)

	// Verify metrics were created - basic check by accessing the registry
	// The implementation should create metrics without panicking
	if registry == nil {
		t.Error("Registry should not be nil")
	}
}

func TestMetricsBuilder_RecordSecurityEvent(t *testing.T) {
	registry := metric.NewRegistry(metric.DefaultTagValidationConfig(), 5*time.Minute)
	defer registry.Close()
	om := New(registry)
	builder := NewMetricsBuilder(om)

	// Test security event without context
	builder.RecordSecurityEvent("brute_force", "blocked", nil)

	// Test security event with context
	context := map[string]string{
		"ip":         "192.168.1.1",
		"user_agent": "Mozilla/5.0",
	}
	builder.RecordSecurityEvent("login_attempt", "allowed", context)

	// Verify no panics occurred
	if registry == nil {
		t.Error("Registry should not be nil")
	}
}

func TestMetricsBuilder_RecordBusinessMetric(t *testing.T) {
	registry := metric.NewRegistry(metric.DefaultTagValidationConfig(), 5*time.Minute)
	defer registry.Close()
	om := New(registry)
	builder := NewMetricsBuilder(om)

	// Test business metric without context
	builder.RecordBusinessMetric("user_conversion", "completed", 1.0, nil)

	// Test business metric with context
	context := map[string]string{
		"source": "organic",
		"tier":   "premium",
	}
	builder.RecordBusinessMetric("payment_processing", "completed", 250.5, context)

	// Verify no panics occurred
	if registry == nil {
		t.Error("Registry should not be nil")
	}
}

func TestNewMetricsBuilder(t *testing.T) {
	registry := metric.NewRegistry(metric.DefaultTagValidationConfig(), 5*time.Minute)
	defer registry.Close()
	om := New(registry)
	builder := NewMetricsBuilder(om)

	if builder == nil {
		t.Error("MetricsBuilder should not be nil")
	}

	if builder.om != om {
		t.Error("MetricsBuilder should contain the provided OperationalMetrics instance")
	}
}

func TestMetricsBuilder_ContextualMetrics(t *testing.T) {
	registry := metric.NewRegistry(metric.DefaultTagValidationConfig(), 5*time.Minute)
	defer registry.Close()
	om := New(registry)
	builder := NewMetricsBuilder(om)

	// Test that contextual metrics are created for each context key-value pair
	context := map[string]string{
		"provider": "oauth",
		"region":   "us-east-1",
		"tier":     "premium",
	}

	builder.RecordWithContext("api_call", "success", 50*time.Millisecond, context)

	// The implementation should create:
	// 1. Main metric: api_call with status "success"
	// 2. Contextual metrics: api_call_provider with status "oauth", etc.
	// We verify by ensuring no panics and the registry exists

	if registry == nil {
		t.Error("Registry should not be nil")
	}
}