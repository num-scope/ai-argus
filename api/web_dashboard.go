package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/xtj/ai-argus/internal/service"
)

func Dashboard(c *gin.Context) {
	dashboard, err := service.GetDashboard(c.Request.Context())
	if err != nil {
		renderError(c, err)
		return
	}
	c.HTML(http.StatusOK, "dashboard.html", dashboardPage{
		pageBase:  pageBase{Page: "dashboard", Notice: c.Query("notice"), Error: c.Query("error")},
		Dashboard: dashboard,
	})
}
