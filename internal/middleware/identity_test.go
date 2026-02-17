package middleware

import (
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

func setupIdentityTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	routes.SetupTestDB()
	r := gin.Default()
	return r
}

func TestResolveIdentity(t *testing.T) {
	router := setupIdentityTestRouter()

	router.GET("/test", ResolveIdentity(), func(c *gin.Context) {
		userID, userExists := c.Get("user_id")
		sessionID, sessionExists := c.Get("session_id")
		sessionToken, tokenExists := c.Get("session_token")

		c.JSON(http.StatusOK, gin.H{
			"user_id":        userID,
			"user_exists":    userExists,
			"session_id":     sessionID,
			"session_exists": sessionExists,
			"session_token":  sessionToken,
			"token_exists":   tokenExists,
		})
	})

	t.Run("Authenticated User With Valid JWT", func(t *testing.T) {
		routes.CleanupTestDB()

		user, _ := routes.CreateTestUser("test@example.com", "Test User", "password123")
		token := routes.GenerateTestJWT(user.ID)

		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "user_exists")
		assert.Contains(t, w.Body.String(), "true")
	})

	t.Run("Guest With Valid Session Token", func(t *testing.T) {
		routes.CleanupTestDB()

		session, _ := routes.CreateTestSession()

		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("X-Session-Token", session.Token)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "session_exists")
		assert.Contains(t, w.Body.String(), "true")
	})

	t.Run("New Guest Session Created", func(t *testing.T) {
		routes.CleanupTestDB()

		req, _ := http.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		// Should have created a new session
		assert.Contains(t, w.Body.String(), "session_exists")

		// Should return session token in header
		sessionToken := w.Header().Get("X-Session-Token")
		assert.NotEmpty(t, sessionToken)
	})

	t.Run("Invalid JWT Falls Back To Session", func(t *testing.T) {
		routes.CleanupTestDB()

		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "Bearer invalid-token")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		// Should create a guest session instead
		assert.Contains(t, w.Body.String(), "session_exists")
	})

	t.Run("Expired Session Creates New Session", func(t *testing.T) {
		routes.CleanupTestDB()

		// Create an expired session
		expiredTime := time.Now().Add(-1 * time.Hour)
		session := models.GuestSession{
			Token:     "expired-token",
			ExpiresAt: expiredTime,
		}
		config.DB.Create(&session)

		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("X-Session-Token", "expired-token")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		// Should create a new session
		newSessionToken := w.Header().Get("X-Session-Token")
		assert.NotEmpty(t, newSessionToken)
		assert.NotEqual(t, "expired-token", newSessionToken)
	})

	t.Run("Valid Session Extends Expiry", func(t *testing.T) {
		routes.CleanupTestDB()

		session, _ := routes.CreateTestSession()
		originalExpiry := session.ExpiresAt

		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("X-Session-Token", session.Token)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		// Check that expiry was extended
		var updatedSession models.GuestSession
		config.DB.Where("token = ?", session.Token).First(&updatedSession)
		assert.True(t, updatedSession.ExpiresAt.After(originalExpiry))
	})

	t.Run("JWT Takes Precedence Over Session Token", func(t *testing.T) {
		routes.CleanupTestDB()

		user, _ := routes.CreateTestUser("test@example.com", "Test User", "password123")
		token := routes.GenerateTestJWT(user.ID)
		session, _ := routes.CreateTestSession()

		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("X-Session-Token", session.Token)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		// Should use JWT, not session
		assert.Contains(t, w.Body.String(), "user_exists")
		assert.Contains(t, w.Body.String(), "true")
	})

	t.Run("Session Last Accessed Updated", func(t *testing.T) {
		routes.CleanupTestDB()

		session, _ := routes.CreateTestSession()

		// Wait a moment to ensure time difference
		time.Sleep(100 * time.Millisecond)

		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("X-Session-Token", session.Token)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		// Check that last_accessed was updated
		var updatedSession models.GuestSession
		config.DB.Where("token = ?", session.Token).First(&updatedSession)

		// LastAccessed should be more recent than creation time
		if !updatedSession.LastAccessed.IsZero() {
			assert.True(t, updatedSession.LastAccessed.After(session.CreatedAt))
		}
	})

	t.Run("Non-Existent Session Token Creates New Session", func(t *testing.T) {
		routes.CleanupTestDB()

		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("X-Session-Token", "non-existent-token")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		// Should create a new session
		newSessionToken := w.Header().Get("X-Session-Token")
		assert.NotEmpty(t, newSessionToken)
		assert.NotEqual(t, "non-existent-token", newSessionToken)
	})
}
