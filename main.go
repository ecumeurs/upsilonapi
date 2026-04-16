package main

import (
	"runtime/debug"

	"github.com/ecumeurs/upsilonapi/handler"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func getGitRevision() string {
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range info.Settings {
			if setting.Key == "vcs.revision" {
				return setting.Value
			}
		}
	}
	return "unknown"
}

func main() {
	r := gin.Default()

	rev := getGitRevision()
	logrus.Infof("Starting UpsilonAPI server on :8081 (rev: %s)", rev)

	// @spec-link [[api_go_health_check]]
	// Health check endpoint (used by Docker healthcheck in CI)
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "revision": rev})
	})

	// Internal Arena Management
	internal := r.Group("/internal")
	{
		internal.POST("/arena/start", handler.HandleArenaStart)
		internal.POST("/arena/:id/action", handler.HandleArenaAction)
	}

	// V1 API
	v1 := r.Group("/v1")
	{
		v1.GET("/match/stats/active", handler.HandleGetActiveMatchStats)
	}

	if err := r.Run(":8081"); err != nil {
		logrus.Fatalf("Failed to start server: %v", err)
	}
}
