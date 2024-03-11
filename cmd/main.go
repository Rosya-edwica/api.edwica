package main

import (
	"log"

	"github.com/Rosya-edwica/api.edwica/config"
	"github.com/Rosya-edwica/api.edwica/internal/database"
	"github.com/Rosya-edwica/api.edwica/internal/server"
)

// FIXME: Если не получится по какой-то причине обновить токен superjob, то мы уйдем в рекурсию
// FIXME: Сделать правильную обработку ошибок, а то мы их только выводим в консоль
// FIXME: Сообщения об ошибках в телеграм
// FIXME: Правильная выдача ошибок пользователям
// FIXME: Оптимизиация хранения ненайденных книг и видео

func main() {
	dbConfig, err := config.LoadDBConfig("")
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

	s := server.NewServer(httpConfig)
	s.Run()
}
