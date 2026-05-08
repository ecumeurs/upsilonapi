// Package main provides the entry point for the UpsilonAPI server.
// This server acts as an HTTP bridge to the UpsilonBattle engine.
// @spec-link [[upsilonapi:module_upsilonapi]]
// @spec-link [[upsilonapi:api_bridge_orchestration]]
package main

import (
	"runtime/debug"

	"github.com/ecumeurs/upsilonapi/handler"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// getGitRevision retrieves the current VCS revision from the build info.
// It iterates through the build settings to find the "vcs.revision" key.
// Returns "unknown" if the build info is unavailable or the revision is not found.
func getGitRevision() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "unknown"
	}

	for _, setting := range info.Settings {
		if setting.Key == "vcs.revision" {
			return setting.Value
		}
	}
	return "unknown"
}

// main is the primary entry point for the UpsilonAPI service.
// It initializes the Gin router, sets up the API routes, and starts the server on port 8081.
// The server handles arena lifecycle management, match statistics, and skill generation.
func main() {
	// Initialize the Gin router with default middleware (logger and recovery).
	r := gin.Default()

	// Capture the current Git revision for logging and health checks.
	rev := getGitRevision()
	logrus.Infof("Starting UpsilonAPI server on :8081 (rev: %s)", rev)

	// @spec-link [[api_go_health_check]]
	// Health check endpoint (used by Docker healthcheck in CI and orchestration tools).
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "revision": rev})
	})

	// V1 API Group: All game orchestration endpoints are versioned under /v1.
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

	// Start the HTTP server. This call is blocking unless an error occurs.
	if err := r.Run(":8081"); err != nil {
		logrus.Fatalf("Failed to start server: %v", err)
	}
}

