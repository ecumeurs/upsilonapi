package bridge

/*
 * @spec-link [[module_upsilonapi]]
 */

import (
	"bytes"
	"encoding/json"
	"net/http"

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
}

func NewHTTPController(id uuid.UUID, matchID uuid.UUID, callbackURL string) *HTTPController {
	hc := &HTTPController{
		Controller:  controller.NewController(id),
		CallbackURL: callbackURL,
		MatchID:     matchID,
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

	return hc
}

func (hc *HTTPController) BattleStart(ctx actor.NotificationContext) {
	logrus.Infof("HTTPController %s: BattleStart received, notifying BattleReady", hc.ID)
	hc.forwardToWebhook(ctx)
	if hc.Ruler != nil {
		hc.Ruler.NotifyActor(message.Create(nil, rulermethods.ControllerBattleReady{
			ControllerID: hc.ID,
		}, nil))
	} else {
		logrus.Warnf("HTTPController %s: Ruler is nil in BattleStart", hc.ID)
	}
}

func (hc *HTTPController) forwardToWebhook(ctx actor.NotificationContext) {
	var action *api.ActionFeedback
	switch d := ctx.Msg.TargetMethod.(type) {
	case rulermethods.ControllerAttacked:
		action = &api.ActionFeedback{
			Type:     "attack",
			ActorID:  d.Attacker.ID.String(),
			TargetID: d.Entity.ID.String(),
			Damage:   d.Damage,
			PrevHP:   d.PrevHP,
			NewHP:    d.NewHP,
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
	}

	bs, err := Get().GetBoardState(hc.MatchID, action)
	if err != nil {
		logrus.Errorf("Failed to get board state for webhook: %v", err)
		return
	}

	eventName := hc.getEventName(ctx.Msg.TargetMethod)

	// @spec-link [[mech_game_state_versioning]]
	if !Get().TrySendWebhook(hc.MatchID, bs.Version, eventName) {
		// Version and event type already sent by another controller belonging to the same match.
		return
	}

	payload := api.ArenaEvent{
		MatchID:   hc.MatchID.String(),
		EventType: eventName,
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

	resp, err := http.Post(hc.CallbackURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		logrus.Errorf("Failed to send webhook: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logrus.Warnf("Webhook returned non-OK status: %d", resp.StatusCode)
	}

	// @spec-link [[mech_arena_lifecycle]]
	if payload.EventType == "game.ended" {
		logrus.Infof("Battle %s ended, triggering arena destruction", hc.MatchID)
		Get().DestroyArena(hc.MatchID)
	}
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
		return "board.updated" // or "attacked"? user says board.update standardized for front usage
	case rulermethods.ControllerMoved:
		return "board.updated"
	case rulermethods.ControllerPassed:
		return "board.updated"
	default:
		return "unknown"
	}
}
