// Package bridge provides internal test infrastructure and factory helpers for the UpsilonAPI.
// These helpers streamline the creation of complex test payloads for bridge-level validation.
// @test-link [[api_go_battle_engine]]
// @test-link [[mechanic_mec_skill_payload_resolution]]
package bridge

import (
	"github.com/ecumeurs/upsilonapi/api"
	"github.com/google/uuid"
)

// floatPtr is a utility to create a pointer to a float64 value.
func floatPtr(v float64) *float64 {
	// 1. Memory Management: Allocate float on heap and return reference.
	return &v
}

// boolPtr is a utility to create a pointer to a bool value.
func boolPtr(v bool) *bool {
	// 1. Memory Management: Allocate bool on heap and return reference.
	return &v
}

// stringPtr is a utility to create a pointer to a string value.
func stringPtr(v string) *string {
	// 1. Memory Management: Allocate string on heap and return reference.
	return &v
}

// createTestRequest builds a baseline ArenaStartRequest for integration testing.
func createTestRequest(matchID uuid.UUID) api.ArenaStartRequest {
	// 1. Payload Assembly: Return a standard match request with identified players and entities.
	return api.ArenaStartRequest{
		MatchID:     matchID.String(),
		CallbackURL: "http://localhost/webhook",
		Players:     []api.Player{{ID: uuid.New().String(), Team: 1, IA: true, Entities: []api.Entity{createTestEntity()}}},
	}
}

// createTestEntity generates a standard character entity with baseline combat stats.
func createTestEntity() api.Entity {
	// 1. Entity Construction: Return a high-HP character DTO for structural testing.
	return api.Entity{
		ID: uuid.New().String(), Name: "TestHero", HP: 10, MaxHP: 10, Move: 3, MaxMove: 3, Attack: 5, Defense: 2,
	}
}
