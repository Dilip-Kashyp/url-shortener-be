package routes

import (
	"url-shortener/middleware"
	"url-shortener/services"

	"github.com/gin-gonic/gin"
)

func URLRoutes(r *gin.RouterGroup) {

	r.POST("/shorten", middleware.ResolveIdentity(), services.ShortenURL)
	r.GET("/history", middleware.ResolveIdentity(), services.GetHistory)
	r.GET("/:shortCode",middleware.ResolveIdentity(), services.RedirectURL)
	r.DELETE("/:shortCode",middleware.ResolveIdentity(), services.DeleteURL)

}
