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
								Targeting: api.Flex[api.PropertyMap]{Data: api.PropertyMap{
									"Accuracy": api.PropertyDTO{Value: intPtr(80)},
								}},
								Costs: api.Flex[api.PropertyMap]{Data: api.PropertyMap{
									"Delay": api.PropertyDTO{
										Value: intPtr(0),
										Max:   intPtr(3),
									},
								}},
								Effect: api.Flex[api.PropertyMap]{Data: api.PropertyMap{
									"Damage": api.PropertyDTO{Value: intPtr(120)},
								}},
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
								Properties: api.Flex[api.PropertyMap]{Data: api.PropertyMap{
									"WeaponBaseDamage": api.PropertyDTO{Value: intPtr(3)},
								}},
							},
						},
						EquippedSkills: []api.EquippedSkill{
							{
								SkillID:   inventorySkillID.String(),
								Name:      "Dodge",
								Behavior:  "Reaction",
								Targeting: api.Flex[api.PropertyMap]{Data: api.PropertyMap{}},
								Costs:     api.Flex[api.PropertyMap]{Data: api.PropertyMap{}},
								Effect: api.Flex[api.PropertyMap]{Data: api.PropertyMap{
									"Dodge": api.PropertyDTO{Value: intPtr(50)},
								}},
								Origin: "inventory",
							},
							{
								SkillID:   itemSkillID.String(),
								Name:      "Launch Grenade",
								Behavior:  "Direct",
								Targeting: api.Flex[api.PropertyMap]{Data: api.PropertyMap{"Accuracy": api.PropertyDTO{Value: intPtr(75)}}},
								Costs:     api.Flex[api.PropertyMap]{Data: api.PropertyMap{}},
								Effect:    api.Flex[api.PropertyMap]{Data: api.PropertyMap{"Damage": api.PropertyDTO{Value: intPtr(200)}}},
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
								Effect: api.Flex[api.PropertyMap]{Data: api.PropertyMap{"Damage": api.PropertyDTO{Value: intPtr(50)}}},
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
