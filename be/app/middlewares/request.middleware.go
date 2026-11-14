package middlewares

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

const requestIDKey = "request_id"

// RequestIDMiddleware propagates `X-Request-Id` (or creates one) and echoes it in the response.
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader("X-Request-Id")
		if id == "" || len(id) > 64 {
			id = uuid.NewString()
		}
		c.Set(requestIDKey, id)
		c.Header("X-Request-Id", id)
		c.Next()
	}
}

func GetRequestID(c *gin.Context) string {
	return c.GetString(requestIDKey)
}

// RequestLoggerMiddleware writes one structured log line per request.
func RequestLoggerMiddleware(logger zerolog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		status := c.Writer.Status()
		event := logger.Info()
		if status >= 500 {
			event = logger.Error()
		} else if status >= 400 {
			event = logger.Warn()
		}
		event.
			Str("request_id", GetRequestID(c)).
			Str("method", c.Request.Method).
			Str("path", c.FullPath()).
			Int("status", status).
			Dur("latency", time.Since(start)).
			Str("ip", c.ClientIP()).
			Msg("http_request")
	}
}
