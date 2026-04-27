package bridge

import (
	"testing"

	"github.com/ecumeurs/upsilonapi/api"
	"github.com/ecumeurs/upsilontypes/property"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// @test-link [[mec_skill_payload_resolution]]
// @test-link [[api_character_skill_inventory]]

func TestArenaInit_EquippedSkillRegistered(t *testing.T) {
	b := Get()
	matchID := uuid.New()
	skillID := uuid.New()

	req := api.ArenaStartRequest{
		MatchID:     matchID.String(),
		CallbackURL: "http://localhost/webhook",
		Players: []api.Player{
			{
				ID:   uuid.New().String(),
				Team: 1,
				IA:   true,
				Entities: []api.Entity{
					{
						ID:      uuid.New().String(),
						Name:    "Mage",
						HP:      10,
						MaxHP:   10,
						Move:    3,
						MaxMove: 3,
						EquippedSkills: []api.EquippedSkill{
							{
								SkillID:  skillID.String(),
								Name:     "Fireball",
								Behavior: "Direct",
								Targeting: map[string]interface{}{
									"Accuracy": float64(80),
								},
								Costs: map[string]interface{}{
									"Delay": map[string]interface{}{
										"value": float64(0),
										"max":   float64(3),
									},
								},
								Effect: map[string]interface{}{
									"Damage": float64(120),
								},
								Origin: "inventory",
							},
						},
					},
				},
			},
		},
	}

	_, _, entities, _, _, _, err := b.StartArena(req)
	assert.NoError(t, err)
	defer b.DestroyArena(matchID)

	assert.Len(t, entities, 1)
	ent := entities[0]

	assert.Len(t, ent.Skills, 1, "entity should have 1 registered skill")

	s, ok := ent.Skills[skillID]
	assert.True(t, ok, "skill should be registered with the payload skill_id")
	assert.Equal(t, "Fireball", s.Name)
	assert.True(t, s.IsDirect())

	dmg := s.Effect.GetProperty(property.Damage)
	assert.NotNil(t, dmg)
	assert.Equal(t, 120, dmg.(property.IntProperty).I())
}

func TestArenaInit_ItemSkillAndInventorySkillCoexist(t *testing.T) {
	b := Get()
	matchID := uuid.New()
	inventorySkillID := uuid.New()
	itemSkillID := uuid.New()

	req := api.ArenaStartRequest{
		MatchID:     matchID.String(),
		CallbackURL: "http://localhost/webhook",
		Players: []api.Player{
			{
				ID:   uuid.New().String(),
				Team: 1,
				IA:   true,
				Entities: []api.Entity{
					{
						ID:      uuid.New().String(),
						Name:    "Gunner",
						HP:      10,
						MaxHP:   10,
						Move:    3,
						MaxMove: 3,
						EquippedItems: []api.EquippedItem{
							{
								ItemID: uuid.New().String(),
								Name:   "Grenade Launcher",
								Slot:   "weapon",
								Properties: map[string]any{
									"WeaponBaseDamage": float64(3),
								},
							},
						},
						EquippedSkills: []api.EquippedSkill{
							{
								SkillID:   inventorySkillID.String(),
								Name:      "Dodge",
								Behavior:  "Reaction",
								Targeting: map[string]interface{}{},
								Costs:     map[string]interface{}{},
								Effect: map[string]interface{}{
									"Dodge": float64(50),
								},
								Origin: "inventory",
							},
							{
								SkillID:   itemSkillID.String(),
								Name:      "Launch Grenade",
								Behavior:  "Direct",
								Targeting: map[string]interface{}{"Accuracy": float64(75)},
								Costs:     map[string]interface{}{},
								Effect:    map[string]interface{}{"Damage": float64(200)},
								Origin:    "item:" + uuid.New().String(),
							},
						},
					},
				},
			},
		},
	}

	_, _, entities, _, _, _, err := b.StartArena(req)
	assert.NoError(t, err)
	defer b.DestroyArena(matchID)

	ent := entities[0]
	assert.Len(t, ent.Skills, 2, "should have both inventory and item-derived skills")

	invSkill, ok := ent.Skills[inventorySkillID]
	assert.True(t, ok)
	assert.True(t, invSkill.IsReaction())

	itemSkill, ok := ent.Skills[itemSkillID]
	assert.True(t, ok)
	assert.True(t, itemSkill.IsDirect())
}

func TestArenaInit_InvalidSkillUUIDSkipped(t *testing.T) {
	b := Get()
	matchID := uuid.New()

	req := api.ArenaStartRequest{
		MatchID:     matchID.String(),
		CallbackURL: "http://localhost/webhook",
		Players: []api.Player{
			{
				ID:   uuid.New().String(),
				Team: 1,
				IA:   true,
				Entities: []api.Entity{
					{
						ID:      uuid.New().String(),
						Name:    "Fighter",
						HP:      10,
						MaxHP:   10,
						Move:    3,
						MaxMove: 3,
						EquippedSkills: []api.EquippedSkill{
							{
								SkillID:  "not-a-uuid",
								Name:     "Bad Skill",
								Behavior: "Direct",
								Effect:   map[string]interface{}{"Damage": float64(50)},
							},
						},
					},
				},
			},
		},
	}

	_, _, entities, _, _, _, err := b.StartArena(req)
	assert.NoError(t, err)
	defer b.DestroyArena(matchID)

	ent := entities[0]
	assert.Empty(t, ent.Skills, "invalid UUID should be silently skipped")
}
