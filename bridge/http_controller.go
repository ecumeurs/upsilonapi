package bridge

/*
 * @spec-link [[module_upsilonapi]]
 */

import (
	"bytes"
	"encoding/json"
	"net/http"
	"reflect"
	"time"

	"github.com/ecumeurs/upsilonapi/api"
	"github.com/ecumeurs/upsilonapi/stdmessage"
	"github.com/ecumeurs/upsilonbattle/battlearena/controller"
	"github.com/ecumeurs/upsilonbattle/battlearena/ruler/rulermethods"
	"github.com/ecumeurs/upsilontools/tools/actor"
	"github.com/ecumeurs/upsilontools/tools/messagequeue/message"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type HTTPController struct {
	*controller.Controller
	CallbackURL string
	MatchID     uuid.UUID
	Players     []api.Player
	PlayerIDs   []uuid.UUID // all human player IDs this controller represents
}

type webhookContext struct {
	Action    *api.ActionFeedback
	EventName string
}

// NewHTTPController creates a single HTTPController for all human players in a match.
// It registers under multiple player IDs via AddController, keeping one actor/queue per match.
func NewHTTPController(matchID uuid.UUID, callbackURL string, players []api.Player, playerIDs []uuid.UUID) *HTTPController {
	hc := &HTTPController{
		Controller:  controller.NewController(matchID),
		CallbackURL: callbackURL,
		MatchID:     matchID,
		Players:     players,
		PlayerIDs:   playerIDs,
	}

	// Override or add methods to handle Ruler's broadcasts
	hc.AddNotificationHandler(rulermethods.ControllerNextTurn{}, hc.forwardToWebhook, nil)
	hc.AddNotificationHandler(rulermethods.BattleStart{}, hc.BattleStart, nil)
	hc.AddNotificationHandler(rulermethods.BattleEnd{}, hc.forwardToWebhook, nil)
	hc.AddNotificationHandler(rulermethods.EntitiesStateChanged{}, hc.forwardToWebhook, nil)
	hc.AddNotificationHandler(rulermethods.ControllerSkillUsed{}, hc.forwardToWebhook, nil)
	hc.AddNotificationHandler(rulermethods.ControllerAttacked{}, hc.forwardToWebhook, nil)
	hc.AddNotificationHandler(rulermethods.ControllerMoved{}, hc.forwardToWebhook, nil)
	hc.AddNotificationHandler(rulermethods.ControllerPassed{}, hc.forwardToWebhook, nil)

	hc.AddReplyHandler(rulermethods.GetBoardStateReply{}, hc.handleBoardStateReply, nil)

	return hc
}

func (hc *HTTPController) BattleStart(ctx actor.NotificationContext) {
	logrus.Infof("HTTPController %s: BattleStart received, notifying BattleReady for %d players", hc.MatchID, len(hc.PlayerIDs))
	hc.forwardToWebhook(ctx)
	if hc.Ruler != nil {
		for _, pid := range hc.PlayerIDs {
			hc.Ruler.NotifyActor(message.Create(nil, rulermethods.ControllerBattleReady{
				ControllerID: pid,
			}, nil))
		}
	} else {
		logrus.Warnf("HTTPController %s: Ruler is nil in BattleStart", hc.MatchID)
	}
}

func (hc *HTTPController) mapCredits(awards []rulermethods.CreditAward) []api.CreditAward {
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

func (hc *HTTPController) mapResults(results []rulermethods.ActionResult) []api.ActionResult {
	if len(results) == 0 {
		return nil
	}
	res := make([]api.ActionResult, len(results))
	for i, r := range results {
		res[i] = api.ActionResult{
			TargetID: r.TargetID.String(),
			Damage:   r.Damage,
			Heal:     r.Heal,
			PrevHP:   r.PrevHP,
			NewHP:    r.NewHP,
			Credits:  hc.mapCredits(r.CreditAwards),
		}
	}
	return res
}

func (hc *HTTPController) forwardToWebhook(ctx actor.NotificationContext) {
	var action *api.ActionFeedback
	switch d := ctx.Msg.TargetMethod.(type) {
	case rulermethods.ControllerAttacked:
		action = &api.ActionFeedback{
			Type:     "attack",
			ActorID:  d.AttackerControllerID.String(),
			TargetID: d.Entity.ID.String(),
			Results: []api.ActionResult{
				{
					TargetID: d.Entity.ID.String(),
					Damage:   d.Damage,
					PrevHP:   d.PrevHP,
					NewHP:    d.NewHP,
					Credits:  hc.mapCredits(d.CreditAwards),
				},
			},
			Credits: hc.mapCredits(d.CreditAwards),
		}
	case rulermethods.ControllerSkillUsed:
		action = &api.ActionFeedback{
			Type:    "skill",
			ActorID: d.EmitterControllerID.String(),
			Results: []api.ActionResult{
				{
					TargetID: d.Entity.ID.String(),
					Credits:  hc.mapCredits(d.CreditAwards),
				},
			},
			Credits: hc.mapCredits(d.CreditAwards),
		}
	case rulermethods.ControllerMoved:
		action = &api.ActionFeedback{
			Type:    "move",
			ActorID: d.EntityID.String(),
			Path:    d.Path,
		}
	case rulermethods.ControllerPassed:
		action = &api.ActionFeedback{
			Type:    "pass",
			ActorID: d.EntityID.String(),
		}
	default:
		// ISS-057: Log unhandled event types to aid debugging
		logrus.WithFields(logrus.Fields{
			"eventType": hc.getEventName(ctx.Msg.TargetMethod),
			"method":    reflect.TypeOf(ctx.Msg.TargetMethod).String(),
		}).Debug("Forwarding notification with no specific action feedback")
	}

	eventName := hc.getEventName(ctx.Msg.TargetMethod)

	// Extract version from notification if available (v2 versioned notifications)
	var version int64
	switch d := ctx.Msg.TargetMethod.(type) {
	case rulermethods.ControllerAttacked: version = d.Version
	case rulermethods.ControllerMoved: version = d.Version
	case rulermethods.ControllerPassed: version = d.Version
	case rulermethods.EntitiesStateChanged: version = d.Version
	case rulermethods.ControllerNextTurn: version = d.Version
	case rulermethods.BattleStart: version = d.Version
	case rulermethods.BattleEnd: version = d.Version
	case rulermethods.ControllerSkillUsed: version = d.Version
	}

	// This prevents redundant engine calls when multiple controllers receive the same broadcast.
	if version >= 0 && !Get().TrySendWebhook(hc.MatchID, version, eventName) {
		return
	}

	if hc.Ruler == nil {
		logrus.Errorf("HTTPController %s: Ruler is nil, cannot get board state", hc.ID)
		return
	}

	// @spec-link [[api_go_battle_action]]
	// Request safe board state from Ruler
	logrus.Debugf("Requesting board state for %s (%s)", hc.MatchID, eventName)
	hc.Ruler.SendActor(message.Create(hc.Actor, rulermethods.GetBoardState{
		ActionContext: &webhookContext{
			Action:    action,
			EventName: eventName,
		},
	}, rulermethods.GetBoardStateReply{}), hc.CallbackChan)
}


func (hc *HTTPController) handleBoardStateReply(ctx actor.ReplyContext) {
	reply, ok := ctx.Msg.Content.(rulermethods.GetBoardStateReply)
	if !ok {
		logrus.Errorf("HTTPController %s: Received invalid reply type for board state", hc.ID)
		return
	}

	logrus.Debugf("Received board state reply for %s", hc.MatchID)


	wctx, ok := reply.ActionContext.(*webhookContext)
	if !ok {
		// This reply was not triggered by a tactical event requiring a webhook (e.g. manual sync or concurrent bridge call).
		// We skip it for the HTTPController as it only needs to relay tactical broadcasts.
		return
	}

	// Construct board state from safe data
	bs := api.NewBoardState(
		hc.MatchID,
		reply.Grid,
		reply.Entities,
		hc.Players,
		reply.TurnState,
		time.Now(),
		time.Now().Add(30*time.Second),
		reply.WinnerTeamID,
		reply.Version,
		wctx.Action,
	)

	// Note: TrySendWebhook check was moved to forwardToWebhook for optimization, 
	// but we kept version tracking logic in the bridge.

	payload := api.ArenaEvent{
		MatchID:   hc.MatchID.String(),
		EventType: wctx.EventName,
		Data:      bs,
		Version:   bs.Version,
		Timeout:   bs.Timeout,
	}

	// @spec-link [[api_standard_envelope]]
	wrapped := stdmessage.New("Arena Event", true, payload)

	jsonData, err := json.Marshal(wrapped)
	if err != nil {
		logrus.Errorf("Failed to marshal webhook payload: %v", err)
		return
	}

	// @spec-link [[mech_arena_lifecycle]]
	// @spec-link [[mech_webhook_delivery]]
	// Asynchronous webhook delivery (ISS-099) to prevent blocking the actor loop.
	go func() {
		resp, err := http.Post(hc.CallbackURL, "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			logrus.Errorf("Failed to send webhook: %v", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			logrus.Warnf("Webhook returned non-OK status: %d", resp.StatusCode)
		}

		if payload.EventType == "game.ended" {
			logrus.Infof("Battle %s ended, triggering arena destruction", hc.MatchID)
			Get().DestroyArena(hc.MatchID)
		}
	}()
}

func (hc *HTTPController) getEventName(content interface{}) string {
	switch content.(type) {
	case rulermethods.ControllerNextTurn:
		return "turn.started"
	case rulermethods.BattleStart:
		return "game.started"
	case rulermethods.BattleEnd:
		return "game.ended"
	case rulermethods.EntitiesStateChanged:
		return "board.updated"
	case rulermethods.ControllerAttacked:
		return "board.updated"
	case rulermethods.ControllerMoved:
		return "board.updated"
	case rulermethods.ControllerPassed:
		return "board.updated"
	default:
		return "unknown"
	}
}
