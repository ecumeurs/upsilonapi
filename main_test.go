package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ecumeurs/upsilonapi/api"
	"github.com/ecumeurs/upsilonapi/handler"
	"github.com/ecumeurs/upsilonapi/stdmessage"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func setupRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	internal := r.Group("/internal")
	{
		internal.POST("/arena/start", handler.HandleArenaStart)
		internal.POST("/arena/:id/action", handler.HandleArenaAction)
	}
	return r
}

// @spec-link [[api_go_battle_engine]]
func TestArenaStartEndpoint(t *testing.T) {
	router := setupRouter()

	// Setup mock webhook receiver
	webhookEvents := make(chan map[string]interface{}, 10)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var event map[string]interface{}
		json.Unmarshal(body, &event)
		webhookEvents <- event
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()
	defer close(webhookEvents)

	id := uuid.New().String()
	mid := uuid.New().String()
	w := httptest.NewRecorder()
	players := []api.Player{
		api.Player{
			ID:   uuid.NewString(),
			Team: 1,
			Entities: []api.Entity{
				api.Entity{
					ID:       uuid.NewString(),
					PlayerID: "",
					Name:     "P1E1",
					HP:       10,
					Attack:   3,
					Defense:  1,
					MaxHP:    10,
					Move:     2,
					MaxMove:  2,
					Position: api.Position{ // note this position is fully arbitrary as it will be rolled by ruler.
						X: 0,
						Y: 5}}},
			IA: false, // Must be false to trigger HTTPController webhooks
		},
		api.Player{
			ID:   uuid.NewString(),
			Team: 2,
			Entities: []api.Entity{
				api.Entity{
					ID:       uuid.NewString(),
					PlayerID: "",
					Name:     "P2E1",
					HP:       10,
					Attack:   3,
					Defense:  1,
					MaxHP:    10,
					Move:     2,
					MaxMove:  2,
					Position: api.Position{ // note this position is fully arbitrary...
						X: 5,
						Y: 0}}},
			IA: true}}

	reqBody, _ := json.Marshal(api.ArenaStartMessage{
		RequestID: id,
		Message:   "Start",
		Success:   true,
		Data: api.ArenaStartRequest{
			MatchID:     mid,
			CallbackURL: ts.URL, // Use mock server URL
			Players:     players,
		},
		Meta: stdmessage.MetaNil{},
	})
	req, _ := http.NewRequest("POST", "/internal/arena/start", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp api.ArenaStartResponseMessage
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	log.Printf("Json Response: %s", w.Body.Bytes())

	assert.NoError(t, err)
	assert.Equal(t, resp.RequestID, id)
	assert.Equal(t, resp.Success, true)
	assert.NotEmpty(t, resp.Data.ArenaID)
	assert.NotEmpty(t, resp.Data.InitialState)

	// Verify webhooks
	expectedEvents := map[string]bool{
		"game.started": false,
		"turn.started": false,
	}

	for range 2 {
		select {
		case event := <-webhookEvents:
			data, ok := event["data"].(map[string]interface{})
			if ok {
				eventType, ok := data["event_type"].(string)
				if ok {
					if _, exists := expectedEvents[eventType]; exists {
						expectedEvents[eventType] = true
					}
				}
			}
		case <-time.After(2 * time.Second):
			t.Errorf("Timed out waiting for webhook event")
		}
	}

	assert.True(t, expectedEvents["game.started"], "Should have received game.started event")
	assert.True(t, expectedEvents["turn.started"], "Should have received turn.started event")
}

// @spec-link [[api_go_battle_engine]]
// @spec-link [[us_take_combat_turn]]
func TestBattleFullRoundtrip(t *testing.T) {
	router := setupRouter()

	// Setup mock webhook receiver
	webhookEvents := make(chan map[string]interface{}, 20)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var event map[string]interface{}
		json.Unmarshal(body, &event)
		log.Printf("Webhook received FULL: %s", string(body))
		webhookEvents <- event
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()
	defer close(webhookEvents)

	id := uuid.New().String()
	mid := uuid.New().String()
	players := []api.Player{
		{
			ID:   uuid.NewString(), // P1
			Team: 1,
			Entities: []api.Entity{
				{
					ID:      uuid.NewString(),
					Name:    "P1E1",
					HP:      10,
					Attack:  3,
					Defense: 1,
					MaxHP:   10,
					Move:    2,
					MaxMove: 2,
					Position: api.Position{
						X: 0,
						Y: 0}}},
			IA: false,
		},
		{
			ID:   uuid.NewString(), // P2
			Team: 2,
			Entities: []api.Entity{
				{
					ID:      uuid.NewString(),
					Name:    "P2E1",
					HP:      10,
					Attack:  3,
					Defense: 1,
					MaxHP:   10,
					Move:    2,
					MaxMove: 2,
					Position: api.Position{
						X: 1,
						Y: 1}}},
			IA: true}}

	// 1. Start Arena
	reqBody, _ := json.Marshal(api.ArenaStartMessage{
		RequestID: id,
		Message:   "Start",
		Success:   true,
		Data: api.ArenaStartRequest{
			MatchID:     mid,
			CallbackURL: ts.URL,
			Players:     players,
		},
		Meta: stdmessage.MetaNil{},
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/internal/arena/start", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var startResp api.ArenaStartResponseMessage
	json.Unmarshal(w.Body.Bytes(), &startResp)
	arenaID := startResp.Data.ArenaID
	p1ID := players[0].ID
	p1e1ID := players[0].Entities[0].ID
	
	initialStateMarshaled, _ := json.MarshalIndent(startResp.Data.InitialState, "", "  ")
	log.Printf("GO TEST INITIAL STATE:\n%s", string(initialStateMarshaled))

	// Wait for game.started and turn.started
	waitForWebhook(t, webhookEvents, "game.started")
	waitForWebhook(t, webhookEvents, "turn.started")

	// 2. Discover actual positions
	var p1e1Pos api.Position
	var p2e1Pos api.Position
	
	allEntities := []api.Entity{}
	for _, p := range startResp.Data.InitialState.Players {
		allEntities = append(allEntities, p.Entities...)
	}

	for _, e := range allEntities {
		if e.ID == p1e1ID {
			p1e1Pos = e.Position
		}
		if e.ID == players[1].Entities[0].ID {
			p2e1Pos = e.Position
		}
	}
	log.Printf("P1E1 actual pos: %+v, P2E1 actual pos: %+v", p1e1Pos, p2e1Pos)

	// 3. Move P1E1 to an adjacent tile (e.g., X+1)
	targetMove := api.Position{X: p1e1Pos.X + 1, Y: p1e1Pos.Y}
	if targetMove.X >= 10 { // boundary check for the 10x10 map used in tests
		targetMove.X = p1e1Pos.X - 1
	}

	log.Printf("Executing MOVE action to %+v...", targetMove)
	moveReqBody, _ := json.Marshal(api.ArenaActionMessage{
		RequestID: uuid.NewString(),
		Data: api.ArenaActionRequest{
			PlayerID: p1ID,
			EntityID: p1e1ID,
			Type:     "move",
			TargetCoords: []api.Position{
				targetMove,
			},
		},
	})
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/internal/arena/"+arenaID+"/action", bytes.NewBuffer(moveReqBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	log.Printf("MOVE status: %d, response: %s", w.Code, w.Body.String())

	assert.Equal(t, http.StatusOK, w.Code, "Move action should succeed")

	// 4. Attack P2E1 at its actual position
	log.Printf("Executing ATTACK action on %+v...", p2e1Pos)
	attackReqBody, _ := json.Marshal(api.ArenaActionMessage{
		RequestID: uuid.NewString(),
		Data: api.ArenaActionRequest{
			PlayerID: p1ID,
			EntityID: p1e1ID,
			Type:     "attack",
			TargetCoords: []api.Position{
				p2e1Pos,
			},
		},
	})
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/internal/arena/"+arenaID+"/action", bytes.NewBuffer(attackReqBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	log.Printf("ATTACK status: %d, response: %s", w.Code, w.Body.String())

	// Note: attack might fail with "Out of range" depending on random spawn distance, 
	// but 200/400 is expected depending on logic.
	if w.Code != http.StatusOK {
		log.Printf("Attack failed (likely out of range): %s", w.Body.String())
	}

	// 5. Pass turn
	log.Printf("Executing PASS action...")
	passReqBody, _ := json.Marshal(api.ArenaActionMessage{
		RequestID: uuid.NewString(),
		Data: api.ArenaActionRequest{
			PlayerID: p1ID,
			EntityID: p1e1ID,
			Type:     "pass",
		},
	})
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/internal/arena/"+arenaID+"/action", bytes.NewBuffer(passReqBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	log.Printf("PASS status: %d, response: %s", w.Code, w.Body.String())
	assert.Equal(t, http.StatusOK, w.Code, "Pass action should succeed")
	waitForWebhook(t, webhookEvents, "turn.started")
}

func waitForWebhook(t *testing.T, events chan map[string]interface{}, expectedType string) {
	timeout := time.After(5 * time.Second)
	for {
		select {
		case event := <-events:
			data, ok := event["data"].(map[string]interface{})
			if ok {
				if data["event_type"] == expectedType {
					return
				}
			}
		case <-timeout:
			t.Fatalf("Timed out waiting for webhook event: %s", expectedType)
		}
	}
}
