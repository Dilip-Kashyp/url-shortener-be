package routes

import (
	"os"
	"time"
	"url-shortener/internal/config"
	"url-shortener/internal/models"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// SetupTestDB initializes an in-memory SQLite database for testing
func SetupTestDB() {
	// Disable GORM logger to reduce noise in tests
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		panic("failed to connect to test database: " + err.Error())
	}

	// Override the global DB connection
	config.DB = db

	// Run migrations
	err = config.DB.AutoMigrate(&models.User{}, &models.URL{}, &models.GuestSession{}, &models.Click{})
	if err != nil {
		panic("failed to migrate test database: " + err.Error())
	}

	// Setup test environment variables
	if os.Getenv("JWT_SECRET") == "" {
		os.Setenv("JWT_SECRET", "test-secret-key-for-testing")
	}
	if os.Getenv("SERVER_URL") == "" {
		os.Setenv("SERVER_URL", "http://localhost:8080")
	}
	if os.Getenv("REDIS_ADDR") == "" {
		os.Setenv("REDIS_ADDR", "localhost:6379")
	}
	if os.Getenv("REDIS_PASSWORD") == "" {
		os.Setenv("REDIS_PASSWORD", "")
	}

	// Initialize Redis client (will fail silently if Redis is not available)
	// This prevents nil pointer errors in tests
	config.ConnectRedis()
}

// CreateTestUser creates a test user in the database
func CreateTestUser(email, name, password string) (*models.User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		Email:    email,
		Name:     name,
		Password: string(hashedPassword),
	}

	if err := config.DB.Create(user).Error; err != nil {
		return nil, err
	}

	return user, nil
}

// GenerateTestJWT generates a JWT token for testing
func GenerateTestJWT(userID uint) string {
	// Set a default JWT secret if not set
	if os.Getenv("JWT_SECRET") == "" {
		os.Setenv("JWT_SECRET", "test-secret-key")
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	})

	tokenString, _ := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	return tokenString
}

// CreateTestSession creates a test guest session
func CreateTestSession() (*models.GuestSession, error) {
	token := uuid.NewString()
	expiresAt := time.Now().Add(30 * 24 * time.Hour)

	session := &models.GuestSession{
		Token:     token,
		ExpiresAt: expiresAt,
	}

	if err := config.DB.Create(session).Error; err != nil {
		return nil, err
	}

	return session, nil
}

// CreateTestURL creates a test URL in the database
func CreateTestURL(originalURL, shortCode string, userID *uint, sessionID *uint) (*models.URL, error) {
	url := &models.URL{
		OriginalURL: originalURL,
		ShortCode:   shortCode,
		UserID:      userID,
		SessionID:   sessionID,
		Clicks:      0,
	}

	if err := config.DB.Create(url).Error; err != nil {
		return nil, err
	}

	return url, nil
}

// CleanupTestDB clears all data from the test database
func CleanupTestDB() {
	config.DB.Exec("DELETE FROM clicks")
	config.DB.Exec("DELETE FROM urls")
	config.DB.Exec("DELETE FROM guest_sessions")
	config.DB.Exec("DELETE FROM users")
}


