package services

import (
	"net/http"
	"time"
	"url-shortener/config"
	"url-shortener/models"
	"url-shortener/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func ShortenURL(c *gin.Context) {
	var input struct {
		OriginalURL string `json:"original_url" validate:"required,url"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, utils.ResponseError(err.Error()))
		return
	}

	shortCode := utils.GenerateShortCode()

	url := models.URL{
		OriginalURL: input.OriginalURL,
		ShortCode:   shortCode,
	}

	if userID, ok := c.Get("user_id"); ok {
		userIDValue := userID.(uint)
		// Validate that the user exists before associating the URL
		var user models.User
		if err := config.DB.First(&user, userIDValue).Error; err != nil {
			c.JSON(http.StatusUnauthorized, utils.ResponseError("user not found"))
			return
		}
		url.UserID = &userIDValue
	} else {
		sessionID := c.GetUint("session_id")
		url.SessionID = &sessionID
	}

	if err := config.DB.Create(&url).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.ResponseError(err.Error()))
		return
	}

	c.JSON(http.StatusCreated, utils.ResponseSuccess(gin.H{
		"short_code": shortCode,
		"short_url":  "http://localhost:8080/api/v1/" + shortCode,
	}))
}

func RedirectURL(c *gin.Context) {
	shortCode := c.Param("shortCode")

	// Check cache
	cachedURL, err := config.RedisClient.Get(config.RedisClient.Context(), shortCode).Result()
	if err == nil {
		// Find URL first to get ID for click record
		var url models.URL
		if err := config.DB.Where("short_code = ?", shortCode).First(&url).Error; err != nil {
			c.JSON(http.StatusNotFound, utils.ResponseError("URL not found"))
			return
		}
		if url.ExpiresAt != nil && time.Now().After(*url.ExpiresAt) {
			c.JSON(http.StatusGone, utils.ResponseError("URL expired"))
			return
		}

		// Create click record
		ip := c.ClientIP()
		userAgent := c.GetHeader("User-Agent")
		click := models.Click{
			URLID:     url.ID,
			IP:        ip,
			UserAgent: userAgent,
		}
		if err := config.DB.Create(&click).Error; err != nil {
			// Log error but don't fail the redirect
		}
		config.DB.Model(&url).Update("clicks", gorm.Expr("clicks + 1"))
		c.Redirect(http.StatusMovedPermanently, cachedURL)
		return
	}

	var url models.URL
	if err := config.DB.Where("short_code = ?", shortCode).First(&url).Error; err != nil {
		c.JSON(http.StatusNotFound, utils.ResponseError("URL not found"))
		return
	}

	if url.ExpiresAt != nil && time.Now().After(*url.ExpiresAt) {
		c.JSON(http.StatusGone, utils.ResponseError("URL expired"))
		return
	}

	// Create click record
	ip := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")
	click := models.Click{
		URLID:     url.ID,
		IP:        ip,
		UserAgent: userAgent,
	}
	if err := config.DB.Create(&click).Error; err != nil {
		// Log error but don't fail the redirect
	}

	// Cache
	config.RedisClient.Set(config.RedisClient.Context(), shortCode, url.OriginalURL, time.Hour)

	config.DB.Model(&url).Update("clicks", url.Clicks+1)
	c.Redirect(http.StatusMovedPermanently, url.OriginalURL)
}

func GetHistory(c *gin.Context) {
	var urls []models.URL

	page := utils.ParseInt(c.DefaultQuery("page", "1"))
	limit := utils.ParseInt(c.DefaultQuery("limit", "10"))

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}

	offset := (page - 1) * limit

	query := config.DB.Model(&models.URL{})

	if userID, ok := c.Get("user_id"); ok {
		query = query.Where("user_id = ?", userID)
	} else {
		sessionID := c.GetUint("session_id")
		if sessionID == 0 {
			c.JSON(http.StatusUnauthorized, utils.ResponseError("unauthorized"))
			return
		}
		query = query.Where("session_id = ?", sessionID)
	}

	if err := query.
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&urls).Error; err != nil {

		c.JSON(http.StatusInternalServerError, utils.ResponseError(err.Error()))
		return
	}

	history := make([]models.HistoryItem, 0, len(urls))
	for _, u := range urls {
		history = append(history, models.HistoryItem{
			ID:          u.ID,
			OriginalURL: u.OriginalURL,
			ShortCode:   u.ShortCode,
			ShortURL:    "http://localhost:8080/api/v1/" + u.ShortCode,
			Clicks:      u.Clicks,
			ExpiresAt:   u.ExpiresAt,
			CreatedAt:   u.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, utils.ResponseSuccess(gin.H{
		"history": history,
		"meta": gin.H{
			"page":  page,
			"limit": limit,
			"count": len(history),
		},
	}))
}

func DeleteURL(c *gin.Context) {
	shortCode := c.Param("shortCode")
	var url models.URL
	query := config.DB.Where("short_code = ?", shortCode)

	if userID, ok := c.Get("user_id"); ok {
		query = query.Where("user_id = ?", userID)
	} else {
		sessionID := c.GetUint("session_id")
		if sessionID == 0 {
			c.JSON(http.StatusUnauthorized, utils.ResponseError("unauthorized"))
			return
		}
		query = query.Where("session_id = ?", sessionID)
	}

	if err := query.First(&url).Error; err != nil {
		c.JSON(http.StatusNotFound, utils.ResponseError("URL not found"))
		return
	}

	if err := config.DB.Where("url_id = ?", url.ID).Delete(&models.Click{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.ResponseError(err.Error()))
		return
	}

	if err := config.DB.Delete(&url).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.ResponseError(err.Error()))
		return
	}

	config.RedisClient.Del(config.RedisClient.Context(), shortCode)

	c.JSON(http.StatusOK, utils.ResponseSuccess(gin.H{
		"message": "URL deleted successfully",
	}))
}
