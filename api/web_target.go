package api

import (
	"net/http"
	"net/url"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/xtj/ai-argus/internal/dto"
	"github.com/xtj/ai-argus/internal/service"
)

func Targets(c *gin.Context) {
	targets, err := service.ListTargets(c.Request.Context())
	if err != nil {
		renderError(c, err)
		return
	}
	c.HTML(http.StatusOK, "targets.html", targetsPage{
		pageBase: pageBase{Page: "targets", Notice: c.Query("notice"), Error: c.Query("error")},
		Targets:  targets,
	})
}

func CreateTarget(c *gin.Context) {
	var req dto.TargetRequest
	if err := c.ShouldBind(&req); err != nil {
		redirectError(c, "/targets", "目标参数无效")
		return
	}
	if _, err := service.CreateTarget(c.Request.Context(), req); err != nil {
		redirectServiceError(c, "/targets", err)
		return
	}
	c.Redirect(http.StatusSeeOther, "/targets?notice="+url.QueryEscape("推理目标已保存"))
}

func DeleteTarget(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		redirectError(c, "/targets", "目标编号无效")
		return
	}
	if err := service.DeleteTarget(c.Request.Context(), id); err != nil {
		redirectServiceError(c, "/targets", err)
		return
	}
	c.Redirect(http.StatusSeeOther, "/targets?notice="+url.QueryEscape("推理目标已删除"))
}
