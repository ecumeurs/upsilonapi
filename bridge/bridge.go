package bridge

// @spec-link [[rule_team_mechanics]]
// @spec-link [[rule_forfeit_battle]]

// @spec-link [[module_upsilonapi]]

import (
	"fmt"
	"log"
	"strings"
	"sync"

	"time"

	"github.com/ecumeurs/upsilonapi/api"
	"github.com/ecumeurs/upsilonbattle/battlearena"
	"github.com/ecumeurs/upsilonbattle/battlearena/controller/controllers"
	"github.com/ecumeurs/upsilonbattle/battlearena/entity"
	"github.com/ecumeurs/upsilonbattle/battlearena/entity/entitygenerator"
	"github.com/ecumeurs/upsilonbattle/battlearena/ruler"
	"github.com/ecumeurs/upsilonbattle/battlearena/ruler/rulermethods"
	"github.com/ecumeurs/upsilonbattle/battlearena/ruler/turner"
	"github.com/ecumeurs/upsilonmapdata/grid"
	"github.com/ecumeurs/upsilonmapdata/grid/position"
	"github.com/ecumeurs/upsilonmapmaker/gridgenerator"
	"github.com/ecumeurs/upsilontools/tools/actor"
	"github.com/ecumeurs/upsilontools/tools/messagequeue/message"
	"github.com/google/uuid"
)

type ArenaBridge struct {
	mu     sync.RWMutex
	arenas map[uuid.UUID]*battlearena.BattleArena
	// @spec-link [[mech_game_state_versioning]]
	lastSentWebhookVersion map[uuid.UUID]int64
}

var bridge = &ArenaBridge{
	arenas:                 make(map[uuid.UUID]*battlearena.BattleArena),
	lastSentWebhookVersion: make(map[uuid.UUID]int64),
}

func Get() *ArenaBridge {
	return bridge
}

func (b *ArenaBridge) StartArena(start api.ArenaStartRequest) (uuid.UUID, *grid.Grid, []entity.Entity, []api.Player, turner.TurnState, int64, error) {
	matchID := uuid.MustParse(start.MatchID)
	battleArena := battlearena.NewBattleArena(matchID)
	battleArena.Metadata["CallbackURL"] = start.CallbackURL
	battleArena.Metadata["Players"] = start.Players

	// Ensure Ruler ID matches MatchID as per caller expectations
	battleArena.Ruler.ID = matchID

	b.mu.Lock()
	b.arenas[matchID] = battleArena
	b.mu.Unlock()

	// this bypass actor's owning resource, we should probably use the SetGrid message instead (doesn't exist yet).
	battleArena.Ruler.SetGrid(gridgenerator.GeneratePlainSquare(10, 10))
	battleArena.Ruler.SetNbControllers(len(start.Players))

	// We need to wait for the reply to get the initial state
	respChan := make(chan *message.Message)
	defer close(respChan)

	count := len(start.Players)

	for _, p := range start.Players {

		for _, ee := range p.Entities {
			e := entitygenerator.GenerateRandomEntity()
			e.Type = entity.Character
			e.Name = ee.Name
			e.ID = uuid.MustParse(ee.ID)
			e.ControllerID = uuid.MustParse(p.ID)

			e.RepsertPropertyCMaxValue("HP", ee.MaxHP)
			e.RepsertPropertyCValue("HP", ee.HP)
			e.RepsertPropertyCMaxValue("Movement", ee.MaxMove)
			e.RepsertPropertyCValue("Movement", ee.Move)
			e.RepsertPropertyValue("Attack", ee.Attack)
			e.RepsertPropertyValue("Defense", ee.Defense)
			e.RepsertPropertyValue("TeamID", p.Team)

			// this bypass actor's owning resource, we should probably use the AddEntity message instead (doesn't exist yet).
			battleArena.Ruler.AddEntity(e)
		}
	}

	// Start the Ruler actor now that initial configuration is complete.
	battleArena.Ruler.Start()

	for _, p := range start.Players {
		if p.IA {
			ctrl := controllers.NewAggressiveController(uuid.MustParse(p.ID), fmt.Sprintf("AggressiveController-%s", p.ID))
			ctrl.Start()

			msg := message.Create(ctrl, rulermethods.AddController{
				Controller:   ctrl,
				ControllerID: ctrl.ID,
			}, nil)

			battleArena.Ruler.SendActor(msg, respChan)

		} else {

			// We need at least one controller to get the initial state
			// In the future, we might add multiple based on players payload
			hc := NewHTTPController(uuid.MustParse(p.ID), matchID, start.CallbackURL)
			hc.Start()

			msg := message.Create(hc, rulermethods.AddController{
				Controller:   hc,
				ControllerID: hc.ID,
			}, nil)

			battleArena.Ruler.SendActor(msg, respChan)
		}
	}

	for i := 0; i < count; i++ {
		log.Printf("[ArenaBridge] Waiting for controller reply (%d/%d) for match %s", i+1, count, matchID)
		select {
		case msg := <-respChan:
			log.Printf("[ArenaBridge] Received reply from controller for match %s (Error: %v)", matchID, msg.HasError)
		case <-time.After(10 * time.Second):
			log.Printf("[ArenaBridge] TIMEOUT waiting for controller reply for match %s", matchID)
		}
	}

	res := make([]entity.Entity, 0, 6)
	for _, v := range battleArena.Ruler.GameState.Entities {
		res = append(res, v)
	}

	return matchID,
		battleArena.Ruler.GameState.Grid,
		res,
		start.Players,
		battleArena.Ruler.GameState.Turner.GetTurnState(),
		battleArena.Ruler.GameState.Version,
		nil
}

func (b *ArenaBridge) GetBoardState(matchID uuid.UUID, action *api.ActionFeedback) (api.BoardState, error) {
	b.mu.RLock()
	arena, ok := b.arenas[matchID]
	b.mu.RUnlock()
	if !ok {
		return api.BoardState{}, fmt.Errorf("arena %s not found", matchID)
	}

	res := make([]entity.Entity, 0, len(arena.Ruler.GameState.Entities))
	for _, v := range arena.Ruler.GameState.Entities {
		res = append(res, v)
	}

	players, _ := arena.Metadata["Players"].([]api.Player)

	return api.NewBoardState(matchID, arena.Ruler.GameState.Grid, res, players, arena.Ruler.GameState.Turner.GetTurnState(), time.Now(), time.Now().Add(30*time.Second), arena.Ruler.GameState.WinnerTeamID, arena.Ruler.GameState.Version, action), nil
}

type webhookSentKey struct {
	matchID   uuid.UUID
	version   int64
	eventType string
}

var lastSentWebhook = make(map[webhookSentKey]bool)
var lastSentMu sync.Mutex

// TrySendWebhook checks if a webhook for this version and event type has already been sent.
// Returns true if this is the first time this combination is being processed.
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

func (b *ArenaBridge) ArenaAction(arenaID uuid.UUID, req api.ArenaActionMessage) (bool, string, interface{}) {
	r, ok := b.GetArena(arenaID)
	if !ok {
		return false, "arena not found", nil
	}

	respChan := make(chan *message.Message)
	defer close(respChan)
	// Translate HTTP action to Ruler message
	// This is a simplified mapping; more logic needed for full support
	// Normalize type to lowercase for case-insensitive matching
	actionType := strings.ToLower(req.Data.Type)

	switch actionType {
	case "attack":
		r.SendActor(message.Create(nil, rulermethods.ControllerAttack{
			ControllerID: uuid.MustParse(req.Data.PlayerID),
			EntityID:     uuid.MustParse(req.Data.EntityID),
			Target:       position.New(req.Data.TargetCoords[0].X, req.Data.TargetCoords[0].Y, 1),
		}, nil), respChan)
	case "pass":
		r.SendActor(message.Create(nil, rulermethods.EndOfTurn{
			ControllerID: uuid.MustParse(req.Data.PlayerID),
			EntityID:     uuid.MustParse(req.Data.EntityID),
		}, nil), respChan)
	case "move":
		path := make([]position.Position, len(req.Data.TargetCoords))
		for i, c := range req.Data.TargetCoords {
			path[i] = position.New(c.X, c.Y, 1)
		}
		r.SendActor(message.Create(nil, rulermethods.ControllerMove{
			ControllerID: uuid.MustParse(req.Data.PlayerID),
			EntityID:     uuid.MustParse(req.Data.EntityID),
			Path:         path,
		}, nil), respChan)
	case "forfeit":
		entityID := uuid.Nil
		if req.Data.EntityID != "" {
			if uid, err := uuid.Parse(req.Data.EntityID); err == nil {
				entityID = uid
			}
		}
		r.SendActor(message.Create(nil, rulermethods.ControllerForfeit{
			ControllerID: uuid.MustParse(req.Data.PlayerID),
			EntityID:     entityID,
		}, nil), respChan)
	default:
		// Just notify the ruler for now with a generic message if type matches?
		// Better to implement specific methods

		r.SendActor(message.Create(nil, rulermethods.EndOfTurn{
			ControllerID: uuid.MustParse(req.Data.PlayerID),
			EntityID:     uuid.MustParse(req.Data.EntityID),
		}, nil), respChan)
	}

	// Wait for the reply
	res := <-respChan

	if res.HasError {
		return false, res.ErrorMessage, nil
	}

	return true, fmt.Sprintf("action %s accepted", req.Data.Type), res.Content
}

func (b *ArenaBridge) GetArena(id uuid.UUID) (*ruler.Ruler, bool) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	r, ok := b.arenas[id]
	if !ok {
		return nil, false
	}
	return r.Ruler, ok
}

// DestroyArena stops the Ruler and all controllers, then removes the arena from memory.
// @spec-link [[mech_arena_lifecycle]]
func (b *ArenaBridge) DestroyArena(matchID uuid.UUID) {
	b.mu.Lock()
	arena, ok := b.arenas[matchID]
	if ok {
		delete(b.arenas, matchID)
		delete(b.lastSentWebhookVersion, matchID)
	}
	b.mu.Unlock()

	if ok && arena.Ruler != nil {
		log.Printf("[ArenaBridge] Destroying arena %s", matchID)
		// Sending ActorStop to Ruler triggers cascading shutdown of controllers
		arena.Ruler.NotifyActor(message.Create(nil, actor.ActorStop{}, nil))
	}
}

// GetActiveMatchCount returns the number of active arenas.
func (b *ArenaBridge) GetActiveMatchCount() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.arenas)
}
