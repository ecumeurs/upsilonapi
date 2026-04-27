package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ecumeurs/upsilonapi/api"
	"github.com/ecumeurs/upsilonapi/stdmessage"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// TestArenaStart1v1PvP imitates a true battle start request for a 1v1 PvP setup.
// Pass condition: receive a "game.started" event via webhook.
// @spec-link [[api_go_battle_engine]]
func TestArenaStart1v1PvP(t *testing.T) {
	router := setupRouter()

	// Setup mock webhook receiver to capture engine events
	webhookEvents := make(chan map[string]interface{}, 10)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var wrapped stdmessage.StandardMessage[api.ArenaEvent, stdmessage.MetaNil]
		if err := json.Unmarshal(body, &wrapped); err != nil {
			t.Errorf("Failed to unmarshal webhook payload: %v", err)
			return
		}
		
		// Convert to map for easier generic checking if needed, but we can also use the struct
		var event map[string]interface{}
		json.Unmarshal(body, &event)
		webhookEvents <- event
		w.WriteHeader(http.StatusOK)
	}))
	defer func() {
		// Small delay to ensure any in-flight POSTs from the engine's HTTPController are settled
		time.Sleep(100 * time.Millisecond)
		ts.Close()
		close(webhookEvents)
	}()

	matchID := uuid.New().String()
	requestID := uuid.New().String()

	// 1v1 PvP Setup: Two human players (IA: false) with one entity each
	players := []api.Player{
		{
			ID:       uuid.NewString(),
			Team:     1,
			Nickname: "PlayerOne",
			IA:       false,
			Entities: []api.Entity{
				{
					ID:      uuid.NewString(),
					Name:    "Warrior",
					HP:      20,
					MaxHP:   20,
					Attack:  5,
					Defense: 2,
					Move:    3,
					MaxMove: 3,
				},
			},
		},
		{
			ID:       uuid.NewString(),
			Team:     2,
			Nickname: "PlayerTwo",
			IA:       false,
			Entities: []api.Entity{
				{
					ID:      uuid.NewString(),
					Name:    "Mage",
					HP:      15,
					MaxHP:   15,
					Attack:  7,
					Defense: 1,
					Move:    3,
					MaxMove: 3,
				},
			},
		},
	}

	startRequest := api.ArenaStartMessage{
		RequestID: requestID,
		Message:   "Start PvP Arena",
		Success:   true,
		Data: api.ArenaStartRequest{
			MatchID:     matchID,
			CallbackURL: ts.URL,
			Players:     players,
		},
		Meta: stdmessage.MetaNil{},
	}

	reqBody, _ := json.Marshal(startRequest)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/v1/arena/start", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	
	// Execute the request
	router.ServeHTTP(w, req)

	// Assert immediate API response
	assert.Equal(t, http.StatusOK, w.Code, "API should return 200 OK")
	
	var resp api.ArenaStartResponseMessage
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.True(t, resp.Success)
	assert.Equal(t, matchID, resp.Data.ArenaID, "Arena ID should match MatchID")

	// Verify asynchronous "game.started" event
	receivedGameStarted := false
	timeout := time.After(5 * time.Second)

	for !receivedGameStarted {
		select {
		case event := <-webhookEvents:
			// Drill down into standard envelope: data -> event_type
			data, ok := event["data"].(map[string]interface{})
			if !ok {
				continue
			}
			
			eventType, _ := data["event_type"].(string)
			if eventType == "game.started" {
				receivedGameStarted = true
				
				// Assert event data integrity
				assert.Equal(t, matchID, data["match_id"], "Event MatchID mismatch")
				boardData, ok := data["data"].(map[string]interface{})
				assert.True(t, ok, "Event should contain board data")
				assert.Len(t, boardData["players"], 2, "Board state should have 2 players")
			}
		case <-timeout:
			t.Fatal("Timed out waiting for 'game.started' event")
		}
	}

	assert.True(t, receivedGameStarted, "Should have received 'game.started' event")
}
