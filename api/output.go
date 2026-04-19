package api

import (
	"time"

	"github.com/ecumeurs/upsilonapi/stdmessage"
	"github.com/ecumeurs/upsilonbattle/battlearena/entity"
	"github.com/ecumeurs/upsilonbattle/battlearena/ruler/turner"
	"github.com/ecumeurs/upsilonmapdata/grid"
	"github.com/ecumeurs/upsilonmapdata/grid/cell"
	"github.com/ecumeurs/upsilonmapdata/grid/position"
	"github.com/ecumeurs/upsilonbattle/battlearena/property"
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

// @spec-link [[entity_grid]]

type Cell struct {
	EntityID string `json:"entity_id"` // if any
	Obstacle bool   `json:"obstacle"`  // if any
}

// Grid: A 2D array of cells; for our purpose as in this implementation, the height will be fixed at 1 for every cell giving us a flat map.
type Grid struct {
	Width  int      `json:"width"`
	Height int      `json:"height"`
	Cells  [][]Cell `json:"cells"` // Cells are stored in width-major order.
}

type Turn struct {
	PlayerID string `json:"player_id"`
	Delay    int    `json:"delay"`
	EntityID string `json:"entity_id"`
}

// ActionFeedback provides explicit data about the last tactical action.
// @spec-link [[api_go_action_feedback]]
type ActionFeedback struct {
	Type     string              `json:"type"` // "move", "attack", "pass"
	ActorID  string              `json:"actor_id"`
	TargetID string              `json:"target_id,omitempty"`
	Path     []position.Position `json:"path,omitempty"`
	Damage   int                 `json:"damage,omitempty"`
	PrevHP   int                 `json:"prev_hp,omitempty"`
	NewHP    int                 `json:"new_hp,omitempty"`
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

	return Entity{
		ID:       entity.ID.String(),
		PlayerID: entity.ControllerID.String(),
		Team:     team,
		Name:     entity.Name,
		HP:       hp,
		MaxHP:    maxHP,
		Attack:   entity.GetPropertyI(property.Attack).I(),
		Defense:  entity.GetPropertyI(property.Defense).I(),
		Move:     move,
		MaxMove:  maxMove,
		Position: Position{X: entity.Position.X, Y: entity.Position.Y},
	}
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

	// Map Grid
	bs.Grid = Grid{
		Width:  g.Width,
		Height: g.Length,
		Cells:  make([][]Cell, g.Width),
	}
	for x := 0; x < g.Width; x++ {
		bs.Grid.Cells[x] = make([]Cell, g.Length)
		for y := 0; y < g.Length; y++ {
			z := g.TopMostCellAt(x, y)
			cl, ok := g.CellAt(position.New(x, y, z))
			if ok {
				bs.Grid.Cells[x][y] = Cell{
					EntityID: cl.EntityID.String(),
					Obstacle: cl.Type == cell.Obstacle,
				}
				if cl.EntityID == uuid.Nil {
					bs.Grid.Cells[x][y].EntityID = ""
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
					// Note: the dead flag will be added by the Laravel Gateway resource transformation.
					bs.Players[i].Entities[j].HP = 0
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
