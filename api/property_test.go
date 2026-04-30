package api

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPropertyDTO_Serialization(t *testing.T) {
	tests := []struct {
		name     string
		dto      PropertyDTO
		expected string
	}{
		{
			name:     "Int value",
			dto:      PropertyDTO{Value: intPtr(42)},
			expected: `{"value":42}`,
		},
		{
			name:     "Int counter (value + max)",
			dto:      PropertyDTO{Value: intPtr(10), Max: intPtr(20)},
			expected: `{"value":10,"max":20}`,
		},
		{
			name:     "Float value",
			dto:      PropertyDTO{FValue: floatPtr(3.14)},
			expected: `{"fvalue":3.14}`,
		},
		{
			name:     "Bool value",
			dto:      PropertyDTO{BValue: boolPtr(true)},
			expected: `{"bvalue":true}`,
		},
		{
			name:     "String value",
			dto:      PropertyDTO{SValue: stringPtr("hello")},
			expected: `{"svalue":"hello"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.dto)
			assert.NoError(t, err)
			assert.JSONEq(t, tt.expected, string(data))

			var decoded PropertyDTO
			err = json.Unmarshal(data, &decoded)
			assert.NoError(t, err)
			assert.Equal(t, tt.dto, decoded)
		})
	}
}

func TestPropertyDTO_PolymorphicUnmarshal(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected PropertyDTO
	}{
		{
			name:     "Unmarshal raw int",
			input:    `42`,
			expected: PropertyDTO{Value: intPtr(42)},
		},
		{
			name:     "Unmarshal raw bool",
			input:    `true`,
			expected: PropertyDTO{BValue: boolPtr(true)},
		},
		{
			name:     "Unmarshal raw string",
			input:    `"hello"`,
			expected: PropertyDTO{SValue: stringPtr("hello")},
		},
		{
			name:     "Unmarshal structured object",
			input:    `{"value": 10, "max": 20}`,
			expected: PropertyDTO{Value: intPtr(10), Max: intPtr(20)},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var decoded PropertyDTO
			err := json.Unmarshal([]byte(tt.input), &decoded)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, decoded)
		})
	}
}

func intPtr(i int) *int { return &i }
func floatPtr(f float64) *float64 { return &f }
func boolPtr(b bool) *bool { return &b }
func stringPtr(s string) *string { return &s }
