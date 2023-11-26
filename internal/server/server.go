package server

import (
	"fmt"
	"log"

	"github.com/Rosya-edwica/api.edwica/config"
	"github.com/Rosya-edwica/api.edwica/internal/server/routes"
	"github.com/Rosya-edwica/api.edwica/pkg/logger"

	"github.com/gin-gonic/gin"
)

type Server struct {
	port   string
	server *gin.Engine
}

func NewServer(cfg *config.Server) Server {
	return Server{
		port:   cfg.Port,
		server: gin.Default(),
	}
}

func (s *Server) Run() {
	router := routes.ConfigRoutes(s.server)
	logger.Log.Info(fmt.Sprintf("Server running at port: %v", s.port))
	log.Fatal(router.Run(":" + s.port))
}
