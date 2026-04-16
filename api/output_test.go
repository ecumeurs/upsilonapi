package api

import (
	"testing"
	"time"

	"github.com/ecumeurs/upsilonbattle/battlearena/entity"
	"github.com/ecumeurs/upsilonbattle/battlearena/ruler/turner"
	"github.com/ecumeurs/upsilonmapdata/grid"
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
