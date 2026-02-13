package routes

import (
	"url-shortener/middleware"
	"url-shortener/services"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.RouterGroup) {

	r.POST("/users", services.RegisterUser)
	r.POST("/login", services.LoginUser)
	r.GET("/users", middleware.AuthRequired(), services.GetUsers)
}
