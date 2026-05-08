// Package main provides integration tests for the UpsilonAPI server.
// It validates the full end-to-end flow from HTTP request to engine execution and back via webhooks.
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

// TestArenaStart1v1PvP verifies that a 1v1 PvP match can be started and triggers a game.started event.
// It validates the full request-response-callback cycle including UUID consistency.
// @test-link [[api_go_battle_engine]]
// @test-link [[api_go_battle_start]]
func TestArenaStart1v1PvP(t *testing.T) {
	// 1. Setup Phase: Initialize the API router and a mock webhook receiver channel.
	router := setupRouter()
	webhookEvents := make(chan map[string]interface{}, 10)
	
	// 2. Server Mocking: Start a local HTTP server to act as the Laravel callback target.
	ts := setupMockWebhookServer(t, webhookEvents)
	defer func() {
		time.Sleep(100 * time.Millisecond)
		ts.Close()
		close(webhookEvents)
	}()

	// 3. Identity Generation: Create unique identifiers for the match and the request context.
	matchID := uuid.New().String()
	requestID := uuid.New().String()
	players := getTestPlayers()

	// 4. Request Construction: Define the start message following the StandardMessage envelope.
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

	// 5. Execution Phase: Post the start request to the /v1/arena/start endpoint.
	reqBody, _ := json.Marshal(startRequest)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/v1/arena/start", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	// 6. Immediate Validation: Assert that the API acknowledged the request with 200 OK.
	assert.Equal(t, http.StatusOK, w.Code, "API must return 200 OK on successful start")
	
	var resp api.ArenaStartResponseMessage
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, matchID, resp.Data.ArenaID, "Returned arena ID must match requested ID")

	// 7. Asynchronous Validation: Wait for the game.started event to arrive via the webhook.
	verifyAsyncGameStart(t, webhookEvents, matchID)
}

// verifyAsyncGameStart blocks until a game.started event for the given match is received.
// It uses a timeout to prevent deadlocking the test suite if the engine fails to respond.
func verifyAsyncGameStart(t *testing.T, events chan map[string]interface{}, matchID string) {
	timeout := time.After(5 * time.Second)
	for {
		select {
		case event := <-events:
			// 1. Event Audit: Check if the received payload represents a game start.
			if verifyGameStartedEvent(t, event, matchID) {
				return
			}
		case <-timeout:
			// 2. Failure: Terminate if the engine fails to broadcast within 5 seconds.
			t.Fatal("Timed out waiting for 'game.started' event")
		}
	}
}

// setupMockWebhookServer initializes a local HTTP server to receive engine events.
func setupMockWebhookServer(t *testing.T, events chan<- map[string]interface{}) *httptest.Server {
	// 1. Handler Definition: Create an inline handler that unmarshals standard envelopes.
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		
		// 2. Structural Validation: Ensure the engine is sending valid StandardMessage containers.
		var wrapped stdmessage.StandardMessage[api.ArenaEvent, stdmessage.MetaNil]
		if err := json.Unmarshal(body, &wrapped); err != nil {
			t.Errorf("Malformed webhook payload received: %v", err)
			return
		}
		
		// 3. Dispatch: Send the raw map to the test's event loop for detailed inspection.
		var event map[string]interface{}
		_ = json.Unmarshal(body, &event)
		events <- event
		w.WriteHeader(http.StatusOK)
	}))
}

// verifyGameStartedEvent checks if the given event is a "game.started" event for the expected match.
func verifyGameStartedEvent(t *testing.T, event map[string]interface{}, matchID string) bool {
	// 1. Data Extraction: Pull the 'data' block and verify its existence in the envelope.
	data, ok := event["data"].(map[string]interface{})
	if !ok { return false }
	
	// 2. Type Check: Ensure the engine flagged this as a 'game.started' event.
	eventType, _ := data["event_type"].(string)
	if eventType != "game.started" { return false }

	// 3. ID Consistency: Verify the match ID in the payload matches our active test.
	assert.Equal(t, matchID, data["match_id"], "Event MatchID mismatch")
	
	// 4. State Integrity: Confirm the board snapshot contains the correct player count.
	boardData, ok := data["data"].(map[string]interface{})
	assert.True(t, ok, "Event must contain board data")
	assert.Len(t, boardData["players"], 2, "Board state must have 2 players in 1v1")
	
	return true
}

// getTestPlayers returns a standard 1v1 PvP player configuration for testing.
func getTestPlayers() []api.Player {
	// 1. Player One: Initialize a human player on team 1 with a tanky entity.
	p1 := api.Player{
		ID: uuid.NewString(), Team: 1, Nickname: "PlayerOne", IA: false,
		Entities: []api.Entity{{ID: uuid.NewString(), Name: "Warrior", HP: 20, MaxHP: 20, Attack: 5, Defense: 2, Move: 3, MaxMove: 3}},
	}
	
	// 2. Player Two: Initialize a human player on team 2 with a high-damage entity.
	p2 := api.Player{
		ID: uuid.NewString(), Team: 2, Nickname: "PlayerTwo", IA: false,
		Entities: []api.Entity{{ID: uuid.NewString(), Name: "Mage", HP: 15, MaxHP: 15, Attack: 7, Defense: 1, Move: 3, MaxMove: 3}},
	}
	
	// 3. Roster Assembly: Return the consolidated player list.
	return []api.Player{p1, p2}
}
