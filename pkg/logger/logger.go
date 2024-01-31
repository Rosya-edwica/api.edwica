package logger

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

type Logger struct {
	Log *log.Logger
}

func (l *Logger) Info(msg string) {
	l.Log.SetPrefix("\u001b[34mINFO: \u001B[0m")
	l.Log.Println(msg)
}

func (l *Logger) Warning(msg string) {
	l.Log.SetPrefix("\u001b[33mWARNING: \u001B[0m")
	l.Log.Println(msg)
}

func (l *Logger) Error(msg string) {
	l.Log.SetPrefix("\u001b[31mERROR: \u001b[0m")
	l.Log.Println(msg)
}

func (l *Logger) Fatal(msg string) {
	l.Log.SetPrefix("\u001b[31mERROR: \u001b[0m")
	l.Log.Fatal(msg)
}

var (
	Log      Logger
	folder   = "logs"
	filename = "info"
)

func init() {
	startTime := time.Now().Unix()
	createFolder()
	logFile, err := os.Create(fmt.Sprintf("%s/%s_%d.log", folder, filename, startTime))
	if err != nil {
		panic(err)
	}
	// Записываем в лог файл и консоль инфу о запросах (время, адрес, статус)
	gin.DisableConsoleColor()
	gin.DefaultWriter = io.MultiWriter(os.Stdout, logFile)

	// Записываем логи приложения в файл и консоль
	Log.Log = log.New(logFile, "", log.LstdFlags|log.Lshortfile)
	mw := io.MultiWriter(os.Stdout, logFile)
	Log.Log.SetOutput(mw)
}

func createFolder() {
	if _, err := os.Stat(folder); errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir(folder, os.ModePerm)
		if err != nil {
			log.Println(err)
		}
	}

}
