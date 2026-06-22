package bridge

// @spec-link [[rule_team_mechanics]]
// @spec-link [[mechanic_item_buff_application]]
// @spec-link [[api_character_skill_inventory]]

import (
	"fmt"
	"log"
	"time"

	"github.com/ecumeurs/upsilonapi/api"
	"github.com/ecumeurs/upsilonbattle/battlearena"
	"github.com/ecumeurs/upsilontypes/entity"
	"github.com/ecumeurs/upsilontypes/entity/skill"
	"github.com/ecumeurs/upsilonmapdata/grid"
	"github.com/ecumeurs/upsilonmapdata/grid/position"
	"github.com/ecumeurs/upsilontypes/property"
	"github.com/ecumeurs/upsilontypes/property/def"
	"github.com/ecumeurs/upsilonbattle/battlearena/ruler/rulermethods"
	"github.com/ecumeurs/upsilonbattle/battlearena/ruler/turner"
	"github.com/ecumeurs/upsilonmapmaker/gridgenerator"
	"github.com/ecumeurs/upsilontools/tools"
	"github.com/ecumeurs/upsilontools/tools/actor"
	"github.com/ecumeurs/upsilontools/tools/messagequeue/message"
	"github.com/google/uuid"
)

// StartArena initializes a new battle arena from a start request.
// It generates the grid, configures entities, and starts the ruler actor.
func (b *ArenaBridge) StartArena(start api.ArenaStartRequest) (uuid.UUID, *grid.Grid, []entity.Entity, []api.Player, turner.TurnState, int64, error) {
	if start.MatchID == "" {
		return uuid.Nil, nil, nil, nil, turner.TurnState{}, 0, fmt.Errorf("mandatory field match_id is missing")
	}
	if err := validateTeamComposition(start.Players); err != nil {
		return uuid.Nil, nil, nil, nil, turner.TurnState{}, 0, err
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
	battleArena.Ruler.ID = matchID

	b.mu.Lock()
	b.arenas[matchID] = battleArena
	b.mu.Unlock()

	// 1. Grid Generation
	gg := gridgenerator.GridGenerator{
		Width:               tools.NewIntRange(7, 8),
		Length:              tools.NewIntRange(7, 8),
		Height:              tools.NewIntRange(2, 3),
		Type:                gridgenerator.Flat,
		GenerateObstrcution: true,
		ObstructionRate:     tools.NewIntRange(2, 8),
	}
	battleArena.Ruler.SetGrid(gg.Generate())
	battleArena.Ruler.SetNbControllers(len(start.Players))

	// 2. Entity Configuration
	if err := b.configureArenaEntities(battleArena, start.Players); err != nil {
		return uuid.Nil, nil, nil, nil, turner.TurnState{}, 0, err
	}

	// 3. Ruler & Controller Startup
	battleArena.Ruler.Start()
	if err := b.setupControllers(battleArena, start.CallbackURL, start.Players); err != nil {
		return uuid.Nil, nil, nil, nil, turner.TurnState{}, 0, err
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

// configureArenaEntities maps API player/entity data into the engine.
// When Entity.AutoGen is true the entity's stats and skills are generated from
// the resolved archetype + grade; otherwise the explicit values are used as before.
func (b *ArenaBridge) configureArenaEntities(ba *battlearena.BattleArena, players []api.Player) error {
	teamSlugs := make(map[int][]string)
	for _, p := range players {
		playerID, err := uuid.Parse(p.ID)
		if err != nil {
			return fmt.Errorf("invalid player_id for player %s: %w", p.Nickname, err)
		}
		if err := b.configurePlayerEntities(ba, p, playerID, teamSlugs); err != nil {
			return err
		}
	}
	return nil
}

// configurePlayerEntities adds all entities for one player to the arena.
func (b *ArenaBridge) configurePlayerEntities(ba *battlearena.BattleArena, p api.Player, playerID uuid.UUID, teamSlugs map[int][]string) error {
	playerGrade := resolvePlayerGrade(p)
	for _, ee := range p.Entities {
		entID, err := uuid.Parse(ee.ID)
		if err != nil {
			return fmt.Errorf("invalid entity_id for entity %s: %w", ee.Name, err)
		}
		if ee.AutoGen {
			if err := b.addAutoGenEntity(ba, p, ee, entID, playerID, playerGrade, teamSlugs); err != nil {
				return err
			}
			continue
		}
		if err := b.addExplicitEntity(ba, p, ee, entID, playerID); err != nil {
			return err
		}
	}
	return nil
}

// addAutoGenEntity generates an entity from archetype+grade and registers it in the arena.
func (b *ArenaBridge) addAutoGenEntity(ba *battlearena.BattleArena, p api.Player, ee api.Entity, entID, playerID uuid.UUID, playerGrade string, teamSlugs map[int][]string) error {
	slug := resolveArchetypeSlug(ee.Archetype, p.Archetype, teamSlugs[p.Team])
	teamSlugs[p.Team] = append(teamSlugs[p.Team], slug)
	entGrade := resolveEntityGrade(ee, playerGrade)

	startPos := position.Position{}
	if ee.Position.X != 0 || ee.Position.Y != 0 {
		startPos = position.New(ee.Position.X, ee.Position.Y, ba.Ruler.GameState.Grid.TopMostGroundAt(ee.Position.X, ee.Position.Y))
	}

	e, err := generateEntityFromArchetype(entID, ee.Name, p.Team, playerID, slug, entGrade, startPos)
	if err != nil {
		return fmt.Errorf("auto-gen failed for entity %s: %w", ee.Name, err)
	}
	ba.Ruler.AddEntity(e)
	return nil
}

// addExplicitEntity builds an entity from explicitly provided stats and registers it in the arena.
func (b *ArenaBridge) addExplicitEntity(ba *battlearena.BattleArena, p api.Player, ee api.Entity, entID, playerID uuid.UUID) error {
	if ee.MaxHP <= 0 {
		return fmt.Errorf("entity %s must have max_hp > 0", ee.Name)
	}
	e := entity.Entity{
		ID:           entID,
		Type:         entity.Character,
		Name:         ee.Name,
		ControllerID: playerID,
	}
	e.Properties = make(map[string]property.Property)
	e.Skills = make(map[uuid.UUID]skill.Skill)

	if ee.Position.X != 0 || ee.Position.Y != 0 {
		e.Position = position.New(ee.Position.X, ee.Position.Y, ba.Ruler.GameState.Grid.TopMostGroundAt(ee.Position.X, ee.Position.Y))
	}

	e.RepsertPropertyCMaxValue(property.HP, ee.MaxHP)
	e.RepsertPropertyCValue(property.HP, ee.HP)
	e.RepsertPropertyCMaxValue(property.Movement, ee.MaxMove)
	e.RepsertPropertyCValue(property.Movement, ee.Move)
	e.RepsertPropertyValue(property.Attack, ee.Attack)
	e.RepsertPropertyValue(property.Defense, ee.Defense)
	e.RepsertPropertyValue(property.TeamID, p.Team)

	for _, item := range ee.EquippedItems {
		b.applyItemAsBuff(&e, item)
	}
	for _, es := range ee.EquippedSkills {
		b.registerEntitySkill(&e, es)
	}
	ba.Ruler.AddEntity(e)
	return nil
}

// applyItemAsBuff converts an equipped item from the API into a permanent engine buff.
// It maps item properties, effects, and zones into the entity's property set.
func (b *ArenaBridge) applyItemAsBuff(e *entity.Entity, item api.EquippedItem) {
	itemID, err := uuid.Parse(item.ItemID)
	if err != nil {
		log.Printf("[ArenaBridge] Skipping item %s: invalid UUID", item.Name)
		return
	}

	buff := property.TemporaryProperties{
		Forever:        true,
		OriginEntityID: itemID,
		Properties:     make(map[string]property.Property),
	}

	if len(item.Effect.Data) > 0 {
		eff := buildSkillEffect(item.Effect.Data)
		buff.Properties[property.PropertyToString(property.Effect)] = def.MakeEffectProperty(&eff, property.Analyser)
	}
	if item.Zone != nil {
		zp := def.DefaultZone()
		zp.Set(*item.Zone)
		buff.Properties[property.PropertyToString(property.Zone)] = zp
	}

	for key, dto := range item.Properties.Data {
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
		if p != nil && setSkillPropValue(p, dto) {
			buff.Properties[property.PropertyToString(effectiveKey)] = p
		}
	}
	e.RegisterBuff(buff)
}

// registerEntitySkill maps an equipped skill from the API into the engine's skill system.
// It configures behavior, targeting, costs, and effects for the entity.
func (b *ArenaBridge) registerEntitySkill(e *entity.Entity, es api.EquippedSkill) {
	skillID, err := uuid.Parse(es.SkillID)
	if err != nil {
		log.Printf("[ArenaBridge] Skipping skill %s: invalid UUID", es.Name)
		return
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
	if es.Zone != nil {
		zp := def.DefaultZone()
		zp.Set(*es.Zone)
		s.Targeting[property.PropertyToString(property.Zone)] = zp
	}
	e.RegisterSkill(s)
}

// setupControllers initializes and registers HTTP and IA controllers.
// It creates a shared HTTPController for all human players and individual IA controllers.
func (b *ArenaBridge) setupControllers(ba *battlearena.BattleArena, callbackURL string, players []api.Player) error {
	// Identify all human players to be managed by the shared HTTP controller.
	var humanPlayerIDs []uuid.UUID
	for _, p := range players {
		if !p.IA {
			humanPlayerIDs = append(humanPlayerIDs, uuid.MustParse(p.ID))
		}
	}

	// Initialize the shared HTTP controller if any human players are present.
	var sharedHC *HTTPController
	if len(humanPlayerIDs) > 0 {
		sharedHC = NewHTTPController(ba.Ruler.ID, callbackURL, players, humanPlayerIDs)
		sharedHC.Ruler = ba.Ruler
		sharedHC.Start()
	}

	// Use a response channel to synchronize with the Ruler's registration process.
	respChan := make(chan *message.Message)
	defer close(respChan)

	// Register each player's controller with the Ruler actor.
	for _, p := range players {
		var ctrl actor.Communication
		pID := uuid.MustParse(p.ID)
		if p.IA {
			// Determine the archetype for this AI player (first entity's resolved slug, or player-level).
			slug := p.Archetype
			if slug == "" && len(p.Entities) > 0 {
				slug = p.Entities[0].Archetype
			}
			gradeStr := resolvePlayerGrade(p)
			iac := newAIControllerForArchetype(pID, fmt.Sprintf("AIController-%s", p.ID), slug, gradeStr)
			iac.Start()
			ctrl = iac
		} else {
			// Reuse the shared HTTP controller for human players.
			ctrl = sharedHC
		}

		msg := message.Create(ctrl, rulermethods.AddController{
			Controller:   ctrl,
			ControllerID: pID,
		}, rulermethods.AddControllerReply{})
		ba.Ruler.SendActor(msg, respChan)
	}

	// Wait for all controllers to be acknowledged by the Ruler.
	for i := 0; i < len(players); i++ {
		select {
		case <-respChan:
			// Controller registration successful.
		case <-time.After(5 * time.Second):
			// Safety timeout to prevent initialization hangs.
			return fmt.Errorf("timeout waiting for controller registration")
		}
	}
	return nil
}
