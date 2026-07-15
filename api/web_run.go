package api

import (
	"errors"
	"net/http"
	"net/url"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/xtj/ai-argus/internal/common"
	"github.com/xtj/ai-argus/internal/dto"
	"github.com/xtj/ai-argus/internal/service"
)

func Runs(c *gin.Context) {
	workspace, err := service.GetRunWorkspace(c.Request.Context(), 100)
	if err != nil {
		renderError(c, err)
		return
	}
	c.HTML(http.StatusOK, "runs.html", runsPage{
		pageBase: pageBase{Page: "runs", Notice: c.Query("notice"), Error: c.Query("error")},
		Runs:     workspace.Runs, Targets: workspace.Targets, Scenarios: workspace.Scenarios,
	})
}

func CreateRun(c *gin.Context) {
	var req dto.StartRunRequest
	if err := c.ShouldBind(&req); err != nil {
		redirectError(c, "/runs", "请选择目标和场景")
		return
	}
	run, err := service.StartRun(c.Request.Context(), req)
	if err != nil {
		redirectServiceError(c, "/runs", err)
		return
	}
	c.Redirect(http.StatusSeeOther, "/runs/"+strconv.FormatInt(run.ID, 10))
}

func RunDetail(c *gin.Context) {
	renderRun(c, "run_detail.html", "runs")
}

func Report(c *gin.Context) {
	renderRun(c, "report.html", "reports")
}

func CancelRun(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		redirectError(c, "/runs", "任务编号无效")
		return
	}
	if err := service.CancelRun(c.Request.Context(), id); err != nil {
		redirectServiceError(c, "/runs/"+strconv.FormatInt(id, 10), err)
		return
	}
	c.Redirect(http.StatusSeeOther, "/runs/"+strconv.FormatInt(id, 10)+"?notice="+url.QueryEscape("停止指令已发送"))
}

func renderRun(c *gin.Context, templateName, page string) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.String(http.StatusBadRequest, "任务编号无效")
		return
	}
	detail, err := service.GetRunDetail(c.Request.Context(), id, 100)
	if err != nil {
		if errors.Is(err, common.ErrNotFound) {
			c.String(http.StatusNotFound, common.ErrNotFound.Error())
			return
		}
		renderError(c, err)
		return
	}
	c.HTML(http.StatusOK, templateName, runPage{
		pageBase: pageBase{Page: page, Notice: c.Query("notice"), Error: c.Query("error"), Charts: templateName == "run_detail.html"},
		Detail:   detail,
	})
}
