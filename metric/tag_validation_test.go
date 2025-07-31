package metric

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestValidateTags(t *testing.T) {
	config := DefaultTagValidationConfig()

	tests := []struct {
		name    string
		tags    Tags
		config  TagValidationConfig
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid tags",
			tags:    Tags{"service": "api", "environment": "prod"},
			config:  config,
			wantErr: false,
		},
		{
			name:    "empty tags",
			tags:    Tags{},
			config:  config,
			wantErr: false,
		},
		{
			name:    "too many tags",
			tags:    make(Tags),
			config:  TagValidationConfig{MaxKeys: 2, MaxKeyLength: 100, MaxValueLength: 200},
			wantErr: true,
			errMsg:  "too many tags",
		},
		{
			name:    "key too long",
			tags:    Tags{strings.Repeat("a", 101): "value"},
			config:  TagValidationConfig{MaxKeys: 10, MaxKeyLength: 100, MaxValueLength: 200},
			wantErr: true,
			errMsg:  "exceeds maximum length",
		},
		{
			name:    "value too long",
			tags:    Tags{"key": strings.Repeat("a", 201)},
			config:  TagValidationConfig{MaxKeys: 10, MaxKeyLength: 100, MaxValueLength: 200},
			wantErr: true,
			errMsg:  "exceeds maximum length",
		},
		{
			name:    "empty key",
			tags:    Tags{"": "value"},
			config:  config,
			wantErr: true,
			errMsg:  "tag keys cannot be empty",
		},
		{
			name:    "disallowed key",
			tags:    Tags{"secret": "value"},
			config:  TagValidationConfig{MaxKeys: 10, MaxKeyLength: 100, MaxValueLength: 200, DisallowedKeys: []string{"secret", "password"}},
			wantErr: true,
			errMsg:  "is not allowed",
		},
	}

	// Set up the "too many tags" test
	for i := 0; i < 3; i++ {
		tests[2].tags[fmt.Sprintf("key%d", i)] = "value"
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTags(tt.tags, tt.config)
			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("Expected error message to contain '%s', got: %s", tt.errMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}
		})
	}
}

func TestRegistryTagValidation(t *testing.T) {
	// Test registry with custom validation config
	config := TagValidationConfig{
		MaxKeys:        2,
		MaxKeyLength:   10,
		MaxValueLength: 20,
		MaxCardinality: 3,
	}
	registry := NewRegistry(config, 5*time.Minute)

	// Valid tags should work
	counter1 := registry.Counter(Options{
		Name: "test_counter",
		Tags: Tags{"env": "prod"},
	})
	if counter1 == nil {
		t.Error("Expected valid counter creation")
	}

	// Test tag validation failure
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic for invalid tags")
		}
	}()
	
	// This should panic due to key being too long
	registry.Counter(Options{
		Name: "test_counter2",
		Tags: Tags{"very_long_key": "value"},
	})
}

func TestRegistryCardinalityLimit(t *testing.T) {
	config := TagValidationConfig{
		MaxKeys:        10,
		MaxKeyLength:   100,
		MaxValueLength: 200,
		MaxCardinality: 2, // Very low limit for testing
	}
	registry := NewRegistry(config, 5*time.Minute)

	// Create first metric instance
	counter1 := registry.Counter(Options{Name: "test_counter"})
	if counter1 == nil {
		t.Error("Expected first counter creation to succeed")
	}

	// Create second metric instance (different type, same name should count toward cardinality)
	gauge1 := registry.Gauge(Options{Name: "test_counter"})
	if gauge1 == nil {
		t.Error("Expected second metric creation to succeed")
	}

	// Third metric instance should fail due to cardinality limit
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic for cardinality limit exceeded")
		} else {
			// Verify the panic message contains cardinality information
			panicMsg := fmt.Sprintf("%v", r)
			if !strings.Contains(panicMsg, "cardinality limit exceeded") {
				t.Errorf("Expected panic message about cardinality, got: %s", panicMsg)
			}
		}
	}()
	
	registry.Histogram(Options{Name: "test_counter"})
}

func TestDefaultTagValidationConfig(t *testing.T) {
	config := DefaultTagValidationConfig()
	
	if config.MaxKeys != 10 {
		t.Errorf("Expected MaxKeys=10, got %d", config.MaxKeys)
	}
	if config.MaxKeyLength != 100 {
		t.Errorf("Expected MaxKeyLength=100, got %d", config.MaxKeyLength)
	}
	if config.MaxValueLength != 200 {
		t.Errorf("Expected MaxValueLength=200, got %d", config.MaxValueLength)
	}
	if config.MaxCardinality != 1000 {
		t.Errorf("Expected MaxCardinality=1000, got %d", config.MaxCardinality)
	}
	if len(config.DisallowedKeys) != 0 {
		t.Errorf("Expected empty DisallowedKeys, got %v", config.DisallowedKeys)
	}
}

func TestRegistryWithDefaultValidation(t *testing.T) {
	registry := NewDefaultRegistry()
	
	// Should work with reasonable tags
	counter := registry.Counter(Options{
		Name: "test_counter",
		Tags: Tags{
			"service":     "api",
			"environment": "production",
			"version":     "1.0.0",
		},
	})
	
	if counter == nil {
		t.Error("Expected counter creation with valid tags")
	}
	
	// Verify the counter works
	counter.Inc()
	if counter.Value() != 1 {
		t.Errorf("Expected counter value 1, got %d", counter.Value())
	}
}