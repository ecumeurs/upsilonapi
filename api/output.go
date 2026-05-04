package api

import (
	"time"
	"fmt"
	"github.com/ecumeurs/upsilonapi/stdmessage"
	"github.com/ecumeurs/upsilontypes/entity"
	"github.com/ecumeurs/upsilonbattle/battlearena/ruler/turner"
	"github.com/ecumeurs/upsilonmapdata/grid"
	"github.com/ecumeurs/upsilonmapdata/grid/cell"
	"github.com/ecumeurs/upsilonmapdata/grid/position"
	"github.com/ecumeurs/upsilontypes/property"
	"github.com/ecumeurs/upsilontypes/property/def"
	"github.com/google/uuid"
)

// @spec-link [[api_go_battle_engine]]

type ArenaActionResponse struct {
	Status string `json:"status"`
}

type ArenaStartResponse struct {
	ArenaID      string     `json:"arena_id"`
	InitialState BoardState `json:"initial_state"`
}

type ActiveMatchStatsResponse struct {
	ActiveCount int `json:"active_count"`
}

type ArenaExistsResponse struct {
	Exists bool `json:"exists"`
}

// SkillGenerateResponse is the payload returned by POST /v1/skills/generate.
// @spec-link [[api_skill_generate_engine]]
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

// @spec-link [[entity_grid]]

// Cell is the topmost cell at a given (x, y) column of the engine grid.
// Cave/underground navigation is not exposed by this iteration; clients
// must treat the cell as the walkable surface at that column.
type Cell struct {
	EntityID string `json:"entity_id"`         // if any
	Obstacle bool   `json:"obstacle"`          // if any
	Height   int    `json:"height"`            // Z index of the topmost cell at (x, y); surface elevation
}

// Grid: A 2D projection of the engine's 3D grid. Each cell is the topmost
// cell at that column (see Cell). MaxHeight exposes the Z ceiling so clients
// can scale elevation rendering without guessing.
type Grid struct {
	Width     int      `json:"width"`
	Height    int      `json:"height"`
	MaxHeight int      `json:"max_height"` // ceiling Z of the engine grid (exclusive upper bound)
	Cells     [][]Cell `json:"cells"`      // Cells are stored in width-major order.
}

type Turn struct {
	PlayerID string `json:"player_id"`
	Delay    int    `json:"delay"`
	EntityID string `json:"entity_id"`
}

type CreditAward struct {
	PlayerID string `json:"player_id"`
	Amount   int    `json:"amount"`
	Source   string `json:"source"` // damage, healing, status
}

// ActionResult provides explicit data about the impact on a single target.
type ActionResult struct {
	TargetID string         `json:"target_id"`
	Damage   int            `json:"damage,omitempty"`
	Heal     int            `json:"heal,omitempty"`
	PrevHP   int            `json:"prev_hp"`
	NewHP    int            `json:"new_hp"`
	Credits  []CreditAward  `json:"credits,omitempty"`
}

// ActionFeedback provides explicit data about the last tactical action.
// @spec-link [[api_go_action_feedback]]
type ActionFeedback struct {
	Type     string              `json:"type"` // "move", "attack", "skill", "pass"
	ActorID  string              `json:"actor_id"`
	TargetID string              `json:"target_id,omitempty"` // Legacy/Primary target
	Path     []position.Position `json:"path,omitempty"`
	Results  []ActionResult      `json:"results,omitempty"`
	Credits  []CreditAward       `json:"credits,omitempty"` // Global action credits
}

// BoardState represents the current state of the board.
// @spec-link [[battleui_api_dtos]]
type BoardState struct {
	Players         []Player  `json:"players"`         // Consolidated roster
	Grid            Grid      `json:"grid"`
	Turn            []Turn    `json:"turn"`
	CurrentPlayerID string    `json:"current_player_id"`
	CurrentEntityID string    `json:"current_entity_id"`
	Timeout         time.Time `json:"timeout"` 
	StartTime       time.Time `json:"start_time"`
	WinnerTeamID    *int            `json:"winner_team_id"`
	Action          *ActionFeedback `json:"action,omitempty"`
	Version         int64           `json:"version"`
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

type ArenaActionResponseMessage = stdmessage.StandardMessage[ArenaActionResponse, stdmessage.MetaNil]
type ArenaStartResponseMessage = stdmessage.StandardMessage[ArenaStartResponse, stdmessage.MetaNil]
type ArenaEventMessage = stdmessage.StandardMessage[ArenaEvent, stdmessage.MetaNil]

// NewError creates a new StandardMessage with the given error.
func NewError(requestId string, err string) stdmessage.StandardMessage[stdmessage.DataNil, stdmessage.MetaNil] {
	return stdmessage.StandardMessage[stdmessage.DataNil, stdmessage.MetaNil]{
		RequestID: requestId,
		Message:   err,
		Meta:      stdmessage.MetaNil{},
		Success:   false,
		Data:      stdmessage.DataNil{},
	}
}

// NewErrorWithKey returns a standard-envelope error message that also carries
// an `error_key` inside `meta`. This is how the engine's ruler error keys
// (entity.path.obstacle, entity.turn.missmatch, ...) are surfaced to external
// clients without extending the envelope schema — `meta` is the sanctioned
// debug/test slot per [[api_standard_envelope]].
func NewErrorWithKey(requestId string, err string, errorKey string) stdmessage.StandardMessage[stdmessage.DataNil, stdmessage.MetaNil] {
	meta := stdmessage.MetaNil{}
	if errorKey != "" {
		meta["error_key"] = errorKey
	}
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
	return stdmessage.StandardMessage[T, stdmessage.MetaNil]{
		RequestID: requestId,
		Message:   msg,
		Meta:      stdmessage.MetaNil{},
		Success:   true,
		Data:      data,
	}
}

// NewEntity creates a new Entity from the given entity (upsilonbattle's)
func NewEntity(entity entity.Entity) Entity {
	team := 0
	if prop := entity.GetPropertyI(property.TeamID); prop != nil {
		team = prop.I()
	}

	hp := 0
	maxHP := 0
	if prop := entity.GetProperty(property.HP); prop != nil {
		if cp, ok := prop.(property.IntCounterProperty); ok {
			hp = cp.GetValue()
			maxHP = cp.GetMaxValue()
		} else {
			hp = prop.Get().(int)
			maxHP = hp
		}
	}

	move := 0
	maxMove := 0
	if prop := entity.GetProperty(property.Movement); prop != nil {
		if cp, ok := prop.(property.IntCounterProperty); ok {
			move = cp.GetValue()
			maxMove = cp.GetMaxValue()
		} else {
			move = prop.Get().(int)
			maxMove = move
		}
	}

	equippedItems := make([]EquippedItem, 0)
	buffs := make([]Buff, 0, len(entity.Buffs))
	for _, b := range entity.Buffs {
		if b.OriginEntityID != uuid.Nil {
			// Actually, let's check for Effect or Zone property to identify items/complex buffs
			_, hasEffect := b.Properties[property.PropertyToString(property.Effect)]
			_, hasZone := b.Properties[property.PropertyToString(property.Zone)]

			if hasEffect || hasZone {
				var zone *string
				if zp, ok := b.Properties[property.PropertyToString(property.Zone)].(*def.ZoneProperty); ok {
					zone = &zp.PatternType
				}
				var effProps PropertyMap
				if ep, ok := b.Properties[property.PropertyToString(property.Effect)].(*def.EffectProperty); ok && ep.Effect != nil {
					effProps = convertPropertySlice(ep.Effect.Properties)
				}

				equippedItems = append(equippedItems, EquippedItem{
					ItemID:     b.OriginEntityID.String(),
					Name:       "Equipped Item", // Placeholder as engine doesn't store name
					Properties: Flex[PropertyMap]{Data: convertPropertyMap(b.Properties)},
					Effect:     Flex[PropertyMap]{Data: effProps},
					Zone:       zone,
				})
				continue // Don't show as a separate buff if it's shown as an item
			}
		}

		buffs = append(buffs, Buff{
			OriginID:   b.OriginEntityID.String(),
			Forever:    b.Forever,
			Properties: Flex[PropertyMap]{Data: convertPropertyMap(b.Properties)},
		})
	}

	skills := make([]EquippedSkill, 0, len(entity.Skills))
	for _, s := range entity.Skills {
		var zone *string
		if zp, ok := s.Targeting[property.PropertyToString(property.Zone)].(*def.ZoneProperty); ok {
			zone = &zp.PatternType
		}

		skills = append(skills, EquippedSkill{
			SkillID:   s.ID.String(),
			Name:      s.Name,
			Behavior:  behaviorName(def.BehaviorType(s.Behavior.Get().(string))),
			Targeting: Flex[PropertyMap]{Data: convertPropertyMap(s.Targeting)},
			Costs:     Flex[PropertyMap]{Data: convertPropertyMap(s.Costs)},
			Effect:    Flex[PropertyMap]{Data: convertPropertySlice(s.Effect.Properties)},
			Zone:      zone,
		})
	}

	return Entity{
		ID:             entity.ID.String(),
		PlayerID:       entity.ControllerID.String(),
		Team:           team,
		Name:           entity.Name,
		HP:             hp,
		MaxHP:          maxHP,
		Attack:         entity.GetPropertyI(property.Attack).I(),
		Defense:        entity.GetPropertyI(property.Defense).I(),
		Move:           move,
		MaxMove:        maxMove,
		Position:       Position{X: entity.Position.X, Y: entity.Position.Y},
		Buffs:          buffs,
		EquippedItems:  equippedItems,
		EquippedSkills: skills,
		IsSelf:         false, // Handled by Laravel gateway
		Dead:           hp <= 0,
	}
}

func convertPropertyMap(props map[string]property.Property) PropertyMap {
	out := make(PropertyMap, len(props))
	for k, v := range props {
		// Skip special fields that are handled at the top level of the DTO
		if k == property.PropertyToString(property.Effect) || k == property.PropertyToString(property.Zone) {
			continue
		}
		out[k] = convertProperty(v)
	}
	return out
}

func convertPropertySlice(props []property.Property) PropertyMap {
	out := make(PropertyMap, len(props))
	for _, v := range props {
		out[v.Name(property.GameMaster)] = convertProperty(v)
	}
	return out
}

func convertProperty(v property.Property) PropertyDTO {
	dto := PropertyDTO{}
	val := v.Get()
	if i, ok := val.(int); ok {
		dto.Value = &i
	} else if f, ok := val.(float64); ok {
		dto.FValue = &f
	} else if bv, ok := val.(bool); ok {
		dto.BValue = &bv
	} else {
		// Handle named string types (enums)
		if sv, ok := val.(string); ok {
			dto.SValue = &sv
		} else {
			// Try to convert to string if it's a named string type
			s := fmt.Sprintf("%v", val)
			// But only if it's not a complex struct that happened to have a String() method we don't want
			// For now, let's just check for the specific types we know or use reflection to check underlying type.
			// Actually, TargetTypes and TargetingMechanicsType are what we care about.
			dto.SValue = &s
		}
	}

	if cp, ok := v.(property.IntCounterProperty); ok {
		mv := cp.GetMaxValue()
		dto.Max = &mv
	}
	return dto
}

// NewBoardState creates a new BoardState DTO from internal state.
func NewBoardState(matchID uuid.UUID, g *grid.Grid, entities []entity.Entity, players []Player, ts turner.TurnState, startTime time.Time, timeout time.Time, winnerTeamID int, version int64, action *ActionFeedback) BoardState {
	bs := BoardState{
		StartTime:       startTime,
		Timeout:         timeout,
		CurrentEntityID: ts.CurrentEntityTurn.String(),
		Players:         players,
		Action:          action,
		Version:         version,
	}

	if winnerTeamID > 0 {
		bs.WinnerTeamID = &winnerTeamID
	}

	// Map Grid. The engine is a true 3D grid; we expose the topmost cell per
	// (x, y) column so clients (CLI 2D ASCII / battleui 3D) share one source
	// of truth for the walkable surface. Z information is carried in
	// Cell.Height, and the Z ceiling via Grid.MaxHeight.
	bs.Grid = Grid{
		Width:     g.Width,
		Height:    g.Length,
		MaxHeight: g.Height,
		Cells:     make([][]Cell, g.Width),
	}

	// Create character lookup map for cell entity resolution
	// @spec-link [[mechanic_multi_entity_cell_system]]
	charMap := make(map[uuid.UUID]bool)
	for _, e := range entities {
		if e.Type == entity.Character || e.Type == entity.Monster {
			charMap[e.ID] = true
		}
	}

	for x := 0; x < g.Width; x++ {
		bs.Grid.Cells[x] = make([]Cell, g.Length)
		for y := 0; y < g.Length; y++ {
			z := g.TopMostCellAt(x, y)
			cl, ok := g.CellAt(position.New(x, y, z))
			if ok {
				var charID string
				// Only output the first character entity found in the cell.
				// Effects and other non-character entities are filtered out for the API's grid view.
				for _, eid := range cl.EntityIDs {
					if charMap[eid] {
						charID = eid.String()
						break
					}
				}

				bs.Grid.Cells[x][y] = Cell{
					EntityID: charID,
					Obstacle: cl.Type == cell.Obstacle,
					Height:   z,
				}
			}
		}
	}

	entityToPlayer := make(map[uuid.UUID]string)
	entityMap := make(map[uuid.UUID]Entity)
	for _, e := range entities {
		entityToPlayer[e.ID] = e.ControllerID.String()
		apiEntity := NewEntity(e)
		entityMap[e.ID] = apiEntity

		if e.ID == ts.CurrentEntityTurn {
			bs.CurrentPlayerID = e.ControllerID.String()
		}
	}

	// Update Players' entity lists with actual engine data (fixes coordinate desync)
	for i := range bs.Players {
		for j := range bs.Players[i].Entities {
			entID, err := uuid.Parse(bs.Players[i].Entities[j].ID)
			if err == nil {
				if actual, found := entityMap[entID]; found {
					bs.Players[i].Entities[j] = actual
				} else {
					// Entity is dead/removed from engine, ensure HP is 0
					bs.Players[i].Entities[j].HP = 0
					bs.Players[i].Entities[j].Dead = true
					bs.Players[i].Entities[j].EquippedSkills = make([]EquippedSkill, 0)
					bs.Players[i].Entities[j].Buffs = make([]Buff, 0)
					bs.Players[i].Entities[j].EquippedItems = make([]EquippedItem, 0)
				}
			}
		}
	}

	for _, t := range ts.RemainingTurns {
		bs.Turn = append(bs.Turn, Turn{
			EntityID: t.EntityId.String(),
			PlayerID: entityToPlayer[t.EntityId],
			Delay:    t.Delay,
		})
	}

	return bs
}

func behaviorName(bt def.BehaviorType) string {
	return string(bt)
}
