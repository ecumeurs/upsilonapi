package handler

import (
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
func HandleArenaStart(c *gin.Context) {
	var req api.ArenaStartMessage

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, api.NewError("", err.Error()))
		return
	}

	id, g, entities, players, turner, version, err := bridge.Get().StartArena(req.Data)
	if err != nil {
		c.JSON(http.StatusBadRequest, api.NewError(req.RequestID, err.Error()))
		return
	}

	bs := api.NewBoardState(id, g, entities, players, turner, time.Now(), time.Now().Add(30*time.Second), 0, version, nil)

	c.JSON(http.StatusOK, api.NewSuccess(req.RequestID, "Arena started", api.ArenaStartResponse{
		ArenaID:      id.String(),
		InitialState: bs,
	}))
}

// HandleArenaAction handles an action in an arena; sends the action to the ruler.
func HandleArenaAction(c *gin.Context) {
	// extract StandardMessage first .

	var req api.ArenaActionMessage
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, api.NewError("", err.Error()))
		return
	}
	idStr := c.Param("id")
	arenaID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, api.NewError(req.RequestID, "invalid arena id"))
		return
	}

	ok, msg, errKey, data := bridge.Get().ArenaAction(arenaID, req)
	if !ok {
		c.JSON(http.StatusPreconditionFailed, api.NewErrorWithKey(req.RequestID, msg, errKey))
		return
	}

	var res interface{}

	switch d := data.(type) {
	case rulermethods.ControllerAttackReply:
		// Map detailed results for synchronous feedback
		results := make([]api.ActionResult, len(d.Results))
		for i, r := range d.Results {
			results[i] = api.ActionResult{
				TargetID: r.TargetID.String(),
				Damage:   r.Damage,
				PrevHP:   r.PrevHP,
				NewHP:    r.NewHP,
				Credits:  mapCreditsToApi(r.CreditAwards),
			}
		}
		res = gin.H{
			"attacker": api.NewEntity(d.Attacker),
			"results":  results,
		}

	case rulermethods.ControllerUseSkillReply:
		// Map detailed results for synchronous feedback
		results := make([]api.ActionResult, len(d.Results))
		for i, r := range d.Results {
			results[i] = api.ActionResult{
				TargetID: r.TargetID.String(),
				Damage:   r.Damage,
				Heal:     r.Heal,
				PrevHP:   r.PrevHP,
				NewHP:    r.NewHP,
				Credits:  mapCreditsToApi(r.CreditAwards),
			}
		}
		res = gin.H{
			"attacker": api.NewEntity(d.Attacker),
			"results":  results,
		}

	case rulermethods.ControllerMoveReply:
		res = gin.H{
			"entity": api.NewEntity(d.Entity),
		}

	default:
		// end of turn
		res = stdmessage.DataNil{}
	}

	c.JSON(http.StatusOK, api.NewSuccess(req.RequestID, msg, res))
}

// HandleArenaForfeit handles a player conceding the match.
// @spec-link [[api_go_battle_forfeit]]
func HandleArenaForfeit(c *gin.Context) {
	var req api.ArenaForfeitMessage
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, api.NewError("", err.Error()))
		return
	}

	idStr := c.Param("id")
	arenaID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, api.NewError(req.RequestID, "invalid arena id"))
		return
	}

	playerID, err := uuid.Parse(req.Data.PlayerID)
	if err != nil {
		c.JSON(http.StatusBadRequest, api.NewError(req.RequestID, "invalid player id"))
		return
	}

	ok, msg, errKey, _ := bridge.Get().ArenaForfeit(arenaID, playerID)
	if !ok {
		c.JSON(http.StatusPreconditionFailed, api.NewErrorWithKey(req.RequestID, msg, errKey))
		return
	}

	c.JSON(http.StatusOK, api.NewSuccess(req.RequestID, "Forfeit accepted", stdmessage.DataNil{}))
}

// HandleGetActiveMatchStats returns the number of active matches.
func HandleGetActiveMatchStats(c *gin.Context) {
	count := bridge.Get().GetActiveMatchCount()
	c.JSON(http.StatusOK, api.NewSuccess("", "Active match stats retrieved", api.ActiveMatchStatsResponse{
		ActiveCount: count,
	}))
}

func mapCreditsToApi(awards []rulermethods.CreditAward) []api.CreditAward {
	if len(awards) == 0 {
		return nil
	}
	res := make([]api.CreditAward, len(awards))
	for i, a := range awards {
		res[i] = api.CreditAward{
			PlayerID: a.PlayerID.String(),
			Amount:   a.Amount,
			Source:   a.Source,
		}
	}
	return res
}
