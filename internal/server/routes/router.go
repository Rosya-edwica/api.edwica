package routes

import (
	"github.com/Rosya-edwica/api.edwica/internal/controllers"

	"github.com/gin-gonic/gin"
)

func ConfigRoutes(router *gin.Engine) *gin.Engine {
	main := router.Group("api/v1")
	{
		videos := main.Group("videos")
		{
			videos.GET("/", controllers.GetVideos)
		}
		books := main.Group("books")
		{
			books.GET("/", controllers.GetBooks)
		}
	}
	return router
}
