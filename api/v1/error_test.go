package v1

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/xtj/ai-argus/internal/common"
)

func TestWriteServiceErrorMapsKnownErrors(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct {
		name       string
		err        error
		statusCode int
		message    string
	}{
		{name: "not found", err: common.ErrNotFound, statusCode: http.StatusNotFound, message: common.ErrNotFound.Error()},
		{name: "already exists", err: common.ErrAlreadyExists, statusCode: http.StatusConflict, message: common.ErrAlreadyExists.Error()},
		{name: "finished", err: common.ErrRunAlreadyFinished, statusCode: http.StatusConflict, message: common.ErrRunAlreadyFinished.Error()},
		{name: "invalid", err: fmt.Errorf("%w: bad value", common.ErrInvalidRequestBody), statusCode: http.StatusBadRequest, message: "bad value"},
		{name: "internal", err: errors.New("database path secret"), statusCode: http.StatusInternalServerError, message: "服务内部错误"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(recorder)
			c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/test", nil)
			writeServiceError(c, test.err)
			if recorder.Code != test.statusCode {
				t.Fatalf("expected status %d, got %d", test.statusCode, recorder.Code)
			}
			if !strings.Contains(recorder.Body.String(), test.message) {
				t.Fatalf("expected message %q, got %s", test.message, recorder.Body.String())
			}
			if strings.Contains(recorder.Body.String(), "database path secret") {
				t.Fatalf("internal error leaked: %s", recorder.Body.String())
			}
		})
	}
}
