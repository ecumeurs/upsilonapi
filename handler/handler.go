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

	ok, msg, data := bridge.Get().ArenaAction(arenaID, req)
	if !ok {
		c.JSON(http.StatusPreconditionFailed, api.NewError(req.RequestID, msg))
		return
	}

	var res interface{}

	switch d := data.(type) {
	case rulermethods.ControllerAttackReply:
		res = api.NewEntity(d.Entity)

	case rulermethods.ControllerMoveReply:
		res = api.NewEntity(d.Entity)

	default:
		// end of turn
		res = stdmessage.DataNil{}
	}

	c.JSON(http.StatusOK, api.NewSuccess(req.RequestID, msg, res))
}

// HandleGetActiveMatchStats returns the number of active matches.
func HandleGetActiveMatchStats(c *gin.Context) {
	count := bridge.Get().GetActiveMatchCount()
	c.JSON(http.StatusOK, api.NewSuccess("", "Active match stats retrieved", api.ActiveMatchStatsResponse{
		ActiveCount: count,
	}))
}
