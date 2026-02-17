package service

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"url-shortener/internal/config"
	"url-shortener/internal/models"
	routes "url-shortener/test"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupUserTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	routes.SetupTestDB()
	r := gin.Default()
	return r
}

func TestRegisterUser(t *testing.T) {
	router := setupUserTestRouter()
	router.POST("/register", RegisterUser)

	t.Run("Valid Registration", func(t *testing.T) {
		routes.CleanupTestDB()

		user := map[string]string{
			"email":    "test@example.com",
			"name":     "Test User",
			"password": "password123",
		}
		body, _ := json.Marshal(user)

		req, _ := http.NewRequest("POST", "/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.True(t, response["status"].(bool))
	})

	t.Run("Duplicate Email", func(t *testing.T) {
		routes.CleanupTestDB()

		// Create first user
		routes.CreateTestUser("test@example.com", "Test User", "password123")

		// Try to register with same email
		user := map[string]string{
			"email":    "test@example.com",
			"name":     "Another User",
			"password": "password456",
		}
		body, _ := json.Marshal(user)

		req, _ := http.NewRequest("POST", "/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusConflict, w.Code)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.False(t, response["status"].(bool))
		assert.Contains(t, response["message"], "email already registered")
	})

	t.Run("Invalid Email Format", func(t *testing.T) {
		routes.CleanupTestDB()

		user := map[string]string{
			"email":    "invalid-email",
			"name":     "Test User",
			"password": "password123",
		}
		body, _ := json.Marshal(user)

		req, _ := http.NewRequest("POST", "/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Missing Required Fields", func(t *testing.T) {
		routes.CleanupTestDB()

		user := map[string]string{
			"email": "test@example.com",
			// Missing name and password
		}
		body, _ := json.Marshal(user)

		req, _ := http.NewRequest("POST", "/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Password Too Short", func(t *testing.T) {
		routes.CleanupTestDB()

		user := map[string]string{
			"email":    "test@example.com",
			"name":     "Test User",
			"password": "12345", // Less than 6 characters
		}
		body, _ := json.Marshal(user)

		req, _ := http.NewRequest("POST", "/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestLoginUser(t *testing.T) {
	router := setupUserTestRouter()
	router.POST("/login", LoginUser)

	t.Run("Valid Login", func(t *testing.T) {
		routes.CleanupTestDB()

		// Create a test user
		routes.CreateTestUser("test@example.com", "Test User", "password123")

		loginData := map[string]string{
			"email":    "test@example.com",
			"password": "password123",
		}
		body, _ := json.Marshal(loginData)

		req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.True(t, response["status"].(bool))

		data := response["data"].(map[string]interface{})
		assert.NotEmpty(t, data["token"])
	})

	t.Run("Invalid Email", func(t *testing.T) {
		routes.CleanupTestDB()

		loginData := map[string]string{
			"email":    "nonexistent@example.com",
			"password": "password123",
		}
		body, _ := json.Marshal(loginData)

		req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.False(t, response["status"].(bool))
		assert.Contains(t, response["message"], "invalid credentials")
	})

	t.Run("Invalid Password", func(t *testing.T) {
		routes.CleanupTestDB()

		// Create a test user
		routes.CreateTestUser("test@example.com", "Test User", "password123")

		loginData := map[string]string{
			"email":    "test@example.com",
			"password": "wrongpassword",
		}
		body, _ := json.Marshal(loginData)

		req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.False(t, response["status"].(bool))
		assert.Contains(t, response["message"], "invalid credentials")
	})

	t.Run("Missing Fields", func(t *testing.T) {
		routes.CleanupTestDB()

		loginData := map[string]string{
			"email": "test@example.com",
			// Missing password
		}
		body, _ := json.Marshal(loginData)

		req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestGetUsers(t *testing.T) {
	router := setupUserTestRouter()
	router.GET("/users", GetUsers)

	t.Run("Get User With Valid Token", func(t *testing.T) {
		routes.CleanupTestDB()

		// Create a test user
		user, _ := routes.CreateTestUser("test@example.com", "Test User", "password123")
		token := routes.GenerateTestJWT(user.ID)

		req, _ := http.NewRequest("GET", "/users", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		// Manually set user_id in context for this test
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Set("user_id", user.ID)

		GetUsers(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.True(t, response["status"].(bool))

		data := response["data"].([]interface{})
		assert.NotEmpty(t, data)
	})

	t.Run("Get User Without Token", func(t *testing.T) {
		routes.CleanupTestDB()

		req, _ := http.NewRequest("GET", "/users", nil)
		w := httptest.NewRecorder()

		c, _ := gin.CreateTestContext(w)
		c.Request = req

		GetUsers(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("Password Not Returned", func(t *testing.T) {
		routes.CleanupTestDB()

		// Create a test user
		user, _ := routes.CreateTestUser("test@example.com", "Test User", "password123")

		req, _ := http.NewRequest("GET", "/users", nil)
		w := httptest.NewRecorder()

		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Set("user_id", user.ID)

		GetUsers(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)

		data := response["data"].([]interface{})
		if len(data) > 0 {
			userData := data[0].(map[string]interface{})
			// Password should be empty string
			assert.Equal(t, "", userData["Password"])
		}
	})
}

func TestPasswordHashing(t *testing.T) {
	routes.SetupTestDB()
	routes.CleanupTestDB()

	t.Run("Password Is Hashed In Database", func(t *testing.T) {
		plainPassword := "password123"
		user, _ := routes.CreateTestUser("test@example.com", "Test User", plainPassword)

		// Retrieve user from database
		var dbUser models.User
		config.DB.First(&dbUser, user.ID)

		// Password should not be stored in plain text
		assert.NotEqual(t, plainPassword, dbUser.Password)
		assert.NotEmpty(t, dbUser.Password)
	})
}
