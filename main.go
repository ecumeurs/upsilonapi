// Package main provides the entry point for the UpsilonAPI server.
// This server acts as an HTTP bridge to the UpsilonBattle engine.
// @spec-link [[api_go_battle_engine]]
// @spec-link [[api_go_health_check]]
package main

import (
	"runtime/debug"

	"github.com/ecumeurs/upsilonapi/handler"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// getGitRevision retrieves the current VCS revision from the build info.
// It iterates through the build settings to find the "vcs.revision" key.
// This allows the API to report its exact build version via the /health endpoint.
func getGitRevision() string {
	// 1. Build Info Retrieval: Access the embedded runtime build metadata.
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "unknown"
	}

	// 2. Setting Lookup: Iterate through the KV pairs to find the git hash.
	for _, setting := range info.Settings {
		if setting.Key == "vcs.revision" {
			return setting.Value
		}
	}
	return "unknown"
}

// main is the primary entry point for the UpsilonAPI service.
// It initializes the Gin router, sets up the API routes, and starts the server on port 8081.
func main() {
	// 1. Framework Setup: Initialize the Gin router with default logger and recovery middleware.
	r := gin.Default()

	// 2. Metadata Initialization: Capture the current Git revision for health reporting.
	rev := getGitRevision()
	logrus.Infof("Starting UpsilonAPI server on :8081 (rev: %s)", rev)

	// 3. Health Monitoring: Register the root health endpoint for orchestration heartbeats.
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "revision": rev})
	})

	// 4. API Routing: Register the V1 route group for match lifecycle and tactical operations.
	registerV1Routes(r)

	// 5. Server Startup: Launch the HTTP server on the designated bridge port.
	if err := r.Run(":8081"); err != nil {
		logrus.Fatalf("Failed to start server: %v", err)
	}
}

// registerV1Routes defines the routing table for version 1 of the Upsilon API.
// It maps REST endpoints to their respective handlers in the handler package.
// @spec-link [[api_go_routing_table]]
func registerV1Routes(r *gin.Engine) {
	v1 := r.Group("/v1")
	{
		// Arena lifecycle management endpoints.
		v1.POST("/arena/start", handler.HandleArenaStart)
		v1.POST("/arena/:id/action", handler.HandleArenaAction)
		v1.POST("/arena/:id/forfeit", handler.HandleArenaForfeit)
		v1.GET("/arena/:id/exists", handler.HandleArenaExists)
		v1.POST("/arena/:id/resurrect", handler.HandleArenaResurrect)

		// Match statistics and telemetry endpoints.
		v1.GET("/match/stats/active", handler.HandleGetActiveMatchStats)

		// Procedural skill generation endpoints.
		v1.POST("/skills/generate", handler.HandleSkillGenerate)
	}
}
