package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/xtj/ai-argus/pkg/logger"
	"github.com/xtj/ai-argus/pkg/response"
	"go.uber.org/zap"
)

func Recovery() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered any) {
		logger.L().Error("HTTP request panic", zap.String("path", c.Request.URL.Path), zap.Any("error", recovered))
		response.Error(c, http.StatusInternalServerError, "服务内部错误")
		c.Abort()
	})
}
