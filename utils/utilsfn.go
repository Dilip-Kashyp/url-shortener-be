package utils

import (
	"crypto/rand"
	"encoding/base64"
	"strconv"
	"url-shortener/config"
	"url-shortener/models"

	"time"
)

func GenerateShortCode() string {
	bytes := make([]byte, 6)
	rand.Read(bytes)
	return base64.URLEncoding.EncodeToString(bytes)[:8]
}

func ParseInt(s string) int {
	i, _ := strconv.Atoi(s)
	return i
}

func ResponseSuccess(data interface{}) models.APIResponse {
	return models.APIResponse{
		Status: true,
		Data:   data,
	}
}

func ResponseError(message string) models.APIResponse {
	return models.APIResponse{
		Status:  false,
		Message: message,
	}
}

func CleanupExpiredData() {
	config.DB.Where("expires_at IS NOT NULL AND expires_at < ?", time.Now()).Delete(&models.URL{})

	config.DB.Where("expires_at < ?", time.Now()).Delete(&models.GuestSession{})

	oneYearAgo := time.Now().AddDate(-1, 0, 0)
	config.DB.Where("created_at < ?", oneYearAgo).Delete(&models.Click{})
}
