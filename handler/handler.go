package handler

import (
	"fmt"
	"net/http"
	"time"

	"github.com/ecumeurs/upsilonapi/api"
	"github.com/ecumeurs/upsilonapi/bridge"
	"github.com/ecumeurs/upsilonapi/stdmessage"
	"github.com/ecumeurs/upsilonbattle/battlearena/ruler/rulermethods"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// @spec-link [[api_go_battle_engine]]

// HandleArenaStart handles the start of a new arena; initializes a new ruler and returns the initial state.
// @spec-link [[api_go_battle_start]]
func HandleArenaStart(c *gin.Context) {
	// Parse and validate the incoming arena start request payload.
	var req api.ArenaStartMessage

	if err := c.ShouldBindJSON(&req); err != nil {
		// Log binding failures for diagnostic visibility.
		fmt.Printf("HandleArenaStart bind error: %v\n", err)
		c.JSON(http.StatusBadRequest, api.NewError("", err.Error()))
		return
	}

	// Delegate arena initialization to the bridge, which orchestrates engine-level setup.
	id, g, entities, players, turner, version, err := bridge.Get().StartArena(req.Data)
	if err != nil {
		// Return engine-side initialization errors to the caller.
		c.JSON(http.StatusBadRequest, api.NewError(req.RequestID, err.Error()))
		return
	}

	// Construct the initial board state snapshot for the response.
	// We use the current server time for lastUpdate and nextUpdate timestamps.
	bs := api.NewBoardState(id, g, entities, players, turner, time.Now(), time.Now().Add(30*time.Second), 0, version, nil)

	// Return successful initialization data including the new arena ID.
	c.JSON(http.StatusOK, api.NewSuccess(req.RequestID, "Arena started", api.ArenaStartResponse{
		ArenaID:      id.String(),
		InitialState: bs,
	}))
}

// HandleArenaAction handles an action in an arena; sends the action to the ruler.
// @spec-link [[api_go_battle_action]]
func HandleArenaAction(c *gin.Context) {
	// Initialize request binding for the polymorphic ArenaActionMessage DTO.
	var req api.ArenaActionMessage
	// Validate JSON body structure and required fields.
	if err := c.ShouldBindJSON(&req); err != nil {
		// Return 400 Bad Request if the payload is malformed.
		c.JSON(http.StatusBadRequest, api.NewError("", err.Error()))
		return
	}

	// Retrieve the arena identifier from the URL path.
	idStr := c.Param("id")
	// Parse the string ID into a formal UUID object.
	arenaID, err := uuid.Parse(idStr)
	if err != nil {
		// Return 400 if the ID format is invalid.
		c.JSON(http.StatusBadRequest, api.NewError(req.RequestID, "invalid arena id"))
		return
	}

	// Forward the validated action to the engine via the bridge.
	// We capture status, human-readable message, error key, and result data.
	ok, msg, errKey, data := bridge.Get().ArenaAction(arenaID, req)
	if !ok {
		// Handle logical failures (e.g., entity out of range) with status 412.
		c.JSON(http.StatusPreconditionFailed, api.NewErrorWithKey(req.RequestID, msg, errKey))
		return
	}

	// Prepare a generic interface to hold the mapped response payload.
	var res interface{}

	// Perform polymorphic mapping based on the engine's reply type.
	switch d := data.(type) {
	case rulermethods.ControllerAttackReply:
		// Map detailed attack results for granular UI feedback.
		results := make([]api.ActionResult, len(d.Results))
		for i, r := range d.Results {
			// Iterate through each affected target and map their state changes.
			results[i] = api.ActionResult{
				TargetID: r.TargetID.String(),
				Damage:   r.Damage,
				PrevHP:   r.PrevHP,
				NewHP:    r.NewHP,
				Credits:  mapCreditsToApi(r.CreditAwards),
			}
		}
		// Wrap the results in a Gin H map for serialization.
		// The frontend uses the attacker state to refresh the primary entity view.
		res = gin.H{
			"attacker": api.NewEntity(d.Attacker),
			"results":  results,
		}

	case rulermethods.ControllerUseSkillReply:
		// Map complex skill execution results including area-of-effect impacts.
		// Skills can affect multiple targets simultaneously with varying effects.
		results := make([]api.ActionResult, len(d.Results))
		for i, r := range d.Results {
			// Calculate and map the delta for both damage and healing components.
			// This covers both offensive and supportive skill outcomes.
			results[i] = api.ActionResult{
				TargetID: r.TargetID.String(),
				Damage:   r.Damage,
				Heal:     r.Heal,
				PrevHP:   r.PrevHP,
				NewHP:    r.NewHP,
				Credits:  mapCreditsToApi(r.CreditAwards),
			}
		}
		// Return the updated attacker state and the list of outcomes.
		// Attacker state update is required for resource cost (MP/AP) visibility.
		res = gin.H{
			"attacker": api.NewEntity(d.Attacker),
			"results":  results,
		}

	case rulermethods.ControllerMoveReply:
		// Map a successful movement action to the updated entity position.
		// This provides immediate feedback on the new coordinates.
		// Coordinates are returned in standard (X, Y) format.
		res = gin.H{
			"entity": api.NewEntity(d.Entity),
		}

	default:
		// Default to an empty data object for actions that return no specific data.
		// This ensures a consistent JSON structure even for generic successes.
		res = stdmessage.DataNil{}
	}

	// Finalize the response with an OK status and the mapped data.
	// We use the original request ID to facilitate client-side tracking.
	c.JSON(http.StatusOK, api.NewSuccess(req.RequestID, msg, res))
}

// HandleArenaForfeit handles a player conceding the match.
// @spec-link [[api_go_battle_forfeit]]
func HandleArenaForfeit(c *gin.Context) {
	// Parse the forfeit intent from the request body.
	var req api.ArenaForfeitMessage
	// Perform JSON binding.
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, api.NewError("", err.Error()))
		return
	}

	// Extract the arena ID from the route.
	idStr := c.Param("id")
	// Parse UUID.
	arenaID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, api.NewError(req.RequestID, "invalid arena id"))
		return
	}

	// Identify which player is forfeiting.
	playerID, err := uuid.Parse(req.Data.PlayerID)
	if err != nil {
		// Handle invalid player UUIDs.
		c.JSON(http.StatusBadRequest, api.NewError(req.RequestID, "invalid player id"))
		return
	}

	// Execute the forfeit logic in the engine.
	// This marks the match as resolved for the relevant team.
	ok, msg, errKey, _ := bridge.Get().ArenaForfeit(arenaID, playerID)
	if !ok {
		// Return failure if the action violates current match rules.
		c.JSON(http.StatusPreconditionFailed, api.NewErrorWithKey(req.RequestID, msg, errKey))
		return
	}

	// Confirm receipt and processing to the management layer.
	c.JSON(http.StatusOK, api.NewSuccess(req.RequestID, "Forfeit accepted", stdmessage.DataNil{}))
}

// HandleGetActiveMatchStats returns the number of active matches.
// @spec-link [[api_go_health_check]]
func HandleGetActiveMatchStats(c *gin.Context) {
	// Query the global match registry for the current count.
	count := bridge.Get().GetActiveMatchCount()
	// Return the count in a structured JSON response.
	c.JSON(http.StatusOK, api.NewSuccess("", "Active match stats retrieved", api.ActiveMatchStatsResponse{
		ActiveCount: count,
	}))
}

// HandleArenaResurrect rebuilds a crashed arena from a persisted board state.
// Called by Laravel when it detects the engine lost in-memory state for an active match.
// @spec-link [[api_go_battle_engine]]
func HandleArenaResurrect(c *gin.Context) {
	// Get target match ID from URL.
	idStr := c.Param("id")
	matchID, err := uuid.Parse(idStr)
	if err != nil {
		// Abort if ID is invalid.
		c.JSON(http.StatusBadRequest, api.NewError("", "invalid arena id"))
		return
	}

	// Bind the full state dump from the request body.
	var req api.ArenaResurrectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// Return error on malformed state.
		c.JSON(http.StatusBadRequest, api.NewError("", err.Error()))
		return
	}
	// Inject the URL match ID to ensure consistency.
	req.MatchID = matchID.String()

	// Instruct the bridge to hydrate the engine with the provided state.
	// This is a critical recovery path for high-availability matches.
	bs, err := bridge.Get().ResurrectArena(req)
	if err != nil {
		// Fail if hydration fails.
		c.JSON(http.StatusBadRequest, api.NewError("", err.Error()))
		return
	}

	// Acknowledge restoration with the current engine state view.
	// The client can use this state to resume UI rendering.
	c.JSON(http.StatusOK, api.NewSuccess("", "Arena resurrected", api.ArenaStartResponse{
		ArenaID:      matchID.String(),
		InitialState: bs,
	}))
}

// HandleArenaExists checks if an arena exists.
// @spec-link [[api_arena_existence_check]]
func HandleArenaExists(c *gin.Context) {
	// Parse the probe target.
	idStr := c.Param("id")
	arenaID, err := uuid.Parse(idStr)
	if err != nil {
		// Return 400 for bad ID.
		c.JSON(http.StatusBadRequest, api.NewError("", "invalid arena id"))
		return
	}

	// Verify presence in the bridge's local registry.
	_, exists := bridge.Get().GetArena(arenaID)

	// Return boolean flag.
	c.JSON(http.StatusOK, api.NewSuccess("", "Existence check complete", api.ArenaExistsResponse{
		Exists: exists,
	}))
}

// mapCreditsToApi converts engine credit awards into API credit awards.
func mapCreditsToApi(awards []rulermethods.CreditAward) []api.CreditAward {
	// Optimize for no-award cases.
	// This happens when an action doesn't trigger any achievements.
	if len(awards) == 0 {
		return nil
	}
	// Allocate a result slice of the correct size.
	res := make([]api.CreditAward, len(awards))
	// Convert each internal award into an API-serializable DTO.
	for i, a := range awards {
		// Translate engine-specific fields to API-standard types.
		res[i] = api.CreditAward{
			PlayerID: a.PlayerID.String(),
			Amount:   a.Amount,
			Source:   a.Source,
		}
	}
	// Return the populated slice to the caller.
	return res
}
