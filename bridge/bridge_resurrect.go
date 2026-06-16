// Package bridge provides the resurrection logic to recover engine state after service interruptions.
// It ensures that matches can be seamlessly resumed with full entity, buff, and initiative persistence.
package bridge

import (
	"fmt"
	"log"
	"time"

	"github.com/ecumeurs/upsilonapi/api"
	"github.com/ecumeurs/upsilonbattle/battlearena"
	"github.com/ecumeurs/upsilonbattle/battlearena/controller/controllers"
	"github.com/ecumeurs/upsilontypes/entity"
	"github.com/ecumeurs/upsilontypes/entity/skill"
	"github.com/ecumeurs/upsilonmapdata/grid"
	"github.com/ecumeurs/upsilonmapdata/grid/cell"
	"github.com/ecumeurs/upsilonmapdata/grid/position"
	"github.com/ecumeurs/upsilonserializer"
	"github.com/ecumeurs/upsilontypes/property"
	"github.com/ecumeurs/upsilontypes/property/def"
	"github.com/ecumeurs/upsilontools/tools/actor"
	"github.com/ecumeurs/upsilontools/tools/messagequeue/message"
	"github.com/ecumeurs/upsilonbattle/battlearena/ruler"
	"github.com/ecumeurs/upsilonbattle/battlearena/ruler/rulermethods"
	"github.com/ecumeurs/upsilonbattle/battlearena/ruler/turner"
	"github.com/google/uuid"
)

// @spec-link [[api_go_battle_engine]]
// @spec-link [[mechanic_mech_arena_lifecycle]]

// ResurrectArena rebuilds a crashed arena from a persisted board state (ISS-054).
func (b *ArenaBridge) ResurrectArena(req api.ArenaResurrectRequest) (api.BoardState, error) {
	// 1. Validation: Ensure all mandatory identifiers and rosters are present.
	matchID, err := validateResurrectRequest(req)
	if err != nil { return api.BoardState{}, err }
	// 2. Idempotency: Reject recovery if the match is already active in memory.
	if b.isArenaActive(matchID) {
		return api.BoardState{}, fmt.Errorf("arena %s is already running — resurrection not needed", matchID)
	}
	// 3. Schema-version guard (Crash-Early): Refuse blobs whose embedded serializer
	//    version is absent (zero) or does not match the current engine schema.
	//    A mismatch indicates a stale or incompatible blob that would silently
	//    mis-deserialize engine state (audit risk R7 / WP-D2).
	if err := validateSerializerVersion(req.SerializerVersion); err != nil { return api.BoardState{}, err }
	// 4. Grid Reconstruction: Rebuild the 3D engine grid from the 2D serialized projection.
	g := resurrectGrid(req.Grid)
	// 5. Arena Setup: Initialize the core BattleArena container and its metadata.
	battleArena := b.initResurrectedArena(matchID, g, req)
	// 6. Entity Restoration: Hydrate the engine with saved characters, stats, and buffs.
	if err := b.restoreEntities(battleArena, req); err != nil { return api.BoardState{}, err }
	// 7. State Recovery: Restore the initiative timeline and versioning metadata.
	currentEntityID := b.restoreEngineState(battleArena, req)
	// 8. Lifecycle: Start the Ruler actor so it can process controller registrations.
	battleArena.Ruler.Start()
	// 9. Connectivity: Re-establish human and AI controllers for the session.
	b.reconnectControllers(battleArena, matchID, req)
	// 10. Hand-off: Signal the Ruler to resume tactical turn execution.
	b.registerAndHandOff(matchID, battleArena, currentEntityID)
	// 11. Response: Return a full board state snapshot for client synchronization.
	return b.buildResurrectionBoardState(matchID, battleArena, req), nil
}

// validateSerializerVersion enforces the Crash-Early schema-version guard (WP-D2 / audit R7).
// It returns a descriptive error when the blob's embedded serializer_version is absent (zero)
// or does not match the engine's current schema, preventing silent mis-deserialization.
func validateSerializerVersion(found int) error {
	want := upsilonserializer.CurrentSerializerVersion
	if found == 0 {
		return fmt.Errorf(
			"resurrection refused: serializer_version is absent in the persisted blob — "+
				"this blob was written before schema versioning was introduced; "+
				"expected serializer_version=%d, got 0 (treat as stale, conclude the match server-side)",
			want,
		)
	}
	if found != want {
		return fmt.Errorf(
			"resurrection refused: serializer_version mismatch — "+
				"expected %d, got %d; the persisted blob shape is incompatible with the current engine",
			want, found,
		)
	}
	return nil
}

// validateResurrectRequest checks for missing fields and malformed UUIDs.
func validateResurrectRequest(req api.ArenaResurrectRequest) (uuid.UUID, error) {
	// 1. ID Verification: Ensure match_id exists and is a valid UUID string.
	if req.MatchID == "" { return uuid.Nil, fmt.Errorf("mandatory field match_id is missing") }
	matchID, err := uuid.Parse(req.MatchID)
	if err != nil { return uuid.Nil, fmt.Errorf("invalid match_id format: %w", err) }
	// 2. Target Routing: Verify that a callback URL is provided for event delivery.
	if req.CallbackURL == "" { return uuid.Nil, fmt.Errorf("mandatory field callback_url is missing") }
	// 3. Roster Check: Resurrection requires at least one participating player.
	if len(req.Players) == 0 { return uuid.Nil, fmt.Errorf("arena roster must not be empty") }
	return matchID, nil
}

// isArenaActive checks the thread-safe registry for an existing arena instance.
func (b *ArenaBridge) isArenaActive(matchID uuid.UUID) bool {
	// 1. Thread Safety: Use RLock for safe concurrent registry access.
	b.mu.RLock()
	defer b.mu.RUnlock()
	_, exists := b.arenas[matchID]
	return exists
}

// initResurrectedArena creates a new BattleArena with the restored grid and metadata.
func (b *ArenaBridge) initResurrectedArena(matchID uuid.UUID, g *grid.Grid, req api.ArenaResurrectRequest) *battlearena.BattleArena {
	// 1. Setup: Instantiate core BattleArena with the match ID.
	ba := battlearena.NewBattleArena(matchID)
	// 2. Metadata: Inject callback URL and original player list.
	ba.Metadata["CallbackURL"] = req.CallbackURL
	ba.Metadata["Players"] = req.Players
	// 3. Ruler Config: Hydrate Ruler with reconstructed grid and player count.
	ba.Ruler.ID = matchID
	ba.Ruler.SetGrid(g)
	ba.Ruler.SetNbControllers(len(req.Players))
	return ba
}

// restoreEntities populates the engine with entities from the resurrection request.
func (b *ArenaBridge) restoreEntities(ba *battlearena.BattleArena, req api.ArenaResurrectRequest) error {
	// 1. Population: Re-create entities for every player in the roster.
	for _, p := range req.Players {
		pID, _ := uuid.Parse(p.ID)
		for _, ee := range p.Entities {
			// 2. Vitality: Only resurrect entities that have positive HP and are not dead.
			if ee.Dead || ee.HP <= 0 { continue }
			// 3. Conversion: Map the DTO back into the engine's internal Entity model.
			ent, err := b.dtoToEntity(pID, ee, p.Team, ba.Ruler.GameState.Grid)
			if err != nil { return err }
			// 4. Registration: Push the hydrated entity into the Ruler state.
			ba.Ruler.AddEntity(ent)
		}
	}
	return nil
}

// dtoToEntity converts a single entity DTO into an internal engine Entity.
func (b *ArenaBridge) dtoToEntity(pID uuid.UUID, ee api.Entity, team int, g *grid.Grid) (entity.Entity, error) {
	// 1. Core: Initialize internal Entity with base stats and calculated height-position.
	entID, err := uuid.Parse(ee.ID)
	if err != nil { return entity.Entity{}, err }
	e := entity.Entity{
		ID: entID, Type: entity.Character, Name: ee.Name, ControllerID: pID,
		Properties: make(map[string]property.Property), Skills: make(map[uuid.UUID]skill.Skill),
		Position: position.New(ee.Position.X, ee.Position.Y, g.TopMostGroundAt(ee.Position.X, ee.Position.Y)),
	}
	// 2. Properties: Restore current and max HP, Movement, Attack, Defense, and Team ID.
	e.RepsertPropertyCMaxValue(property.HP, ee.MaxHP)
	e.RepsertPropertyCValue(property.HP, ee.HP)
	e.RepsertPropertyCMaxValue(property.Movement, ee.MaxMove)
	e.RepsertPropertyCValue(property.Movement, ee.Move)
	e.RepsertPropertyValue(property.Attack, ee.Attack)
	e.RepsertPropertyValue(property.Defense, ee.Defense)
	e.RepsertPropertyValue(property.TeamID, team)
	// 3. Accessories: Hydrate buffs and skills for the entity.
	b.restoreEntityBuffs(&e, ee.Buffs)
	b.restoreEntitySkills(&e, ee.EquippedSkills)
	return e, nil
}

// restoreEntityBuffs re-injects saved buffs into an entity during resurrection.
func (b *ArenaBridge) restoreEntityBuffs(e *entity.Entity, buffs []api.Buff) {
	// 1. Buff Hydration: Iterate through every saved status effect.
	for _, b := range buffs {
		originID, err := uuid.Parse(b.OriginID)
		if err != nil { continue }
		buff := property.TemporaryProperties{
			Forever: b.Forever, OriginEntityID: originID, Properties: make(map[string]property.Property),
		}
		// 2. Property Mapping: Resolve each buffered property back to its engine type.
		for key, dto := range b.Properties.Data {
			hydrateSingleBuffProperty(key, dto, &buff)
		}
		e.RegisterBuff(buff)
	}
}

// hydrateSingleBuffProperty performs the individual property resolution for a buff.
func hydrateSingleBuffProperty(key string, dto api.PropertyDTO, buff *property.TemporaryProperties) {
	// 1. Alias Resolution: Check for property naming consistency.
	effectiveKey := key
	if alias, ok := propertyAliasMap[effectiveKey]; ok { effectiveKey = alias }
	// 2. Definition Lookup: Identify if property belongs to Item or Entity family.
	var p property.Property
	if prop := def.ItemProperty(property.ItemProperties(effectiveKey)); prop != nil {
		p = prop
	} else if prop := def.EntityProperty(property.EntityProperties(effectiveKey)); prop != nil {
		p = prop
	}
	// 3. Hydration: If valid, set the value and register in the buff container.
	if p != nil && setSkillPropValue(p, dto) {
		buff.Properties[property.PropertyToString(effectiveKey)] = p
	}
}

// restoreEntitySkills populates an entity's skill map from the resurrection payload.
func (b *ArenaBridge) restoreEntitySkills(e *entity.Entity, skills []api.EquippedSkill) {
	// 1. Skill Hydration: Map every tactical ability back into the entity registry.
	for _, es := range skills {
		skillID, err := uuid.Parse(es.SkillID)
		if err != nil { continue }
		// 2. Behavior Config: Reconstruct the behavior and targeting logic.
		bh := def.DefaultBehavior()
		bh.SetS(string(parseBehaviorType(es.Behavior)))
		s := skill.Skill{
			ID: skillID, Name: es.Name, Behavior: bh,
			Targeting: buildSkillPropertyMap(es.Targeting.Data),
			Costs:     buildSkillPropertyMap(es.Costs.Data),
			Effect:    buildSkillEffect(es.Effect.Data),
		}
		// 3. Registration: Inject into engine state.
		e.RegisterSkill(s)
	}
}

// restoreEngineState recovers the initiative timeline and current turn pointers.
func (b *ArenaBridge) restoreEngineState(ba *battlearena.BattleArena, req api.ArenaResurrectRequest) uuid.UUID {
	// 1. Timeline: Map saved turns into internal initiative queue structs.
	var turnerTurns []turner.EntityTurn
	for _, t := range req.Turns {
		entID, err := uuid.Parse(t.EntityID)
		if err != nil { continue }
		turnerTurns = append(turnerTurns, turner.EntityTurn{EntityId: entID, Delay: t.Delay})
	}
	// 2. Pointer Resolution: Identify the current acting entity and versioning.
	currentEntityID := uuid.Nil
	if req.CurrentEntityID != "" { currentEntityID, _ = uuid.Parse(req.CurrentEntityID) }
	// 3. Hydration: Push the initiative state into the Ruler.
	ba.Ruler.Resurrect(turnerTurns, currentEntityID, req.Version)
	return currentEntityID
}

// reconnectControllers re-instantiates human and AI controllers for the match.
func (b *ArenaBridge) reconnectControllers(ba *battlearena.BattleArena, matchID uuid.UUID, req api.ArenaResurrectRequest) {
	// 1. Human Players: Identify represented player IDs for the shared proxy.
	var humanIDs []uuid.UUID
	for _, p := range req.Players {
		if !p.IA { humanIDs = append(humanIDs, uuid.MustParse(p.ID)) }
	}
	// 2. Proxy Initialization: Re-create the HTTPController for web callbacks.
	var sharedHC *HTTPController
	if len(humanIDs) > 0 {
		sharedHC = NewHTTPController(matchID, req.CallbackURL, req.Players, humanIDs)
		sharedHC.Ruler = ba.Ruler
		sharedHC.Start()
	}
	// 3. Handshake: Connect every player to the Ruler actor.
	respChan := make(chan *message.Message, len(req.Players))
	for _, p := range req.Players {
		b.registerSingleController(ba.Ruler, p, sharedHC, respChan)
	}
	// 4. Synchronization: Wait for Ruler acknowledgements.
	b.waitForControllerReplies(matchID, len(req.Players), respChan)
}

// registerSingleController connects a single player to the engine ruler.
func (b *ArenaBridge) registerSingleController(r *ruler.Ruler, p api.Player, sharedHC *HTTPController, resp chan *message.Message) {
	// 1. Actor Choice: Instantiate Aggressive (AI) or Proxy (Human) controller.
	var ctrl actor.Communication
	pID := uuid.MustParse(p.ID)
	if p.IA {
		iac := controllers.NewAggressiveController(pID, fmt.Sprintf("AggressiveController-%s", p.ID))
		iac.Start()
		ctrl = iac
	} else {
		ctrl = sharedHC
	}
	// 2. Messaging: Send AddController request to the Ruler actor.
	msg := message.Create(ctrl, rulermethods.AddController{Controller: ctrl, ControllerID: pID}, rulermethods.AddControllerReply{})
	r.SendActor(msg, resp)
}

// waitForControllerReplies blocks until the Ruler confirms all player connections.
func (b *ArenaBridge) waitForControllerReplies(matchID uuid.UUID, count int, resp chan *message.Message) {
	// 1. Barrier: Block until all controller-registration replies are received.
	for i := 0; i < count; i++ {
		select {
		case <-resp:
		case <-time.After(5 * time.Second):
			log.Printf("[ArenaBridge.Resurrect] TIMEOUT: %s", matchID)
		}
	}
}

// registerAndHandOff finishes the resurrection by handshaking with the engine logic.
func (b *ArenaBridge) registerAndHandOff(matchID uuid.UUID, ba *battlearena.BattleArena, currentEntityID uuid.UUID) {
	// 1. Registry: Add the live arena to the bridge singleton.
	b.mu.Lock()
	b.arenas[matchID] = ba
	b.mu.Unlock()
	// 2. Hand-off: Signal the Ruler to resume tactical turn execution.
	ba.Ruler.NotifyActor(message.Create(nil, rulermethods.Resurrect{CurrentEntityID: currentEntityID}, nil))
}

// buildResurrectionBoardState generates the final board state snapshot after recovery.
func (b *ArenaBridge) buildResurrectionBoardState(matchID uuid.UUID, ba *battlearena.BattleArena, req api.ArenaResurrectRequest) api.BoardState {
	// 1. Entity Extraction: Gather current engine entity snapshots.
	entities := make([]entity.Entity, 0, len(ba.Ruler.GameState.Entities))
	for _, v := range ba.Ruler.GameState.Entities { entities = append(entities, v) }
	// 2. Projection: Build the API-standard BoardState snapshot.
	return api.NewBoardState(
		matchID, ba.Ruler.GameState.Grid, entities, req.Players,
		ba.Ruler.GameState.Turner.GetTurnState(), time.Now(), time.Now().Add(30*time.Second),
		0, req.Version, nil,
	)
}

// resurrectGrid reconstructs a 3D engine grid from the 2D serialized projection.
func resurrectGrid(rg api.ResurrectGrid) *grid.Grid {
	// 1. Dimensioning: Setup core grid with saved width and depth limits.
	g := &grid.Grid{
		Width: rg.Width, Length: rg.Height, Height: rg.MaxHeight,
		Cells: make(map[position.Position]*cell.Cell),
	}
	// 2. Projection: Process each column to rebuild vertical stacks.
	for x := 0; x < rg.Width; x++ {
		for y := 0; y < rg.Height; y++ {
			resurrectGridColumn(g, rg, x, y)
		}
	}
	return g
}

// resurrectGridColumn rebuilds a single (x,y) vertical stack of cells.
func resurrectGridColumn(g *grid.Grid, rg api.ResurrectGrid, x int, y int) {
	// 1. Surface: Extract height and obstacle status for the topmost tile.
	surfaceZ, isObstacle := 0, false
	if x < len(rg.Cells) && y < len(rg.Cells[x]) {
		surfaceZ = rg.Cells[x][y].Height
		isObstacle = rg.Cells[x][y].Obstacle
	}
	// 2. Foundation: Create solid dirt volume up to the surface elevation.
	for z := 0; z < surfaceZ; z++ {
		pos := position.New(x, y, z)
		g.Cells[pos] = cell.NewCell(cell.Dirt, pos)
	}
	// 3. Playable Tile: Place Ground or Obstacle cell at the target Z.
	pos := position.New(x, y, surfaceZ)
	cellType := cell.Ground
	if isObstacle { cellType = cell.Obstacle }
	g.Cells[pos] = cell.NewCell(cellType, pos)
}
