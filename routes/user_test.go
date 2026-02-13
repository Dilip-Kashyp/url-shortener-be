package routes

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"url-shortener/config"
	"url-shortener/models"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB() {
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	config.DB = db
	config.DB.AutoMigrate(&models.User{})
}

func TestRegisterUser(t *testing.T) {
	setupTestDB()
	r := gin.Default()
	RegisterRoutes(r)

	userJSON := `{"email":"test@example.com","name":"Test","password":"password123"}`
	req, _ := http.NewRequest("POST", "/users", bytes.NewBufferString(userJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestLoginUser(t *testing.T) {
	setupTestDB()
	r := gin.Default()
	RegisterRoutes(r)

	// First register
	userJSON := `{"email":"test@example.com","name":"Test","password":"password123"}`
	req, _ := http.NewRequest("POST", "/users", bytes.NewBufferString(userJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Then login
	loginJSON := `{"email":"test@example.com","password":"password123"}`
	req, _ = http.NewRequest("POST", "/login", bytes.NewBufferString(loginJSON))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "token")
}