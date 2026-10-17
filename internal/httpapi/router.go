// Package httpapi adapts HTTP requests to Kotoba's application logic using Gin.
package httpapi

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// NewRouter builds the local learning API. Authentication is not implemented yet.
func NewRouter(logger *slog.Logger) *gin.Engine {
	return NewRouterWithReadiness(logger, nil)
}

// NewRouterWithReadiness adds a dependency check without exposing driver details.
func NewRouterWithReadiness(logger *slog.Logger, ready func(context.Context) error) *gin.Engine {
	router := gin.New()
	// No reverse proxy is trusted in the local-only first slice.
	_ = router.SetTrustedProxies(nil)
	router.HandleMethodNotAllowed = true
	router.Use(func(c *gin.Context) {
		started := time.Now()
		c.Next()
		// Route templates avoid recording user input in URLs or request bodies.
		logger.Info("http request", "method", c.Request.Method,
			"route", c.FullPath(), "status", c.Writer.Status(),
			"duration", time.Since(started))
	})
	router.Use(gin.CustomRecoveryWithWriter(io.Discard, func(c *gin.Context, _ any) {
		logger.Error("request panic recovered")
		writeError(c, http.StatusInternalServerError, "internal_error", "Something went wrong.", nil)
	}))
	router.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "kotoba-api"})
	})
	router.GET("/readyz", func(c *gin.Context) {
		if ready == nil {
			writeError(c, http.StatusServiceUnavailable, "database_not_configured", "Database is not configured.", nil)
			return
		}
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()
		if err := ready(ctx); err != nil {
			writeError(c, http.StatusServiceUnavailable, "database_unavailable", "Database is unavailable.", nil)
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ready"})
	})
	v1 := router.Group("/api/v1")
	v1.POST("/words/validate", validateWord)
	router.NoRoute(func(c *gin.Context) {
		writeError(c, http.StatusNotFound, "not_found", "This endpoint does not exist.", nil)
	})
	router.NoMethod(func(c *gin.Context) {
		writeError(c, http.StatusMethodNotAllowed, "method_not_allowed", "This method is not supported.", nil)
	})
	return router
}
