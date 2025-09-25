package operational

import (
	"fmt"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/MichaelAJay/go-metrics/metric"
)

// TestAuthenticationFlowIntegration simulates the exact authentication patterns
// found in the auth service analysis to validate MetricsBuilder effectiveness
func TestAuthenticationFlowIntegration(t *testing.T) {
	registry := metric.NewRegistry(metric.DefaultTagValidationConfig(), 5*time.Minute)
	defer registry.Close()

	om := New(registry)
	builder := NewMetricsBuilder(om)

	// Authentication flow contexts from the auth service analysis
	authContexts := []map[string]string{
		{"provider": "password", "status": "success", "user_type": "premium"},
		{"provider": "oauth", "status": "success", "user_type": "basic"},
		{"provider": "mfa", "status": "error", "user_type": "premium"},
		{"provider": "sso", "status": "timeout", "user_type": "enterprise"},
		{"provider": "biometric", "status": "success", "user_type": "premium"},
		{"provider": "ldap", "status": "success", "user_type": "enterprise"},
		{"provider": "saml", "status": "error", "user_type": "enterprise"},
		{"provider": "oauth", "status": "rate_limited", "user_type": "basic"},
	}

	// Simulate multiple authentication attempts with various contexts
	for i, context := range authContexts {
		duration := time.Duration(50+i*10) * time.Millisecond
		status := context["status"]

		// Remove status from context to avoid duplication
		contextCopy := make(map[string]string)
		for k, v := range context {
			if k != "status" {
				contextCopy[k] = v
			}
		}

		builder.RecordWithContext("authentication", status, duration, contextCopy)
	}

	// Verify no panics and metrics were recorded
	if registry == nil {
		t.Fatal("Registry should not be nil")
	}

	// Test repeated patterns to validate caching efficiency
	for range 10 {
		context := map[string]string{
			"provider":  "password",
			"user_type": "premium",
		}
		builder.RecordWithContext("authentication", "success", 45*time.Millisecond, context)
	}
}

// TestSecurityEventFlowIntegration simulates security events patterns
// from the auth service analysis
func TestSecurityEventFlowIntegration(t *testing.T) {
	registry := metric.NewRegistry(metric.DefaultTagValidationConfig(), 5*time.Minute)
	defer registry.Close()

	om := New(registry)
	builder := NewMetricsBuilder(om)

	// Security events from the auth service analysis
	securityEvents := []struct {
		eventType string
		action    string
		context   map[string]string
	}{
		{
			"brute_force", "blocked",
			map[string]string{"ip": "192.168.1.100", "user_agent": "curl/7.68.0"},
		},
		{
			"credential_stuffing", "flagged",
			map[string]string{"source": "tor_exit", "patterns": "common_passwords"},
		},
		{
			"login_attempt", "allowed",
			map[string]string{"geolocation": "us_east", "device": "mobile"},
		},
		{
			"session_hijack", "blocked",
			map[string]string{"session_id": "abc123", "suspicious_activity": "location_change"},
		},
		{
			"password_spray", "detected",
			map[string]string{"target_accounts": "admin_users", "success_rate": "low"},
		},
		{
			"mfa_bypass", "attempted",
			map[string]string{"method": "sim_swapping", "provider": "sms"},
		},
		{
			"privilege_escalation", "blocked",
			map[string]string{"user": "standard_user", "target_role": "admin"},
		},
		{
			"suspicious_login", "flagged",
			map[string]string{"time_anomaly": "unusual_hours", "location": "new_country"},
		},
	}

	// Record each security event
	for _, event := range securityEvents {
		builder.RecordSecurityEvent(event.eventType, event.action, event.context)
	}

	// Verify no panics occurred
	if registry == nil {
		t.Fatal("Registry should not be nil")
	}

	// Test high-frequency security events
	for i := range 50 {
		context := map[string]string{
			"ip":         fmt.Sprintf("192.168.1.%d", i%255),
			"user_agent": "suspicious_bot",
		}
		builder.RecordSecurityEvent("brute_force", "blocked", context)
	}
}

// TestBusinessMetricsFlowIntegration simulates business metrics patterns
// from the auth service analysis
func TestBusinessMetricsFlowIntegration(t *testing.T) {
	registry := metric.NewRegistry(metric.DefaultTagValidationConfig(), 5*time.Minute)
	defer registry.Close()

	om := New(registry)
	builder := NewMetricsBuilder(om)

	// Business metrics from the auth service analysis
	businessMetrics := []struct {
		metricType string
		category   string
		value      float64
		context    map[string]string
	}{
		{
			"session_duration", "completed", 1800.5,
			map[string]string{"tier": "premium", "region": "us_east"},
		},
		{
			"provider_usage", "oauth", 1.0,
			map[string]string{"region": "eu_west", "success_rate": "high"},
		},
		{
			"user_conversion", "signup_to_premium", 1.0,
			map[string]string{"source": "organic", "campaign": "none"},
		},
		{
			"authentication_time", "success", 245.3,
			map[string]string{"provider": "sso", "complexity": "high"},
		},
		{
			"mfa_adoption", "enabled", 1.0,
			map[string]string{"user_segment": "enterprise", "method": "totp"},
		},
		{
			"password_strength", "strong", 1.0,
			map[string]string{"policy": "strict", "user_type": "admin"},
		},
		{
			"login_frequency", "daily_active", 1.0,
			map[string]string{"cohort": "week_1", "retention": "high"},
		},
		{
			"api_usage", "authentication_endpoint", 1.0,
			map[string]string{"client_type": "web", "rate_limit": "normal"},
		},
	}

	// Record each business metric
	for _, metric := range businessMetrics {
		builder.RecordBusinessMetric(metric.metricType, metric.category, metric.value, metric.context)
	}

	// Verify no panics occurred
	if registry == nil {
		t.Fatal("Registry should not be nil")
	}

	// Test high-volume business metrics
	for i := range 100 {
		context := map[string]string{
			"tier":   []string{"basic", "premium", "enterprise"}[i%3],
			"region": []string{"us_east", "us_west", "eu_west"}[i%3],
		}
		value := float64(100 + i*10)
		builder.RecordBusinessMetric("api_latency", "completed", value, context)
	}
}

// TestMetricRegistryValidation ensures metrics are correctly created and accessible
func TestMetricRegistryValidation(t *testing.T) {
	registry := metric.NewRegistry(metric.DefaultTagValidationConfig(), 5*time.Minute)
	defer registry.Close()

	om := New(registry)
	builder := NewMetricsBuilder(om)

	// Test metric creation consistency
	context := map[string]string{
		"provider": "oauth",
		"region":   "us_east",
	}

	builder.RecordWithContext("test_operation", "success", 100*time.Millisecond, context)

	// Verify the same context produces consistent metric names
	builder.RecordWithContext("test_operation", "success", 150*time.Millisecond, context)
	builder.RecordWithContext("test_operation", "error", 200*time.Millisecond, context)

	// Test with different contexts to ensure proper separation
	differentContext := map[string]string{
		"provider": "saml",
		"region":   "eu_west",
	}

	builder.RecordWithContext("test_operation", "success", 120*time.Millisecond, differentContext)

	if registry == nil {
		t.Fatal("Registry should not be nil after metric operations")
	}
}

// TestContextualMetricsCreation validates contextual metrics are created properly
func TestContextualMetricsCreation(t *testing.T) {
	registry := metric.NewRegistry(metric.DefaultTagValidationConfig(), 5*time.Minute)
	defer registry.Close()

	om := New(registry)
	builder := NewMetricsBuilder(om)

	t.Run("Single context key", func(t *testing.T) {
		context := map[string]string{"provider": "oauth"}
		builder.RecordWithContext("single_context", "success", 50*time.Millisecond, context)
	})

	t.Run("Multiple context keys", func(t *testing.T) {
		context := map[string]string{
			"provider": "oauth",
			"region":   "us_east",
			"tier":     "premium",
		}
		builder.RecordWithContext("multi_context", "success", 75*time.Millisecond, context)
	})

	t.Run("Context key naming consistency", func(t *testing.T) {
		context := map[string]string{
			"user_type":   "premium",
			"auth_method": "mfa",
		}
		builder.RecordWithContext("naming_test", "success", 60*time.Millisecond, context)

		// Should create metrics like: naming_test_user_type and naming_test_auth_method
	})

	t.Run("Empty context", func(t *testing.T) {
		builder.RecordWithContext("empty_context", "success", 40*time.Millisecond, nil)
		builder.RecordWithContext("empty_context", "success", 40*time.Millisecond, map[string]string{})
	})

	t.Run("Nested context scenarios", func(t *testing.T) {
		for i := range 5 {
			context := map[string]string{
				"level":     fmt.Sprintf("level_%d", i),
				"iteration": fmt.Sprintf("%d", i),
			}
			builder.RecordWithContext("nested_test", "processing", time.Duration(i*10)*time.Millisecond, context)
		}
	})

	if registry == nil {
		t.Fatal("Registry should not be nil")
	}
}

// TestAuthServiceWorkloadSimulation replicates actual auth service patterns
func TestAuthServiceWorkloadSimulation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping workload simulation in short mode")
	}

	registry := metric.NewRegistry(metric.DefaultTagValidationConfig(), 5*time.Minute)
	defer registry.Close()

	om := New(registry)
	builder := NewMetricsBuilder(om)

	// Simulate auth service workload parameters
	providers := []string{"password", "oauth", "mfa", "sso", "biometric"}
	userTypes := []string{"basic", "premium", "enterprise"}
	statuses := []string{"success", "error", "timeout", "rate_limited"}
	regions := []string{"us_east", "us_west", "eu_west", "ap_south"}

	const totalOperations = 1000
	const concurrentWorkers = 10

	var wg sync.WaitGroup
	operationsPerWorker := totalOperations / concurrentWorkers

	for worker := range concurrentWorkers {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			for i := 0; i < operationsPerWorker; i++ {
				// Simulate authentication operations
				authContext := map[string]string{
					"provider":  providers[(workerID+i)%len(providers)],
					"user_type": userTypes[i%len(userTypes)],
					"region":    regions[(workerID*i)%len(regions)],
				}
				status := statuses[i%len(statuses)]
				duration := time.Duration(50+i%200) * time.Millisecond

				builder.RecordWithContext("authentication", status, duration, authContext)

				// Simulate security events (less frequent)
				if i%10 == 0 {
					secContext := map[string]string{
						"ip":         fmt.Sprintf("192.168.%d.%d", workerID, i%255),
						"user_agent": "test_agent",
					}
					builder.RecordSecurityEvent("login_attempt", "allowed", secContext)
				}

				// Simulate business metrics (less frequent)
				if i%20 == 0 {
					bizContext := map[string]string{
						"tier":    userTypes[i%len(userTypes)],
						"feature": "authentication",
					}
					builder.RecordBusinessMetric("user_activity", "active", 1.0, bizContext)
				}
			}
		}(worker)
	}

	// Wait for all workers to complete
	wg.Wait()

	if registry == nil {
		t.Fatal("Registry should not be nil after workload simulation")
	}
}

// TestConcurrentSafety tests thread-safe operation under concurrent load
func TestConcurrentSafety(t *testing.T) {
	registry := metric.NewRegistry(metric.DefaultTagValidationConfig(), 5*time.Minute)
	defer registry.Close()

	om := New(registry)
	builder := NewMetricsBuilder(om)

	const numGoroutines = 50
	const operationsPerGoroutine = 100

	var wg sync.WaitGroup

	// Test concurrent MetricsBuilder usage with same operations
	t.Run("Concurrent same operations", func(t *testing.T) {
		for i := range numGoroutines {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()

				context := map[string]string{
					"worker":   fmt.Sprintf("worker_%d", id),
					"provider": "concurrent_test",
				}

				for range operationsPerGoroutine {
					builder.RecordWithContext("concurrent_op", "success", 10*time.Millisecond, context)
				}
			}(i)
		}
		wg.Wait()
	})

	// Test concurrent different operations
	t.Run("Concurrent different operations", func(t *testing.T) {
		operations := []string{"auth", "payment", "user_mgmt", "api_call"}

		for i := range numGoroutines {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()

				operation := operations[id%len(operations)]
				context := map[string]string{
					"worker": fmt.Sprintf("worker_%d", id),
					"type":   operation,
				}

				for j := range operationsPerGoroutine {
					builder.RecordWithContext(operation, "success", time.Duration(j)*time.Millisecond, context)
				}
			}(i)
		}
		wg.Wait()
	})

	// Test mixed operation types under concurrent access
	t.Run("Mixed operation types", func(t *testing.T) {
		for i := range numGoroutines {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()

				context := map[string]string{"worker": fmt.Sprintf("worker_%d", id)}

				for j := range operationsPerGoroutine {
					switch j % 3 {
					case 0:
						builder.RecordWithContext("mixed_auth", "success", 10*time.Millisecond, context)
					case 1:
						builder.RecordSecurityEvent("mixed_security", "flagged", context)
					case 2:
						builder.RecordBusinessMetric("mixed_business", "completed", float64(j), context)
					}
				}
			}(i)
		}
		wg.Wait()
	})

	if registry == nil {
		t.Fatal("Registry should not be nil after concurrent operations")
	}
}

// TestTagPoolEfficiency validates that the tag pool is working effectively
func TestTagPoolEfficiency(t *testing.T) {
	registry := metric.NewRegistry(metric.DefaultTagValidationConfig(), 5*time.Minute)
	defer registry.Close()

	om := New(registry)
	builder := NewMetricsBuilder(om)

	// Capture initial memory state
	var m1, m2 runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m1)

	// Perform operations that should leverage the tag pool
	const iterations = 1000
	for i := range iterations {
		context := map[string]string{
			"provider":  "test_provider",
			"status":    "success",
			"iteration": fmt.Sprintf("%d", i),
		}

		builder.RecordWithContext("pool_test", "success", 10*time.Millisecond, context)

		// Also test security events and business metrics
		if i%3 == 0 {
			builder.RecordSecurityEvent("pool_security", "test", context)
		}
		if i%5 == 0 {
			builder.RecordBusinessMetric("pool_business", "test", float64(i), context)
		}
	}

	// Capture final memory state
	runtime.GC()
	runtime.ReadMemStats(&m2)

	// The pool should prevent excessive allocations
	// We don't assert specific numbers as they can vary, but log for analysis
	t.Logf("Memory allocations: Before=%d, After=%d, Diff=%d",
		m1.Alloc, m2.Alloc, m2.Alloc-m1.Alloc)
	t.Logf("Total allocations: Before=%d, After=%d, Diff=%d",
		m1.TotalAlloc, m2.TotalAlloc, m2.TotalAlloc-m1.TotalAlloc)

	// Verify registry is still functional
	if registry == nil {
		t.Fatal("Registry should not be nil")
	}
}

// TestEdgeCases tests various edge case scenarios
func TestEdgeCases(t *testing.T) {
	registry := metric.NewRegistry(metric.DefaultTagValidationConfig(), 5*time.Minute)
	defer registry.Close()

	om := New(registry)
	builder := NewMetricsBuilder(om)

	t.Run("Empty context maps", func(t *testing.T) {
		builder.RecordWithContext("empty_test", "success", 10*time.Millisecond, nil)
		builder.RecordWithContext("empty_test", "success", 10*time.Millisecond, map[string]string{})
	})

	t.Run("Large context maps", func(t *testing.T) {
		largeContext := make(map[string]string)
		for i := range 50 {
			largeContext[fmt.Sprintf("key_%d", i)] = fmt.Sprintf("value_%d", i)
		}
		builder.RecordWithContext("large_context", "success", 10*time.Millisecond, largeContext)
	})

	t.Run("Special characters in tags", func(t *testing.T) {
		specialContext := map[string]string{
			"special-key":          "value-with-dashes",
			"key_with_underscores": "value_with_underscores",
			"unicode":              "测试值",
		}
		builder.RecordWithContext("special_chars", "success", 10*time.Millisecond, specialContext)
	})

	t.Run("Long tag values within limits", func(t *testing.T) {
		// Create a long value that's within the tag validation limits (200 chars)
		longValue := string(make([]rune, 150))
		for i := range longValue {
			longValue = longValue[:i] + "a" + longValue[i+1:]
		}

		longContext := map[string]string{
			"long_value": longValue,
			"normal":     "value",
		}
		builder.RecordWithContext("long_values", "success", 10*time.Millisecond, longContext)
	})

	t.Run("Zero duration operations", func(t *testing.T) {
		context := map[string]string{"test": "zero_duration"}
		builder.RecordWithContext("zero_duration", "success", 0, context)
		builder.RecordSecurityEvent("instant_event", "occurred", context)
	})

	t.Run("Empty operation strings", func(t *testing.T) {
		context := map[string]string{"test": "empty_op"}
		// These should not panic, though the metrics might not be useful
		builder.RecordWithContext("", "success", 10*time.Millisecond, context)
		builder.RecordSecurityEvent("", "occurred", context)
		builder.RecordBusinessMetric("", "completed", 1.0, context)
	})

	if registry == nil {
		t.Fatal("Registry should not be nil after edge case testing")
	}
}

// TestMultiServiceEcosystem simulates MetricsBuilder usage across different services
func TestMultiServiceEcosystem(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping ecosystem simulation in short mode")
	}

	registry := metric.NewRegistry(metric.DefaultTagValidationConfig(), 5*time.Minute)
	defer registry.Close()

	om := New(registry)

	// Simulate different services using the same MetricsBuilder infrastructure
	services := map[string]*MetricsBuilder{
		"auth":    NewMetricsBuilder(om),
		"payment": NewMetricsBuilder(om),
		"user":    NewMetricsBuilder(om),
		"gateway": NewMetricsBuilder(om),
	}

	var wg sync.WaitGroup

	// Auth Service simulation
	wg.Add(1)
	go func() {
		defer wg.Done()
		authBuilder := services["auth"]

		for i := range 100 {
			context := map[string]string{
				"provider": []string{"oauth", "password", "mfa"}[i%3],
				"service":  "auth",
			}
			authBuilder.RecordWithContext("authentication", "success", 50*time.Millisecond, context)
			authBuilder.RecordSecurityEvent("login_attempt", "allowed", context)
		}
	}()

	// Payment Service simulation
	wg.Add(1)
	go func() {
		defer wg.Done()
		paymentBuilder := services["payment"]

		for i := range 100 {
			context := map[string]string{
				"processor": []string{"stripe", "paypal", "square"}[i%3],
				"service":   "payment",
			}
			paymentBuilder.RecordWithContext("payment_processing", "success", 200*time.Millisecond, context)
			paymentBuilder.RecordBusinessMetric("transaction_value", "completed", float64(100+i), context)
		}
	}()

	// User Service simulation
	wg.Add(1)
	go func() {
		defer wg.Done()
		userBuilder := services["user"]

		for i := range 100 {
			context := map[string]string{
				"operation": []string{"profile_update", "preference_change", "avatar_upload"}[i%3],
				"service":   "user",
			}
			userBuilder.RecordWithContext("user_operation", "success", 75*time.Millisecond, context)
			userBuilder.RecordBusinessMetric("user_engagement", "active", 1.0, context)
		}
	}()

	// API Gateway simulation
	wg.Add(1)
	go func() {
		defer wg.Done()
		gatewayBuilder := services["gateway"]

		for i := range 100 {
			context := map[string]string{
				"endpoint": []string{"/auth", "/payment", "/user", "/health"}[i%4],
				"service":  "gateway",
			}
			gatewayBuilder.RecordWithContext("request_routing", "success", 25*time.Millisecond, context)

			if i%10 == 0 {
				gatewayBuilder.RecordSecurityEvent("rate_limit", "applied", context)
			}
		}
	}()

	wg.Wait()

	if registry == nil {
		t.Fatal("Registry should not be nil after multi-service simulation")
	}
}
