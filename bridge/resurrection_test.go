// Package bridge provides unit tests for the arena resurrection and state recovery logic.
// It ensures that matches can be successfully resumed from a serialized snapshot after a service restart.
// @test-link [[api_go_battle_engine]]
// @test-link [[api_character_skill_inventory]]
package bridge

import (
	"strings"
	"testing"
	"time"

	"github.com/ecumeurs/upsilonapi/api"
	"github.com/ecumeurs/upsilonserializer"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// boardStateToResurrectReq converts a captured BoardState back into an ArenaResurrectRequest.
// This mirrors the behavior of the external management layer when performing state recovery.
func boardStateToResurrectReq(matchID uuid.UUID, callbackURL string, players []api.Player, bs api.BoardState) api.ArenaResurrectRequest {
	// 1. Grid Projection: Map the 2D projected API grid back into the 3D resurrection format.
	cells := make([][]api.ResurrectCell, bs.Grid.Width)
	for x := 0; x < bs.Grid.Width; x++ {
		cells[x] = make([]api.ResurrectCell, bs.Grid.Height)
		for y := 0; y < bs.Grid.Height; y++ {
			c := bs.Grid.Cells[x][y]
			cells[x][y] = api.ResurrectCell{Obstacle: c.Obstacle, Height: c.Height}
		}
	}

	// 2. Timeline Reconstruction: Map the initiative turns back to the engine's expected format.
	turns := make([]api.ResurrectTurn, len(bs.Turn))
	for i, t := range bs.Turn {
		turns[i] = api.ResurrectTurn{EntityID: t.EntityID, Delay: t.Delay}
	}

	// 3. Final Construction: Assemble the full recovery payload, including the embedded schema version.
	return api.ArenaResurrectRequest{
		MatchID:           matchID.String(),
		CallbackURL:       callbackURL,
		Players:           players,
		Grid:              api.ResurrectGrid{Width: bs.Grid.Width, Height: bs.Grid.Height, MaxHeight: bs.Grid.MaxHeight, Cells: cells},
		Turns:             turns,
		CurrentEntityID:   bs.CurrentEntityID,
		Version:           bs.Version,
		SerializerVersion: bs.SerializerVersion,
	}
}

// TestArenaResurrection_StatePreserved verifies that a reconstructed arena maintains its original data.
func TestArenaResurrection_StatePreserved(t *testing.T) {
	// 1. Setup Phase: Initialize a standard match with two competing entities.
	b := Get()
	matchID := uuid.New()
	callbackURL := "http://localhost/webhook"
	players := getResurrectionTestPlayers()

	_, g, entities, _, ts, version, err := b.StartArena(api.ArenaStartRequest{
		MatchID: matchID.String(), CallbackURL: callbackURL, Players: players,
	})
	require.NoError(t, err)
	time.Sleep(300 * time.Millisecond)

	// 2. Snapshot Phase: Capture the board state and initiative metadata before simulated failure.
	bs := api.NewBoardState(matchID, g, entities, players, ts, time.Now(), time.Now().Add(30*time.Second), 0, version, nil)
	preCrashVersion := version
	preCrashCurrentEntity := bs.CurrentEntityID

	// 3. Simulated Crash: Explicitly remove the arena from the bridge memory.
	b.DestroyArena(matchID)
	time.Sleep(200 * time.Millisecond)

	// 4. Recovery Phase: Trigger the resurrection logic using the previously captured snapshot.
	resurrectReq := boardStateToResurrectReq(matchID, callbackURL, players, bs)
	newBS, err := b.ResurrectArena(resurrectReq)
	require.NoError(t, err)

	// 5. Validation Phase: Confirm that the arena is back online and its critical metrics match the pre-crash state.
	assert.Equal(t, preCrashVersion, newBS.Version, "match version must be incremented or preserved exactly")
	assert.Equal(t, preCrashCurrentEntity, newBS.CurrentEntityID, "turn ownership must persist across resurrection")

	verifyEntityVitals(t, newBS.Players)
	b.DestroyArena(matchID)
}

// getResurrectionTestPlayers returns a baseline set of players for resurrection testing.
func getResurrectionTestPlayers() []api.Player {
	return []api.Player{
		{ID: uuid.New().String(), Team: 1, IA: true, Entities: []api.Entity{
			{ID: uuid.New().String(), Name: "Hero", HP: 8, MaxHP: 10, Move: 2, MaxMove: 3, Attack: 5, Defense: 2},
		}},
		{ID: uuid.New().String(), Team: 2, IA: true, Entities: []api.Entity{
			{ID: uuid.New().String(), Name: "Villain", HP: 10, MaxHP: 10, Move: 3, MaxMove: 3, Attack: 4, Defense: 1},
		}},
	}
}

// verifyEntityVitals ensures that no entity is resurrected with zero health unless previously dead.
func verifyEntityVitals(t *testing.T, players []api.Player) {
	for _, p := range players {
		for _, e := range p.Entities {
			if !e.Dead {
				assert.Greater(t, e.HP, 0, "active entity %s must have positive HP after resurrection", e.Name)
			}
		}
	}
}

// TestArenaResurrection_Idempotent ensures that the bridge rejects resurrection for active arenas.
func TestArenaResurrection_Idempotent(t *testing.T) {
	// 1. Setup Phase: Start a fresh match.
	b := Get()
	matchID := uuid.New()
	players := getResurrectionTestPlayers()

	_, g, entities, _, ts, version, err := b.StartArena(api.ArenaStartRequest{
		MatchID: matchID.String(), CallbackURL: "http://localhost/webhook", Players: players,
	})
	require.NoError(t, err)
	time.Sleep(150 * time.Millisecond)

	// 2. Snapshot Phase: Capture the current state.
	bs := api.NewBoardState(matchID, g, entities, players, ts, time.Now(), time.Now().Add(30*time.Second), 0, version, nil)
	req := boardStateToResurrectReq(matchID, "http://localhost/webhook", players, bs)

	// 3. Execution Phase: Attempt to resurrect while the match is still running.
	// The bridge must identify the ID conflict and return an error.
	_, err = b.ResurrectArena(req)
	assert.Error(t, err, "should reject resurrection requests for existing arena IDs")

	b.DestroyArena(matchID)
}

// TestArenaResurrection_GridObstaclesPreserved verifies that the 3D grid layout is correctly restored.
func TestArenaResurrection_GridObstaclesPreserved(t *testing.T) {
	// 1. Setup Phase: Start an arena and calculate its initial obstacle density.
	b := Get()
	matchID := uuid.New()
	players := getResurrectionTestPlayers()

	_, g, entities, _, ts, version, err := b.StartArena(api.ArenaStartRequest{
		MatchID: matchID.String(), CallbackURL: "http://localhost/webhook", Players: players,
	})
	require.NoError(t, err)
	time.Sleep(150 * time.Millisecond)

	bs := api.NewBoardState(matchID, g, entities, players, ts, time.Now(), time.Now().Add(30*time.Second), 0, version, nil)
	origObstacles := countObstacles(bs.Grid)

	// 2. Execution Phase: Simulate service restart and perform recovery.
	b.DestroyArena(matchID)
	req := boardStateToResurrectReq(matchID, "http://localhost/webhook", players, bs)
	newBS, err := b.ResurrectArena(req)
	require.NoError(t, err)

	// 3. Validation Phase: Ensure the obstacle layout remains identical.
	newObstacles := countObstacles(newBS.Grid)
	assert.Equal(t, origObstacles, newObstacles, "grid obstacle topology must be preserved exactly")

	b.DestroyArena(matchID)
}

// countObstacles is a helper to audit the total number of blocked cells in a grid.
func countObstacles(grid api.Grid) int {
	count := 0
	for x := range grid.Cells {
		for _, c := range grid.Cells[x] {
			if c.Obstacle { count++ }
		}
	}
	return count
}

// ── Serializer version guard tests (WP-D2 / audit risk R7) ───────────────────

// TestBoardState_StampsSerializerVersion verifies that NewBoardState embeds the current schema version.
func TestBoardState_StampsSerializerVersion(t *testing.T) {
	// 1. Setup: Start a real arena and capture its first board state.
	b := Get()
	matchID := uuid.New()
	players := getResurrectionTestPlayers()

	_, g, entities, _, ts, version, err := b.StartArena(api.ArenaStartRequest{
		MatchID: matchID.String(), CallbackURL: "http://localhost/webhook", Players: players,
	})
	require.NoError(t, err)
	time.Sleep(150 * time.Millisecond)

	// 2. Snapshot: Build a BoardState from the live engine.
	bs := api.NewBoardState(matchID, g, entities, players, ts, time.Now(), time.Now().Add(30*time.Second), 0, version, nil)

	// 3. Assertion: The stamped version must match the canonical constant.
	assert.Equal(t, upsilonserializer.CurrentSerializerVersion, bs.SerializerVersion,
		"NewBoardState must stamp SerializerVersion=%d into every blob", upsilonserializer.CurrentSerializerVersion)

	b.DestroyArena(matchID)
}

// TestResurrectArena_AbsentVersion_Rejected verifies that a blob with no serializer_version (zero) is refused.
func TestResurrectArena_AbsentVersion_Rejected(t *testing.T) {
	// 1. Setup: Build a valid resurrection request but omit the serializer version (simulates a
	//    blob persisted before versioning was introduced — the field unmarshals as zero).
	b := Get()
	matchID := uuid.New()
	players := getResurrectionTestPlayers()

	_, g, entities, _, ts, version, err := b.StartArena(api.ArenaStartRequest{
		MatchID: matchID.String(), CallbackURL: "http://localhost/webhook", Players: players,
	})
	require.NoError(t, err)
	time.Sleep(150 * time.Millisecond)

	bs := api.NewBoardState(matchID, g, entities, players, ts, time.Now(), time.Now().Add(30*time.Second), 0, version, nil)
	b.DestroyArena(matchID)
	time.Sleep(100 * time.Millisecond)

	req := boardStateToResurrectReq(matchID, "http://localhost/webhook", players, bs)
	// 2. Tamper: Erase the serializer version to simulate an unversioned legacy blob.
	req.SerializerVersion = 0

	// 3. Execution: Resurrection must be refused with a descriptive error.
	_, err = b.ResurrectArena(req)
	require.Error(t, err, "resurrection of an unversioned blob must return an explicit error")
	assert.True(t, strings.Contains(err.Error(), "serializer_version is absent"),
		"error must mention absent serializer_version; got: %s", err.Error())
}

// TestResurrectArena_WrongVersion_Rejected verifies that a blob with a mismatched serializer_version is refused.
func TestResurrectArena_WrongVersion_Rejected(t *testing.T) {
	// 1. Setup: Build a valid resurrection request, then set an incompatible schema version.
	b := Get()
	matchID := uuid.New()
	players := getResurrectionTestPlayers()

	_, g, entities, _, ts, version, err := b.StartArena(api.ArenaStartRequest{
		MatchID: matchID.String(), CallbackURL: "http://localhost/webhook", Players: players,
	})
	require.NoError(t, err)
	time.Sleep(150 * time.Millisecond)

	bs := api.NewBoardState(matchID, g, entities, players, ts, time.Now(), time.Now().Add(30*time.Second), 0, version, nil)
	b.DestroyArena(matchID)
	time.Sleep(100 * time.Millisecond)

	req := boardStateToResurrectReq(matchID, "http://localhost/webhook", players, bs)
	// 2. Tamper: Inject a future/unknown schema version (CurrentSerializerVersion + 99).
	req.SerializerVersion = upsilonserializer.CurrentSerializerVersion + 99

	// 3. Execution: Resurrection must be refused with a clear version-mismatch error.
	_, err = b.ResurrectArena(req)
	require.Error(t, err, "resurrection of an incompatible-version blob must return an explicit error")
	assert.True(t, strings.Contains(err.Error(), "serializer_version mismatch"),
		"error must report serializer_version mismatch; got: %s", err.Error())
}

// TestResurrectArena_CorrectVersion_Succeeds verifies that round-trip resurrection still works for current-version blobs.
func TestResurrectArena_CorrectVersion_Succeeds(t *testing.T) {
	// 1. Setup: Full start → snapshot → destroy → resurrect cycle with an unmodified blob.
	b := Get()
	matchID := uuid.New()
	players := getResurrectionTestPlayers()

	_, g, entities, _, ts, version, err := b.StartArena(api.ArenaStartRequest{
		MatchID: matchID.String(), CallbackURL: "http://localhost/webhook", Players: players,
	})
	require.NoError(t, err)
	time.Sleep(300 * time.Millisecond)

	bs := api.NewBoardState(matchID, g, entities, players, ts, time.Now(), time.Now().Add(30*time.Second), 0, version, nil)
	b.DestroyArena(matchID)
	time.Sleep(200 * time.Millisecond)

	req := boardStateToResurrectReq(matchID, "http://localhost/webhook", players, bs)

	// 2. Execution: A blob with the correct serializer_version must be accepted.
	newBS, err := b.ResurrectArena(req)
	require.NoError(t, err, "resurrection must succeed when serializer_version matches current schema")
	assert.Equal(t, upsilonserializer.CurrentSerializerVersion, newBS.SerializerVersion,
		"resurrected board state must carry the current serializer version")

	b.DestroyArena(matchID)
}
