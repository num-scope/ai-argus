package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/xtj/ai-argus/api/v1"
	"github.com/xtj/ai-argus/internal/middleware"
	"github.com/xtj/ai-argus/pkg/response"
	"github.com/xtj/ai-argus/web"
)

func NewRouter() (*gin.Engine, error) {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(middleware.RequestLog(), middleware.Recovery())

	templates, err := web.Templates()
	if err != nil {
		return nil, err
	}
	router.SetHTMLTemplate(templates)
	staticFiles, err := web.StaticFS()
	if err != nil {
		return nil, err
	}
	router.StaticFS("/static", http.FS(staticFiles))

	router.GET("/", Dashboard)
	router.GET("/targets", Targets)
	router.POST("/targets", CreateTarget)
	router.POST("/targets/:id/delete", DeleteTarget)
	router.GET("/scenarios", Scenarios)
	router.POST("/scenarios", CreateScenario)
	router.POST("/scenarios/:id/delete", DeleteScenario)
	router.GET("/runs", Runs)
	router.POST("/runs", CreateRun)
	router.GET("/runs/:id", RunDetail)
	router.POST("/runs/:id/cancel", CancelRun)
	router.GET("/runs/:id/export", ExportRun)
	router.GET("/reports/:id", Report)

	apiGroup := router.Group("/api/v1")
	apiGroup.GET("/targets", v1.GetTargetList)
	apiGroup.GET("/scenarios", v1.GetScenarioList)
	apiGroup.GET("/runs/:id", v1.GetRun)
	apiGroup.POST("/runs/:id/cancel", v1.CancelRun)

	router.GET("/health", func(c *gin.Context) {
		response.Success(c, gin.H{"status": "up"})
	})
	return router, nil
}
