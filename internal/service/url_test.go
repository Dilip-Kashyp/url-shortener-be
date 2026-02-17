package service

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
	"url-shortener/internal/config"
	"url-shortener/internal/models"
	routes "url-shortener/test"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupURLTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	routes.SetupTestDB()
	r := gin.Default()
	return r
}

func TestShortenURL(t *testing.T) {
	router := setupURLTestRouter()
	router.POST("/shorten", ShortenURL)

	t.Run("Shorten URL With Authenticated User", func(t *testing.T) {
		routes.CleanupTestDB()

		// Create a test user
		user, _ := routes.CreateTestUser("test@example.com", "Test User", "password123")

		urlData := map[string]string{
			"original_url": "https://www.example.com",
		}
		body, _ := json.Marshal(urlData)

		req, _ := http.NewRequest("POST", "/shorten", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Set("user_id", user.ID)

		ShortenURL(c)

		assert.Equal(t, http.StatusCreated, w.Code)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.True(t, response["status"].(bool))

		data := response["data"].(map[string]interface{})
		assert.NotEmpty(t, data["short_code"])
		assert.NotEmpty(t, data["short_url"])
	})

	t.Run("Shorten URL With Guest Session", func(t *testing.T) {
		routes.CleanupTestDB()

		// Create a test session
		session, _ := routes.CreateTestSession()

		urlData := map[string]string{
			"original_url": "https://www.example.com",
		}
		body, _ := json.Marshal(urlData)

		req, _ := http.NewRequest("POST", "/shorten", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Set("session_id", session.ID)

		ShortenURL(c)

		assert.Equal(t, http.StatusCreated, w.Code)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.True(t, response["status"].(bool))
	})

	t.Run("Invalid URL Format", func(t *testing.T) {
		routes.CleanupTestDB()

		user, _ := routes.CreateTestUser("test@example.com", "Test User", "password123")

		urlData := map[string]string{
			"original_url": "not-a-valid-url",
		}
		body, _ := json.Marshal(urlData)

		req, _ := http.NewRequest("POST", "/shorten", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Set("user_id", user.ID)

		ShortenURL(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Missing Original URL", func(t *testing.T) {
		routes.CleanupTestDB()

		user, _ := routes.CreateTestUser("test@example.com", "Test User", "password123")

		urlData := map[string]string{}
		body, _ := json.Marshal(urlData)

		req, _ := http.NewRequest("POST", "/shorten", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Set("user_id", user.ID)

		ShortenURL(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("User Not Found", func(t *testing.T) {
		routes.CleanupTestDB()

		urlData := map[string]string{
			"original_url": "https://www.example.com",
		}
		body, _ := json.Marshal(urlData)

		req, _ := http.NewRequest("POST", "/shorten", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Set("user_id", uint(99999)) // Non-existent user ID

		ShortenURL(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestRedirectURL(t *testing.T) {
	router := setupURLTestRouter()
	router.GET("/redirect/:code", RedirectURL)

	t.Run("Valid Short Code Redirect", func(t *testing.T) {
		routes.CleanupTestDB()

		user, _ := routes.CreateTestUser("test@example.com", "Test User", "password123")
		userID := user.ID
		url, _ := routes.CreateTestURL("https://www.example.com", "abc123", &userID, nil)

		req, _ := http.NewRequest("GET", "/redirect/"+url.ShortCode, nil)
		w := httptest.NewRecorder()

		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Params = gin.Params{{Key: "code", Value: url.ShortCode}}

		RedirectURL(c)

		assert.Equal(t, http.StatusMovedPermanently, w.Code)
		assert.Equal(t, "https://www.example.com", w.Header().Get("Location"))
	})

	t.Run("Non-Existent Short Code", func(t *testing.T) {
		routes.CleanupTestDB()

		req, _ := http.NewRequest("GET", "/redirect/nonexistent", nil)
		w := httptest.NewRecorder()

		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Params = gin.Params{{Key: "code", Value: "nonexistent"}}

		RedirectURL(c)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("Expired URL", func(t *testing.T) {
		routes.CleanupTestDB()

		user, _ := routes.CreateTestUser("test@example.com", "Test User", "password123")
		userID := user.ID

		// Create URL with expired time
		expiredTime := time.Now().Add(-1 * time.Hour)
		url := &models.URL{
			OriginalURL: "https://www.example.com",
			ShortCode:   "expired123",
			UserID:      &userID,
			ExpiresAt:   &expiredTime,
		}
		config.DB.Create(url)

		req, _ := http.NewRequest("GET", "/redirect/expired123", nil)
		w := httptest.NewRecorder()

		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Params = gin.Params{{Key: "code", Value: "expired123"}}

		RedirectURL(c)

		assert.Equal(t, http.StatusGone, w.Code)
	})

	t.Run("Click Tracking", func(t *testing.T) {
		routes.CleanupTestDB()

		user, _ := routes.CreateTestUser("test@example.com", "Test User", "password123")
		userID := user.ID
		url, _ := routes.CreateTestURL("https://www.example.com", "track123", &userID, nil)

		req, _ := http.NewRequest("GET", "/redirect/track123", nil)
		req.Header.Set("User-Agent", "Test Browser")
		w := httptest.NewRecorder()

		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Params = gin.Params{{Key: "code", Value: "track123"}}

		RedirectURL(c)

		// Check that click was recorded
		var clicks []models.Click
		config.DB.Where("url_id = ?", url.ID).Find(&clicks)
		assert.NotEmpty(t, clicks)
		assert.Equal(t, "Test Browser", clicks[0].UserAgent)
	})
}

func TestGetHistory(t *testing.T) {
	router := setupURLTestRouter()
	router.GET("/history", GetHistory)

	t.Run("Get User History", func(t *testing.T) {
		routes.CleanupTestDB()

		user, _ := routes.CreateTestUser("test@example.com", "Test User", "password123")
		userID := user.ID
		routes.CreateTestURL("https://www.example.com", "abc123", &userID, nil)
		routes.CreateTestURL("https://www.google.com", "def456", &userID, nil)

		req, _ := http.NewRequest("GET", "/history", nil)
		w := httptest.NewRecorder()

		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Set("user_id", userID)

		GetHistory(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.True(t, response["status"].(bool))

		data := response["data"].(map[string]interface{})
		history := data["history"].([]interface{})
		assert.Equal(t, 2, len(history))
	})

	t.Run("Get Guest Session History", func(t *testing.T) {
		routes.CleanupTestDB()

		session, _ := routes.CreateTestSession()
		sessionID := session.ID
		routes.CreateTestURL("https://www.example.com", "guest123", nil, &sessionID)

		req, _ := http.NewRequest("GET", "/history", nil)
		w := httptest.NewRecorder()

		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Set("session_id", sessionID)

		GetHistory(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.True(t, response["status"].(bool))

		data := response["data"].(map[string]interface{})
		history := data["history"].([]interface{})
		assert.Equal(t, 1, len(history))
	})

	t.Run("Pagination", func(t *testing.T) {
		routes.CleanupTestDB()

		user, _ := routes.CreateTestUser("test@example.com", "Test User", "password123")
		userID := user.ID

		// Create 15 URLs
		for i := 0; i < 15; i++ {
			routes.CreateTestURL("https://www.example.com", "url"+string(rune(i)), &userID, nil)
		}

		req, _ := http.NewRequest("GET", "/history?page=1&limit=10", nil)
		w := httptest.NewRecorder()

		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Set("user_id", userID)

		GetHistory(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)

		data := response["data"].(map[string]interface{})
		history := data["history"].([]interface{})
		meta := data["meta"].(map[string]interface{})

		assert.LessOrEqual(t, len(history), 10)
		assert.Equal(t, float64(1), meta["page"])
		assert.Equal(t, float64(10), meta["limit"])
	})

	t.Run("Empty History", func(t *testing.T) {
		routes.CleanupTestDB()

		user, _ := routes.CreateTestUser("test@example.com", "Test User", "password123")

		req, _ := http.NewRequest("GET", "/history", nil)
		w := httptest.NewRecorder()

		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Set("user_id", user.ID)

		GetHistory(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)

		data := response["data"].(map[string]interface{})
		history := data["history"].([]interface{})
		assert.Equal(t, 0, len(history))
	})

	t.Run("Unauthorized Without Session", func(t *testing.T) {
		routes.CleanupTestDB()

		req, _ := http.NewRequest("GET", "/history", nil)
		w := httptest.NewRecorder()

		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Set("session_id", uint(0))

		GetHistory(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestDeleteURL(t *testing.T) {
	router := setupURLTestRouter()
	router.DELETE("/:code", DeleteURL)

	t.Run("Delete Own URL As User", func(t *testing.T) {
		routes.CleanupTestDB()

		user, _ := routes.CreateTestUser("test@example.com", "Test User", "password123")
		userID := user.ID
		url, _ := routes.CreateTestURL("https://www.example.com", "delete123", &userID, nil)

		req, _ := http.NewRequest("DELETE", "/"+url.ShortCode, nil)
		w := httptest.NewRecorder()

		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Params = gin.Params{{Key: "code", Value: url.ShortCode}}
		c.Set("user_id", userID)

		DeleteURL(c)

		assert.Equal(t, http.StatusOK, w.Code)

		// Verify URL is deleted
		var deletedURL models.URL
		err := config.DB.Where("short_code = ?", url.ShortCode).First(&deletedURL).Error
		assert.Error(t, err)
	})

	t.Run("Delete Own URL As Guest", func(t *testing.T) {
		routes.CleanupTestDB()

		session, _ := routes.CreateTestSession()
		sessionID := session.ID
		url, _ := routes.CreateTestURL("https://www.example.com", "guest-del", nil, &sessionID)

		req, _ := http.NewRequest("DELETE", "/"+url.ShortCode, nil)
		w := httptest.NewRecorder()

		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Params = gin.Params{{Key: "code", Value: url.ShortCode}}
		c.Set("session_id", sessionID)

		DeleteURL(c)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Cannot Delete Other User's URL", func(t *testing.T) {
		routes.CleanupTestDB()

		user1, _ := routes.CreateTestUser("user1@example.com", "User 1", "password123")
		user2, _ := routes.CreateTestUser("user2@example.com", "User 2", "password123")

		user1ID := user1.ID
		url, _ := routes.CreateTestURL("https://www.example.com", "other123", &user1ID, nil)

		req, _ := http.NewRequest("DELETE", "/"+url.ShortCode, nil)
		w := httptest.NewRecorder()

		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Params = gin.Params{{Key: "code", Value: url.ShortCode}}
		c.Set("user_id", user2.ID)

		DeleteURL(c)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("Delete Non-Existent URL", func(t *testing.T) {
		routes.CleanupTestDB()

		user, _ := routes.CreateTestUser("test@example.com", "Test User", "password123")

		req, _ := http.NewRequest("DELETE", "/nonexistent", nil)
		w := httptest.NewRecorder()

		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Params = gin.Params{{Key: "code", Value: "nonexistent"}}
		c.Set("user_id", user.ID)

		DeleteURL(c)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("Missing Code Parameter", func(t *testing.T) {
		routes.CleanupTestDB()

		user, _ := routes.CreateTestUser("test@example.com", "Test User", "password123")

		req, _ := http.NewRequest("DELETE", "/", nil)
		w := httptest.NewRecorder()

		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Params = gin.Params{{Key: "code", Value: ""}}
		c.Set("user_id", user.ID)

		DeleteURL(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Cascade Delete Clicks", func(t *testing.T) {
		routes.CleanupTestDB()

		user, _ := routes.CreateTestUser("test@example.com", "Test User", "password123")
		userID := user.ID
		url, _ := routes.CreateTestURL("https://www.example.com", "cascade123", &userID, nil)

		// Create some clicks
		click := models.Click{
			URLID:     url.ID,
			IP:        "127.0.0.1",
			UserAgent: "Test",
		}
		config.DB.Create(&click)

		req, _ := http.NewRequest("DELETE", "/cascade123", nil)
		w := httptest.NewRecorder()

		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Params = gin.Params{{Key: "code", Value: "cascade123"}}
		c.Set("user_id", userID)

		DeleteURL(c)

		assert.Equal(t, http.StatusOK, w.Code)

		// Verify clicks are deleted
		var clicks []models.Click
		config.DB.Where("url_id = ?", url.ID).Find(&clicks)
		assert.Empty(t, clicks)
	})
}
