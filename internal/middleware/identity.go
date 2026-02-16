package middleware

import (
	"os"
	"strings"
	"time"

	"url-shortener/internal/config"
	"url-shortener/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func ResolveIdentity() gin.HandlerFunc {
	return func(c *gin.Context) {

		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				return []byte(os.Getenv("JWT_SECRET")), nil
			})

			if err == nil && token.Valid {
				if claims, ok := token.Claims.(jwt.MapClaims); ok {
					c.Set("user_id", uint(claims["user_id"].(float64)))
					c.Next()
					return
				}
			}
		}
		sessionToken := c.GetHeader("X-Session-Token")
		if sessionToken != "" {
			var session models.GuestSession
			err := config.DB.
				Where("token = ?", sessionToken).
				First(&session).Error

			if err == nil && !session.ExpiresAt.IsZero() && time.Now().Before(session.ExpiresAt) {

				newExpiresAt := time.Now().Add(30 * 24 * time.Hour)
				now := time.Now()

				config.DB.Model(&session).Updates(map[string]interface{}{
					"last_accessed": now,
					"expires_at":    newExpiresAt,
				})

				c.Set("session_id", session.ID)
				c.Set("session_token", session.Token)
				c.Next()
				return
			}
		}

		token := uuid.NewString()
		expiresAt := time.Now().Add(30 * 24 * time.Hour)
		session := models.GuestSession{
			Token:     token,
			ExpiresAt: expiresAt,
		}
		config.DB.Create(&session)

		c.Set("session_id", session.ID)
		c.Set("session_token", token)

		c.Header("X-Session-Token", token)
		c.Next()
	}
}
