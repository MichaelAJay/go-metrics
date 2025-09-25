package operational

import (
	"context"
	"fmt"
	"math/rand"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/MichaelAJay/go-metrics/metric"
)

// TestRealWorldAuthServiceSimulation implements Phase 3: Real-World Simulation
// Simulates sustained auth service workload at 1000 req/sec for 5 minutes
func TestRealWorldAuthServiceSimulation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping long-running simulation test in short mode")
	}

	registry := metric.NewRegistry(metric.DefaultTagValidationConfig(), 10*time.Minute)
	defer registry.Close()

	om := New(registry)
	builder := NewMetricsBuilder(om)

	// Test parameters based on test plan requirements
	const (
		targetRPS   = 1000              // 1000 requests per second
		testDuration = 5 * time.Minute  // 5 minutes sustained load
		workerCount = 20                // Number of concurrent workers
	)

	// Auth service simulation data
	providers := []string{"password", "oauth", "mfa", "sso", "biometric", "ldap", "saml"}
	statuses := []string{"success", "error", "timeout", "rate_limited", "blocked"}
	userTypes := []string{"basic", "premium", "enterprise", "trial"}
	regions := []string{"us-east", "us-west", "eu-central", "ap-southeast", "ca-central"}

	// Security event types from analysis
	securityEvents := []string{"brute_force", "credential_stuffing", "login_attempt",
		"session_hijack", "password_spray", "mfa_bypass", "privilege_escalation", "suspicious_login"}
	securityActions := []string{"blocked", "flagged", "allowed", "detected", "attempted"}

	// Business metrics from analysis
	businessMetrics := []string{"session_duration", "provider_usage", "user_conversion",
		"authentication_success_rate", "mfa_adoption"}
	businessCategories := []string{"completed", "abandoned", "organic", "converted", "premium"}

	ctx, cancel := context.WithTimeout(context.Background(), testDuration)
	defer cancel()

	var (
		totalOps      int64
		authOps       int64
		securityOps   int64
		businessOps   int64
		wg            sync.WaitGroup
		startTime     = time.Now()
	)

	t.Logf("Starting auth service workload simulation: %d workers, %d req/sec target, %v duration",
		workerCount, targetRPS, testDuration)

	// Launch worker goroutines
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			rng := rand.New(rand.NewSource(int64(workerID) + time.Now().UnixNano()))
			opsPerWorker := targetRPS / workerCount
			ticker := time.NewTicker(time.Second / time.Duration(opsPerWorker))
			defer ticker.Stop()

			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					// Simulate auth service operation mix
					operationType := rng.Intn(10)

					switch {
					case operationType < 6: // 60% authentication operations
						simulateAuthOperation(builder, rng, providers, statuses, userTypes, regions)
						atomic.AddInt64(&authOps, 1)

					case operationType < 8: // 20% security events
						simulateSecurityEvent(builder, rng, securityEvents, securityActions)
						atomic.AddInt64(&securityOps, 1)

					default: // 20% business metrics
						simulateBusinessMetric(builder, rng, businessMetrics, businessCategories, regions)
						atomic.AddInt64(&businessOps, 1)
					}

					atomic.AddInt64(&totalOps, 1)
				}
			}
		}(i)
	}

	// Progress reporting goroutine
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				elapsed := time.Since(startTime)
				ops := atomic.LoadInt64(&totalOps)
				rate := float64(ops) / elapsed.Seconds()
				t.Logf("Progress: %v elapsed, %d ops, %.1f ops/sec", elapsed.Truncate(time.Second), ops, rate)
			}
		}
	}()

	// Wait for completion
	wg.Wait()
	elapsed := time.Since(startTime)

	// Report final statistics
	totalOperations := atomic.LoadInt64(&totalOps)
	authOperations := atomic.LoadInt64(&authOps)
	securityOperations := atomic.LoadInt64(&securityOps)
	businessOperations := atomic.LoadInt64(&businessOps)
	actualRPS := float64(totalOperations) / elapsed.Seconds()

	t.Logf("Auth Service Workload Simulation Complete:")
	t.Logf("  Duration: %v", elapsed)
	t.Logf("  Total Operations: %d", totalOperations)
	t.Logf("  Authentication Ops: %d (%.1f%%)", authOperations, float64(authOperations)/float64(totalOperations)*100)
	t.Logf("  Security Ops: %d (%.1f%%)", securityOperations, float64(securityOperations)/float64(totalOperations)*100)
	t.Logf("  Business Ops: %d (%.1f%%)", businessOperations, float64(businessOperations)/float64(totalOperations)*100)
	t.Logf("  Actual Rate: %.1f req/sec (target: %d)", actualRPS, targetRPS)
	t.Logf("  Rate Achievement: %.1f%%", (actualRPS/float64(targetRPS))*100)

	// Validate minimum performance criteria
	if actualRPS < float64(targetRPS)*0.8 {
		t.Errorf("Actual rate %.1f req/sec is below 80%% of target %d req/sec", actualRPS, targetRPS)
	}

	if totalOperations == 0 {
		t.Fatal("No operations were recorded during simulation")
	}

	// Memory check
	runtime.GC()
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	t.Logf("Memory Stats: Alloc=%d KB, TotalAlloc=%d KB, Sys=%d KB, NumGC=%d",
		bToKb(m.Alloc), bToKb(m.TotalAlloc), bToKb(m.Sys), m.NumGC)
}

// TestRealWorldMultiServiceEcosystemSimulation implements Phase 3: Multi-Service Ecosystem Simulation
// Tests MetricsBuilder reusability across different service types
func TestRealWorldMultiServiceEcosystemSimulation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping long-running ecosystem simulation test in short mode")
	}

	registry := metric.NewRegistry(metric.DefaultTagValidationConfig(), 10*time.Minute)
	defer registry.Close()

	om := New(registry)
	builder := NewMetricsBuilder(om)

	// Test parameters
	const (
		testDuration = 2 * time.Minute // Shorter than auth-only test but covers multiple services
		serviceCount = 4               // Auth, Payment, User, API Gateway
		workersPerService = 5          // Concurrent workers per service
	)

	ctx, cancel := context.WithTimeout(context.Background(), testDuration)
	defer cancel()

	var (
		totalOps    int64
		serviceOps  [serviceCount]int64
		wg          sync.WaitGroup
		startTime   = time.Now()
	)

	serviceNames := []string{"Auth", "Payment", "User", "Gateway"}

	t.Logf("Starting multi-service ecosystem simulation: %d services, %d workers each, %v duration",
		serviceCount, workersPerService, testDuration)

	// Launch workers for each service type
	for serviceID := 0; serviceID < serviceCount; serviceID++ {
		for workerID := 0; workerID < workersPerService; workerID++ {
			wg.Add(1)
			go func(svcID, wkrID int) {
				defer wg.Done()

				rng := rand.New(rand.NewSource(int64(svcID*100+wkrID) + time.Now().UnixNano()))

				// Each service has different operation patterns
				switch svcID {
				case 0: // Auth Service
					simulateAuthServiceWorkload(ctx, builder, rng, &serviceOps[svcID])
				case 1: // Payment Service
					simulatePaymentServiceWorkload(ctx, builder, rng, &serviceOps[svcID])
				case 2: // User Service
					simulateUserServiceWorkload(ctx, builder, rng, &serviceOps[svcID])
				case 3: // API Gateway
					simulateGatewayServiceWorkload(ctx, builder, rng, &serviceOps[svcID])
				}
			}(serviceID, workerID)
		}
	}

	// Progress reporting
	go func() {
		ticker := time.NewTicker(20 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				elapsed := time.Since(startTime)
				ops := int64(0)
				for i := 0; i < serviceCount; i++ {
					ops += atomic.LoadInt64(&serviceOps[i])
				}
				rate := float64(ops) / elapsed.Seconds()
				t.Logf("Ecosystem progress: %v elapsed, %d ops, %.1f ops/sec", elapsed.Truncate(time.Second), ops, rate)
			}
		}
	}()

	// Wait for completion
	wg.Wait()
	elapsed := time.Since(startTime)

	// Calculate totals
	for i := 0; i < serviceCount; i++ {
		totalOps += atomic.LoadInt64(&serviceOps[i])
	}

	// Report final statistics
	actualRPS := float64(totalOps) / elapsed.Seconds()

	t.Logf("Multi-Service Ecosystem Simulation Complete:")
	t.Logf("  Duration: %v", elapsed)
	t.Logf("  Total Operations: %d", totalOps)
	for i := 0; i < serviceCount; i++ {
		ops := atomic.LoadInt64(&serviceOps[i])
		t.Logf("  %s Service: %d ops (%.1f%%)", serviceNames[i], ops, float64(ops)/float64(totalOps)*100)
	}
	t.Logf("  Overall Rate: %.1f req/sec", actualRPS)

	// Validate that all services recorded operations
	for i := 0; i < serviceCount; i++ {
		if serviceOps[i] == 0 {
			t.Errorf("Service %s recorded no operations", serviceNames[i])
		}
	}

	if totalOps == 0 {
		t.Fatal("No operations were recorded during ecosystem simulation")
	}

	// Memory check
	runtime.GC()
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	t.Logf("Ecosystem Memory Stats: Alloc=%d KB, TotalAlloc=%d KB, Sys=%d KB, NumGC=%d",
		bToKb(m.Alloc), bToKb(m.TotalAlloc), bToKb(m.Sys), m.NumGC)
}

// Service-specific workload simulation functions

func simulateAuthOperation(builder *MetricsBuilder, rng *rand.Rand, providers, statuses, userTypes, regions []string) {
	provider := providers[rng.Intn(len(providers))]
	status := statuses[rng.Intn(len(statuses))]
	userType := userTypes[rng.Intn(len(userTypes))]
	region := regions[rng.Intn(len(regions))]

	// Realistic auth operation duration (50-500ms)
	duration := time.Duration(50+rng.Intn(450)) * time.Millisecond

	context := map[string]string{
		"provider":  provider,
		"user_type": userType,
		"region":    region,
	}

	builder.RecordWithContext("authentication", status, duration, context)
}

func simulateSecurityEvent(builder *MetricsBuilder, rng *rand.Rand, events, actions []string) {
	event := events[rng.Intn(len(events))]
	action := actions[rng.Intn(len(actions))]

	context := map[string]string{
		"severity": []string{"low", "medium", "high", "critical"}[rng.Intn(4)],
		"source":   []string{"internal", "external", "automated", "manual"}[rng.Intn(4)],
	}

	builder.RecordSecurityEvent(event, action, context)
}

func simulateBusinessMetric(builder *MetricsBuilder, rng *rand.Rand, metrics, categories, regions []string) {
	metric := metrics[rng.Intn(len(metrics))]
	category := categories[rng.Intn(len(categories))]
	region := regions[rng.Intn(len(regions))]

	// Realistic business metric value
	value := 100.0 + rng.Float64()*1000.0

	context := map[string]string{
		"region": region,
		"tier":   []string{"basic", "premium", "enterprise"}[rng.Intn(3)],
	}

	builder.RecordBusinessMetric(metric, category, value, context)
}

// Service workload simulators for ecosystem test

func simulateAuthServiceWorkload(ctx context.Context, builder *MetricsBuilder, rng *rand.Rand, counter *int64) {
	providers := []string{"oauth", "saml", "ldap", "password"}
	statuses := []string{"success", "error", "timeout"}

	ticker := time.NewTicker(10 * time.Millisecond) // 100 ops/sec per worker
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// Mix of auth operations and security events
			if rng.Intn(10) < 7 { // 70% auth operations
				provider := providers[rng.Intn(len(providers))]
				status := statuses[rng.Intn(len(statuses))]
				duration := time.Duration(20+rng.Intn(200)) * time.Millisecond

				context := map[string]string{
					"provider": provider,
					"client":   fmt.Sprintf("client_%d", rng.Intn(10)),
				}

				builder.RecordWithContext("auth_login", status, duration, context)
			} else { // 30% security events
				events := []string{"failed_login", "suspicious_activity"}
				event := events[rng.Intn(len(events))]
				action := []string{"blocked", "flagged"}[rng.Intn(2)]

				context := map[string]string{
					"ip_range": fmt.Sprintf("192.168.%d.0/24", rng.Intn(10)),
				}

				builder.RecordSecurityEvent(event, action, context)
			}
			atomic.AddInt64(counter, 1)
		}
	}
}

func simulatePaymentServiceWorkload(ctx context.Context, builder *MetricsBuilder, rng *rand.Rand, counter *int64) {
	methods := []string{"credit_card", "paypal", "bank_transfer", "crypto"}
	statuses := []string{"completed", "failed", "pending", "cancelled"}

	ticker := time.NewTicker(15 * time.Millisecond) // ~67 ops/sec per worker
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			method := methods[rng.Intn(len(methods))]
			status := statuses[rng.Intn(len(statuses))]
			amount := 10.0 + rng.Float64()*1000.0 // $10-$1000
			duration := time.Duration(100+rng.Intn(2000)) * time.Millisecond

			context := map[string]string{
				"method":   method,
				"currency": []string{"USD", "EUR", "GBP"}[rng.Intn(3)],
				"merchant": fmt.Sprintf("merchant_%d", rng.Intn(5)),
			}

			builder.RecordWithContext("payment_processing", status, duration, context)
			builder.RecordBusinessMetric("transaction_amount", status, amount, context)

			atomic.AddInt64(counter, 2) // Two operations recorded
		}
	}
}

func simulateUserServiceWorkload(ctx context.Context, builder *MetricsBuilder, rng *rand.Rand, counter *int64) {
	operations := []string{"profile_update", "preferences_change", "account_deletion", "password_reset"}
	statuses := []string{"success", "error", "validation_failed"}

	ticker := time.NewTicker(20 * time.Millisecond) // 50 ops/sec per worker
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			operation := operations[rng.Intn(len(operations))]
			status := statuses[rng.Intn(len(statuses))]
			duration := time.Duration(30+rng.Intn(300)) * time.Millisecond

			context := map[string]string{
				"user_tier": []string{"free", "premium", "enterprise"}[rng.Intn(3)],
				"region":    []string{"us", "eu", "asia"}[rng.Intn(3)],
			}

			builder.RecordWithContext(operation, status, duration, context)
			atomic.AddInt64(counter, 1)
		}
	}
}

func simulateGatewayServiceWorkload(ctx context.Context, builder *MetricsBuilder, rng *rand.Rand, counter *int64) {
	routes := []string{"/api/v1/users", "/api/v1/orders", "/api/v1/products", "/api/v1/auth"}
	methods := []string{"GET", "POST", "PUT", "DELETE"}
	statuses := []string{"200", "400", "401", "404", "500", "503"}

	ticker := time.NewTicker(5 * time.Millisecond) // 200 ops/sec per worker (gateway handles most traffic)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			route := routes[rng.Intn(len(routes))]
			method := methods[rng.Intn(len(methods))]
			status := statuses[rng.Intn(len(statuses))]
			duration := time.Duration(1+rng.Intn(50)) * time.Millisecond // Fast gateway responses

			context := map[string]string{
				"route":        route,
				"method":       method,
				"client_type":  []string{"web", "mobile", "api"}[rng.Intn(3)],
				"rate_limited": []string{"true", "false"}[rng.Intn(2)], // 50/50 true/false for context
			}

			builder.RecordWithContext("gateway_request", status, duration, context)

			// Rate limiting metrics
			if rng.Intn(100) < 5 { // 5% rate limiting events
				builder.RecordSecurityEvent("rate_limit", "applied", map[string]string{
					"client_ip": fmt.Sprintf("10.0.%d.%d", rng.Intn(255), rng.Intn(255)),
					"reason":    []string{"burst", "sustained", "suspicious"}[rng.Intn(3)],
				})
				atomic.AddInt64(counter, 1)
			}

			atomic.AddInt64(counter, 1)
		}
	}
}

// Utility function
func bToKb(b uint64) uint64 {
	return b / 1024
}