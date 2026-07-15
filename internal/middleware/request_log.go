package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xtj/ai-argus/pkg/logger"
	"go.uber.org/zap"
)

func RequestLog() gin.HandlerFunc {
	return func(c *gin.Context) {
		startedAt := time.Now()
		c.Next()
		logger.L().Info("HTTP request completed",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Int("status", c.Writer.Status()),
			zap.Duration("duration", time.Since(startedAt)),
		)
	}
}
