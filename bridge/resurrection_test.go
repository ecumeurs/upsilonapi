package bridge

// @test-link [[api_go_battle_engine]]

import (
	"testing"
	"time"

	"github.com/ecumeurs/upsilonapi/api"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// boardStateToResurrectReq converts a captured BoardState back into an ArenaResurrectRequest,
// mirroring exactly what the Laravel side will do with the cached game_state_cache.
func boardStateToResurrectReq(matchID uuid.UUID, callbackURL string, players []api.Player, bs api.BoardState) api.ArenaResurrectRequest {
	cells := make([][]api.ResurrectCell, bs.Grid.Width)
	for x := 0; x < bs.Grid.Width; x++ {
		cells[x] = make([]api.ResurrectCell, bs.Grid.Height)
		for y := 0; y < bs.Grid.Height; y++ {
			c := bs.Grid.Cells[x][y]
			cells[x][y] = api.ResurrectCell{
				Obstacle: c.Obstacle,
				Height:   c.Height,
			}
		}
	}

	turns := make([]api.ResurrectTurn, len(bs.Turn))
	for i, t := range bs.Turn {
		turns[i] = api.ResurrectTurn{
			EntityID: t.EntityID,
			Delay:    t.Delay,
		}
	}

	return api.ArenaResurrectRequest{
		MatchID:         matchID.String(),
		CallbackURL:     callbackURL,
		Players:         players,
		Grid:            api.ResurrectGrid{Width: bs.Grid.Width, Height: bs.Grid.Height, MaxHeight: bs.Grid.MaxHeight, Cells: cells},
		Turns:           turns,
		CurrentEntityID: bs.CurrentEntityID,
		Version:         bs.Version,
	}
}

func TestArenaResurrection_StatePreserved(t *testing.T) {
	b := Get()
	matchID := uuid.New()
	callbackURL := "http://localhost/webhook"

	players := []api.Player{
		{ID: uuid.New().String(), Team: 1, IA: true, Entities: []api.Entity{
			{ID: uuid.New().String(), Name: "Hero", HP: 8, MaxHP: 10, Move: 2, MaxMove: 3, Attack: 5, Defense: 2},
		}},
		{ID: uuid.New().String(), Team: 2, IA: true, Entities: []api.Entity{
			{ID: uuid.New().String(), Name: "Villain", HP: 10, MaxHP: 10, Move: 3, MaxMove: 3, Attack: 4, Defense: 1},
		}},
	}

	_, g, entities, _, ts, version, err := b.StartArena(api.ArenaStartRequest{
		MatchID:     matchID.String(),
		CallbackURL: callbackURL,
		Players:     players,
	})
	require.NoError(t, err)

	// Allow first turn to be dispatched.
	time.Sleep(300 * time.Millisecond)

	bs := api.NewBoardState(matchID, g, entities, players, ts, time.Now(), time.Now().Add(30*time.Second), 0, version, nil)
	preCrashVersion := version
	preCrashCurrentEntity := bs.CurrentEntityID

	// Simulate crash: destroy the arena.
	b.DestroyArena(matchID)
	time.Sleep(200 * time.Millisecond)

	// Resurrect from the captured state.
	resurrectReq := boardStateToResurrectReq(matchID, callbackURL, players, bs)
	newBS, err := b.ResurrectArena(resurrectReq)
	require.NoError(t, err)

	// Arena is alive again — verify state.
	_, ok := b.GetArena(matchID)
	assert.True(t, ok)
	assert.Equal(t, preCrashVersion, newBS.Version)
	assert.Equal(t, preCrashCurrentEntity, newBS.CurrentEntityID)

	for _, p := range newBS.Players {
		for _, e := range p.Entities {
			if !e.Dead {
				assert.Greater(t, e.HP, 0, "entity %s must have HP > 0 after resurrection", e.Name)
			}
		}
	}

	time.Sleep(200 * time.Millisecond)
	b.DestroyArena(matchID)
}

func TestArenaResurrection_Idempotent(t *testing.T) {
	b := Get()
	matchID := uuid.New()

	players := []api.Player{
		{ID: uuid.New().String(), Team: 1, IA: true, Entities: []api.Entity{
			{ID: uuid.New().String(), Name: "A", HP: 5, MaxHP: 5, Move: 2, MaxMove: 2, Attack: 3, Defense: 1},
		}},
		{ID: uuid.New().String(), Team: 2, IA: true, Entities: []api.Entity{
			{ID: uuid.New().String(), Name: "B", HP: 5, MaxHP: 5, Move: 2, MaxMove: 2, Attack: 3, Defense: 1},
		}},
	}

	_, g, entities, _, ts, version, err := b.StartArena(api.ArenaStartRequest{
		MatchID:     matchID.String(),
		CallbackURL: "http://localhost/webhook",
		Players:     players,
	})
	require.NoError(t, err)
	time.Sleep(150 * time.Millisecond)

	bs := api.NewBoardState(matchID, g, entities, players, ts, time.Now(), time.Now().Add(30*time.Second), 0, version, nil)
	req := boardStateToResurrectReq(matchID, "http://localhost/webhook", players, bs)

	// Resurrect while arena is alive should fail.
	_, err = b.ResurrectArena(req)
	assert.Error(t, err, "should reject resurrect when arena is already running")

	b.DestroyArena(matchID)
	time.Sleep(100 * time.Millisecond)
}

func TestArenaResurrection_GridObstaclesPreserved(t *testing.T) {
	b := Get()
	matchID := uuid.New()

	players := []api.Player{
		{ID: uuid.New().String(), Team: 1, IA: true, Entities: []api.Entity{
			{ID: uuid.New().String(), Name: "A", HP: 10, MaxHP: 10, Move: 3, MaxMove: 3, Attack: 5, Defense: 2},
		}},
		{ID: uuid.New().String(), Team: 2, IA: true, Entities: []api.Entity{
			{ID: uuid.New().String(), Name: "B", HP: 10, MaxHP: 10, Move: 3, MaxMove: 3, Attack: 5, Defense: 2},
		}},
	}

	_, g, entities, _, ts, version, err := b.StartArena(api.ArenaStartRequest{
		MatchID:     matchID.String(),
		CallbackURL: "http://localhost/webhook",
		Players:     players,
	})
	require.NoError(t, err)
	time.Sleep(150 * time.Millisecond)

	bs := api.NewBoardState(matchID, g, entities, players, ts, time.Now(), time.Now().Add(30*time.Second), 0, version, nil)

	origObstacles := 0
	for x := range bs.Grid.Cells {
		for _, c := range bs.Grid.Cells[x] {
			if c.Obstacle {
				origObstacles++
			}
		}
	}

	b.DestroyArena(matchID)
	time.Sleep(100 * time.Millisecond)

	req := boardStateToResurrectReq(matchID, "http://localhost/webhook", players, bs)
	newBS, err := b.ResurrectArena(req)
	require.NoError(t, err)

	newObstacles := 0
	for x := range newBS.Grid.Cells {
		for _, c := range newBS.Grid.Cells[x] {
			if c.Obstacle {
				newObstacles++
			}
		}
	}

	assert.Equal(t, origObstacles, newObstacles, "obstacle count must be preserved after resurrection")

	time.Sleep(100 * time.Millisecond)
	b.DestroyArena(matchID)
}
