package bridge

import (
	"testing"

	"github.com/ecumeurs/upsilonapi/api"
	"github.com/ecumeurs/upsilontypes/property"
	"github.com/ecumeurs/upsilontypes/property/def"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestMapping_ZoneAndEffect(t *testing.T) {
	b := Get()
	matchID := uuid.New()
	skillID := uuid.New()
	itemID := uuid.New()

	zone := "Neighbours"
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
						Name:    "Hero",
						HP:      10,
						MaxHP:   10,
						EquippedItems: []api.EquippedItem{
							{
								ItemID: itemID.String(),
								Name:   "Magic Ring",
								Slot:   "ring",
								Effect: api.Flex[api.PropertyMap]{Data: api.PropertyMap{
									"Heal": api.PropertyDTO{Value: intPtr(5)},
								}},
								Zone: &zone,
							},
						},
						EquippedSkills: []api.EquippedSkill{
							{
								SkillID:  skillID.String(),
								Name:     "Explosion",
								Behavior: "Direct",
								Effect: api.Flex[api.PropertyMap]{Data: api.PropertyMap{
									"Damage": api.PropertyDTO{Value: intPtr(50)},
								}},
								Zone: &zone,
							},
						},
					},
				},
			},
		},
	}

	// 1. Test Mapping IN (Bridge.StartArena)
	_, _, entities, _, _, _, err := b.StartArena(req)
	assert.NoError(t, err)
	defer b.DestroyArena(matchID)

	ent := entities[0]
	
	// Check Skill Zone
	s, ok := ent.Skills[skillID]
	assert.True(t, ok)
	zp, ok := s.Targeting[property.PropertyToString(property.Zone)].(*def.ZoneProperty)
	assert.True(t, ok, "Zone should be a ZoneProperty")
	assert.Equal(t, "Neighbours", zp.PatternType)
	assert.Equal(t, 27, len(zp.ZonePattern), "Neighbours pattern should have 27 positions (3x3x3)")

	// Check Item Effect & Zone (via Buffs)
	assert.Len(t, ent.Buffs, 1)
	buff := ent.Buffs[0]
	assert.Equal(t, itemID, buff.OriginEntityID)
	
	ep, ok := buff.Properties[property.PropertyToString(property.Effect)].(*def.EffectProperty)
	assert.True(t, ok, "Effect should be an EffectProperty")
	assert.NotNil(t, ep.Effect)
	assert.Equal(t, 1, len(ep.Effect.Properties))
	heal := ep.Effect.Properties[0].(property.IntProperty)
	assert.Equal(t, 5, heal.I())

	zp2, ok := buff.Properties[property.PropertyToString(property.Zone)].(*def.ZoneProperty)
	assert.True(t, ok, "Zone should be a ZoneProperty in buff")
	assert.Equal(t, "Neighbours", zp2.PatternType)

	// 2. Test Mapping OUT (api.NewEntity)
	dto := api.NewEntity(ent)
	
	// Check Skill DTO
	assert.Len(t, dto.EquippedSkills, 1)
	assert.Equal(t, "Neighbours", *dto.EquippedSkills[0].Zone)
	assert.Equal(t, 50, *dto.EquippedSkills[0].Effect.Data["Damage"].Value)
	
	// Check Item DTO
	assert.Len(t, dto.EquippedItems, 1)
	assert.Equal(t, itemID.String(), dto.EquippedItems[0].ItemID)
	assert.Equal(t, "Neighbours", *dto.EquippedItems[0].Zone)
	assert.Equal(t, 5, *dto.EquippedItems[0].Effect.Data["Heal"].Value)
	
	// Ensure Zone and Effect are EXCLUDED from the Properties/Targeting maps
	assert.NotContains(t, dto.EquippedSkills[0].Targeting.Data, "Zone")
	assert.NotContains(t, dto.EquippedItems[0].Properties.Data, "Zone")
	assert.NotContains(t, dto.EquippedItems[0].Properties.Data, "Effect")
}
