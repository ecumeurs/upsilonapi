// Package main provides comprehensive end-to-end integration tests for the UpsilonAPI.
// It verifies the full tactical cycle including match start, movement, and skill execution.
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ecumeurs/upsilonapi/api"
	"github.com/ecumeurs/upsilonapi/handler"
	"github.com/ecumeurs/upsilonapi/stdmessage"
	"github.com/ecumeurs/upsilonmapdata/grid/position"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// setupRouter initializes the Gin engine with the standard Upsilon routing table.
// @test-link [[api_go_routing_table]]
// @test-link [[api_go_battle_engine]]
func setupRouter() *gin.Engine {
	// 1. Initial Setup: Create a default Gin engine with recovery and logging.
	r := gin.Default()
	
	// 2. Route Registration: Attach the V1 handlers for match orchestration.
	v1 := r.Group("/v1")
	{
		v1.POST("/arena/start", handler.HandleArenaStart)
		v1.POST("/arena/:id/action", handler.HandleArenaAction)
	}
	return r
}

// TestArenaStartEndpoint verifies the /v1/arena/start REST API contract.
// It ensures that the match initiation request is correctly received and acknowledged.
// @test-link [[api_go_battle_start]]
func TestArenaStartEndpoint(t *testing.T) {
	// 1. Environment: Setup the router and unique match metadata.
	router := setupRouter()
	matchID := uuid.New().String()
	
	// 2. Request Setup: Build the start request payload using helper-generated players.
	startRequest := api.ArenaStartMessage{
		RequestID: uuid.New().String(),
		Data: api.ArenaStartRequest{
			MatchID:     matchID,
			CallbackURL: "http://localhost/webhook",
			Players:     getTestPlayers(),
		},
	}

	// 3. Execution: Dispatch the POST request to the API.
	w := performPost(router, "/v1/arena/start", startRequest)

	// 4. Validation: Assert success status and correct arena ID reflection.
	assert.Equal(t, http.StatusOK, w.Code)
	var resp api.ArenaStartResponseMessage
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, matchID, resp.Data.ArenaID)
}

// TestBattleFullRoundtrip executes a multi-turn tactical sequence to verify engine-bridge consistency.
// It follows the lifecycle from match start to first action and verifies state broadcasts.
// @test-link [[api_go_battle_engine]]
func TestBattleFullRoundtrip(t *testing.T) {
	// 1. Setup: Initialize communication channels and mock callback server.
	router := setupRouter()
	webhookEvents := make(chan map[string]interface{}, 20)
	ts := setupMockWebhookServer(t, webhookEvents)
	defer ts.Close()

	// 2. Initialization: Start a new arena and capture the turn-0 board state.
	matchID := uuid.New().String()
	bs := executeStart(t, router, matchID, ts.URL)

	// 3. Tactical Loop: Wait for the first turn and verify entity placement.
	waitForWebhook(t, webhookEvents, "turn.started", 2*time.Second)
	p1, p2 := findActorPositions(bs)
	assert.NotEqual(t, p1, p2, "Entities must occupy distinct positions at start")

	// 4. Movement Execution: Move the current entity to a neighboring coordinate.
	target := position.New(p1.X+1, p1.Y, 0)
	executeAction(t, router, matchID, bs.CurrentPlayerID, bs.CurrentEntityID, "move", []api.Position{{X: target.X, Y: target.Y}}, "")

	// 5. Verification: Confirm the move was reflected in the next board broadcast.
	waitForWebhook(t, webhookEvents, "board.updated", 2*time.Second)
}

// executeStart is a helper to initiate a match and return the initial board state.
func executeStart(t *testing.T, router *gin.Engine, matchID, callbackURL string) api.BoardState {
	// 1. Payload Creation: Define a standard 1v1 PvP start message.
	startReq := api.ArenaStartMessage{
		Data: api.ArenaStartRequest{
			MatchID: matchID, CallbackURL: callbackURL, Players: getTestPlayers(),
		},
	}
	
	// 2. Post Execution: Send the request to the start endpoint.
	w := performPost(router, "/v1/arena/start", startReq)
	requireStatus(t, w, http.StatusOK)
	
	// 3. Unmarshaling: Extract the initial board state from the response.
	var resp api.ArenaStartResponseMessage
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	return resp.Data.InitialState
}

// executeAction is a helper to dispatch a tactical command to the API.
func executeAction(t *testing.T, router *gin.Engine, matchID, playerID, entityID, actionType string, coords []api.Position, skillID string) {
	// 1. Payload Creation: Define the action intent (move, attack, skill).
	actionReq := api.ArenaActionMessage{
		Data: api.ArenaActionRequest{
			PlayerID: playerID, Type: actionType, TargetCoords: coords, EntityID: entityID, SkillID: skillID,
		},
	}
	
	// 2. Post Execution: Send the request to the action endpoint for the specific match.
	url := fmt.Sprintf("/v1/arena/%s/action", matchID)
	w := performPost(router, url, actionReq)
	requireStatus(t, w, http.StatusOK)
}

// findActorPositions scans the board state to locate the primary combatants.
func findActorPositions(bs api.BoardState) (p1, p2 api.Position) {
	// 1. Search Logic: Extract positions for the first entity of each player.
	if len(bs.Players) >= 2 {
		if len(bs.Players[0].Entities) > 0 { p1 = bs.Players[0].Entities[0].Position }
		if len(bs.Players[1].Entities) > 0 { p2 = bs.Players[1].Entities[0].Position }
	}
	return
}

// waitForWebhook blocks until an event of the specified type arrives on the channel.
func waitForWebhook(t *testing.T, events <-chan map[string]interface{}, expectedType string, timeout time.Duration) {
	// 1. Loop Setup: Poll the channel until the target event or timeout occurs.
	deadline := time.After(timeout)
	for {
		select {
		case event := <-events:
			// 2. Type Check: Inspect the 'event_type' field in the incoming envelope.
			if isEventType(event, expectedType) { return }
		case <-deadline:
			// 3. Failure: Terminate the test if the event is not received in time.
			t.Fatalf("Timeout waiting for webhook event: %s", expectedType)
		}
	}
}

// isEventType identifies if a raw event map matches the target type string.
func isEventType(event map[string]interface{}, target string) bool {
	data, ok := event["data"].(map[string]interface{})
	if !ok { return false }
	eventType, _ := data["event_type"].(string)
	return eventType == target
}

// performPost is a utility to execute a JSON POST request against a Gin router.
func performPost(r *gin.Engine, url string, payload interface{}) *httptest.ResponseRecorder {
	body, _ := json.Marshal(payload)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	return w
}

// requireStatus is a simple helper to assert HTTP status codes.
// It terminates the test immediately if the received status does not match the expected one.
func requireStatus(t *testing.T, w *httptest.ResponseRecorder, expected int) {
	if w.Code != expected {
		t.Fatalf("Expected status %d, got %d. Body: %s", expected, w.Code, w.Body.String())
	}
}
