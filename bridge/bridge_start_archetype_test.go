package bridge

// @test-link [[mec_ai_archetype_system]]
// @test-link [[rule_ai_team_composition_rules]]
// @test-link [[mechanic_ai_progression_matching]]

import (
	"testing"

	"github.com/ecumeurs/upsilonapi/api"
	"github.com/ecumeurs/upsilontypes/property"
	"github.com/ecumeurs/upsilonmapdata/grid/position"
	"github.com/google/uuid"
)

// ── validateTeamComposition ───────────────────────────────────────────────────

func TestValidateTeamCompositionAllowsOneSupportPerTeam(t *testing.T) {
	players := []api.Player{
		{ID: uuid.New().String(), Team: 2, IA: true, Entities: []api.Entity{{ID: uuid.New().String(), Name: "A", AutoGen: true, Archetype: "support"}}},
		{ID: uuid.New().String(), Team: 2, IA: true, Entities: []api.Entity{{ID: uuid.New().String(), Name: "B", AutoGen: true, Archetype: "fighter"}}},
	}
	if err := validateTeamComposition(players); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestValidateTeamCompositionRejectsTwoSupports verifies that two support archetypes on the same team are rejected.
func TestValidateTeamCompositionRejectsTwoSupports(t *testing.T) {
	players := []api.Player{
		{ID: uuid.New().String(), Team: 2, IA: true, Entities: []api.Entity{{ID: uuid.New().String(), Name: "A", AutoGen: true, Archetype: "support"}}},
		{ID: uuid.New().String(), Team: 2, IA: true, Entities: []api.Entity{{ID: uuid.New().String(), Name: "B", AutoGen: true, Archetype: "support"}}},
	}
	if err := validateTeamComposition(players); err == nil {
		t.Error("expected error for 2 supports on same team")
	}
}

// TestValidateTeamCompositionRejectsTwoSneaks verifies that two sneak archetypes on the same team are rejected.
func TestValidateTeamCompositionRejectsTwoSneaks(t *testing.T) {
	players := []api.Player{
		{ID: uuid.New().String(), Team: 2, IA: true, Entities: []api.Entity{{ID: uuid.New().String(), Name: "A", AutoGen: true, Archetype: "sneak"}}},
		{ID: uuid.New().String(), Team: 2, IA: true, Entities: []api.Entity{{ID: uuid.New().String(), Name: "B", AutoGen: true, Archetype: "sneak"}}},
	}
	if err := validateTeamComposition(players); err == nil {
		t.Error("expected error for 2 sneaks on same team")
	}
}

// TestValidateTeamCompositionDifferentTeamsAllowed verifies that the same constrained archetype is allowed on different teams.
func TestValidateTeamCompositionDifferentTeamsAllowed(t *testing.T) {
	players := []api.Player{
		{ID: uuid.New().String(), Team: 1, IA: true, Entities: []api.Entity{{ID: uuid.New().String(), Name: "A", AutoGen: true, Archetype: "support"}}},
		{ID: uuid.New().String(), Team: 2, IA: true, Entities: []api.Entity{{ID: uuid.New().String(), Name: "B", AutoGen: true, Archetype: "support"}}},
	}
	if err := validateTeamComposition(players); err != nil {
		t.Errorf("unexpected error for supports on separate teams: %v", err)
	}
}

// ── generateEntityFromArchetype ───────────────────────────────────────────────

func TestGenerateEntityFromArchetypeProducesValidEntity(t *testing.T) {
	for _, slug := range []string{"fighter", "ranger", "support", "sneak"} {
		e, err := generateEntityFromArchetype(uuid.New(), "TestBot", 2, uuid.New(), slug, "I", position.Position{})
		if err != nil {
			t.Errorf("%s Grade I: unexpected error: %v", slug, err)
			continue
		}
		if e.GetPropertyI(property.HP).I() <= 0 {
			t.Errorf("%s: HP must be > 0", slug)
		}
		if e.GetPropertyI(property.Attack).I() <= 0 {
			t.Errorf("%s: Attack must be > 0", slug)
		}
		if e.GetPropertyI(property.TeamID).I() != 2 {
			t.Errorf("%s: TeamID mismatch", slug)
		}
	}
}

// TestGenerateEntityGradeVHasHigherStatsThanGradeI verifies that grade-V entities have strictly higher HP than grade-I entities.
func TestGenerateEntityGradeVHasHigherStatsThanGradeI(t *testing.T) {
	entI, _ := generateEntityFromArchetype(uuid.New(), "Bot", 1, uuid.New(), "fighter", "I", position.Position{})
	entV, _ := generateEntityFromArchetype(uuid.New(), "Bot", 1, uuid.New(), "fighter", "V", position.Position{})

	hpI := entI.GetPropertyI(property.HP).I()
	hpV := entV.GetPropertyI(property.HP).I()
	if hpV <= hpI {
		t.Errorf("Grade V HP (%d) should be > Grade I HP (%d)", hpV, hpI)
	}
}

// TestGenerateEntitySetsAIArchetypeProperty verifies that the AIArchetype property is populated on auto-generated entities.
func TestGenerateEntitySetsAIArchetypeProperty(t *testing.T) {
	e, _ := generateEntityFromArchetype(uuid.New(), "Bot", 1, uuid.New(), "sneak", "II", position.Position{})
	if e.Properties[string(property.AIArchetype)] == nil {
		t.Error("AIArchetype property not set")
	}
}

// TestGenerateEntitySkillsScaleWithGrade verifies that higher-grade entities receive at least as many skills as lower-grade entities.
func TestGenerateEntitySkillsScaleWithGrade(t *testing.T) {
	entI, _ := generateEntityFromArchetype(uuid.New(), "Bot", 1, uuid.New(), "ranger", "I", position.Position{})
	entV, _ := generateEntityFromArchetype(uuid.New(), "Bot", 1, uuid.New(), "ranger", "V", position.Position{})

	if len(entV.Skills) < len(entI.Skills) {
		t.Errorf("Grade V should have >= skills as Grade I (got V=%d, I=%d)", len(entV.Skills), len(entI.Skills))
	}
}

// ── bridge integration: AutoGen entity startup ────────────────────────────────

// TestStartArenaWithAutoGenEntity verifies that StartArena correctly generates an entity from archetype+grade when AutoGen is true.
func TestStartArenaWithAutoGenEntity(t *testing.T) {
	matchID := uuid.New()
	req := buildAutoGenArenaRequest(matchID.String(), "fighter", 10)

	_, _, entities, _, _, _, err := b.StartArena(req)
	if err != nil {
		t.Fatalf("StartArena failed: %v", err)
	}
	if len(entities) == 0 {
		t.Fatal("no entities returned")
	}
	b.DestroyArena(matchID)
}

// TestStartArenaTeamCompViolationReturnsError verifies that StartArena rejects a request with two sneaks on the same team.
func TestStartArenaTeamCompViolationReturnsError(t *testing.T) {
	req := api.ArenaStartRequest{
		MatchID:     uuid.New().String(),
		CallbackURL: "http://localhost/webhook",
		Players: []api.Player{
			{
				ID: uuid.New().String(), Team: 2, IA: true,
				Entities: []api.Entity{{ID: uuid.New().String(), Name: "A", AutoGen: true, Archetype: "sneak"}},
			},
			{
				ID: uuid.New().String(), Team: 2, IA: true,
				Entities: []api.Entity{{ID: uuid.New().String(), Name: "B", AutoGen: true, Archetype: "sneak"}},
			},
			{
				ID:       uuid.New().String(), Team: 1, IA: false,
				Entities: []api.Entity{createTestEntity()},
			},
		},
	}
	_, _, _, _, _, _, err := b.StartArena(req)
	if err == nil {
		t.Error("expected team-composition error for 2 sneaks on same team")
	}
}

// buildAutoGenArenaRequest builds a minimal ArenaStartRequest with one human player and one AI auto-gen player.
func buildAutoGenArenaRequest(matchID, archetype string, totalWins int) api.ArenaStartRequest {
	autoGenEntity := api.Entity{ID: uuid.New().String(), Name: "Bot", AutoGen: true}
	aiPlayer := api.Player{
		ID:        uuid.New().String(),
		Team:      2,
		IA:        true,
		Archetype: archetype,
		TotalWins: totalWins,
		Entities:  []api.Entity{autoGenEntity},
	}
	humanPlayer := api.Player{
		ID:       uuid.New().String(),
		Team:     1,
		IA:       false,
		Entities: []api.Entity{createTestEntity()},
	}
	return api.ArenaStartRequest{
		MatchID:     matchID,
		CallbackURL: "http://localhost/webhook",
		Players:     []api.Player{humanPlayer, aiPlayer},
	}
}
