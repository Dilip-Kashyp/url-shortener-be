package services

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"url-shortener/utils"
)

func PingService(c *gin.Context) {
	c.JSON(http.StatusOK, utils.ResponseSuccess(gin.H{
		"message": "pong",
	}))
}
