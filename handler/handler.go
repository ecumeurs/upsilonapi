// Package handler provides the HTTP request handlers for the Upsilon Hub API.
// It acts as the primary entry point for external match orchestration and tactical inputs.
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

// HandleArenaStart handles the start of a new arena; initializes a new ruler and returns the initial state.
// @spec-link [[api_go_battle_start]]
func HandleArenaStart(c *gin.Context) {
	// 1. Request Validation: Parse the incoming arena start request payload into the message envelope.
	var req api.ArenaStartMessage
	if err := c.ShouldBindJSON(&req); err != nil {
		fmt.Printf("HandleArenaStart bind error: %v\n", err)
		c.JSON(http.StatusBadRequest, api.NewError("", err.Error()))
		return
	}
	// 2. Engine Initialization: Delegate setup to the bridge, which orchestrates the battle engine startup.
	id, g, entities, players, turner, version, err := bridge.Get().StartArena(req.Data)
	if err != nil {
		c.JSON(http.StatusBadRequest, api.NewError(req.RequestID, err.Error()))
		return
	}
	// 3. State Snapshot: Construct the initial board state snapshot for the response payload.
	bs := api.NewBoardState(id, g, entities, players, turner, time.Now(), time.Now().Add(30*time.Second), 0, version, nil)
	// 4. Response Dispatch: Return successful initialization data including the new arena ID and state.
	c.JSON(http.StatusOK, api.NewSuccess(req.RequestID, "Arena started", api.ArenaStartResponse{
		ArenaID: id.String(), InitialState: bs,
	}))
}

// HandleArenaAction handles tactical inputs (move, attack, skill) and forwards them to the engine.
// @spec-link [[api_go_battle_action]]
func HandleArenaAction(c *gin.Context) {
	// 1. Input Processing: Bind the polymorphic ArenaActionMessage DTO from the request body.
	var req api.ArenaActionMessage
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, api.NewError("", err.Error()))
		return
	}
	// 2. Identity Verification: Retrieve and parse the arena identifier from the URL path parameter.
	idStr := c.Param("id")
	arenaID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, api.NewError(req.RequestID, "invalid arena id"))
		return
	}
	// 3. Execution: Forward the validated action to the engine via the bridge for processing.
	ok, msg, errKey, data := bridge.Get().ArenaAction(arenaID, req)
	if !ok {
		c.JSON(http.StatusPreconditionFailed, api.NewErrorWithKey(req.RequestID, msg, errKey))
		return
	}
	// 4. Response Mapping: Transform engine-specific results into the API's standard DTO format.
	res := mapActionReplyToApi(data)
	// 5. Response Dispatch: Finalize with an OK status and the original request ID for tracking.
	c.JSON(http.StatusOK, api.NewSuccess(req.RequestID, msg, res))
}

// mapActionReplyToApi translates various engine reply types into serializable API responses.
func mapActionReplyToApi(data interface{}) interface{} {
	// 1. Type Switch: Identify the specific engine reply type to map the corresponding payload.
	switch d := data.(type) {
	case rulermethods.ControllerAttackReply:
		// 2. Damage Outcome: Convert the list of affected targets and their health changes.
		results := make([]api.ActionResult, len(d.Results))
		for i, r := range d.Results {
			results[i] = api.ActionResult{
				TargetID: r.TargetID.String(), Damage: r.Damage, PrevHP: r.PrevHP, NewHP: r.NewHP, Credits: mapCreditsToApi(r.CreditAwards),
			}
		}
		return gin.H{"attacker": api.NewEntity(d.Attacker), "results": results}
	case rulermethods.ControllerUseSkillReply:
		// 3. Skill Outcome: Handle complex area-of-effect impacts (Damage and Healing).
		results := make([]api.ActionResult, len(d.Results))
		for i, r := range d.Results {
			results[i] = api.ActionResult{
				TargetID: r.TargetID.String(), Damage: r.Damage, Heal: r.Heal, PrevHP: r.PrevHP, NewHP: r.NewHP, Credits: mapCreditsToApi(r.CreditAwards),
			}
		}
		return gin.H{"attacker": api.NewEntity(d.Attacker), "results": results}
	case rulermethods.ControllerMoveReply:
		// 4. Movement: Provide the updated coordinates and status of the moved entity.
		return gin.H{"entity": api.NewEntity(d.Entity)}
	default:
		// 5. Default: Return an empty structure for actions without explicit results.
		return stdmessage.DataNil{}
	}
}

// HandleArenaForfeit handles a player conceding the match.
// @spec-link [[api_go_battle_forfeit]]
func HandleArenaForfeit(c *gin.Context) {
	// 1. Input Processing: Parse the forfeit intent and player ID from the message body.
	var req api.ArenaForfeitMessage
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, api.NewError("", err.Error()))
		return
	}
	// 2. Route Parsing: Extract the match and player identifiers from parameters and data.
	arenaID, _ := uuid.Parse(c.Param("id"))
	playerID, err := uuid.Parse(req.Data.PlayerID)
	if err != nil {
		c.JSON(http.StatusBadRequest, api.NewError(req.RequestID, "invalid player id"))
		return
	}
	// 3. Execution: Execute the forfeit logic in the engine to resolve the match state.
	ok, msg, errKey, _ := bridge.Get().ArenaForfeit(arenaID, playerID)
	if !ok {
		c.JSON(http.StatusPreconditionFailed, api.NewErrorWithKey(req.RequestID, msg, errKey))
		return
	}
	// 4. Response: Confirm the forfeit acceptance to the management layer.
	c.JSON(http.StatusOK, api.NewSuccess(req.RequestID, "Forfeit accepted", stdmessage.DataNil{}))
}

// HandleGetActiveMatchStats returns quantitative metrics about current engine load.
func HandleGetActiveMatchStats(c *gin.Context) {
	// 1. Telemetry: Query the bridge for the count of active matches in memory.
	count := bridge.Get().GetActiveMatchCount()
	// 2. Response: Return the count in a structured success envelope.
	c.JSON(http.StatusOK, api.NewSuccess("", "Active match stats retrieved", api.ActiveMatchStatsResponse{
		ActiveCount: count,
	}))
}

// HandleArenaResurrect rebuilds a crashed arena from a persisted state dump.
func HandleArenaResurrect(c *gin.Context) {
	// 1. Request Parsing: Validate the match ID and the enveloped state body,
	// like every other arena endpoint ([[api_standard_envelope]]).
	matchID, _ := uuid.Parse(c.Param("id"))
	var req api.ArenaResurrectMessage
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, api.NewError("", err.Error()))
		return
	}
	req.Data.MatchID = matchID.String()
	// 2. State Hydration: Reconstruct the engine's in-memory structures from the provided DTO.
	bs, err := bridge.Get().ResurrectArena(req.Data)
	if err != nil {
		c.JSON(http.StatusBadRequest, api.NewError(req.RequestID, err.Error()))
		return
	}
	// 3. Response: Acknowledge restoration with the current engine state view for resumption.
	c.JSON(http.StatusOK, api.NewSuccess(req.RequestID, "Arena resurrected", api.ArenaStartResponse{
		ArenaID: matchID.String(), InitialState: bs,
	}))
}

// HandleArenaExists checks if a specific match is currently being simulated.
func HandleArenaExists(c *gin.Context) {
	// 1. Discovery: Verify the presence of the arena ID in the local bridge registry.
	arenaID, _ := uuid.Parse(c.Param("id"))
	_, exists := bridge.Get().GetArena(arenaID)
	// 2. Response: Return the existence flag to the caller.
	c.JSON(http.StatusOK, api.NewSuccess("", "Existence check complete", api.ArenaExistsResponse{
		Exists: exists,
	}))
}

// mapCreditsToApi converts engine-side achievement awards into API-serializable DTOs.
func mapCreditsToApi(awards []rulermethods.CreditAward) []api.CreditAward {
	// 1. Optimization: Return nil if the input slice is empty.
	if len(awards) == 0 { return nil }
	// 2. Transformation: Map each award into the standard CreditAward DTO format.
	res := make([]api.CreditAward, len(awards))
	for i, a := range awards {
		res[i] = api.CreditAward{
			PlayerID: a.PlayerID.String(), Amount: a.Amount, Source: a.Source,
		}
	}
	return res
}
