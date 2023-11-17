package main

import (
	"fmt"
	"log"

	"github.com/Rosya-edwica/api.edwica/internal/database"
	"github.com/Rosya-edwica/api.edwica/internal/server"
	"github.com/ilyakaznacheev/cleanenv"
)

func main() {
	dbConfig := LoadDBConfig()
	httpConfig := LoadHTTPConfig()
	fmt.Println(dbConfig)
	fmt.Println(httpConfig)
	_, err := database.New(dbConfig)
	if err != nil {
		log.Fatal(err)
	}

	server := server.NewServer(httpConfig)
	server.Run()
}

func LoadDBConfig() *database.Config {
	var cfg database.Config
	err := cleanenv.ReadConfig("config/configdb.yaml", &cfg)
	if err != nil {
		log.Fatal(err)
	}
	return &cfg
}

func LoadHTTPConfig() *server.Config {
	var cfg server.Config
	err := cleanenv.ReadConfig("config/confighttp.yaml", &cfg)
	if err != nil {
		log.Fatal(err)
	}
	return &cfg
}
