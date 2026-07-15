package v1

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/xtj/ai-argus/internal/common"
	"github.com/xtj/ai-argus/internal/service"
	"github.com/xtj/ai-argus/pkg/response"
)

func GetRun(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, common.ErrInvalidParam.Error())
		return
	}
	detail, err := service.GetRunDetail(c.Request.Context(), id, 100)
	if err != nil {
		writeServiceError(c, err)
		return
	}
	response.Success(c, detail)
}

func CancelRun(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, common.ErrInvalidParam.Error())
		return
	}
	if err := service.CancelRun(c.Request.Context(), id); err != nil {
		writeServiceError(c, err)
		return
	}
	response.Success(c, gin.H{"cancelled": true})
}
