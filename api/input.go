// Package api provides the input Data Transfer Objects (DTOs) for the Upsilon Hub API.
// It defines the request schemas for starting arenas, executing actions, and performing state recovery.
package api

import (
	"encoding/json"
	"fmt"
	"github.com/ecumeurs/upsilonapi/stdmessage"
)

// Flex handles inconsistent JSON from external systems (e.g. Laravel)
// where an empty object might be represented as an empty array [].
// This is critical for maintaining compatibility with PHP-to-Go JSON serialization quirks.
type Flex[T any] struct {
	Data T
}

// MarshalJSON satisfies the json.Marshaler interface.
func (f Flex[T]) MarshalJSON() ([]byte, error) {
	return json.Marshal(f.Data)
}

// UnmarshalJSON satisfies the json.Unmarshaler interface.
// It explicitly handles the "[]" case to avoid unmarshaling errors for empty objects.
func (f *Flex[T]) UnmarshalJSON(data []byte) error {
	if string(data) == "[]" {
		// Return zero value for T to handle PHP's empty-map-to-array conversion.
		return nil
	}
	return json.Unmarshal(data, &f.Data)
}

// PropertyDTO represents a single property value in a strictly typed manner.
// It supports integers (with optional max for counters), floats, booleans, and strings.
// @spec-link [[mechanic_mec_skill_payload_resolution]]
type PropertyDTO struct {
	Value  *int     `json:"value,omitempty"`
	FValue *float64 `json:"fvalue,omitempty"`
	Max    *int     `json:"max,omitempty"`
	BValue *bool    `json:"bvalue,omitempty"`
	SValue *string  `json:"svalue,omitempty"`
}

// MarshalJSON satisfies the json.Marshaler interface.
func (p PropertyDTO) MarshalJSON() ([]byte, error) {
	type alias PropertyDTO
	return json.Marshal(alias(p))
}

// UnmarshalJSON satisfies the json.Unmarshaler interface.
// It implements polymorphic unmarshaling to handle both structured DTOs and primitives.
func (p *PropertyDTO) UnmarshalJSON(data []byte) error {
	// 1. Structural Attempt: Try unmarshaling as a structured object first.
	type alias PropertyDTO
	var a alias
	if err := json.Unmarshal(data, &a); err == nil && (a.Value != nil || a.Max != nil || a.BValue != nil || a.SValue != nil || a.FValue != nil) {
		*p = PropertyDTO(a)
		return nil
	}

	// 2. Primitive Fallback: Attempt to parse the raw JSON data into known primitive types.
	var i int
	if err := json.Unmarshal(data, &i); err == nil {
		p.Value = &i
		return nil
	}
	var f float64
	if err := json.Unmarshal(data, &f); err == nil {
		p.FValue = &f
		return nil
	}
	var b bool
	if err := json.Unmarshal(data, &b); err == nil {
		p.BValue = &b
		return nil
	}
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		p.SValue = &s
		return nil
	}

	return fmt.Errorf("invalid property format: %s", string(data))
}

// PropertyMap is a utility type for collections of named engine properties.
type PropertyMap = map[string]PropertyDTO

// Position represents a 2D coordinate on the engine grid.
type Position struct {
	X int `json:"x"`
	Y int `json:"y"`
}

// Casting is the client-facing projection of a channeling entity's in-flight cast.
// It carries enough for the UI to render a "Channeling: <skill>" indicator and an
// interruption gauge, plus the target (an entity id for entity-target channels, a
// tile for tile-target channels). @spec-link [[upsilonbattle:mechanic_channeling_mechanic]]
type Casting struct {
	SkillID      string    `json:"skill_id"`
	SkillName    string    `json:"skill_name"`
	TargetEntity string    `json:"target_entity,omitempty"` // entity-target channels
	TargetTile   *Position `json:"target_tile,omitempty"`   // tile-target channels
	Interruption int       `json:"interruption"`            // 0-100
}

// ArenaActionRequest defines the payload for executing a tactical engine action.
type ArenaActionRequest struct {
	PlayerID     string     `json:"player_id"`
	Type         string     `json:"type"` // move, attack, skill, pass, forfeit
	TargetCoords []Position `json:"target_coords"`
	EntityID     string     `json:"entity_id"`
	SkillID      string     `json:"skill_id,omitempty"`
}

// Entity represents a combatant within the engine.
// @spec-link [[entity_character]]
type Entity struct {
	ID             string          `json:"id"`
	PlayerID       string          `json:"player_id"`
	Team           int             `json:"team"`
	Name           string          `json:"name"`
	HP             int             `json:"hp"`
	MaxHP          int             `json:"max_hp"`
	Attack         int             `json:"attack"`
	Defense        int             `json:"defense"`
	Move           int             `json:"move"`
	MaxMove        int             `json:"max_move"`
	Position       Position        `json:"position"`
	EquippedItems  []EquippedItem  `json:"equipped_items"`
	Buffs          []Buff          `json:"buffs"`
	EquippedSkills []EquippedSkill `json:"equipped_skills"`
	IsSelf         bool            `json:"is_self"`
	Dead           bool            `json:"dead"`
	// IsCasting is set while the entity is channeling a skill; nil/absent otherwise.
	// @spec-link [[upsilonbattle:mechanic_channeling_mechanic]]
	IsCasting *Casting `json:"is_casting,omitempty"`
	Archetype string   `json:"archetype,omitempty"` // per-entity override; inherits Player.Archetype when empty
	Grade     string   `json:"grade,omitempty"`     // per-entity override; inherits Player grade when empty
	AutoGen   bool     `json:"auto_gen,omitempty"`  // true → Go generates stats and skills from archetype+grade
}

// EquippedSkill carries the tactical definition of an entity's ability.
// @spec-link [[api_character_skill_inventory]]
type EquippedSkill struct {
	SkillID   string            `json:"skill_id"`
	Name      string            `json:"name"`
	Behavior  string            `json:"behavior"` // Direct, Zone, Reaction
	Targeting Flex[PropertyMap] `json:"targeting"`
	Costs     Flex[PropertyMap] `json:"costs"`
	Effect    Flex[PropertyMap] `json:"effect"`
	Zone      *string           `json:"zone,omitempty"`
	Origin    string            `json:"origin,omitempty"` // "inventory" | "item:<item_id>"
}

// Buff represents a temporary or permanent property modification on an entity.
type Buff struct {
	OriginID   string            `json:"origin_id"`
	Forever    bool              `json:"forever"`
	Properties Flex[PropertyMap] `json:"properties"`
}

// EquippedItem represents an item that grants stats or skills to an entity.
type EquippedItem struct {
	ItemID     string            `json:"item_id"`
	Name       string            `json:"name"`
	Slot       string            `json:"slot"`
	Properties Flex[PropertyMap] `json:"properties"`
	Effect     Flex[PropertyMap] `json:"effect,omitempty"`
	Zone       *string           `json:"zone,omitempty"`
}

// Player represents a human or AI participant in a match.
// @spec-link [[entity_player]]
type Player struct {
	ID        string   `json:"id"`
	Nickname  string   `json:"nickname"`
	Entities  []Entity `json:"entities"`
	Team      int      `json:"team"`
	IA        bool     `json:"ia"`
	Archetype string   `json:"archetype,omitempty"`  // "fighter"|"ranger"|"support"|"sneak"|"" (random for IA)
	Grade     string   `json:"grade,omitempty"`      // "I".."V"; derived from TotalWins when empty
	TotalWins int      `json:"total_wins,omitempty"` // used to derive Grade when Grade is empty
}

// ArenaStartRequest is the payload for initializing a new battle.
type ArenaStartRequest struct {
	MatchID     string   `json:"match_id"`
	CallbackURL string   `json:"callback_url"`
	Players     []Player `json:"players"`
}

// ArenaForfeitRequest is the payload for a player voluntarily leaving the match.
type ArenaForfeitRequest struct {
	PlayerID string `json:"player_id"`
}

// ArenaResurrectRequest carries persisted board state from Laravel to rebuild a crashed arena.
type ArenaResurrectRequest struct {
	MatchID         string          `json:"match_id"`
	CallbackURL     string          `json:"callback_url"`
	Players         []Player        `json:"players"`
	Grid            ResurrectGrid   `json:"grid"`
	Turns           []ResurrectTurn `json:"turns"`
	CurrentEntityID string          `json:"current_entity_id"`
	Version         int64           `json:"version"`
	// SerializerVersion must match upsilonserializer.CurrentSerializerVersion.
	// Blobs written before versioning was introduced will carry zero (absent field),
	// which is explicitly rejected by the resurrection guard.
	SerializerVersion int `json:"serializer_version"`
}

// ResurrectGrid is the 2D projection of the engine grid sufficient to rebuild pathfinding.
type ResurrectGrid struct {
	Width     int               `json:"width"`
	Height    int               `json:"height"`     // Y dimension (Length)
	MaxHeight int               `json:"max_height"` // Z ceiling
	Cells     [][]ResurrectCell `json:"cells"`
}

// ResurrectCell carries per-column surface info needed to reconstruct the 3D grid.
type ResurrectCell struct {
	Obstacle bool `json:"obstacle"`
	Height   int  `json:"height"` // topmost Z of the surface at this (x,y)
}

// ResurrectTurn represents one entry in the saved turner queue.
type ResurrectTurn struct {
	EntityID string `json:"entity_id"`
	Delay    int    `json:"delay"`
}

type ArenaActionMessage = stdmessage.StandardMessage[ArenaActionRequest, stdmessage.MetaNil]
type ArenaStartMessage = stdmessage.StandardMessage[ArenaStartRequest, stdmessage.MetaNil]
type ArenaForfeitMessage = stdmessage.StandardMessage[ArenaForfeitRequest, stdmessage.MetaNil]
type ArenaResurrectMessage = stdmessage.StandardMessage[ArenaResurrectRequest, stdmessage.MetaNil]
