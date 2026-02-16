package middleware

import (
	"net/http"
	"os"
	"strings"
	"time"
	"url-shortener/internal/config"
	"url-shortener/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return []byte(os.Getenv("JWT_SECRET")), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			c.Set("token_claims", claims)
		}
		c.Next()
	}
}

func GenerateTokens(userID uint) (string, string) {
	accessClaims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(15 * time.Minute).Unix(),
	}

	refreshClaims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(7 * 24 * time.Hour).Unix(),
	}

	accessToken, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims).
		SignedString([]byte(os.Getenv("JWT_SECRET")))

	refreshToken, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).
		SignedString([]byte(os.Getenv("JWT_REFRESH_SECRET")))

	return accessToken, refreshToken
}

func LinkSessionToUser(c *gin.Context, userID uint) {
	sessionToken := c.GetHeader("X-Session-Token")
	if sessionToken == "" {
		return
	}

	// update urls
	config.DB.Model(&models.URL{}).
		Where("session_token = ?", sessionToken).
		Update("user_id", userID)

	// clear session token
	config.DB.Model(&models.URL{}).
		Where("session_token = ?", sessionToken).
		Update("session_token", nil)
}

func RefreshToken(c *gin.Context) {
	var input struct {
		RefreshToken string `json:"refresh_token"`
	}
	c.ShouldBindJSON(&input)

	token, err := jwt.Parse(input.RefreshToken, func(t *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_REFRESH_SECRET")), nil
	})

	if err != nil || !token.Valid {
		c.JSON(401, gin.H{"error": "invalid refresh token"})
		return
	}

	claims := token.Claims.(jwt.MapClaims)
	userID := uint(claims["user_id"].(float64))

	access, refresh := GenerateTokens(userID)
	c.JSON(200, gin.H{
		"access_token":  access,
		"refresh_token": refresh,
	})
}
