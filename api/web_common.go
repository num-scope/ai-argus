package api

import (
	"errors"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/xtj/ai-argus/internal/common"
	"github.com/xtj/ai-argus/internal/dto"
	"github.com/xtj/ai-argus/pkg/logger"
	"go.uber.org/zap"
)

type pageBase struct {
	Page   string
	Notice string
	Error  string
}

type dashboardPage struct {
	pageBase
	Dashboard *dto.Dashboard
}

type targetsPage struct {
	pageBase
	Targets []dto.TargetResponse
}

type scenariosPage struct {
	pageBase
	Scenarios      []dto.ScenarioResponse
	MaxConcurrency int
}

type runsPage struct {
	pageBase
	Runs      []dto.RunResponse
	Targets   []dto.TargetResponse
	Scenarios []dto.ScenarioResponse
}

type runPage struct {
	pageBase
	Detail *dto.RunDetail
}

func redirectError(c *gin.Context, path, message string) {
	c.Redirect(http.StatusSeeOther, path+"?error="+url.QueryEscape(message))
}

func redirectServiceError(c *gin.Context, path string, err error) {
	if errors.Is(err, common.ErrInvalidRequestBody) || errors.Is(err, common.ErrNotFound) || errors.Is(err, common.ErrAlreadyExists) || errors.Is(err, common.ErrRunAlreadyFinished) {
		redirectError(c, path, err.Error())
		return
	}
	logger.L().Error("web command failed", zap.String("path", c.Request.URL.Path), zap.Error(err))
	redirectError(c, path, "操作失败，请稍后重试")
}

func renderError(c *gin.Context, err error) {
	logger.L().Error("web service failed", zap.String("path", c.Request.URL.Path), zap.Error(err))
	c.String(http.StatusInternalServerError, "平台读取失败")
}
