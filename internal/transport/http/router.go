package http

import (
	stdhttp "net/http"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/wyw14/cry-053/internal/config"
	"github.com/wyw14/cry-053/internal/middleware"
	"go.uber.org/zap"
)

type Readiness struct {
	ready atomic.Bool
}

func (r *Readiness) Set(value bool) {
	r.ready.Store(value)
}

func NewRouter(handler *Handler, logger *zap.Logger, timeout time.Duration, readiness *Readiness) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(middleware.RequestID(), middleware.SecurityHeaders(), middleware.CORS([]string{"http://localhost:5173"}), middleware.Recovery(logger), middleware.Timeout(timeout))
	router.GET("/healthz", func(c *gin.Context) {
		c.JSON(stdhttp.StatusOK, gin.H{"status": "ok", "project": config.Info})
	})
	router.GET("/readyz", func(c *gin.Context) {
		if !readiness.ready.Load() {
			c.JSON(stdhttp.StatusServiceUnavailable, gin.H{"status": "not_ready"})
			return
		}
		c.JSON(stdhttp.StatusOK, gin.H{"status": "ready"})
	})
	api := router.Group("/api/v1", middleware.Actor())
	api.POST("/bundles", handler.uploadBundle)
	api.POST("/bundles/:id/validate", handler.validateBundle)
	api.GET("/bundles/:id/preview", handler.previewBundle)
	api.POST("/bundles/:id/approve", handler.approveBundle)
	api.POST("/bundles/:id/publish", handler.publishBundle)
	api.GET("/config-types", handler.listTypes)
	api.POST("/config-types", handler.saveType)
	api.PATCH("/configurations/:id/status", handler.transition)
	api.POST("/configurations/:id/rollback", handler.rollback)
	api.DELETE("/configurations/:id", handler.deleteConfiguration)
	api.GET("/configurations/:id/impact", handler.impactReport)
	api.GET("/audits", handler.listAudits)
	return router
}
