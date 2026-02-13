package routes

import (
	"github.com/gin-gonic/gin"
	"url-shortener/services"
)

func PingRoutes(r *gin.RouterGroup) {
	r.GET("/ping", services.PingService)
}
