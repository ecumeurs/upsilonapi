package bridge

import (
	"testing"

	"github.com/ecumeurs/upsilonapi/api"
	"github.com/ecumeurs/upsilontypes/property"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestArenaInit_EquippedItemsBecomeBuffs(t *testing.T) {
	bridge := Get()
	matchID := uuid.New()
	playerID := uuid.New()
	entityID := uuid.New()
	itemID := uuid.New()

	req := api.ArenaStartRequest{
		MatchID:     matchID.String(),
		CallbackURL: "http://localhost/webhook",
		Players: []api.Player{
			{
				ID:   playerID.String(),
				Team: 1,
				IA:   true,
				Entities: []api.Entity{
					{
						ID:      entityID.String(),
						Name:    "Warrior",
						HP:      10,
						MaxHP:   10,
						Move:    3,
						MaxMove: 3,
						Attack:  5,
						Defense: 2,
						EquippedItems: []api.EquippedItem{
							{
								ItemID: itemID.String(),
								Name:   "Heavy Armor",
								Slot:   "armor",
								Properties: map[string]any{
									"ArmorRating": 5,
								},
							},
						},
					},
				},
			},
		},
	}

	_, _, entities, _, _, _, err := bridge.StartArena(req)
	assert.NoError(t, err)

	// Clean up for other tests
	defer bridge.DestroyArena(matchID)

	assert.Len(t, entities, 1)
	ent := entities[0]

	// Verify the buff is present
	assert.Len(t, ent.Buffs, 1)
	assert.Equal(t, itemID, ent.Buffs[0].OriginEntityID)
	assert.True(t, ent.Buffs[0].Forever)

	// Verify the property is correctly applied
	// ArmorRating in ItemProperties maps to "Armor" string
	armorProp := ent.GetProperty(property.ArmorRating)
	assert.NotNil(t, armorProp)
	assert.Equal(t, 5, armorProp.Get().(int))
}

func TestArenaInit_StatMapping(t *testing.T) {
	bridge := Get()
	matchID := uuid.New()
	itemID := uuid.New()

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
						Name:    "Swift Rogue",
						HP:      10,
						MaxHP:   10,
						Move:    3,
						MaxMove: 3,
						EquippedItems: []api.EquippedItem{
							{
								ItemID: itemID.String(),
								Name:   "Swift Boots",
								Slot:   "utility",
								Properties: map[string]any{
									"Movement": 2, // Movement is an EntityProperty
								},
							},
						},
					},
				},
			},
		},
	}

	_, _, entities, _, _, _, err := bridge.StartArena(req)
	assert.NoError(t, err)
	defer bridge.DestroyArena(matchID)

	ent := entities[0]
	// Movement is a counter property, 3 (base) + 2 (buff) = 5
	mvt := ent.GetProperty(property.Movement)
	assert.Equal(t, 5, mvt.Get().(int))
}
