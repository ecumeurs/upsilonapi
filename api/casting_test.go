package api

import (
	"testing"

	"github.com/ecumeurs/upsilonmapdata/grid/position"
	"github.com/ecumeurs/upsilontypes/entity"
	"github.com/ecumeurs/upsilontypes/entity/skill"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// @test-link [[upsilonbattle:mechanic_channeling_mechanic]]

// TestNewEntity_SerializesEntityTargetCasting: an entity-target channel projects a
// Casting DTO carrying the resolved skill name, the target entity id, and the gauge.
func TestNewEntity_SerializesEntityTargetCasting(t *testing.T) {
	ent := entity.New()
	sk := skill.New()
	sk.Name = "Fireball"
	ent.Skills[sk.ID] = sk

	targetID := uuid.New()
	ent.IsCasting = &entity.CastingState{
		SkillID:      sk.ID,
		TargetEntity: targetID,
		Interruption: 30,
	}

	dto := NewEntity(ent)

	if assert.NotNil(t, dto.IsCasting, "casting entity must serialize is_casting") {
		assert.Equal(t, sk.ID.String(), dto.IsCasting.SkillID)
		assert.Equal(t, "Fireball", dto.IsCasting.SkillName, "skill name resolved for the indicator")
		assert.Equal(t, targetID.String(), dto.IsCasting.TargetEntity)
		assert.Equal(t, 30, dto.IsCasting.Interruption)
		assert.Nil(t, dto.IsCasting.TargetTile, "entity-target channel carries no tile")
	}
}

// TestNewEntity_SerializesTileTargetCasting: a tile-target channel projects the
// fixed target tile instead of an entity.
func TestNewEntity_SerializesTileTargetCasting(t *testing.T) {
	ent := entity.New()
	sk := skill.New()
	sk.Name = "Meteor"
	ent.Skills[sk.ID] = sk

	tile := position.New(3, 4, 3)
	ent.IsCasting = &entity.CastingState{SkillID: sk.ID, TargetPos: &tile}

	dto := NewEntity(ent)

	if assert.NotNil(t, dto.IsCasting) {
		assert.Empty(t, dto.IsCasting.TargetEntity, "tile-target channel carries no entity")
		if assert.NotNil(t, dto.IsCasting.TargetTile) {
			assert.Equal(t, 3, dto.IsCasting.TargetTile.X)
			assert.Equal(t, 4, dto.IsCasting.TargetTile.Y)
		}
	}
}

// TestNewEntity_NoCastingOmitsField: a non-casting entity serializes is_casting as
// absent (nil → omitempty).
func TestNewEntity_NoCastingOmitsField(t *testing.T) {
	dto := NewEntity(entity.New())
	assert.Nil(t, dto.IsCasting)
}
