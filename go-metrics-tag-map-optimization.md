# Tag Map Allocation Optimization

## Problem Statement

**Current Impact**: 150+ allocations per operation from tag map creation
- Every `Record*()` call creates new `map[string]string` for tags
- `maps.Copy` operations cause excessive heap allocations
- Accounts for 40-50% of total allocations in high-frequency scenarios

## Root Cause

```go
// Current pattern causing allocations
func RecordOperation(name, status string, duration time.Duration) {
    tags := map[string]string{          // ❌ New allocation
        "operation": name,
        "status": status,
    }
    // Additional maps.Copy allocations during processing
}
```

## Proposed Solution

**Internal Tag Map Pooling**: Implement transparent pooling within go-metrics

### Core Changes

1. **Internal tag map pool**:
   ```go
   var tagMapPool = sync.Pool{
       New: func() interface{} {
           return make(map[string]string, 8) // Reasonable default capacity
       },
   }
   ```

2. **Transparent pooling in Record methods**:
   ```go
   func (m *Metrics) RecordOperation(name, status string, duration time.Duration) {
       tags := tagMapPool.Get().(map[string]string)
       defer tagMapPool.Put(clearMap(tags))

       tags["operation"] = name
       tags["status"] = status
       // Process with pooled map
   }
   ```

3. **Safe map clearing utility**:
   ```go
   func clearMap(m map[string]string) map[string]string {
       for k := range m {
           delete(m, k)
       }
       return m
   }
   ```

## Expected Impact

- **Eliminate 150+ allocations per operation** in high-frequency scenarios
- **Reduce memory pressure** by 40-50% for metrics-heavy applications
- **Zero API changes** - optimization is completely internal
- **Thread-safe** through sync.Pool design

## Implementation Priority

**HIGH** - This is a foundational performance issue affecting all go-metrics consumers. The optimization should be transparent to existing code while providing significant performance benefits.

## Benchmarking Plan

**Establish baseline measurements at root allocation sources before optimization**

### Core Allocation Benchmarks

Target the closest public APIs that exercise `copyTags()` and tag allocation:

1. **`.With()` method benchmarks** (direct `copyTags()` exercisers) - ✅ **IMPLEMENTED**:
   ```go
   func BenchmarkCounterWith(b *testing.B)     // 416 B/op, 3 allocs/op
   func BenchmarkGaugeWith(b *testing.B)       // 416 B/op, 3 allocs/op
   func BenchmarkHistogramWith(b *testing.B)   // 576 B/op, 4 allocs/op
   func BenchmarkTimerWith(b *testing.B)       // 592 B/op, 5 allocs/op
   ```

   **Baseline Results** (from `metric/with_benchmark_test.go`):
   - Counter/Gauge: 416 bytes and 3 allocations per With() call
   - Histogram: 576 bytes and 4 allocations (includes bucket allocation)
   - Timer: 592 bytes and 5 allocations (highest overhead due to histogram complexity)
   - All benchmarks directly target `copyTags()` allocation source per metrics.go:42-58

2. **`.Tags()` method benchmarks** (defensive copy allocations):
   ```go
   func BenchmarkCounterTags(b *testing.B)
   func BenchmarkGaugeTags(b *testing.B)
   func BenchmarkHistogramTags(b *testing.B)
   func BenchmarkTimerTags(b *testing.B)
   ```

3. **Operational package benchmarks** (high-level allocation patterns):
   ```go
   func BenchmarkRecordError(b *testing.B)
   func BenchmarkRecordOperation(b *testing.B)
   ```

### Allocation-Focused Scenarios

**Scenario 1: Chained `.With()` calls** (compounds `copyTags()` allocations):
```go
func BenchmarkChainedWith(b *testing.B) {
    counter := registry.Counter(Options{Name: "test"})
    for i := 0; i < b.N; i++ {
        counter.With(Tags{"service": "api"}).
                With(Tags{"method": "POST"}).
                With(Tags{"status": "200"}).Inc()
    }
}
```

**Scenario 2: High-frequency operational recording**:
```go
func BenchmarkHighFrequencyOperations(b *testing.B) {
    om := operational.New(registry)
    operations := []string{"auth", "query", "cache"}
    statuses := []string{"success", "error", "timeout"}

    for i := 0; i < b.N; i++ {
        op := operations[i%len(operations)]
        status := statuses[i%len(statuses)]
        om.RecordOperation(op, status, time.Millisecond)
    }
}
```

### Memory Allocation Measurement

**Key metrics to track**:
- `AllocsPerOp` - Total allocations per operation
- `BytesPerOp` - Total bytes allocated per operation
- `MemAllocs` - Allocation count during benchmark
- `MemBytes` - Total memory allocated

**Example measurement**:
```bash
go test -bench=. -benchmem -count=5 ./metric > before.txt
# Apply optimization
go test -bench=. -benchmem -count=5 ./metric > after.txt
benchcmp before.txt after.txt
```

### Success Criteria

**Target improvements**:
- **>90% reduction** in `AllocsPerOp` for `.With()` operations
- **>80% reduction** in `BytesPerOp` for tag-heavy scenarios
- **>70% reduction** in operational package allocation overhead
- **Zero functional regressions** in existing test suite

## Post-Mortem: Benchmarking Methodology Issue

### Issue Discovery

During benchmark implementation, we discovered that the "150+ allocations per operation" problem was primarily a **benchmarking artifact** rather than a production performance issue.

**Root Cause Analysis:**
1. **Operational package uses metric caching** (operational.go:84-90, 122-127, 159-166)
2. **High allocations occur only during metric initialization**, not steady-state operations
3. **Benchmarks were measuring cold-start scenarios** instead of realistic production usage

**Evidence from operational_benchmark_test.go:**

```
Cold Start (unrealistic):
- RecordError: 1096 B/op, 18 allocs/op
- RecordOperation: 1816 B/op, 29 allocs/op

Cached Metrics (production reality):
- RecordError: 96 B/op, 4 allocs/op
- RecordOperation: 88 B/op, 5 allocs/op
```

**Conclusion:** The tag map pooling optimization would provide minimal production benefit since expensive allocations only occur during metric initialization, not during steady-state operation.

### Corrected Benchmarking Methodology

**Problem:** Most Go benchmarks inadvertently measure initialization overhead by creating fresh objects in each iteration.

**Solution:** Pre-warm caches and separate initialization from measurement:

#### Before (Incorrect - measures initialization):
```go
func BenchmarkOperationalMetrics(b *testing.B) {
    registry := metric.NewDefaultRegistry()

    for i := 0; i < b.N; i++ {
        om := New(registry) // ❌ Fresh instance each iteration
        om.RecordError("auth", "validation_error", "invalid_token")
    }
}
```

#### After (Correct - measures steady-state):
```go
func BenchmarkOperationalMetrics(b *testing.B) {
    registry := metric.NewDefaultRegistry()
    om := New(registry)

    // Pre-warm cache (excluded from timing)
    om.RecordError("auth", "validation_error", "invalid_token")

    b.ResetTimer() // ✅ Start timing after initialization
    b.ReportAllocs()

    for i := 0; i < b.N; i++ {
        om.RecordError("auth", "validation_error", "invalid_token")
    }
}
```

#### For Varied Workloads (Pre-warm all combinations):
```go
func BenchmarkVariedOperationalMetrics(b *testing.B) {
    registry := metric.NewDefaultRegistry()
    om := New(registry)

    operations := []string{"auth", "query", "cache"}
    statuses := []string{"success", "error", "timeout"}

    // Pre-warm cache for all combinations
    for _, op := range operations {
        for _, status := range statuses {
            om.RecordOperation(op, status, time.Millisecond)
        }
    }

    b.ResetTimer() // Start timing after cache warm-up
    b.ReportAllocs()

    for i := 0; i < b.N; i++ {
        op := operations[i%len(operations)]
        status := statuses[i%len(statuses)]
        om.RecordOperation(op, status, time.Millisecond)
    }
}
```

### Key Benchmarking Principles

1. **Separate initialization from measurement** using `b.ResetTimer()`
2. **Pre-warm caches** to reflect production usage patterns
3. **Use realistic workload patterns** rather than worst-case scenarios
4. **Measure steady-state performance**, not cold-start overhead
5. **Validate benchmarks against production behavior** before optimizing

### Recommendation

**Abandon the tag map pooling optimization.** The operational package's existing caching strategy already solves the allocation problem for production workloads. Focus optimization efforts on areas with genuine steady-state performance issues.

## Verification

Success metrics:
- Benchmark showing >90% reduction in tag-related allocations
- No functional regressions in existing test suite
- Memory profiling confirms reduced GC pressure