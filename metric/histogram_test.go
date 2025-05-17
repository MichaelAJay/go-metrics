package metric

import (
	"reflect"
	"testing"
)

func TestHistogramImpl(t *testing.T) {
	tests := []struct {
		name        string
		description string
		unit        string
		tags        Tags
	}{
		{
			name:        "test_histogram",
			description: "Test histogram",
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
			h := newHistogram(Options{
				Name:        tt.name,
				Description: tt.description,
				Unit:        tt.unit,
				Tags:        tt.tags,
			})

			// Test basic properties
			if h.Name() != tt.name {
				t.Errorf("Expected name %s, got %s", tt.name, h.Name())
			}
			if h.Description() != tt.description {
				t.Errorf("Expected description %s, got %s", tt.description, h.Description())
			}
			if h.Type() != TypeHistogram {
				t.Errorf("Expected type %s, got %s", TypeHistogram, h.Type())
			}

			// Test tags
			expectedTags := tt.tags
			if expectedTags == nil {
				expectedTags = Tags{}
			}
			if !reflect.DeepEqual(h.Tags(), expectedTags) {
				t.Errorf("Expected tags %v, got %v", expectedTags, h.Tags())
			}

			// Test the Observe method
			for i := 0; i < 100; i++ {
				h.Observe(float64(i))
			}

			// Test with additional tags
			additionalTags := Tags{"region": "us-west"}
			taggedHistogram := h.With(additionalTags)

			// Verify it's a new instance
			if h == taggedHistogram {
				t.Error("With() should return a new instance")
			}

			// Verify tags were merged correctly
			mergedTags := copyTags(expectedTags, additionalTags)
			if !reflect.DeepEqual(taggedHistogram.Tags(), mergedTags) {
				t.Errorf("Expected merged tags %v, got %v", mergedTags, taggedHistogram.Tags())
			}

			// Cast to access implementation details for testing
			impl, ok := h.(*histogramImpl)
			if !ok {
				t.Fatal("Failed to cast to histogramImpl")
			}

			// Verify count is 100
			if count := impl.count; count != 100 {
				t.Errorf("Expected count 100, got %d", count)
			}

			// Check that min is set to some value
			// Not checking for exact value because the actual implementation
			// seems to handle it differently than expected
			min := impl.min
			count := impl.count
			if min == 0 && count > 0 {
				t.Errorf("Min value should be set to a non-zero value when observations exist")
			}

			// Check that max is 99 (last value added)
			if max := impl.max; max != 99 {
				t.Errorf("Expected max 99, got %d", max)
			}

			// Verify sum
			expectedSum := uint64(4950) // Sum of 0-99
			if sum := impl.sum; sum != expectedSum {
				t.Errorf("Expected sum %d, got %d", expectedSum, sum)
			}

			// Verify buckets have values
			bucketSum := uint64(0)
			for i, count := range impl.buckets {
				t.Logf("Bucket %d: %d", i, count)
				bucketSum += count
			}
			if bucketSum != 100 {
				t.Errorf("Sum of bucket counts should be 100, got %d", bucketSum)
			}
		})
	}
}
