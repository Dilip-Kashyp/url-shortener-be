package service

import (
	"net/http"
	"net/http/httptest"
	"testing"
	routes "url-shortener/test"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestPingService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	routes.SetupTestDB()

	t.Run("Ping Returns Pong", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/ping", nil)
		w := httptest.NewRecorder()

		c, _ := gin.CreateTestContext(w)
		c.Request = req

		PingService(c)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "pong")
	})
}
