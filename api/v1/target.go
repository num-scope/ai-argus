package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/xtj/ai-argus/internal/common"
	"github.com/xtj/ai-argus/internal/dto"
	"github.com/xtj/ai-argus/internal/service"
	"github.com/xtj/ai-argus/pkg/response"
)

func GetTargetList(c *gin.Context) {
	var req dto.PageQuery
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, http.StatusBadRequest, common.ErrInvalidParam.Error())
		return
	}
	targets, err := service.ListTargetsPage(c.Request.Context(), req)
	if err != nil {
		writeServiceError(c, err)
		return
	}
	response.Success(c, targets)
}
