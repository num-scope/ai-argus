package v1

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/xtj/ai-argus/internal/common"
	"github.com/xtj/ai-argus/pkg/logger"
	"github.com/xtj/ai-argus/pkg/response"
	"go.uber.org/zap"
)

func writeServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, common.ErrNotFound):
		response.Error(c, http.StatusNotFound, common.ErrNotFound.Error())
	case errors.Is(err, common.ErrAlreadyExists):
		response.Error(c, http.StatusConflict, common.ErrAlreadyExists.Error())
	case errors.Is(err, common.ErrRunAlreadyFinished):
		response.Error(c, http.StatusConflict, common.ErrRunAlreadyFinished.Error())
	case errors.Is(err, common.ErrInvalidRequestBody):
		response.Error(c, http.StatusBadRequest, err.Error())
	default:
		logger.L().Error("api service failed", zap.String("path", c.Request.URL.Path), zap.Error(err))
		response.Error(c, http.StatusInternalServerError, "服务内部错误")
	}
}
