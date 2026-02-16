package handler

import (
	"url-shortener/internal/middleware"
	"url-shortener/internal/service"

	"github.com/gin-gonic/gin"
)

func URLRoutes(r *gin.RouterGroup) {
	u := r.Group("/url")
	u.POST("/shorten", middleware.ResolveIdentity(), service.ShortenURL)
	u.GET("/history", middleware.ResolveIdentity(), service.GetHistory)
	u.GET("/redirect/:code", middleware.ResolveIdentity(), service.RedirectURL)
	u.DELETE("/:code", middleware.ResolveIdentity(), service.DeleteURL)

}
