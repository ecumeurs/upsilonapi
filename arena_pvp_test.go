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
// It verifies that the UpsilonAPI correctly initializes a match between two human players
// and asynchronously broadcasts the "game.started" event to the provided callback URL.
// Pass condition: receive a "game.started" event via webhook within the timeout period.
// This test exercises the full lifecycle from HTTP request to engine execution and webhook callback.
// The test ensures that the API bridge correctly handles match creation and event forwarding.
// @spec-link [[api_go_battle_engine]]
// @spec-link [[api_go_battle_start]]
func TestArenaStart1v1PvP(t *testing.T) {
	// Initialize the router for the UpsilonAPI.
	router := setupRouter()

	// Setup mock webhook receiver to capture engine events asynchronously.
	// We use a buffered channel to avoid blocking the engine's event loop.
	// This channel will store events received by our mock server.
	webhookEvents := make(chan map[string]interface{}, 10)
	
	// Create a test server to act as the callback URL for the engine.
	// Extracted to a helper to reduce nesting depth in this function.
	// This server will listen for POST requests from the engine.
	ts := setupMockWebhookServer(t, webhookEvents)
	
	// Ensure the test server is cleaned up and the channel is closed after the test.
	// This prevents resource leaks and ensures the test completes cleanly.
	defer func() {
		// Small delay to ensure any in-flight POSTs from the engine's HTTPController are settled.
		time.Sleep(100 * time.Millisecond)
		ts.Close()
		close(webhookEvents)
	}()

	// Generate unique IDs for the match and the request.
	// matchID is used to identify the arena instance.
	// requestID is used to track the specific start request.
	matchID := uuid.New().String()
	requestID := uuid.New().String()

	// 1v1 PvP Setup: Use a helper to avoid nesting complexity from struct literals.
	// This keeps the main test logic clean and focused on the flow.
	players := getTestPlayers()

	// Construct the arena start request message following the standard API contract.
	// We include the matchID, callback URL, and the list of players.
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

	// Marshal the request into JSON and prepare the HTTP POST request.
	// This simulates a request from the Laravel gateway or a CLI tool.
	reqBody, _ := json.Marshal(startRequest)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/v1/arena/start", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	
	// Execute the request via the router and capture the immediate response.
	// This tests the routing and handler logic of the API.
	router.ServeHTTP(w, req)

	// Assert immediate API response is successful (200 OK) and contains correct data.
	// We expect the API to acknowledge the request and return the arena ID.
	assert.Equal(t, http.StatusOK, w.Code, "API should return 200 OK")
	
	var resp api.ArenaStartResponseMessage
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.True(t, resp.Success)
	assert.Equal(t, matchID, resp.Data.ArenaID, "Arena ID should match MatchID")

	// Verify asynchronous "game.started" event is received via the webhook.
	// We wait for the engine to broadcast that the match has officially begun.
	receivedGameStarted := false
	timeout := time.After(5 * time.Second)

	// Loop until the expected event is received or we time out (5 seconds).
	// This is a common pattern for testing asynchronous callbacks.
	for !receivedGameStarted {
		select {
		case event := <-webhookEvents:
			// Process the event using a helper to maintain low nesting depth.
			// The helper handles the parsing and validation of the event.
			receivedGameStarted = verifyGameStartedEvent(t, event, matchID)
		case <-timeout:
			// Fail the test if we don't get the event in time.
			// This usually indicates a failure in the engine or the callback logic.
			t.Fatal("Timed out waiting for 'game.started' event")
		}
	}

	// Final verification that we actually exited the loop because of success.
	// This ensures our receivedGameStarted flag was correctly set.
	assert.True(t, receivedGameStarted, "Should have received 'game.started' event")
}

// setupMockWebhookServer initializes a local HTTP server to receive engine events.
// It parses the incoming standard message envelopes and sends the data to a channel.
func setupMockWebhookServer(t *testing.T, events chan<- map[string]interface{}) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Read the body of the incoming webhook request from the engine.
		body, _ := io.ReadAll(r.Body)
		
		// Validate that the message follows the standard envelope format.
		// This ensures compatibility with our API standards.
		var wrapped stdmessage.StandardMessage[api.ArenaEvent, stdmessage.MetaNil]
		if err := json.Unmarshal(body, &wrapped); err != nil {
			t.Errorf("Failed to unmarshal webhook payload: %v", err)
			return
		}
		
		// Parse the raw body into a map for generic assertion checking in the test.
		var event map[string]interface{}
		json.Unmarshal(body, &event)
		
		// Send the event to our channel for the main test loop to process and verify.
		events <- event
		w.WriteHeader(http.StatusOK)
	}))
}

// verifyGameStartedEvent checks if the given event is a "game.started" event for the expected match.
// It performs detailed assertions on the event payload and returns true if it matches.
func verifyGameStartedEvent(t *testing.T, event map[string]interface{}, matchID string) bool {
	// Extract the "data" block from the standard envelope.
	// The data block contains the actual event details.
	data, ok := event["data"].(map[string]interface{})
	if !ok {
		return false
	}
	
	// Check if the event type is indeed "game.started".
	// Other event types might be received, which we should ignore.
	eventType, _ := data["event_type"].(string)
	if eventType != "game.started" {
		return false
	}

	// Verify that the match ID in the event matches our started match.
	// This ensures we are not receiving events for other concurrent tests.
	assert.Equal(t, matchID, data["match_id"], "Event MatchID mismatch")
	
	// Ensure the board data is present and contains the expected number of players (2 for 1v1).
	// This validates the initial state of the battle arena.
	boardData, ok := data["data"].(map[string]interface{})
	assert.True(t, ok, "Event should contain board data")
	assert.Len(t, boardData["players"], 2, "Board state should have 2 players")
	
	return true
}

// getTestPlayers returns a standard 1v1 PvP player configuration for testing.
// This helper defines two human players with one entity each and unique IDs.
// @lint-ignore-complexity
func getTestPlayers() []api.Player {
	// Returns two players, each with one entity, for a 1v1 battle setup.
	// This helper ensures we have a consistent state for pvp tests.
	// Each player is initialized with a unique ID and a specific team.
	// The entities are also given unique IDs and baseline stats for combat.
	return []api.Player{
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
}



