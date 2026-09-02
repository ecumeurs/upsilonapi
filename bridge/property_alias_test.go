// Package bridge provides a regression test pinning that the ArmorRating wire key resolves
// correctly now that propertyAliasMap has been removed (ISS-143), per CODING_RULE §5.
package bridge

import (
	"testing"

	"github.com/ecumeurs/upsilonapi/api"
	"github.com/ecumeurs/upsilontypes/entity"
	"github.com/ecumeurs/upsilontypes/property"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPropertyIngestion_ArmorRatingAliasMustNotBeSilentlyDropped guards a historical
// regression: an earlier revision of propertyAliasMap mapped the incoming "ArmorRating" wire
// key to a stale, dead string ("Armor"), which made resolution fail and silently dropped the
// property instead of registering it on the entity (ISS-143). That alias entry no longer
// exists in propertyAliasMap — propertyAliasMap itself was removed entirely as part of
// ISS-143/ISS-140 (the registry now resolves "ArmorRating" directly, registered ScopeItem in
// registry_item.go). This test is retained as a regression guard: it pins that the real,
// current property.Key wire key ("ArmorRating") keeps resolving correctly with no
// intermediary alias layer at all.
// @test-link [[mechanic_item_buff_application]]
func TestPropertyIngestion_ArmorRatingAliasMustNotBeSilentlyDropped(t *testing.T) {
	// 1. Setup Phase: A bridge, a fresh entity, and an equipped item carrying the
	// ArmorRating wire key, which is the real, current property.Key constant value.
	b := Get()
	e := entity.New()
	item := api.EquippedItem{
		ItemID: uuid.New().String(),
		Name:   "Plate Armor",
		Slot:   "chest",
		Properties: api.Flex[api.PropertyMap]{
			Data: api.PropertyMap{"ArmorRating": {Value: intPtr(5)}},
		},
	}

	// 2. Execution Phase: Apply the item as a buff directly, bypassing StartArena.
	err := b.applyItemAsBuff(&e, item)

	// 3. Validation Phase: The ArmorRating buff must actually land on the entity with
	// its value intact, with no error, now that resolution goes straight through the
	// registry (def.Lookup, scoped to Item) with no alias indirection.
	require.NoError(t, err)
	buffed := e.GetBuffsFor(property.ArmorRating)
	require.NotEmpty(t, buffed, "item property key %q must resolve directly via the registry", "ArmorRating")
	ip, ok := buffed[0].(property.IntProperty)
	require.True(t, ok, "ArmorRating buff property must be an IntProperty")
	assert.Equal(t, 5, ip.I(), "ArmorRating buff value must survive ingestion")
}
