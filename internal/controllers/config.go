package controllers

import (
	"github.com/gin-gonic/gin"
)

func IncorrentParameters(c *gin.Context, param string) map[string]string {
	response := make(map[string]string)
	if param == "text" {
		response["error"] = "missed required param: text"
		response["message"] = "Вы забыли указать обязательный параметр: text. Благодаря нему ведется поиск необходимой информации"
	}
	return response
}

func ErrorListHandler(c *gin.Context, errorsList []error) map[string][]map[string]string {
	response := make(map[string][]map[string]string, len(errorsList))
	for _, err := range errorsList {
		if err == nil {
			continue
		}
		errMap := ErrorHandler(c, err)
		response["errors"] = append(response["errors"], errMap)
	}
	if len(response["errors"]) == 0 {
		return nil
	}
	return response
}

func ErrorHandler(c *gin.Context, err error) map[string]string {
	if err == nil {
		return nil
	}
	errMap := make(map[string]string)
	errMap["error"] = err.Error()
	errMap["message"] = "описание ошибки"
	return errMap
}
