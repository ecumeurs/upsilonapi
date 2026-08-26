// Package api provides unit tests for the polymorphic property DTOs and serialization logic.
// It ensures that engine properties (int, float, bool, string) are correctly mapped to and from JSON.
// @test-link [[api_go_battle_engine]]
// @test-link [[mechanic_skill_payload_resolution]]
package api

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPropertyDTO_Serialization verifies that structured property DTOs are marshaled into correct JSON schemas.
func TestPropertyDTO_Serialization(t *testing.T) {
	// 1. Setup Phase: Create an integer counter property (value + max).
	val := 10
	max := 20
	p := PropertyDTO{Value: &val, Max: &max}

	// 2. Execution Phase: Marshal the DTO into a JSON byte slice.
	data, err := json.Marshal(p)
	require.NoError(t, err)

	// 3. Validation Phase: Confirm the output contains both 'value' and 'max' keys.
	var result map[string]interface{}
	err = json.Unmarshal(data, &result)
	assert.NoError(t, err)
	assert.Equal(t, float64(10), result["value"], "serialized value must match input")
	assert.Equal(t, float64(20), result["max"], "serialized max must match input")
}

// TestPropertyDTO_PolymorphicUnmarshal ensures the DTO can handle both structured objects and primitives.
func TestPropertyDTO_PolymorphicUnmarshal(t *testing.T) {
	// 1. Case: Structured Counter Object.
	// It should correctly identify 'value' and 'max'.
	var p1 PropertyDTO
	err := json.Unmarshal([]byte(`{"value": 5, "max": 10}`), &p1)
	assert.NoError(t, err)
	assert.Equal(t, 5, *p1.Value)
	assert.Equal(t, 10, *p1.Max)

	// 2. Case: Primitive Integer.
	// It should map a raw number to the .Value pointer.
	var p2 PropertyDTO
	err = json.Unmarshal([]byte(`42`), &p2)
	assert.NoError(t, err)
	assert.Equal(t, 42, *p2.Value)

	// 3. Case: Primitive Boolean.
	// It should map a raw boolean to the .BValue pointer.
	var p3 PropertyDTO
	err = json.Unmarshal([]byte(`true`), &p3)
	assert.NoError(t, err)
	assert.True(t, *p3.BValue)

	// 4. Case: Primitive Float.
	// It should map a raw decimal to the .FValue pointer.
	var p4 PropertyDTO
	err = json.Unmarshal([]byte(`12.5`), &p4)
	assert.NoError(t, err)
	assert.Equal(t, 12.5, *p4.FValue)
}

// floatPtr is a utility to create a pointer to a float64 value.
func floatPtr(v float64) *float64 {
	// 1. Memory Allocation: Create a new float64 on the heap.
	// 2. Assignment: Return the address to the caller.
	return &v
}

// boolPtr is a utility to create a pointer to a bool value.
func boolPtr(v bool) *bool {
	// 1. Memory Allocation: Create a new boolean on the heap.
	// 2. Assignment: Return the address to the caller.
	return &v
}

// stringPtr is a utility to create a pointer to a string value.
func stringPtr(v string) *string {
	// 1. Memory Allocation: Create a new string on the heap.
	// 2. Assignment: Return the address to the caller.
	return &v
}
