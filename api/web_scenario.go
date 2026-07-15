package api

import (
	"net/http"
	"net/url"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/xtj/ai-argus/internal/dto"
	"github.com/xtj/ai-argus/internal/service"
)

func Scenarios(c *gin.Context) {
	scenarios, err := service.ListScenarios(c.Request.Context())
	if err != nil {
		renderError(c, err)
		return
	}
	c.HTML(http.StatusOK, "scenarios.html", scenariosPage{
		pageBase:       pageBase{Page: "scenarios", Notice: c.Query("notice"), Error: c.Query("error")},
		Scenarios:      scenarios,
		MaxConcurrency: service.RunConcurrencyLimit(),
	})
}

func CreateScenario(c *gin.Context) {
	var req dto.ScenarioRequest
	if err := c.ShouldBind(&req); err != nil {
		redirectError(c, "/scenarios", "场景参数无效")
		return
	}
	if _, err := service.CreateScenario(c.Request.Context(), req); err != nil {
		redirectServiceError(c, "/scenarios", err)
		return
	}
	c.Redirect(http.StatusSeeOther, "/scenarios?notice="+url.QueryEscape("评测场景已保存"))
}

func DeleteScenario(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		redirectError(c, "/scenarios", "场景编号无效")
		return
	}
	if err := service.DeleteScenario(c.Request.Context(), id); err != nil {
		redirectServiceError(c, "/scenarios", err)
		return
	}
	c.Redirect(http.StatusSeeOther, "/scenarios?notice="+url.QueryEscape("评测场景已删除"))
}
