// Package api provides the Data Transfer Objects (DTOs) and conversion logic for the Upsilon Hub API.
// It bridges the internal engine structures with the external JSON-based communication layer.
package api

import (
	"fmt"
	"time"

	"github.com/ecumeurs/upsilonapi/stdmessage"
	"github.com/ecumeurs/upsilonbattle/battlearena/ruler/turner"
	"github.com/ecumeurs/upsilonmapdata/grid"
	"github.com/ecumeurs/upsilonmapdata/grid/cell"
	"github.com/ecumeurs/upsilonmapdata/grid/position"
	"github.com/ecumeurs/upsilonserializer"
	"github.com/ecumeurs/upsilontypes/entity"
	"github.com/ecumeurs/upsilontypes/entity/skill"
	"github.com/ecumeurs/upsilontypes/property"
	"github.com/ecumeurs/upsilontypes/property/def"
	"github.com/google/uuid"
)

// @spec-link [[api_go_battle_engine]]
// @spec-link [[mechanic_mec_skill_payload_resolution]]

// ArenaActionResponse is the payload returned by POST /v1/matches/{id}/actions.
type ArenaActionResponse struct {
	// Status indicates if the action was accepted ("ok").
	Status string `json:"status"`
}

// ArenaStartResponse is the payload returned by POST /v1/matches/start.
type ArenaStartResponse struct {
	// ArenaID is the unique identifier for the started battle.
	ArenaID string `json:"arena_id"`
	// InitialState contains the full board and entity layout at turn 0.
	InitialState BoardState `json:"initial_state"`
}

// ArenaStartResponseMessage is the standard success envelope for match initialization.
type ArenaStartResponseMessage = stdmessage.StandardMessage[ArenaStartResponse, stdmessage.MetaNil]

// ActiveMatchStatsResponse returns quantitative data about running arenas.
type ActiveMatchStatsResponse struct {
	// ActiveCount is the total number of non-archived matches.
	ActiveCount int `json:"active_count"`
}

// ArenaExistsResponse indicates if a specific match is currently registered.
type ArenaExistsResponse struct {
	// Exists is true if the match ID is found in the bridge memory.
	Exists bool `json:"exists"`
}

// SkillGenerateResponse is the payload returned by POST /v1/skills/generate.
type SkillGenerateResponse struct {
	ID             string            `json:"id"`
	Name           string            `json:"name"`
	Behavior       string            `json:"behavior"`
	Targeting      Flex[PropertyMap] `json:"targeting"`
	Costs          Flex[PropertyMap] `json:"costs"`
	Effect         Flex[PropertyMap] `json:"effect"`
	Grade          string            `json:"grade"`
	Tags           []string          `json:"tags"`
	WeightPositive int               `json:"weight_positive"`
	WeightNegative int               `json:"weight_negative"`
}

// Cell is the topmost cell at a given (x, y) column of the engine grid.
type Cell struct {
	EntityID string `json:"entity_id"` // if any
	Obstacle bool   `json:"obstacle"`  // if any
	Height   int    `json:"height"`    // Z index of the topmost cell at (x, y); surface elevation
}

// Grid: A 2D projection of the engine's 3D grid.
type Grid struct {
	Width     int      `json:"width"`
	Height    int      `json:"height"`
	MaxHeight int      `json:"max_height"` // ceiling Z of the engine grid (exclusive upper bound)
	Cells     [][]Cell `json:"cells"`      // Cells are stored in width-major order.
}

// Turn represents an entry in the initiative timeline.
type Turn struct {
	PlayerID string `json:"player_id"`
	Delay    int    `json:"delay"`
	EntityID string `json:"entity_id"`
}

// CreditAward tracks combat performance metrics for rewards.
type CreditAward struct {
	PlayerID string `json:"player_id"`
	Amount   int    `json:"amount"`
	Source   string `json:"source"` // damage, healing, status
}

// ActionResult provides explicit data about the impact on a single target.
type ActionResult struct {
	TargetID string        `json:"target_id"`
	Damage   int           `json:"damage,omitempty"`
	Heal     int           `json:"heal,omitempty"`
	PrevHP   int           `json:"prev_hp"`
	NewHP    int           `json:"new_hp"`
	Credits  []CreditAward `json:"credits,omitempty"`
}

// ActionFeedback provides explicit data about the last tactical action.
type ActionFeedback struct {
	Type     string              `json:"type"` // "move", "attack", "skill", "pass"
	ActorID  string              `json:"actor_id"`
	TargetID string              `json:"target_id,omitempty"` // Legacy/Primary target
	Path     []position.Position `json:"path,omitempty"`
	Results  []ActionResult      `json:"results,omitempty"`
	Credits  []CreditAward       `json:"credits,omitempty"` // Global action credits
}

// BoardState represents the current state of the board.
type BoardState struct {
	Players           []Player        `json:"players"` // Consolidated roster
	Grid              Grid            `json:"grid"`
	Turn              []Turn          `json:"turn"`
	CurrentPlayerID   string          `json:"current_player_id"`
	CurrentEntityID   string          `json:"current_entity_id"`
	Timeout           time.Time       `json:"timeout"`
	StartTime         time.Time       `json:"start_time"`
	WinnerTeamID      *int            `json:"winner_team_id"`
	Action            *ActionFeedback `json:"action,omitempty"`
	Version           int64           `json:"version"`
	// SerializerVersion is the embedded schema version stamped into every persisted blob.
	// It must match upsilonserializer.CurrentSerializerVersion on resurrection; a mismatch
	// or absent value (zero) triggers an explicit Crash-Early error rather than a silent
	// mis-deserialization.
	SerializerVersion int             `json:"serializer_version"`
}

// ArenaEvent is the payload for the webhook
type ArenaEvent struct {
	MatchID   string     `json:"match_id"`   // targeted match
	EventType string     `json:"event_type"` // Board State Change, Turn Started, Battle Start, Battle End
	PlayerID  string     `json:"player_id"`  // if set, targeted player
	EntityID  string     `json:"entity_id"`  // if set, targeted entity
	Data      BoardState `json:"data"`       // event specific data (board change)
	Version   int64      `json:"version"`    // version number
	Timeout   time.Time  `json:"timeout"`    // End of turn date.
}

// NewError creates a new StandardMessage with the given error.
func NewError(requestId string, err string) stdmessage.StandardMessage[stdmessage.DataNil, stdmessage.MetaNil] {
	// 1. Envelope Initialization: Create the baseline failure container.
	// 2. Failure Identification: Set success flag to false for the consumer.
	// 3. Error Content: Inject the human-readable error description string.
	// 4. Traceability: Pass the original request identifier for correlation.
	// 5. Data Padding: Provide empty map to satisfy polymorphic unmarshalers.
	return stdmessage.StandardMessage[stdmessage.DataNil, stdmessage.MetaNil]{
		RequestID: requestId,
		Message:   err,
		Meta:      stdmessage.MetaNil{},
		Success:   false,
		Data:      stdmessage.DataNil{},
	}
}

// NewErrorWithKey returns a standard-envelope error message that also carries an error_key.
func NewErrorWithKey(requestId string, err string, errorKey string) stdmessage.StandardMessage[stdmessage.DataNil, stdmessage.MetaNil] {
	// 1. Metadata Handling: Initialize meta map for supplemental machine-readable fields.
	meta := stdmessage.MetaNil{}
	// 2. Key Injection: Add the machine-readable error key if provided by the engine.
	if errorKey != "" {
		// 2.1 Meta assignment for error_key.
		meta["error_key"] = errorKey
	}
	// 3. Final Construction: Return the failure envelope enriched with error metadata.
	return stdmessage.StandardMessage[stdmessage.DataNil, stdmessage.MetaNil]{
		RequestID: requestId,
		Message:   err,
		Meta:      meta,
		Success:   false,
		Data:      stdmessage.DataNil{},
	}
}

// NewSuccess creates a new StandardMessage with the given data.
func NewSuccess[T any](requestId string, msg string, data T) stdmessage.StandardMessage[T, stdmessage.MetaNil] {
	// 1. Payload Wrapping: Inject the generic data into the standard success container.
	// 2. Success Identification: Set truthy status flag for consumer branching logic.
	// 3. Acknowledgement: Include the success message for localized UI feedback display.
	// 4. Traceability: Include the request identifier to match with client-side state.
	return stdmessage.StandardMessage[T, stdmessage.MetaNil]{
		RequestID: requestId,
		Message:   msg,
		Meta:      stdmessage.MetaNil{},
		Success:   true,
		Data:      data,
	}
}

// NewEntity creates a new Entity DTO from the given engine entity.
func NewEntity(ent entity.Entity) Entity {
	// 1. Stats Extraction: Retrieve current and maximum HP and Movement capacity values.
	hp, maxHP, move, maxMove := extractEntityStats(ent)
	// 2. Team Affiliation: Identify which squad the entity belongs to via PropertyI.
	team := 0
	if prop := ent.GetPropertyI(property.TeamID); prop != nil {
		// 2.1 Property value extraction.
		team = prop.I()
	}
	// 3. Modifier Extraction: Map all active buffs and equipped item metadata from engine.
	buffs, items := extractEntityBuffsAndItems(ent)
	// 4. Skillset Projection: Convert every registered engine skill into its API DTO form.
	skills := extractEntitySkills(ent.Skills)
	// 5. Final Assembly: Bundle all resolved fields into the final serializable Entity.
	return Entity{
		ID: ent.ID.String(),
		PlayerID: ent.ControllerID.String(),
		Team: team,
		Name: ent.Name,
		HP: hp,
		MaxHP: maxHP,
		Attack: ent.GetPropertyI(property.Attack).I(),
		Defense: ent.GetPropertyI(property.Defense).I(),
		Move: move,
		MaxMove: maxMove,
		Position: Position{X: ent.Position.X, Y: ent.Position.Y},
		Buffs: buffs,
		EquippedItems: items,
		EquippedSkills: skills,
		IsSelf: false,
		Dead: hp <= 0,
	}
}

// extractEntityStats is a helper to retrieve HP and Movement counters with full documentation.
func extractEntityStats(ent entity.Entity) (hp, maxHP, move, maxMove int) {
	// 1. HP Resolution: Check for existence of the health property in engine container.
	if prop := ent.GetProperty(property.HP); prop != nil {
		// 1.1 Counter Logic: Extract current value and the ceiling for health pool.
		if cp, ok := prop.(property.IntCounterProperty); ok {
			hp = cp.GetValue()
			maxHP = cp.GetMaxValue()
		} else {
			hp = prop.Get().(int)
			maxHP = hp
		}
	}
	// 2. Movement Resolution: Check for existence of the traversal stamina property.
	if prop := ent.GetProperty(property.Movement); prop != nil {
		// 2.1 Counter Logic: Extract current value and the ceiling for movement capacity.
		if cp, ok := prop.(property.IntCounterProperty); ok {
			move = cp.GetValue()
			maxMove = cp.GetMaxValue()
		} else {
			move = prop.Get().(int)
			maxMove = move
		}
	}
	return
}

// extractEntityBuffsAndItems processes the entity's buff list into API-facing structures.
func extractEntityBuffsAndItems(ent entity.Entity) ([]Buff, []EquippedItem) {
	// 1. Containers: Initialize slices for standard property buffs and derived items.
	buffs := make([]Buff, 0, len(ent.Buffs))
	items := make([]EquippedItem, 0)
	// 2. Loop: Iterate through every active temporary property modification on entity.
	for _, b := range ent.Buffs {
		// 2.1 Marker Identification: Look for complex Effect or Zone property keys.
		_, hasEffect := b.Properties[property.PropertyToString(property.Effect)]
		_, hasZone := b.Properties[property.PropertyToString(property.Zone)]
		// 2.2 Routing: If the buff has an origin ID and complex props, treat as Item.
		if b.OriginEntityID != uuid.Nil && (hasEffect || hasZone) {
			items = append(items, convertBuffToItem(b))
			continue
		}
		// 2.3 Default: Map remaining blocks as standard property-modifying buffs.
		buffs = append(buffs, Buff{
			OriginID: b.OriginEntityID.String(),
			Forever: b.Forever,
			Properties: Flex[PropertyMap]{Data: convertPropertyMap(b.Properties)},
		})
	}
	return buffs, items
}

// convertBuffToItem transforms a temporary property block into an EquippedItem DTO.
func convertBuffToItem(b property.TemporaryProperties) EquippedItem {
	// 1. Zone Logic: Resolve the pattern identifier string for area-of-effect gear.
	var zone *string
	if zp, ok := b.Properties[property.PropertyToString(property.Zone)].(*def.ZoneProperty); ok {
		// 1.1 Pattern string assignment.
		zone = &zp.PatternType
	}
	// 2. Effect Logic: Resolve the payload property map for complex item-granted buffs.
	var effProps PropertyMap
	if ep, ok := b.Properties[property.PropertyToString(property.Effect)].(*def.EffectProperty); ok && ep.Effect != nil {
		// 2.1 Property slice conversion.
		effProps = convertPropertySlice(ep.Effect.Properties)
	}
	// 3. Construction: Return the Equipment DTO with its properties and effects.
	return EquippedItem{
		ItemID: b.OriginEntityID.String(),
		Name: "Equipped Item",
		Properties: Flex[PropertyMap]{Data: convertPropertyMap(b.Properties)},
		Effect: Flex[PropertyMap]{Data: effProps},
		Zone: zone,
	}
}

// extractEntitySkills transforms a map of engine skills into a serializable slice.
func extractEntitySkills(skills map[uuid.UUID]skill.Skill) []EquippedSkill {
	// 1. Setup: Allocate result slice with capacity matching the engine skill map.
	res := make([]EquippedSkill, 0, len(skills))
	// 2. Projection: Process every registered skill in the entity's repertoire.
	for _, s := range skills {
		var zone *string
		// 2.1 Pattern Detection: Capture the targeting zone type if defined.
		if zp, ok := s.Targeting[property.PropertyToString(property.Zone)].(*def.ZoneProperty); ok {
			zone = &zp.PatternType
		}
		// 2.2 DTO Assembly: Map behavior, costs, and effects into the skill DTO.
		res = append(res, EquippedSkill{
			SkillID: s.ID.String(),
			Name: s.Name,
			Behavior: behaviorName(def.BehaviorType(s.Behavior.Get().(string))),
			Targeting: Flex[PropertyMap]{Data: convertPropertyMap(s.Targeting)},
			Costs: Flex[PropertyMap]{Data: convertPropertyMap(s.Costs)},
			Effect: Flex[PropertyMap]{Data: convertPropertySlice(s.Effect.Properties)},
			Zone: zone,
		})
	}
	return res
}

// convertPropertyMap transforms a map of engine properties into a serializable PropertyMap.
func convertPropertyMap(props map[string]property.Property) PropertyMap {
	// 1. Output Init: Prepare the map for DTO conversion results.
	out := make(PropertyMap, len(props))
	// 2. Iteration: Convert each property while filtering internal technical markers.
	for k, v := range props {
		// 2.1 Filter: Skip Effect and Zone properties which are handled separately.
		if k == property.PropertyToString(property.Effect) || k == property.PropertyToString(property.Zone) {
			continue
		}
		// 2.2 Conversion: Map engine property to DTO.
		out[k] = convertProperty(v)
	}
	return out
}

// convertPropertySlice transforms a slice of engine properties into a serializable PropertyMap.
func convertPropertySlice(props []property.Property) PropertyMap {
	// 1. Setup: Prepare a map indexed by GameMaster-facing property names.
	out := make(PropertyMap, len(props))
	// 2. Traversal: Convert every property in the slice to its DTO counterpart.
	for _, v := range props {
		// 2.1 Indexing: Use GM name for consistent property identification.
		out[v.Name(property.GameMaster)] = convertProperty(v)
	}
	return out
}

// convertProperty transforms a single engine property into a PropertyDTO.
func convertProperty(v property.Property) PropertyDTO {
	// 1. Value Access: Retrieve the underlying primitive value from the engine.
	dto := PropertyDTO{}
	val := v.Get()
	// 2. Type Mapping: Populate the correct DTO pointer based on the runtime type.
	switch t := val.(type) {
	case int:
		dto.Value = &t
	case float64:
		dto.FValue = &t
	case bool:
		dto.BValue = &t
	case string:
		dto.SValue = &t
	default:
		// 2.1 Fallback: Cast unknown types to string representation for safety.
		s := fmt.Sprintf("%v", val)
		dto.SValue = &s
	}
	// 3. Counter Enrichment: If property implements counter, inject the ceiling value.
	if cp, ok := v.(property.IntCounterProperty); ok {
		// 3.1 Max value extraction.
		mv := cp.GetMaxValue()
		dto.Max = &mv
	}
	return dto
}

// NewBoardState creates a new BoardState DTO from internal engine state.
func NewBoardState(matchID uuid.UUID, g *grid.Grid, entities []entity.Entity, players []Player, ts turner.TurnState, startTime time.Time, timeout time.Time, winnerTeamID int, version int64, action *ActionFeedback) BoardState {
	// 1. Snapshot: Capture basic match metadata, initiative, and versioning.
	bs := BoardState{
		StartTime: startTime, Timeout: timeout, CurrentEntityID: ts.CurrentEntityTurn.String(),
		Players: players, Action: action, Version: version,
		// Stamp the schema version into every blob so resurrection can detect incompatible shapes.
		SerializerVersion: upsilonserializer.CurrentSerializerVersion,
	}
	// 2. Victory Condition: Set the winning team if the battle has concluded.
	if winnerTeamID > 0 {
		// 2.1 Winner assignment.
		bs.WinnerTeamID = &winnerTeamID
	}
	// 3. Grid Projection: Map the 3D grid layout to a 2D surface snapshot.
	bs.Grid = mapGrid(g, entities)
	// 4. Roster Sync: Identify players and update their entity snapshots.
	eToP, eMap := mapEntities(entities, ts.CurrentEntityTurn, &bs)
	// 5. Update: Push latest entity stats into the player rosters.
	updatePlayersEntities(&bs, eMap)
	// 6. Initiative: Hydrate the turn queue for client-side visibility.
	for _, t := range ts.RemainingTurns {
		// 6.1 Turn entry creation.
		bs.Turn = append(bs.Turn, Turn{
			EntityID: t.EntityId.String(), PlayerID: eToP[t.EntityId], Delay: t.Delay,
		})
	}
	return bs
}

// mapGrid converts the engine's 3D grid into the API's 2D surface projection.
func mapGrid(g *grid.Grid, entities []entity.Entity) Grid {
	// 1. Metadata Setup: Set width, height, and depth ceiling for projection.
	res := Grid{
		Width: g.Width, Height: g.Length, MaxHeight: g.Height,
		Cells: make([][]Cell, g.Width),
	}
	// 2. Optimization Cache: Index characters to speed up surface resolution.
	charMap := make(map[uuid.UUID]bool)
	for _, e := range entities {
		// 2.1 Character filter logic.
		if e.Type == entity.Character || e.Type == entity.Monster {
			// 2.2 Presence marking.
			charMap[e.ID] = true
		}
	}
	// 3. Projection Loop: Walk every (x,y) column to find its topmost cell.
	for x := 0; x < g.Width; x++ {
		// 3.1 Column allocation.
		res.Cells[x] = make([]Cell, g.Length)
		for y := 0; y < g.Length; y++ {
			// 3.2 Cell resolution.
			res.Cells[x][y] = resolveSurfaceCell(g, x, y, charMap)
		}
	}
	return res
}

// resolveSurfaceCell identifies the walkable properties and character ID at a specific (x,y) column.
func resolveSurfaceCell(g *grid.Grid, x int, y int, charMap map[uuid.UUID]bool) Cell {
	// 1. Z-Scan: Identify the highest coordinate with content in the column stack.
	z := g.TopMostCellAt(x, y)
	cl, ok := g.CellAt(position.New(x, y, z))
	// 2. Existence Check: If cell is missing from map, return empty default.
	if !ok { return Cell{} }
	// 3. Occupancy Search: Find the first character ID present at the surface level.
	var charID string
	for _, eid := range cl.EntityIDs {
		// 3.1 Character identification.
		if charMap[eid] {
			// 3.2 ID string assignment.
			charID = eid.String()
			break
		}
	}
	// 4. Projection: Return Cell with height, obstacle status, and entity ID.
	return Cell{EntityID: charID, Obstacle: cl.Type == cell.Obstacle, Height: z}
}

// mapEntities creates a lookup map for current arena entities.
func mapEntities(entities []entity.Entity, currentTurn uuid.UUID, bs *BoardState) (map[uuid.UUID]string, map[uuid.UUID]Entity) {
	// 1. Setup: Prepare indexing maps for entity-to-player and entity-to-DTO.
	eToP := make(map[uuid.UUID]string)
	eMap := make(map[uuid.UUID]Entity)
	// 2. Transformation: Process every engine entity and mark turn ownership.
	for _, e := range entities {
		// 2.1 Mapping: Associate entity with its controller.
		eToP[e.ID] = e.ControllerID.String()
		// 2.2 Conversion: Map engine entity to API DTO.
		apiEntity := NewEntity(e)
		eMap[e.ID] = apiEntity
		// 2.3 Turn Tracking: Identify acting player for current initiative slot.
		if e.ID == currentTurn { bs.CurrentPlayerID = e.ControllerID.String() }
	}
	return eToP, eMap
}

// updatePlayersEntities synchronizes the player roster with the latest entity data.
func updatePlayersEntities(bs *BoardState, entityMap map[uuid.UUID]Entity) {
	// 1. Iteration: Walk every player in the board state roster.
	for i := range bs.Players {
		// 1.1 Delegate: Update each player's individual entity set.
		updateSinglePlayerEntities(&bs.Players[i], entityMap)
	}
}

// updateSinglePlayerEntities refreshes all entities for a specific player from the live map.
func updateSinglePlayerEntities(p *Player, entityMap map[uuid.UUID]Entity) {
	// 1. Vitality Sync: Overwrite existing entity data with latest engine snapshots.
	for j := range p.Entities {
		entID, _ := uuid.Parse(p.Entities[j].ID)
		// 1.1 Check: Verify if entity still exists in the live engine state.
		if actual, found := entityMap[entID]; found {
			// 1.2 Overwrite: Push latest stats to player entity DTO.
			p.Entities[j] = actual
		} else {
			// 1.3 Removal: Mark as dead and clear volatile states if missing from engine.
			p.Entities[j].HP = 0
			p.Entities[j].Dead = true
			p.Entities[j].EquippedSkills = []EquippedSkill{}
			p.Entities[j].Buffs = []Buff{}
			p.Entities[j].EquippedItems = []EquippedItem{}
		}
	}
}

// behaviorName returns the string representation of a BehaviorType.
func behaviorName(bt def.BehaviorType) string {
	// 1. Translation: Direct string cast of the behavior enum identifier.
	return string(bt)
}
