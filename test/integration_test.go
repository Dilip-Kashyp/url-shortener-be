package routes

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"url-shortener/internal/handler"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupIntegrationRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	SetupTestDB()

	r := gin.Default()

	api := r.Group("/api")
	handler.RegisterRoutes(api)
	handler.URLRoutes(api)
	handler.PingRoutes(api)

	return r
}

func TestCompleteUserFlow(t *testing.T) {
	router := setupIntegrationRouter()
	CleanupTestDB()

	var authToken string
	var shortCode string

	t.Run("1. Register New User", func(t *testing.T) {
		userData := map[string]string{
			"email":    "integration@example.com",
			"name":     "Integration Test User",
			"password": "password123",
		}
		body, _ := json.Marshal(userData)

		req, _ := http.NewRequest("POST", "/api/user/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("2. Login User", func(t *testing.T) {
		loginData := map[string]string{
			"email":    "integration@example.com",
			"password": "password123",
		}
		body, _ := json.Marshal(loginData)

		req, _ := http.NewRequest("POST", "/api/user/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		data := response["data"].(map[string]interface{})
		authToken = data["token"].(string)

		assert.NotEmpty(t, authToken)
	})

	t.Run("3. Create Shortened URL", func(t *testing.T) {
		urlData := map[string]string{
			"original_url": "https://www.example.com/very/long/url/path",
		}
		body, _ := json.Marshal(urlData)

		req, _ := http.NewRequest("POST", "/api/url/shorten", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+authToken)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		data := response["data"].(map[string]interface{})
		shortCode = data["short_code"].(string)

		assert.NotEmpty(t, shortCode)
	})

	t.Run("4. Redirect Using Short Code", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/url/redirect/"+shortCode, nil)
		req.Header.Set("Authorization", "Bearer "+authToken)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusMovedPermanently, w.Code)
		assert.Contains(t, w.Header().Get("Location"), "example.com")
	})

	t.Run("5. Get URL History", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/url/history", nil)
		req.Header.Set("Authorization", "Bearer "+authToken)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		data := response["data"].(map[string]interface{})
		history := data["history"].([]interface{})

		assert.NotEmpty(t, history)
	})

	t.Run("6. Delete URL", func(t *testing.T) {
		req, _ := http.NewRequest("DELETE", "/api/url/"+shortCode, nil)
		req.Header.Set("Authorization", "Bearer "+authToken)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("7. Verify URL Deleted", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/url/redirect/"+shortCode, nil)
		req.Header.Set("Authorization", "Bearer "+authToken)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestGuestSessionFlow(t *testing.T) {
	router := setupIntegrationRouter()
	CleanupTestDB()

	var sessionToken string
	var shortCode string

	t.Run("1. Create URL As Guest", func(t *testing.T) {
		urlData := map[string]string{
			"original_url": "https://www.example.com/guest/url",
		}
		body, _ := json.Marshal(urlData)

		req, _ := http.NewRequest("POST", "/api/url/shorten", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		// Get session token from response header
		sessionToken = w.Header().Get("X-Session-Token")
		assert.NotEmpty(t, sessionToken)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		data := response["data"].(map[string]interface{})
		shortCode = data["short_code"].(string)
	})

	t.Run("2. Get Guest History With Session Token", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/url/history", nil)
		req.Header.Set("X-Session-Token", sessionToken)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		data := response["data"].(map[string]interface{})
		history := data["history"].([]interface{})

		assert.NotEmpty(t, history)
	})

	t.Run("3. Delete Guest URL", func(t *testing.T) {
		req, _ := http.NewRequest("DELETE", "/api/url/"+shortCode, nil)
		req.Header.Set("X-Session-Token", sessionToken)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestAuthenticationFlow(t *testing.T) {
	router := setupIntegrationRouter()
	CleanupTestDB()

	t.Run("Cannot Access Protected Route Without Auth", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/user/get-user", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("Can Access Protected Route With Valid Token", func(t *testing.T) {
		// Create and login user
		CreateTestUser("protected@example.com", "Protected User", "password123")

		loginData := map[string]string{
			"email":    "protected@example.com",
			"password": "password123",
		}
		body, _ := json.Marshal(loginData)

		loginReq, _ := http.NewRequest("POST", "/api/user/login", bytes.NewBuffer(body))
		loginReq.Header.Set("Content-Type", "application/json")
		loginW := httptest.NewRecorder()

		router.ServeHTTP(loginW, loginReq)

		var loginResponse map[string]interface{}
		json.Unmarshal(loginW.Body.Bytes(), &loginResponse)
		data := loginResponse["data"].(map[string]interface{})
		token := data["token"].(string)

		// Access protected route
		req, _ := http.NewRequest("GET", "/api/user/get-user", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestURLOwnershipAndAuthorization(t *testing.T) {
	router := setupIntegrationRouter()
	CleanupTestDB()

	var user1Token, user2Token string
	var user1ShortCode string

	// Setup: Create two users
	t.Run("Setup Users", func(t *testing.T) {
		// User 1
		CreateTestUser("user1@example.com", "User 1", "password123")
		loginData1 := map[string]string{
			"email":    "user1@example.com",
			"password": "password123",
		}
		body1, _ := json.Marshal(loginData1)
		req1, _ := http.NewRequest("POST", "/api/user/login", bytes.NewBuffer(body1))
		req1.Header.Set("Content-Type", "application/json")
		w1 := httptest.NewRecorder()
		router.ServeHTTP(w1, req1)

		var response1 map[string]interface{}
		json.Unmarshal(w1.Body.Bytes(), &response1)
		data1 := response1["data"].(map[string]interface{})
		user1Token = data1["token"].(string)

		// User 2
		CreateTestUser("user2@example.com", "User 2", "password123")
		loginData2 := map[string]string{
			"email":    "user2@example.com",
			"password": "password123",
		}
		body2, _ := json.Marshal(loginData2)
		req2, _ := http.NewRequest("POST", "/api/user/login", bytes.NewBuffer(body2))
		req2.Header.Set("Content-Type", "application/json")
		w2 := httptest.NewRecorder()
		router.ServeHTTP(w2, req2)

		var response2 map[string]interface{}
		json.Unmarshal(w2.Body.Bytes(), &response2)
		data2 := response2["data"].(map[string]interface{})
		user2Token = data2["token"].(string)
	})

	t.Run("User 1 Creates URL", func(t *testing.T) {
		urlData := map[string]string{
			"original_url": "https://www.example.com/user1/url",
		}
		body, _ := json.Marshal(urlData)

		req, _ := http.NewRequest("POST", "/api/url/shorten", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+user1Token)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		data := response["data"].(map[string]interface{})
		user1ShortCode = data["short_code"].(string)
	})

	t.Run("User 2 Cannot Delete User 1's URL", func(t *testing.T) {
		req, _ := http.NewRequest("DELETE", "/api/url/"+user1ShortCode, nil)
		req.Header.Set("Authorization", "Bearer "+user2Token)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("User 1 Can Delete Own URL", func(t *testing.T) {
		req, _ := http.NewRequest("DELETE", "/api/url/"+user1ShortCode, nil)
		req.Header.Set("Authorization", "Bearer "+user1Token)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestPingEndpoint(t *testing.T) {
	router := setupIntegrationRouter()

	t.Run("Ping Endpoint Returns Pong", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/test/ping", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "pong")
	})
}
