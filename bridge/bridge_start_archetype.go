package bridge

// @spec-link [[mec_ai_archetype_system]]
// @spec-link [[rule_ai_team_composition_rules]]
// @spec-link [[mechanic_ai_progression_matching]]

import (
	"fmt"
	"math"

	"github.com/ecumeurs/upsilonapi/api"
	"github.com/ecumeurs/upsilonbattle/battlearena/controller/archetype"
	"github.com/ecumeurs/upsilonbattle/battlearena/controller/controllers"
	"github.com/ecumeurs/upsilontypes/entity"
	"github.com/ecumeurs/upsilontypes/entity/grade"
	"github.com/ecumeurs/upsilontypes/entity/skill"
	"github.com/ecumeurs/upsilontypes/property"
	"github.com/ecumeurs/upsilonmapdata/grid/position"
	"github.com/google/uuid"
)

// resolvePlayerGrade returns the canonical grade string for a player, deriving it
// from TotalWins when Grade is not explicitly set.
func resolvePlayerGrade(p api.Player) string {
	if p.Grade != "" && grade.ValidGrade(p.Grade) {
		return p.Grade
	}
	return grade.GradeFromWins(p.TotalWins)
}

// resolveEntityGrade returns the canonical grade for an entity, falling back to the
// player-level grade when the entity does not override it.
func resolveEntityGrade(e api.Entity, playerGrade string) string {
	if e.Grade != "" && grade.ValidGrade(e.Grade) {
		return e.Grade
	}
	return playerGrade
}

// resolveArchetypeSlug returns the resolved archetype slug for an entity, using the
// player-level slug when the entity has no override, and picking randomly (respecting
// team-composition constraints) when neither is specified.
//
// existingTeamSlugs contains the slugs already assigned to other entities in the same team.
func resolveArchetypeSlug(eSlug, pSlug string, existingTeamSlugs []string) string {
	if eSlug != "" {
		if _, ok := archetype.Get(eSlug); ok {
			return eSlug
		}
	}
	if pSlug != "" {
		if _, ok := archetype.Get(pSlug); ok {
			return pSlug
		}
	}
	return archetype.RandomFor(existingTeamSlugs)
}

// validateTeamComposition returns an error if any AI team violates the composition rules:
//   - at most 1 support per team
//   - at most 1 sneak per team
//
// Only IA players with AutoGen entities are checked; explicit-stats entities are skipped.
//
// @spec-link [[rule_ai_team_composition_rules]]
func validateTeamComposition(players []api.Player) error {
	supportCount := make(map[int]int)
	sneakCount := make(map[int]int)
	for _, p := range players {
		if !p.IA {
			continue
		}
		if err := checkTeamEntities(p, supportCount, sneakCount); err != nil {
			return err
		}
	}
	return nil
}

// checkTeamEntities tallies constrained archetypes for one player's entities
// and returns an error on first violation. Only explicit AutoGen slugs are checked.
func checkTeamEntities(p api.Player, supportCount, sneakCount map[int]int) error {
	for _, e := range p.Entities {
		if !e.AutoGen {
			continue
		}
		slug := e.Archetype
		if slug == "" {
			slug = p.Archetype
		}
		if err := checkArchetypeLimit(slug, p.Team, supportCount, sneakCount); err != nil {
			return err
		}
	}
	return nil
}

// checkArchetypeLimit increments the per-team counter for constrained archetypes
// and returns an error if the limit (1 support, 1 sneak) is exceeded.
func checkArchetypeLimit(slug string, team int, supportCount, sneakCount map[int]int) error {
	switch slug {
	case "support":
		supportCount[team]++
		if supportCount[team] > 1 {
			return fmt.Errorf("team %d: at most 1 support archetype allowed, found >1", team)
		}
	case "sneak":
		sneakCount[team]++
		if sneakCount[team] > 1 {
			return fmt.Errorf("team %d: at most 1 sneak archetype allowed, found >1", team)
		}
	}
	return nil
}

// generateEntityFromArchetype creates a fully configured engine entity (stats + skills)
// driven by archetype and grade. Called when Entity.AutoGen is true.
//
// @spec-link [[mechanic_ai_progression_matching]]
func generateEntityFromArchetype(
	entID uuid.UUID,
	name string,
	teamID int,
	controllerID uuid.UUID,
	slug string,
	gradeStr string,
	startPos position.Position,
) (entity.Entity, error) {
	arch, ok := archetype.Get(slug)
	if !ok {
		return entity.Entity{}, fmt.Errorf("unknown archetype: %q", slug)
	}
	gradeIdx := grade.MustGradeIndex(gradeStr)
	cpPool := grade.CPForGrade(gradeStr)

	// Distribute CP according to archetype stat weights.
	w := arch.StatWeights()
	total := w.HP + w.SP + w.MP + w.Attack + w.Defense + w.Movement + w.AttackRange
	if total <= 0 {
		total = 1
	}
	norm := func(wt float64) int { return int(math.Round(float64(cpPool) * wt / total)) }

	hp := 10 + norm(w.HP)*2
	sp := 10 + norm(w.SP)*2
	mp := 10 + norm(w.MP)*2
	atk := 1 + norm(w.Attack)
	def := norm(w.Defense)
	mvt := 3 + int(math.Round(float64(norm(w.Movement))*0.5))
	atkRange := 1 + int(math.Round(float64(norm(w.AttackRange))*0.25))

	if mvt < 1 {
		mvt = 1
	}
	if atkRange < 1 {
		atkRange = 1
	}

	e := entity.Entity{
		ID:           entID,
		Type:         entity.Character,
		Name:         name,
		ControllerID: controllerID,
		Position:     startPos,
	}
	e.Properties = make(map[string]property.Property)
	e.Skills = make(map[uuid.UUID]skill.Skill)

	e.RepsertPropertyCMaxValue(property.HP, hp)
	e.RepsertPropertyCValue(property.HP, hp)
	e.RepsertPropertyCMaxValue(property.SP, sp)
	e.RepsertPropertyCValue(property.SP, sp)
	e.RepsertPropertyCMaxValue(property.MP, mp)
	e.RepsertPropertyCValue(property.MP, mp)
	e.RepsertPropertyCMaxValue(property.Movement, mvt)
	e.RepsertPropertyCValue(property.Movement, mvt)
	e.RepsertPropertyValue(property.Attack, atk)
	e.RepsertPropertyValue(property.Defense, def)
	e.RepsertPropertyValue(property.TeamID, teamID)
	e.RepsertPropertyValue(property.AttackRange, atkRange)
	e.RepsertPropertyValue(property.AIArchetype, slug)

	// Generate skills scaled to grade. More skills at higher grades.
	numSkills := 2 + gradeIdx/3
	for _, sk := range arch.BuildSkillBundle(gradeStr, numSkills) {
		e.RegisterSkill(sk)
	}

	return e, nil
}

// newAIControllerForArchetype creates an AIController configured for the given archetype and grade.
func newAIControllerForArchetype(id uuid.UUID, name string, slug string, gradeStr string) *controllers.AIController {
	arch, ok := archetype.Get(slug)
	if !ok {
		// Fall back to baseline.
		return controllers.NewAggressiveController(id, name)
	}
	ctl := controllers.NewAIController(id, name, arch.Behavior())
	ctl.SetGrade(grade.MustGradeIndex(gradeStr))
	return ctl
}
