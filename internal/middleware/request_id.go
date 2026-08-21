package middleware

import (
	"crypto/rand"
	"encoding/hex"

	"github.com/gin-gonic/gin"
)

const requestIDKey = "request_id"

func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
		}
		c.Set(requestIDKey, requestID)
		c.Header("X-Request-ID", requestID)
		c.Next()
	}
}

func GetRequestID(c *gin.Context) string {
	value, _ := c.Get(requestIDKey)
	requestID, _ := value.(string)
	return requestID
}

func generateRequestID() string {
	payload := make([]byte, 12)
	if _, err := rand.Read(payload); err != nil {
		return "request-unavailable"
	}
	return hex.EncodeToString(payload)
}
