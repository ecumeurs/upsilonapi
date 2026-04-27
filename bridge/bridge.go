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
	"github.com/ecumeurs/upsilontypes/entity"
	"github.com/ecumeurs/upsilontypes/entity/skill"
	"github.com/ecumeurs/upsilontypes/property"
	"github.com/ecumeurs/upsilontypes/property/def"
	"github.com/ecumeurs/upsilontypes/property/effect"
	"github.com/ecumeurs/upsilonbattle/battlearena/ruler"
	"github.com/ecumeurs/upsilonbattle/battlearena/ruler/rulermethods"
	"github.com/ecumeurs/upsilonbattle/battlearena/ruler/turner"
	"github.com/ecumeurs/upsilonmapdata/grid"
	"github.com/ecumeurs/upsilonmapdata/grid/position"
	"github.com/ecumeurs/upsilonmapmaker/gridgenerator"
	"github.com/ecumeurs/upsilontools/tools"
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

var propertyAliasMap = map[string]string{
	"ArmorRating": "Armor",
	"CritChance":  "CriticalChance",
	"CritDamage":  "CriticalMultiplier",
}

var bridge = &ArenaBridge{
	arenas:                 make(map[uuid.UUID]*battlearena.BattleArena),
	lastSentWebhookVersion: make(map[uuid.UUID]int64),
}

func Get() *ArenaBridge {
	return bridge
}

func (b *ArenaBridge) StartArena(start api.ArenaStartRequest) (uuid.UUID, *grid.Grid, []entity.Entity, []api.Player, turner.TurnState, int64, error) {
	if start.MatchID == "" {
		return uuid.Nil, nil, nil, nil, turner.TurnState{}, 0, fmt.Errorf("mandatory field match_id is missing")
	}
	matchID, err := uuid.Parse(start.MatchID)
	if err != nil {
		return uuid.Nil, nil, nil, nil, turner.TurnState{}, 0, fmt.Errorf("invalid match_id: %w", err)
	}

	if start.CallbackURL == "" {
		return uuid.Nil, nil, nil, nil, turner.TurnState{}, 0, fmt.Errorf("mandatory field callback_url is missing")
	}

	if len(start.Players) == 0 {
		return uuid.Nil, nil, nil, nil, turner.TurnState{}, 0, fmt.Errorf("arena must have at least one player")
	}

	battleArena := battlearena.NewBattleArena(matchID)
	battleArena.Metadata["CallbackURL"] = start.CallbackURL
	battleArena.Metadata["Players"] = start.Players

	// Ensure Ruler ID matches MatchID as per caller expectations
	battleArena.Ruler.ID = matchID

	b.mu.Lock()
	b.arenas[matchID] = battleArena
	b.mu.Unlock()

	// Default to a reliable 7x7 for both tests and production baseline.
	// Reduced from 8x8/Hill to 7x7/Flat to avoid impassable cliffs that break bot navigation (ISS-087).
	gg := gridgenerator.GridGenerator{
		Width:               tools.NewIntRange(7, 8),
		Length:              tools.NewIntRange(7, 8),
		Height:              tools.NewIntRange(2, 3),
		Type:                gridgenerator.Flat,
		GenerateObstrcution: true,
		ObstructionRate:     tools.NewIntRange(2, 8), // 2-8% obstruction for manageable tactical depth
	}
	battleArena.Ruler.SetGrid(gg.Generate())
	battleArena.Ruler.SetNbControllers(len(start.Players))

	// We need to wait for the reply to get the initial state
	respChan := make(chan *message.Message)
	defer close(respChan)

	count := len(start.Players)

	for _, p := range start.Players {
		playerID, err := uuid.Parse(p.ID)
		if err != nil {
			return uuid.Nil, nil, nil, nil, turner.TurnState{}, 0, fmt.Errorf("invalid player_id for player %s: %w", p.Nickname, err)
		}

		for _, ee := range p.Entities {
			entID, err := uuid.Parse(ee.ID)
			if err != nil {
				return uuid.Nil, nil, nil, nil, turner.TurnState{}, 0, fmt.Errorf("invalid entity_id for entity %s: %w", ee.Name, err)
			}

			// CRASH EARLY: Critical stats must be positive
			if ee.MaxHP <= 0 {
				return uuid.Nil, nil, nil, nil, turner.TurnState{}, 0, fmt.Errorf("entity %s must have max_hp > 0", ee.Name)
			}

			// Clean initialization (no random defaults)
			e := entity.Entity{
				ID:           entID,
				Type:         entity.Character,
				Name:         ee.Name,
				ControllerID: playerID,
			}
			e.Properties = make(map[string]property.Property)
			e.Skills = make(map[uuid.UUID]skill.Skill)
			if ee.Position.X != 0 || ee.Position.Y != 0 {
				e.Position = position.New(ee.Position.X, ee.Position.Y, battleArena.Ruler.GameState.Grid.TopMostGroundAt(ee.Position.X, ee.Position.Y))
			}

			e.RepsertPropertyCMaxValue(property.HP, ee.MaxHP)
			e.RepsertPropertyCValue(property.HP, ee.HP)
			e.RepsertPropertyCMaxValue(property.Movement, ee.MaxMove)
			e.RepsertPropertyCValue(property.Movement, ee.Move)
			e.RepsertPropertyValue(property.Attack, ee.Attack)
			e.RepsertPropertyValue(property.Defense, ee.Defense)
			e.RepsertPropertyValue(property.TeamID, p.Team)

			// Load equipped items as buffs
			// @spec-link [[mec_item_buff_application]]
			for _, item := range ee.EquippedItems {
				itemID, err := uuid.Parse(item.ItemID)
				if err != nil {
					log.Printf("[ArenaBridge] Skipping item %s for entity %s: invalid UUID", item.Name, ee.Name)
					continue
				}

				buff := property.TemporaryProperties{
					Forever:        true,
					OriginEntityID: itemID,
					Properties:     make(map[string]property.Property),
				}

				for key, raw := range item.Properties {
					// Handle common aliases (e.g. ArmorRating -> Armor)
					effectiveKey := key
					if alias, ok := propertyAliasMap[effectiveKey]; ok {
						effectiveKey = alias
					}

					// Map string key to Property. Item properties take priority, then Entity properties.
					var p property.Property
					if prop := def.ItemProperty(property.ItemProperties(effectiveKey)); prop != nil {
						p = prop
					} else if prop := def.EntityProperty(property.EntityProperties(effectiveKey)); prop != nil {
						p = prop
					}

					if p != nil {
						// Handle JSON number decoding (float64 to int)
						if f, ok := raw.(float64); ok {
							p.Set(int(f))
						} else if i, ok := raw.(int); ok {
							p.Set(i)
						} else {
							p.Set(raw)
						}
						buff.Properties[property.PropertyToString(effectiveKey)] = p
					}
				}
				e.RegisterBuff(buff)
			}

			// Load equipped skills from payload
			// @spec-link [[mec_skill_payload_resolution]]
			// @spec-link [[api_character_skill_inventory]]
			for _, es := range ee.EquippedSkills {
				skillID, err := uuid.Parse(es.SkillID)
				if err != nil {
					log.Printf("[ArenaBridge] Skipping skill %s for entity %s: invalid UUID", es.Name, ee.Name)
					continue
				}
				s := skill.Skill{
					ID:        skillID,
					Name:      es.Name,
					Behavior:  *def.MakeBehaviorProperty(parseBehaviorType(es.Behavior)),
					Targeting: buildSkillPropertyMap(es.Targeting),
					Costs:     buildSkillPropertyMap(es.Costs),
					Effect:    buildSkillEffect(es.Effect),
				}
				e.RegisterSkill(s)
			}

			battleArena.Ruler.AddEntity(e)
		}
	}

	// Start the Ruler actor now that initial configuration is complete.
	battleArena.Ruler.Start()

	for _, p := range start.Players {
		var ctrl actor.Communication
		pID := uuid.MustParse(p.ID) // Safe here as we already parsed it above
		if p.IA {
			iac := controllers.NewAggressiveController(pID, fmt.Sprintf("AggressiveController-%s", p.ID))
			iac.Start()
			ctrl = iac
		} else {
			hc := NewHTTPController(pID, matchID, start.CallbackURL, start.Players)
			hc.Ruler = battleArena.Ruler
			hc.Start()
			ctrl = hc
		}

		msg := message.Create(ctrl, rulermethods.AddController{
			Controller:   ctrl,
			ControllerID: pID,
		}, rulermethods.AddControllerReply{})

		battleArena.Ruler.SendActor(msg, respChan)
	}

	for i := 0; i < count; i++ {
		log.Printf("[ArenaBridge] Waiting for AddController reply (%d/%d) for match %s", i+1, count, matchID)
		select {
		case msg := <-respChan:
			log.Printf("[ArenaBridge] Received AddController reply for match %s (Error: %v)", matchID, msg.HasError)
		case <-time.After(5 * time.Second):
			log.Printf("[ArenaBridge] TIMEOUT waiting for AddController reply for match %s", matchID)
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

	// Request board state from Ruler via message to avoid data races
	// @spec-link [[api_go_battle_action]]
	respChan := make(chan *message.Message, 1)
	arena.Ruler.SendActor(message.Create(nil, rulermethods.GetBoardState{
		ActionContext: action,
	}, rulermethods.GetBoardStateReply{}), respChan)

	select {
	case res := <-respChan:
		if res.HasError {
			return api.BoardState{}, fmt.Errorf("engine error: %s", res.ErrorMessage)
		}
		reply := res.TargetMethod.(rulermethods.GetBoardStateReply)
		players, _ := arena.Metadata["Players"].([]api.Player)

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

// ArenaAction proxies a tactical command to the Ruler. It returns
// (ok, message, errorKey, data). errorKey is populated only on failure and
// mirrors the engine's ReplyWithError key; it lands on the external envelope
// as `meta.error_key` via [[api_standard_envelope]].
func (b *ArenaBridge) ArenaAction(arenaID uuid.UUID, req api.ArenaActionMessage) (bool, string, string, interface{}) {
	r, ok := b.GetArena(arenaID)
	if !ok {
		return false, "arena not found", "arena.notfound", nil
	}

	playerID, err := uuid.Parse(req.Data.PlayerID)
	if err != nil {
		return false, fmt.Sprintf("invalid player_id: %v", err), "request.player_id.invalid", nil
	}

	entityID, err := uuid.Parse(req.Data.EntityID)
	if err != nil {
		return false, fmt.Sprintf("invalid entity_id: %v", err), "request.entity_id.invalid", nil
	}

	respChan := make(chan *message.Message)
	defer close(respChan)
	// Translate HTTP action to Ruler message
	// This is a simplified mapping; more logic needed for full support
	// Normalize type to lowercase for case-insensitive matching
	actionType := strings.ToLower(req.Data.Type)

	switch actionType {
	case "attack":
		if len(req.Data.TargetCoords) == 0 {
			return false, "attack requires target_coords", "request.target_coords.missing", nil
		}
		r.SendActor(message.Create(nil, rulermethods.ControllerAttack{
			ControllerID: playerID,
			EntityID:     entityID,
			Target:       position.New(req.Data.TargetCoords[0].X, req.Data.TargetCoords[0].Y, r.GameState.Grid.TopMostGroundAt(req.Data.TargetCoords[0].X, req.Data.TargetCoords[0].Y)),
		}, rulermethods.ControllerAttackReply{}), respChan)
	case "pass":
		r.SendActor(message.Create(nil, rulermethods.EndOfTurn{
			ControllerID: playerID,
			EntityID:     entityID,
		}, rulermethods.EndOfTurn{}), respChan)
	case "move":
		if len(req.Data.TargetCoords) == 0 {
			return false, "move requires target_coords", "request.target_coords.missing", nil
		}
		path := make([]position.Position, len(req.Data.TargetCoords))
		for i, c := range req.Data.TargetCoords {
			path[i] = position.New(c.X, c.Y, r.GameState.Grid.TopMostGroundAt(c.X, c.Y))
		}
		r.SendActor(message.Create(nil, rulermethods.ControllerMove{
			ControllerID: playerID,
			EntityID:     entityID,
			Path:         path,
		}, rulermethods.ControllerMoveReply{}), respChan)
	default:
		// Just notify the ruler for now with a generic message if type matches?
		// Better to implement specific methods

		r.SendActor(message.Create(nil, rulermethods.EndOfTurn{
			ControllerID: playerID,
			EntityID:     entityID,
		}, rulermethods.EndOfTurn{}), respChan)
	}

	// Wait for the reply
	res := <-respChan

	if res.HasError {
		return false, res.ErrorMessage, res.ErrorKey, nil
	}

	return true, fmt.Sprintf("action %s accepted", req.Data.Type), "", res.Content
}

// ArenaForfeit allows a player to concede the match without an entity context.
// Returns (ok, message, errorKey, data). errorKey mirrors the engine reply so
// the external envelope can surface it as meta.error_key.
// @spec-link [[api_go_battle_forfeit]]
func (b *ArenaBridge) ArenaForfeit(arenaID uuid.UUID, playerID uuid.UUID) (bool, string, string, interface{}) {
	r, ok := b.GetArena(arenaID)
	if !ok {
		return false, "arena not found", "arena.notfound", nil
	}

	respChan := make(chan *message.Message)
	defer close(respChan)

	r.SendActor(message.Create(nil, rulermethods.ControllerForfeit{
		ControllerID: playerID,
		EntityID:     uuid.Nil, // Forfeiting is team-wide
	}, rulermethods.ControllerForfeit{}), respChan)

	// Wait for the reply
	res := <-respChan

	if res.HasError {
		return false, res.ErrorMessage, res.ErrorKey, nil
	}

	return true, "forfeit accepted", "", res.Content
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

// ── Skill payload helpers ─────────────────────────────────────────────────

// parseBehaviorType converts the wire string to a def.BehaviorType.
// @spec-link [[mec_skill_payload_resolution]]
func parseBehaviorType(s string) def.BehaviorType {
	switch s {
	case "Reaction":
		return def.BehaviorTypeReaction
	case "Passive":
		return def.BehaviorTypePassive
	case "Counter":
		return def.BehaviorTypeCounter
	case "Trap":
		return def.BehaviorTypeTrap
	default:
		return def.BehaviorTypeDirect
	}
}

// setSkillPropValue applies a JSON-decoded value to a property.Property.
// Supports plain float64 (int wire) and {"value":X,"max":Y} for counters.
func setSkillPropValue(prop property.Property, val interface{}) bool {
	switch v := val.(type) {
	case float64:
		prop.Set(int(v))
		return true
	case map[string]interface{}:
		cp, ok := prop.(property.IntCounterProperty)
		if !ok {
			return false
		}
		if raw, ok := v["value"]; ok {
			if f, ok := raw.(float64); ok {
				cp.SetValue(int(f))
			}
		}
		if raw, ok := v["max"]; ok {
			if f, ok := raw.(float64); ok {
				cp.SetMaxValue(int(f))
			}
		}
		return true
	}
	return false
}

// buildSkillPropertyMap reconstructs a Targeting or Costs property map from
// the JSON payload. Unknown keys are silently skipped.
func buildSkillPropertyMap(raw map[string]interface{}) map[string]property.Property {
	result := make(map[string]property.Property)
	for key, val := range raw {
		prop := def.SkillProperty(property.SkillProperties(key))
		if prop == nil {
			continue
		}
		if setSkillPropValue(prop, val) {
			result[key] = prop
		}
	}
	return result
}

// buildSkillEffect reconstructs an effect.Effect from the JSON payload.
func buildSkillEffect(raw map[string]interface{}) effect.Effect {
	eff := *effect.New()
	for key, val := range raw {
		prop := def.SkillProperty(property.SkillProperties(key))
		if prop == nil {
			continue
		}
		if setSkillPropValue(prop, val) {
			eff.Properties = append(eff.Properties, prop)
		}
	}
	return eff
}
