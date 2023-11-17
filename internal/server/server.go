package server

import (
	"github.com/Rosya-edwica/api.edwica/internal/server/routes"
	"log"

	"github.com/gin-gonic/gin"
)

type Config struct {
	Port string `yaml:"port" env-required:"true"`
}

type Server struct {
	port   string
	server *gin.Engine
}

func NewServer(cfg *Config) Server {
	return Server{
		port:   cfg.Port,
		server: gin.Default(),
	}
}

func (s *Server) Run() {
	router := routes.ConfigRoutes(s.server)
	log.Printf("Server running at port: %v", s.port)
	log.Fatal(router.Run(":" + s.port))
}
