package bridge

// @spec-link [[api_go_battle_action]]
// @spec-link [[api_go_battle_forfeit]]
// @spec-link [[rule_forfeit_battle]]
// @spec-link [[mech_game_state_versioning]]

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/ecumeurs/upsilonapi/api"
	"github.com/ecumeurs/upsilonmapdata/grid/position"
	"github.com/ecumeurs/upsilontools/tools/messagequeue/message"
	"github.com/ecumeurs/upsilonbattle/battlearena/ruler/rulermethods"
	"github.com/google/uuid"
)

// GetBoardState requests the current match state from the Ruler.
// It uses a timeout-backed synchronous message call to ensure thread safety
// and avoid data races during state access.
func (b *ArenaBridge) GetBoardState(matchID uuid.UUID, action *api.ActionFeedback) (api.BoardState, error) {
	// Acquire a read lock to safely access the arena map.
	b.mu.RLock()
	arena, ok := b.arenas[matchID]
	b.mu.RUnlock()
	if !ok {
		return api.BoardState{}, fmt.Errorf("arena %s not found", matchID)
	}

	// Request board state from Ruler via message to avoid data races.
	// This ensures we get a consistent snapshot from the actor's internal state.
	// @spec-link [[api_go_battle_action]]
	respChan := make(chan *message.Message, 1)
	arena.Ruler.SendActor(message.Create(nil, rulermethods.GetBoardState{
		ActionContext: action,
	}, rulermethods.GetBoardStateReply{}), respChan)

	select {
	case res := <-respChan:
		// Handle potential engine errors returned by the Ruler actor.
		if res.HasError {
			return api.BoardState{}, fmt.Errorf("engine error: %s", res.ErrorMessage)
		}
		// Extract the reply and construct a fresh BoardState DTO.
		reply := res.TargetMethod.(rulermethods.GetBoardStateReply)
		players, _ := arena.Metadata["Players"].([]api.Player)

		// Map engine entities and grid to the API-standard representation.
		return api.NewBoardState(
			matchID,
			reply.Grid,
			reply.Entities,
			players,
			reply.TurnState,
			time.Now(),
			time.Now().Add(30*time.Second),
			reply.WinnerTeamID,
			reply.Version,
			action,
		), nil
	case <-time.After(2 * time.Second):
		// Safeguard against deadlocked actors or excessive engine load.
		return api.BoardState{}, fmt.Errorf("timeout waiting for ruler state")
	}
}

type webhookSentKey struct {
	matchID   uuid.UUID
	version   int64
	eventType string
}

var lastSentWebhook = make(map[webhookSentKey]bool)
var lastSentMu sync.Mutex

// TrySendWebhook checks if a webhook for this version and event type has already been sent.
// It prevents duplicate deliveries for the same game state version.
// @spec-link [[mech_game_state_versioning]]
func (b *ArenaBridge) TrySendWebhook(matchID uuid.UUID, version int64, eventType string) bool {
	lastSentMu.Lock()
	defer lastSentMu.Unlock()

	key := webhookSentKey{matchID, version, eventType}
	if lastSentWebhook[key] {
		return false
	}

	// Cleanup old versions for this match to prevent memory leak
	for k := range lastSentWebhook {
		if k.matchID == matchID && k.version < version {
			delete(lastSentWebhook, k)
		}
	}

	lastSentWebhook[key] = true
	return true
}

// ArenaAction proxies a tactical command (move, skill, attack, pass) to the Ruler.
// It returns (success, message, errorKey, data) for API response mapping.
// Error keys are propagated to the external envelope as meta.error_key.
// @spec-link [[api_go_battle_action]]
func (b *ArenaBridge) ArenaAction(arenaID uuid.UUID, req api.ArenaActionMessage) (bool, string, string, interface{}) {
	// Retrieve the Ruler actor for the specified match.
	r, ok := b.GetArena(arenaID)
	if !ok {
		return false, "arena not found", "arena.notfound", nil
	}

	// Parse player and entity UUIDs from the request payload.
	playerID, err := uuid.Parse(req.Data.PlayerID)
	if err != nil {
		return false, fmt.Sprintf("invalid player_id: %v", err), "request.player_id.invalid", nil
	}

	entityID, err := uuid.Parse(req.Data.EntityID)
	if err != nil {
		return false, fmt.Sprintf("invalid entity_id: %v", err), "request.entity_id.invalid", nil
	}

	// Create a response channel for the actor's asynchronous reply.
	respChan := make(chan *message.Message)
	defer close(respChan)
	
	// Normalize type to lowercase for case-insensitive matching across client platforms.
	actionType := strings.ToLower(req.Data.Type)

	// Dispatch the request to the appropriate Ruler method based on action type.
	switch actionType {
	case "skill":
		// Skills require a valid skill_id and targeting coordinates.
		if req.Data.SkillID == "" {
			return false, "skill requires skill_id", "request.skill_id.missing", nil
		}
		skillID, err := uuid.Parse(req.Data.SkillID)
		if err != nil {
			return false, fmt.Sprintf("invalid skill_id: %v", err), "request.skill_id.invalid", nil
		}
		if len(req.Data.TargetCoords) == 0 {
			return false, "skill requires target_coords", "request.target_coords.missing", nil
		}
		// Resolve the target position within the current grid context.
		r.SendActor(message.Create(nil, rulermethods.ControllerUseSkill{
			ControllerID: playerID,
			EntityID:     entityID,
			SkillID:      skillID,
			Target:       position.New(req.Data.TargetCoords[0].X, req.Data.TargetCoords[0].Y, r.GameState.Grid.TopMostCellAt(req.Data.TargetCoords[0].X, req.Data.TargetCoords[0].Y)),
		}, rulermethods.ControllerUseSkillReply{}), respChan)
	case "attack":
		// Standard attacks require a single target coordinate.
		if len(req.Data.TargetCoords) == 0 {
			return false, "attack requires target_coords", "request.target_coords.missing", nil
		}
		r.SendActor(message.Create(nil, rulermethods.ControllerAttack{
			ControllerID: playerID,
			EntityID:     entityID,
			Target:       position.New(req.Data.TargetCoords[0].X, req.Data.TargetCoords[0].Y, r.GameState.Grid.TopMostCellAt(req.Data.TargetCoords[0].X, req.Data.TargetCoords[0].Y)),
		}, rulermethods.ControllerAttackReply{}), respChan)
	case "pass":
		// Explicit turn termination by the player.
		r.SendActor(message.Create(nil, rulermethods.EndOfTurn{
			ControllerID: playerID,
			EntityID:     entityID,
		}, rulermethods.EndOfTurn{}), respChan)
	case "move":
		// Movement requires a valid path of coordinates.
		if len(req.Data.TargetCoords) == 0 {
			return false, "move requires target_coords", "request.target_coords.missing", nil
		}
		// Map the 2D API coordinates to 3D engine positions.
		path := make([]position.Position, len(req.Data.TargetCoords))
		for i, c := range req.Data.TargetCoords {
			path[i] = position.New(c.X, c.Y, r.GameState.Grid.TopMostCellAt(c.X, c.Y))
		}
		r.SendActor(message.Create(nil, rulermethods.ControllerMove{
			ControllerID: playerID,
			EntityID:     entityID,
			Path:         path,
		}, rulermethods.ControllerMoveReply{}), respChan)
	default:
		// Fallback to EndOfTurn for unrecognized action types to prevent stuck turns.
		r.SendActor(message.Create(nil, rulermethods.EndOfTurn{
			ControllerID: playerID,
			EntityID:     entityID,
		}, rulermethods.EndOfTurn{}), respChan)
	}

	// Wait for the reply from the Ruler actor (blocks until engine processing is complete).
	res := <-respChan

	// Propagate engine errors back to the caller with appropriate keys.
	if res.HasError {
		return false, res.ErrorMessage, res.ErrorKey, nil
	}

	return true, fmt.Sprintf("action %s accepted", req.Data.Type), "", res.Content
}

// ArenaForfeit allows a player to concede the match.
// It is team-wide and does not require a specific entity context.
// @spec-link [[api_go_battle_forfeit]]
func (b *ArenaBridge) ArenaForfeit(arenaID uuid.UUID, playerID uuid.UUID) (bool, string, string, interface{}) {
	// Look up the active Ruler.
	r, ok := b.GetArena(arenaID)
	if !ok {
		return false, "arena not found", "arena.notfound", nil
	}

	respChan := make(chan *message.Message)
	defer close(respChan)

	// Dispatch the forfeit command to the Ruler.
	r.SendActor(message.Create(nil, rulermethods.ControllerForfeit{
		ControllerID: playerID,
		EntityID:     uuid.Nil, // Forfeiting is team-wide.
	}, rulermethods.ControllerForfeit{}), respChan)

	// Wait for the reply from the Ruler actor.
	res := <-respChan

	if res.HasError {
		return false, res.ErrorMessage, res.ErrorKey, nil
	}

	return true, "forfeit accepted", "", res.Content
}
