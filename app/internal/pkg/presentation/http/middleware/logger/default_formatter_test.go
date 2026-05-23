package logger

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/iPatrushevSergey/metrics/app/internal/pkg/adapters/logger"
	"github.com/stretchr/testify/assert"
)

func TestDefaultLogFormatter_Log(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/ping", nil)

	f := DefaultLogFormatter{}
	f.Log(logger.NewNopLogger(), LogParams{
		Ctx:          c,
		Duration:     time.Millisecond,
		RequestBody:  []byte("req"),
		ResponseBody: httptest.NewRecorder().Body,
	})
	assert.Equal(t, http.StatusOK, w.Code)
}
