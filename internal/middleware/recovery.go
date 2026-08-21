package middleware

import (
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func Recovery(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if recovered := recover(); recovered != nil {
				logger.Error("panic recovered", zap.Any("panic", recovered), zap.ByteString("stack", debug.Stack()), zap.String("request_id", GetRequestID(c)))
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"code": "INTERNAL", "message": "服务发生内部错误", "request_id": GetRequestID(c)})
			}
		}()
		c.Next()
	}
}
