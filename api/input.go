package api

import "github.com/ecumeurs/upsilonapi/stdmessage"

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
	EquippedSkills []string       `json:"equipped_skills"` // reserved for ISS-073
}

type Buff struct {
	OriginID   string         `json:"origin_id"`
	Forever    bool           `json:"forever"`
	Properties map[string]any `json:"properties"`
}

type EquippedItem struct {
	ItemID     string         `json:"item_id"`
	Name       string         `json:"name"`
	Slot       string         `json:"slot"`
	Properties map[string]any `json:"properties"`
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

type ArenaActionMessage = stdmessage.StandardMessage[ArenaActionRequest, stdmessage.MetaNil]
type ArenaStartMessage = stdmessage.StandardMessage[ArenaStartRequest, stdmessage.MetaNil]
type ArenaForfeitMessage = stdmessage.StandardMessage[ArenaForfeitRequest, stdmessage.MetaNil]
