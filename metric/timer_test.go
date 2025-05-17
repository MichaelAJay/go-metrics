package metric

import (
	"reflect"
	"testing"
	"time"
)

func TestTimerImpl(t *testing.T) {
	tests := []struct {
		name        string
		description string
		unit        string
		tags        Tags
	}{
		{
			name:        "test_timer",
			description: "Test timer",
			unit:        "ms",
			tags:        Tags{"service": "test"},
		},
		{
			name:        "empty_details",
			description: "",
			unit:        "",
			tags:        nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			timer := newTimer(Options{
				Name:        tt.name,
				Description: tt.description,
				Unit:        tt.unit,
				Tags:        tt.tags,
			})

			// Test basic properties
			if timer.Name() != tt.name {
				t.Errorf("Expected name %s, got %s", tt.name, timer.Name())
			}
			if timer.Description() != tt.description {
				t.Errorf("Expected description %s, got %s", tt.description, timer.Description())
			}
			if timer.Type() != TypeTimer {
				t.Errorf("Expected type %s, got %s", TypeTimer, timer.Type())
			}

			// Test tags
			expectedTags := tt.tags
			if expectedTags == nil {
				expectedTags = Tags{}
			}
			if !reflect.DeepEqual(timer.Tags(), expectedTags) {
				t.Errorf("Expected tags %v, got %v", expectedTags, timer.Tags())
			}

			// Test Record method
			timer.Record(100 * time.Millisecond)

			// Test RecordSince method
			now := time.Now()
			pastTime := now.Add(-50 * time.Millisecond)
			timer.RecordSince(pastTime)

			// Test Time method
			duration := timer.Time(func() {
				time.Sleep(10 * time.Millisecond)
			})
			if duration < 10*time.Millisecond {
				t.Errorf("Timer recorded less than 10ms: %v", duration)
			}

			// Test with additional tags
			additionalTags := Tags{"region": "us-west"}
			taggedTimer := timer.With(additionalTags)

			// Verify it's a new instance
			if timer == taggedTimer {
				t.Error("With() should return a new instance")
			}

			// Verify tags were merged correctly
			mergedTags := copyTags(expectedTags, additionalTags)
			if !reflect.DeepEqual(taggedTimer.Tags(), mergedTags) {
				t.Errorf("Expected merged tags %v, got %v", mergedTags, taggedTimer.Tags())
			}

			// Cast to access implementation details
			impl, ok := timer.(*timerImpl)
			if !ok {
				t.Fatal("Failed to cast to timerImpl")
			}

			// Verify the histogram was used
			if impl.histogram == nil {
				t.Error("Histogram should not be nil")
			}

			// Verify the histogram type
			histImpl, ok := impl.histogram.(*histogramImpl)
			if !ok {
				t.Fatal("Failed to cast to histogramImpl")
			}

			// Verify the histogram has recorded values
			if count := histImpl.count; count < 3 {
				t.Errorf("Expected at least 3 observations in histogram, got %d", count)
			}
		})
	}
}
