package api

import (
	"encoding/json"
	"fmt"
	"github.com/ecumeurs/upsilonapi/stdmessage"
)

// @spec-link [[rule_dto_strict_typing]]
// Flex handles inconsistent JSON from external systems (e.g. Laravel)
// where an empty object might be represented as an empty array [].
type Flex[T any] struct {
	Data T
}

func (f Flex[T]) MarshalJSON() ([]byte, error) {
	return json.Marshal(f.Data)
}

func (f *Flex[T]) UnmarshalJSON(data []byte) error {
	if string(data) == "[]" {
		// Return zero value for T
		return nil
	}
	return json.Unmarshal(data, &f.Data)
}

// PropertyDTO represents a single property value in a strictly typed manner.
// It supports integers (with optional max for counters), booleans, and strings.
// @spec-link [[rule_dto_strict_typing]]
type PropertyDTO struct {
	Value  *int     `json:"value,omitempty"`
	FValue *float64 `json:"fvalue,omitempty"`
	Max    *int     `json:"max,omitempty"`
	BValue *bool    `json:"bvalue,omitempty"`
	SValue *string  `json:"svalue,omitempty"`
}

func (p PropertyDTO) MarshalJSON() ([]byte, error) {
	// If it's a simple value, we might want to flatten it? 
	// No, the user said "mostly int, and some time counter int. So we may have to have a generic property DTO for this will nullable max."
	// So we should output as a struct if it's a counter (value + max) or just the value?
	// To be safe and contractual, let's always output the struct as defined by the JSON tags.
	type alias PropertyDTO
	return json.Marshal(alias(p))
}

func (p *PropertyDTO) UnmarshalJSON(data []byte) error {
	// Try unmarshaling as a struct first (structured properties)
	type alias PropertyDTO
	var a alias
	if err := json.Unmarshal(data, &a); err == nil && (a.Value != nil || a.Max != nil || a.BValue != nil || a.SValue != nil) {
		*p = PropertyDTO(a)
		return nil
	}

	// Fallback to primitives
	var i int
	if err := json.Unmarshal(data, &i); err == nil {
		p.Value = &i
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

type PropertyMap = map[string]PropertyDTO

// @spec-link [[api_go_battle_engine]]

type Position struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type ArenaActionRequest struct {
	PlayerID     string     `json:"player_id"`
	Type         string     `json:"type"`
	TargetCoords []Position `json:"target_coords"`
	EntityID     string     `json:"entity_id"`
	SkillID      string     `json:"skill_id,omitempty"`
}

// @spec-link [[entity_character]]
type Entity struct {
	ID             string         `json:"id"`
	PlayerID       string         `json:"player_id"`
	Team           int            `json:"team"`
	Name           string         `json:"name"`
	HP             int            `json:"hp"`
	MaxHP          int            `json:"max_hp"`
	Attack         int            `json:"attack"`
	Defense        int            `json:"defense"`
	Move           int            `json:"move"`
	MaxMove        int            `json:"max_move"`
	Position       Position       `json:"position"` // not used at start
	EquippedItems  []EquippedItem `json:"equipped_items"`
	Buffs          []Buff         `json:"buffs"`           // Added for engine state transparency [[mec_item_buff_application]]
	EquippedSkills []EquippedSkill `json:"equipped_skills"`
	IsSelf         bool           `json:"is_self"`
	Dead           bool           `json:"dead"`
}

// @spec-link [[api_character_skill_inventory]]
// @spec-link [[mec_skill_payload_resolution]]
type EquippedSkill struct {
	SkillID   string              `json:"skill_id"`
	Name      string              `json:"name"`
	Behavior  string              `json:"behavior"`
	Targeting Flex[PropertyMap]   `json:"targeting"`
	Costs     Flex[PropertyMap]   `json:"costs"`
	Effect    Flex[PropertyMap]   `json:"effect"`
	Origin    string              `json:"origin,omitempty"` // "inventory" | "item:<item_id>"
}

type Buff struct {
	OriginID   string            `json:"origin_id"`
	Forever    bool              `json:"forever"`
	Properties Flex[PropertyMap] `json:"properties"`
}

type EquippedItem struct {
	ItemID     string            `json:"item_id"`
	Name       string            `json:"name"`
	Slot       string            `json:"slot"`
	Properties Flex[PropertyMap] `json:"properties"`
}

// @spec-link [[entity_player]]
type Player struct {
	ID       string   `json:"id"`
	Nickname string   `json:"nickname"`
	Entities []Entity `json:"entities"`
	Team     int      `json:"team"`
	IA       bool     `json:"ia"`
}

type ArenaStartRequest struct {
	MatchID     string   `json:"match_id"`
	CallbackURL string   `json:"callback_url"`
	Players     []Player `json:"players"`
}

type ArenaForfeitRequest struct {
	PlayerID string `json:"player_id"`
}

// ArenaResurrectRequest carries persisted board state from Laravel to rebuild
// a crashed arena. Players carry entities with current HP/Move/Position/Buffs/Skills.
// ISS-054: HasMoved/HasActed flags are not preserved (accepted mid-turn state loss).
type ArenaResurrectRequest struct {
	MatchID         string             `json:"match_id"`
	CallbackURL     string             `json:"callback_url"`
	Players         []Player           `json:"players"`
	Grid            ResurrectGrid      `json:"grid"`
	Turns           []ResurrectTurn    `json:"turns"`
	CurrentEntityID string             `json:"current_entity_id"`
	Version         int64              `json:"version"`
}

// ResurrectGrid is the 2D projection of the engine grid sufficient to rebuild pathfinding.
type ResurrectGrid struct {
	Width     int                `json:"width"`
	Height    int                `json:"height"`   // Y dimension (Length)
	MaxHeight int                `json:"max_height"` // Z ceiling
	Cells     [][]ResurrectCell  `json:"cells"`
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
