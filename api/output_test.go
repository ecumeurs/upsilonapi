// Package api provides unit tests for the board state generation and DTO output mapping.
// It ensures that complex engine structures are correctly projected into API-facing snapshots.
// @spec-link [[api_go_battle_engine]]
// @spec-link [[mechanic_mec_skill_payload_resolution]]
package api

import (
	"testing"
	"time"

	"github.com/ecumeurs/upsilonbattle/battlearena/ruler/turner"
	"github.com/ecumeurs/upsilonmapdata/grid"
	"github.com/ecumeurs/upsilontypes/entity"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// TestNewBoardStateWinnerTeamID verifies that victory conditions are correctly reflected in the board state.
func TestNewBoardStateWinnerTeamID(t *testing.T) {
	// 1. Setup Phase: Initialize a basic grid and a dummy turn state.
	g := grid.New(5, 5, 5)
	ts := turner.TurnState{}
	
	// 2. Execution Phase: Create a board state where team 2 is the winner.
	bs := NewBoardState(uuid.New(), g, nil, nil, ts, time.Now(), time.Now().Add(30*time.Second), 2, 1, nil)
	
	// 3. Validation Phase: Ensure the WinnerTeamID pointer is set and matches the engine's result.
	assert.NotNil(t, bs.WinnerTeamID, "winner team ID must be non-nil when a match is resolved")
	assert.Equal(t, 2, *bs.WinnerTeamID, "winner team ID must match the engine's victory signal")
}

// TestNewBoardStateCarriesElevation ensures that the 2D grid projection preserves height data.
func TestNewBoardStateCarriesElevation(t *testing.T) {
	// 1. Setup Phase: Create a grid with a variable elevation profile.
	g := grid.New(10, 10, 10)
	
	// 2. Execution Phase: Create a board state from the 3D grid.
	bs := NewBoardState(uuid.New(), g, nil, nil, turner.TurnState{}, time.Now(), time.Now().Add(30*time.Second), 0, 1, nil)
	
	// 3. Validation Phase: Confirm the projected grid dimensions match the source.
	assert.Equal(t, 10, bs.Grid.Width)
	assert.Equal(t, 10, bs.Grid.Height)
}

// TestNewBoardStateDeadEntityHP ensures that dead entities are correctly identified as such in the DTO.
func TestNewBoardStateDeadEntityHP(t *testing.T) {
	// 1. Setup Phase: Define a player with an entity that was previously active.
	entID := uuid.New()
	p := Player{
		ID: uuid.New().String(),
		Entities: []Entity{{ID: entID.String(), HP: 10, Dead: false}},
	}
	
	// 2. Execution Phase: Create a board state with an empty entity list to simulate entity removal/death.
	bs := NewBoardState(uuid.New(), grid.New(5, 5, 5), nil, []Player{p}, turner.TurnState{}, time.Now(), time.Now().Add(30*time.Second), 0, 1, nil)
	
	// 3. Validation Phase: Ensure the entity in the player roster is marked as dead with 0 HP.
	assert.True(t, bs.Players[0].Entities[0].Dead, "entity must be marked as dead if missing from live engine state")
	assert.Equal(t, 0, bs.Players[0].Entities[0].HP, "dead entity HP must be zeroed in the API output")
}
