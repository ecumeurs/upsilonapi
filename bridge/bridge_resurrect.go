package bridge

// @spec-link [[api_battle_proxy]]
// @spec-link [[mechanic_mech_arena_lifecycle]]

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
	"github.com/ecumeurs/upsilontypes/property"
	"github.com/ecumeurs/upsilontypes/property/def"
	"github.com/ecumeurs/upsilontools/tools/actor"
	"github.com/ecumeurs/upsilontools/tools/messagequeue/message"
	"github.com/ecumeurs/upsilonbattle/battlearena/ruler/rulermethods"
	"github.com/ecumeurs/upsilonbattle/battlearena/ruler/turner"
	"github.com/google/uuid"
)

// ResurrectArena rebuilds a crashed arena from a persisted board state (ISS-054).
// It reconstructs the grid, entities, turner queue, and controllers, then hands the
// turn to the entity that was active when the engine went down.
func (b *ArenaBridge) ResurrectArena(req api.ArenaResurrectRequest) (api.BoardState, error) {
	if req.MatchID == "" {
		return api.BoardState{}, fmt.Errorf("mandatory field match_id is missing")
	}
	matchID, err := uuid.Parse(req.MatchID)
	if err != nil {
		return api.BoardState{}, fmt.Errorf("invalid match_id: %w", err)
	}
	if req.CallbackURL == "" {
		return api.BoardState{}, fmt.Errorf("mandatory field callback_url is missing")
	}
	if len(req.Players) == 0 {
		return api.BoardState{}, fmt.Errorf("arena must have at least one player")
	}

	// Idempotency: if arena is already alive (e.g. double-call), reject.
	b.mu.RLock()
	_, exists := b.arenas[matchID]
	b.mu.RUnlock()
	if exists {
		return api.BoardState{}, fmt.Errorf("arena %s is already running — resurrection not needed", matchID)
	}

	// 1. Rebuild grid from serialized 2D projection.
	g := resurrectGrid(req.Grid)

	// 2. Create new arena and configure it.
	battleArena := battlearena.NewBattleArena(matchID)
	battleArena.Metadata["CallbackURL"] = req.CallbackURL
	battleArena.Metadata["Players"] = req.Players
	battleArena.Ruler.ID = matchID
	battleArena.Ruler.SetGrid(g)
	battleArena.Ruler.SetNbControllers(len(req.Players))

	// 3. Restore entities from saved board state.
	for _, p := range req.Players {
		playerID, err := uuid.Parse(p.ID)
		if err != nil {
			return api.BoardState{}, fmt.Errorf("invalid player_id for player %s: %w", p.Nickname, err)
		}
		for _, ee := range p.Entities {
			if ee.Dead || ee.HP <= 0 {
				continue // skip dead entities — they are no longer in the game
			}
			entID, err := uuid.Parse(ee.ID)
			if err != nil {
				return api.BoardState{}, fmt.Errorf("invalid entity_id for entity %s: %w", ee.Name, err)
			}
			if ee.MaxHP <= 0 {
				return api.BoardState{}, fmt.Errorf("entity %s must have max_hp > 0", ee.Name)
			}

			e := entity.Entity{
				ID:           entID,
				Type:         entity.Character,
				Name:         ee.Name,
				ControllerID: playerID,
			}
			e.Properties = make(map[string]property.Property)
			e.Skills = make(map[uuid.UUID]skill.Skill)

			// Restore exact persisted position; no random fallback.
			e.Position = position.New(ee.Position.X, ee.Position.Y, g.TopMostGroundAt(ee.Position.X, ee.Position.Y))

			e.RepsertPropertyCMaxValue(property.HP, ee.MaxHP)
			e.RepsertPropertyCValue(property.HP, ee.HP)
			e.RepsertPropertyCMaxValue(property.Movement, ee.MaxMove)
			e.RepsertPropertyCValue(property.Movement, ee.Move)
			e.RepsertPropertyValue(property.Attack, ee.Attack)
			e.RepsertPropertyValue(property.Defense, ee.Defense)
			e.RepsertPropertyValue(property.TeamID, p.Team)

			// Re-apply persisted buffs (item effects, status effects, etc.).
			for _, b := range ee.Buffs {
				originID, err := uuid.Parse(b.OriginID)
				if err != nil {
					log.Printf("[ArenaBridge.Resurrect] Skipping buff with bad origin_id for entity %s", ee.Name)
					continue
				}
				buff := property.TemporaryProperties{
					Forever:        b.Forever,
					OriginEntityID: originID,
					Properties:     make(map[string]property.Property),
				}
				for key, dto := range b.Properties.Data {
					effectiveKey := key
					if alias, ok := propertyAliasMap[effectiveKey]; ok {
						effectiveKey = alias
					}
					var p property.Property
					if prop := def.ItemProperty(property.ItemProperties(effectiveKey)); prop != nil {
						p = prop
					} else if prop := def.EntityProperty(property.EntityProperties(effectiveKey)); prop != nil {
						p = prop
					}
					if p != nil {
						if setSkillPropValue(p, dto) {
							buff.Properties[property.PropertyToString(effectiveKey)] = p
						}
					}
				}
				e.RegisterBuff(buff)
			}

			// Restore equipped skills.
			for _, es := range ee.EquippedSkills {
				skillID, err := uuid.Parse(es.SkillID)
				if err != nil {
					log.Printf("[ArenaBridge.Resurrect] Skipping skill %s for entity %s: invalid UUID", es.Name, ee.Name)
					continue
				}
				bh := def.DefaultBehavior()
				bh.SetS(string(parseBehaviorType(es.Behavior)))
				s := skill.Skill{
					ID:        skillID,
					Name:      es.Name,
					Behavior:  bh,
					Targeting: buildSkillPropertyMap(es.Targeting.Data),
					Costs:     buildSkillPropertyMap(es.Costs.Data),
					Effect:    buildSkillEffect(es.Effect.Data),
				}
				e.RegisterSkill(s)
			}

			battleArena.Ruler.AddEntity(e)
		}
	}

	// 4. Restore turner queue and mark ruler as in-progress.
	var turnerTurns []turner.EntityTurn
	for _, t := range req.Turns {
		entID, err := uuid.Parse(t.EntityID)
		if err != nil {
			log.Printf("[ArenaBridge.Resurrect] Skipping turn entry with bad entity_id: %v", err)
			continue
		}
		turnerTurns = append(turnerTurns, turner.EntityTurn{EntityId: entID, Delay: t.Delay})
	}
	currentEntityID := uuid.Nil
	if req.CurrentEntityID != "" {
		currentEntityID, err = uuid.Parse(req.CurrentEntityID)
		if err != nil {
			return api.BoardState{}, fmt.Errorf("invalid current_entity_id: %w", err)
		}
	}
	battleArena.Ruler.Resurrect(turnerTurns, currentEntityID, req.Version)

	// 5. Start the Ruler actor.
	battleArena.Ruler.Start()

	// 6. Re-create controllers.
	respChan := make(chan *message.Message)
	defer close(respChan)
	count := len(req.Players)

	var humanPlayerIDs []uuid.UUID
	for _, p := range req.Players {
		if !p.IA {
			humanPlayerIDs = append(humanPlayerIDs, uuid.MustParse(p.ID))
		}
	}
	var sharedHC *HTTPController
	if len(humanPlayerIDs) > 0 {
		sharedHC = NewHTTPController(matchID, req.CallbackURL, req.Players, humanPlayerIDs)
		sharedHC.Ruler = battleArena.Ruler
		sharedHC.Start()
	}

	for _, p := range req.Players {
		var ctrl actor.Communication
		pID := uuid.MustParse(p.ID)
		if p.IA {
			iac := controllers.NewAggressiveController(pID, fmt.Sprintf("AggressiveController-%s", p.ID))
			iac.Start()
			ctrl = iac
		} else {
			ctrl = sharedHC
		}
		msg := message.Create(ctrl, rulermethods.AddController{
			Controller:   ctrl,
			ControllerID: pID,
		}, rulermethods.AddControllerReply{})
		battleArena.Ruler.SendActor(msg, respChan)
	}

	for i := 0; i < count; i++ {
		log.Printf("[ArenaBridge.Resurrect] Waiting for AddController reply (%d/%d) for match %s", i+1, count, matchID)
		select {
		case msg := <-respChan:
			log.Printf("[ArenaBridge.Resurrect] AddController reply for match %s (Error: %v)", matchID, msg.HasError)
		case <-time.After(5 * time.Second):
			log.Printf("[ArenaBridge.Resurrect] TIMEOUT waiting for AddController reply for match %s", matchID)
		}
	}

	// 7. Register arena and trigger resurrection turn hand-off.
	b.mu.Lock()
	b.arenas[matchID] = battleArena
	b.mu.Unlock()

	battleArena.Ruler.NotifyActor(message.Create(nil, rulermethods.Resurrect{
		CurrentEntityID: currentEntityID,
	}, nil))

	res := make([]entity.Entity, 0, 6)
	for _, v := range battleArena.Ruler.GameState.Entities {
		res = append(res, v)
	}
	return api.NewBoardState(
		matchID,
		battleArena.Ruler.GameState.Grid,
		res,
		req.Players,
		battleArena.Ruler.GameState.Turner.GetTurnState(),
		time.Now(),
		time.Now().Add(30*time.Second),
		0,
		req.Version,
		nil,
	), nil
}

// resurrectGrid reconstructs a 3D engine grid from the 2D serialized projection.
// Each (x,y) column gets dirt cells below and a ground (or obstacle) cell at the saved surface Z.
// This ensures the resurrected arena has identical pathfinding constraints to the original.
// @spec-link [[api_battle_proxy]]
func resurrectGrid(rg api.ResurrectGrid) *grid.Grid {
	// Initialize a new grid with dimensions matching the serialized request.
	g := &grid.Grid{
		Width:  rg.Width,
		Length: rg.Height, // DTO field "height" is the Y dimension (Length in engine terms).
		Height: rg.MaxHeight,
		Cells:  make(map[position.Position]*cell.Cell),
	}
	// Iterate through each coordinate in the 2D projection.
	for x := 0; x < rg.Width; x++ {
		for y := 0; y < rg.Height; y++ {
			resurrectGridColumn(g, rg, x, y)
		}
	}
	// Return the reconstructed 3D grid for engine consumption.
	return g
}

// resurrectGridColumn rebuilds a single (x,y) vertical stack of cells.
func resurrectGridColumn(g *grid.Grid, rg api.ResurrectGrid, x int, y int) {
	var surfaceZ int
	var isObstacle bool
	// Extract surface height and obstruction status from the DTO.
	if x < len(rg.Cells) && y < len(rg.Cells[x]) {
		surfaceZ = rg.Cells[x][y].Height
		isObstacle = rg.Cells[x][y].Obstacle
	}
	// Lay dirt cells below the surface to fill the 3D volume.
	for z := 0; z < surfaceZ; z++ {
		pos := position.New(x, y, z)
		g.Cells[pos] = cell.NewCell(cell.Dirt, pos)
	}
	// Create the surface cell (Ground or Obstacle) at the specified height.
	pos := position.New(x, y, surfaceZ)
	cellType := cell.Ground
	if isObstacle {
		cellType = cell.Obstacle
	}
	g.Cells[pos] = cell.NewCell(cellType, pos)
}
