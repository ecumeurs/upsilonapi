package api

import (
	"testing"
	"time"

	"github.com/ecumeurs/upsilontypes/entity"
	"github.com/ecumeurs/upsilonbattle/battlearena/ruler/turner"
	"github.com/ecumeurs/upsilonmapdata/grid"
	"github.com/ecumeurs/upsilonmapmaker/gridgenerator"
	"github.com/ecumeurs/upsilontools/tools"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestNewBoardStateWinnerTeamID(t *testing.T) {
	matchID := uuid.New()
	g := grid.NewGrid(10, 10, 1)
	entities := []entity.Entity{}
	players := []Player{}
	ts := turner.TurnState{}
	startTime := time.Now()
	timeout := time.Now().Add(30 * time.Second)
	winnerTeamID := 2

	// Test with a winner
	bs := NewBoardState(matchID, g, entities, players, ts, startTime, timeout, winnerTeamID, 0, nil)
	assert.Equal(t, winnerTeamID, *bs.WinnerTeamID, "WinnerTeamID should be populated in BoardState")

	// Test without a winner (0)
	bs = NewBoardState(matchID, g, entities, players, ts, startTime, timeout, 0, 0, nil)
	assert.Nil(t, bs.WinnerTeamID, "WinnerTeamID should be nil when 0 is passed")
}

// TestNewBoardStateCarriesElevation verifies the Grid/Cell payload exposes
// topmost-cell elevation when the engine runs a non-flat generator (Hill).
// Regression guard for the 3D rendering plumbing: the API must surface
// Cell.Height and Grid.MaxHeight so battleui can render terrain.
func TestNewBoardStateCarriesElevation(t *testing.T) {
	tools.SeedWith(42)

	gg := gridgenerator.GridGenerator{
		Width:  tools.NewIntRange(15, 16),
		Length: tools.NewIntRange(15, 16),
		Height: tools.NewIntRange(12, 13),
		Type:   gridgenerator.Hill,
	}
	g := gg.Generate()

	bs := NewBoardState(uuid.New(), g, nil, nil, turner.TurnState{}, time.Now(), time.Now(), 0, 0, nil)

	assert.Equal(t, g.Width, bs.Grid.Width)
	assert.Equal(t, g.Length, bs.Grid.Height)
	assert.Equal(t, g.Height, bs.Grid.MaxHeight, "Grid.MaxHeight must expose the engine Z ceiling")

	minH, maxH := -1, -1
	for x := 0; x < bs.Grid.Width; x++ {
		for y := 0; y < bs.Grid.Height; y++ {
			h := bs.Grid.Cells[x][y].Height
			if minH < 0 || h < minH {
				minH = h
			}
			if h > maxH {
				maxH = h
			}
		}
	}
	assert.Greater(t, maxH, minH, "Hill generator must produce varying Cell.Height across the grid")
	assert.LessOrEqual(t, maxH, bs.Grid.MaxHeight, "every Cell.Height must fit under MaxHeight")
}

func TestNewBoardStateDeadEntityHP(t *testing.T) {
	matchID := uuid.New()
	g := grid.NewGrid(10, 10, 1)
	entID := uuid.New()
	
	// Initial roster with 1 entity having 10 HP
	players := []Player{
		{
			ID: uuid.New().String(),
			Entities: []Entity{
				{ID: entID.String(), HP: 10},
			},
		},
	}
	
	// Empty live entities (simulating death/removal)
	entities := []entity.Entity{}
	
	ts := turner.TurnState{}
	startTime := time.Now()
	timeout := time.Now().Add(30 * time.Second)

	bs := NewBoardState(matchID, g, entities, players, ts, startTime, timeout, 0, 0, nil)
	
	assert.Equal(t, 0, bs.Players[0].Entities[0].HP, "Entity not in live map should have HP set to 0")
}
