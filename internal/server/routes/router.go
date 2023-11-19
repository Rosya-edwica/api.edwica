package routes

import (
	"github.com/Rosya-edwica/api.edwica/internal/controllers/book"
	"github.com/Rosya-edwica/api.edwica/internal/controllers/vacancy"
	"github.com/Rosya-edwica/api.edwica/internal/controllers/video"

	"github.com/gin-gonic/gin"
)

func ConfigRoutes(router *gin.Engine) *gin.Engine {
	main := router.Group("api/v1")
	{
		videos := main.Group("videos")
		{
			videos.GET("/", video.GetVideos)
		}
		books := main.Group("books")
		{
			books.GET("/", book.GetBooks)
		}
		vacancies := main.Group("vacancies")
		{
			vacancies.GET("/", vacancy.GetVacancies)
		}
	}
	return router
}
