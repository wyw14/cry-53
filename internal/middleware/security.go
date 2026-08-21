package middleware

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/wyw14/cry-053/internal/domain"
)

const actorKey = "actor"

func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("Referrer-Policy", "no-referrer")
		c.Header("Content-Security-Policy", "default-src 'self'; frame-ancestors 'none'")
		c.Next()
	}
}

func CORS(allowed []string) gin.HandlerFunc {
	allow := make(map[string]struct{}, len(allowed))
	for _, origin := range allowed {
		allow[origin] = struct{}{}
	}
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if _, ok := allow[origin]; ok {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Vary", "Origin")
			c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, Idempotency-Key, X-Request-ID, X-Actor-ID, X-Actor-Roles")
			c.Header("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
		}
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

func Actor() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := strings.TrimSpace(c.GetHeader("X-Actor-ID"))
		roles := splitRoles(c.GetHeader("X-Actor-Roles"))
		if id == "" || len(roles) == 0 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"code": "UNAUTHORIZED", "message": "缺少本地身份头", "request_id": GetRequestID(c)})
			return
		}
		c.Set(actorKey, domain.Actor{ID: id, Name: id, Roles: roles})
		c.Next()
	}
}

func GetActor(c *gin.Context) domain.Actor {
	value, _ := c.Get(actorKey)
	actor, _ := value.(domain.Actor)
	return actor
}

func Timeout(duration time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), duration)
		defer cancel()
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

func splitRoles(raw string) []string {
	result := make([]string, 0)
	seen := make(map[string]struct{})
	for _, role := range strings.Split(raw, ",") {
		role = strings.TrimSpace(role)
		if role == "" {
			continue
		}
		if _, ok := seen[role]; ok {
			continue
		}
		seen[role] = struct{}{}
		result = append(result, role)
	}
	return result
}
