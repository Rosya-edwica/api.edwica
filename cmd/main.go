package main

import (
	"log"

	"github.com/Rosya-edwica/api.edwica/config"
	"github.com/Rosya-edwica/api.edwica/internal/database"
	"github.com/Rosya-edwica/api.edwica/internal/server"
)

func main() {
	dbConfig, err := config.LoadDBConfig()
	if err != nil {
		log.Fatal(err)
	}
	httpConfig, err := config.LoadHTTPConfig()
	if err != nil {
		log.Fatal(err)
	}
	_, err = database.New(dbConfig)
	if err != nil {
		log.Fatal(err)
	}

	server := server.NewServer(httpConfig)
	server.Run()
}
